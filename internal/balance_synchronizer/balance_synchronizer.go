package balance_synchronizer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"time"
)

const (
	int8Key = 8

	messageBalanceProcessorNotInitialized = "balance processor not initialized"
)

type balanceSynchronizer struct {
	ctx                  context.Context
	cfg                  *configs.Config
	log                  logger.Logger
	sb                   block_validity_prover.ServiceBlockchain
	blockSynchronizer    block_validity_prover.BlockSynchronizer
	blockValidityService block_validity_prover.BlockValidityService
	balanceProcessor     balance_prover_service.BalanceProcessor
	syncBalanceProver    *SyncBalanceProver
	userState            UserState
}

func NewSynchronizer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	blockValidityService block_validity_prover.BlockValidityService,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	userState UserState,
) *balanceSynchronizer {
	return &balanceSynchronizer{
		ctx:                  ctx,
		cfg:                  cfg,
		log:                  log,
		sb:                   sb,
		blockSynchronizer:    blockSynchronizer,
		blockValidityService: blockValidityService,
		balanceProcessor:     balanceProcessor,
		syncBalanceProver:    syncBalanceProver,
		userState:            userState,
	}
}

func (s *balanceSynchronizer) CurrentNonce() uint32 {
	return s.userState.Nonce()
}

func (s *balanceSynchronizer) LastBalanceProof() *intMaxTypes.Plonky2Proof {
	return &intMaxTypes.Plonky2Proof{
		Proof:        s.syncBalanceProver.lastBalanceProofBody,
		PublicInputs: s.syncBalanceProver.balanceData.BalanceProofPublicInputs,
	}
}

func (s *balanceSynchronizer) ProveSendTransition(
	sendWitness *balance_prover_service.SendWitness,
	updateWitness *block_validity_prover.UpdateWitness,
) (*balance_prover_service.SenderProofWithPublicInputs, error) {
	publicKey := s.userState.PublicKey()
	lastBalanceProof := s.LastBalanceProof().ProofBase64String()
	return s.balanceProcessor.ProveSendTransition(
		publicKey,
		sendWitness,
		updateWitness,
		&lastBalanceProof,
	)
}

func (s *balanceSynchronizer) Sync(
	intMaxPrivateKey *intMaxAcc.PrivateKey,
) error {
	timeout := 1 * time.Second
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-s.ctx.Done():
			ticker.Stop()
			s.log.Warnf("Received cancel signal from context, stopping...")
			return nil
		case <-ticker.C:
			err := s.syncProcessing(intMaxPrivateKey)
			if err != nil {
				if errors.Is(err, ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber) ||
					errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
					continue
				}

				const msg = "failed to start sync processing: %+v"
				s.log.Fatalf(msg, err.Error())
			}
		}
	}
}

func (s *balanceSynchronizer) syncProcessing(intMaxPrivateKey *intMaxAcc.PrivateKey) (err error) {
	balanceTransitionData, err := balance_prover_service.NewBalanceTransitionData(s.ctx, s.cfg, s.log, intMaxPrivateKey)
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		s.log.Fatalf(msg, err.Error())
	}
	sortedValidUserData, err := balanceTransitionData.SortValidUserData(
		s.log,
		s.blockValidityService,
	)
	if err != nil {
		const msg = "failed to sort valid user data: %+v"
		s.log.Fatalf(msg, err.Error())
	}
	fmt.Printf("size of sortedValidUserData: %v\n", len(sortedValidUserData))
	for _, transition := range sortedValidUserData {
		fmt.Printf("transition block number: %d\n", transition.BlockNumber())
	}

	validityProverInfo, err := s.blockValidityService.FetchValidityProverInfo()
	if err != nil {
		const msg = "failed to fetch validity prover info: %+v"
		s.log.Fatalf(msg, err.Error())
		return errors.Join(ErrLatestSynchronizedBlockNumberFail, err)
	}

	latestSynchronizedBlockNumber := validityProverInfo.BlockNumber
	if latestSynchronizedBlockNumber <= s.syncBalanceProver.LastUpdatedBlockNumber() {
		return ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber
	}

	for _, transition := range sortedValidUserData {
		fmt.Printf("wallet private state commitment (before): %s\n", s.userState.PrivateState().Commitment().String())
		fmt.Printf("valid transition: %v\n", transition)

		switch transition := transition.(type) {
		case balance_prover_service.ValidSentTx:
			err = s.validSentTx(&transition)
			if err != nil {
				const msg = "failed to send transaction: %+v"
				s.log.Warnf(msg, err.Error())
				continue
			}
		case balance_prover_service.ValidReceivedDeposit:
			err = s.validReceivedDeposit(&transition)
			if err != nil {
				if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
					errors.Is(err, ErrApplyReceivedDepositTransitionFail) {
					const msg = "failed to receive deposit: %+v"
					s.log.Warnf(msg, err.Error())
					continue
				} else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
					return err
				}

				const msg = "failed to sync balance prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}
		case balance_prover_service.ValidReceivedTransfer:
			err = s.validReceivedTransfer(&transition)
			if err != nil {
				if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
					errors.Is(err, ErrApplyReceivedTransferTransitionFail) {
					const msg = "failed to receive deposit: %+v"
					s.log.Warnf(msg, err.Error())
					continue
				} else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
					return err
				}

				const msg = "failed to sync balance prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}
		default:
			fmt.Printf("unknown transition: %v\n", transition)
		}
	}

	return nil
}

func (s *balanceSynchronizer) validSentTx(transition *balance_prover_service.ValidSentTx) error {
	fmt.Printf("valid sent transaction: %v\n", transition.TxHash)
	err := applySentTransactionTransition(
		s.log,
		transition.Tx,
		s.blockValidityService,
		s.blockSynchronizer,
		s.balanceProcessor,
		s.syncBalanceProver,
		s.userState,
	)

	if err != nil {
		return errors.Join(ErrValidSentTxFail, err)
	}

	return nil
}

func (s *balanceSynchronizer) validReceivedDeposit(
	transition *balance_prover_service.ValidReceivedDeposit,
) (err error) {
	fmt.Printf("valid received deposit: %v\n", transition.DepositHash)
	transitionBlockNumber := transition.BlockNumber()
	fmt.Printf("transitionBlockNumber: %d", transitionBlockNumber)
	err = s.syncBalanceProver.SyncNoSend(
		s.log,
		s.blockValidityService,
		s.blockSynchronizer,
		s.userState,
		s.balanceProcessor,
		transitionBlockNumber,
	)
	if err != nil {
		return errors.Join(ErrValidReceivedDepositFail, err)
	}

	err = applyReceivedDepositTransition(
		transition.Deposit,
		s.blockValidityService,
		s.balanceProcessor,
		s.syncBalanceProver,
		s.userState,
	)
	if err != nil {
		return errors.Join(ErrApplyReceivedDepositTransitionFail, err)
	}

	return nil
}

func (s *balanceSynchronizer) validReceivedTransfer(
	transition *balance_prover_service.ValidReceivedTransfer,
) (err error) {
	fmt.Printf("valid received transfer: %v\n", transition.TransferHash)
	transitionBlockNumber := transition.BlockNumber()
	fmt.Printf("transitionBlockNumber: %d\n", transitionBlockNumber)
	err = s.syncBalanceProver.SyncNoSend(
		s.log,
		s.blockValidityService,
		s.blockSynchronizer,
		s.userState,
		s.balanceProcessor,
		transitionBlockNumber,
	)
	if err != nil {
		const msg = "failed to sync balance prover: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	err = applyReceivedTransferTransition(
		transition.Transfer,
		s.blockValidityService,
		s.balanceProcessor,
		s.syncBalanceProver,
		s.userState,
	)
	if err != nil {
		return errors.Join(ErrApplyReceivedTransferTransitionFail, err)
	}

	return nil
}

func applyReceivedDepositTransition(
	deposit *balance_prover_service.DepositDetails,
	blockValidityService block_validity_prover.BlockValidityService,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	userState UserState,
) error {
	if deposit == nil {
		return errors.New("deposit is not found")
	}

	fmt.Printf("applyReceivedDepositTransition deposit ID: %d\n", deposit.DepositID)
	depositInfo, err := blockValidityService.GetDepositInfoByHash(deposit.DepositHash)
	if err != nil {
		const msg = "failed to get Deposit Leaf and Index by Hash: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if depositInfo.DepositIndex == nil {
		const msg = "failed to get Deposit Index by Hash: %+v"
		return fmt.Errorf(msg, "depositIndex is nil")
	}

	depositIndex := *depositInfo.DepositIndex
	if !depositInfo.IsSynchronized {
		const msg = "deposit index %d is not synchronized"
		return fmt.Errorf(msg, depositIndex)
	}

	fmt.Printf("deposit index: %d\n", depositIndex)

	depositCase := balance_prover_service.DepositCase{
		Deposit: intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		},
		DepositIndex: depositIndex,
		DepositID:    deposit.DepositID,
		DepositSalt:  *deposit.Salt,
	}
	err = userState.AddDepositCase(depositIndex, &depositCase)
	if err != nil {
		const msg = "failed to add deposit case: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	fmt.Printf("start to prove deposit\n")
	err = syncBalanceProver.ReceiveDeposit(
		userState,
		balanceProcessor,
		blockValidityService,
		depositIndex,
	)
	if err != nil {
		fmt.Printf("prove deposit %v\n", err.Error())
		if err.Error() == messageBalanceProcessorNotInitialized {
			return errors.New(messageBalanceProcessorNotInitialized)
		}

		return fmt.Errorf("failed to receive deposit: %+v", err.Error())
	}

	return nil
}

func applyReceivedTransferTransition(
	transfer *intMaxTypes.TransferDetailsWithProofBody,
	blockValidityService block_validity_prover.BlockValidityService,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	userState UserState,
) error {
	fmt.Printf("transfer hash: %d\n", transfer.TransferDetails.TransferWitness.Transfer.Hash())

	senderLastBalanceProofBody, err := base64.StdEncoding.DecodeString(transfer.SenderLastBalanceProofBody)
	if err != nil {
		var ErrDecodeSenderBalanceProofBody = errors.New("failed to decode sender balance proof body")
		return errors.Join(ErrDecodeSenderBalanceProofBody, err)
	}

	reader := bufio.NewReader(bytes.NewReader(transfer.TransferDetails.SenderLastBalancePublicInputs))
	senderLastBalancePublicInputs, err := intMaxTypes.DecodePublicInputs(reader, uint32(len(transfer.TransferDetails.SenderLastBalancePublicInputs))/int8Key)
	if err != nil {
		var ErrDecodeSenderBalancePublicInputs = errors.New("failed to decode sender balance public inputs")
		return errors.Join(ErrDecodeSenderBalancePublicInputs, err)
	}

	senderLastBalanceProof := intMaxTypes.Plonky2Proof{
		Proof:        senderLastBalanceProofBody,
		PublicInputs: senderLastBalancePublicInputs,
	}

	encodedSenderLastBalanceProof, err := senderLastBalanceProof.Base64String()
	if err != nil {
		var ErrEncodeSenderBalanceProof = errors.New("failed to encode sender balance proof")
		return errors.Join(ErrEncodeSenderBalanceProof, err)
	}

	senderBalanceTransitionProof := intMaxTypes.Plonky2Proof{
		Proof:        senderLastBalanceProofBody,
		PublicInputs: senderLastBalancePublicInputs,
	}

	encodedSenderBalanceTransitionProof, err := senderBalanceTransitionProof.Base64String()
	if err != nil {
		var ErrEncodeSenderBalanceTransitionProof = errors.New("failed to encode sender balance transition proof")
		return errors.Join(ErrEncodeSenderBalanceTransitionProof, err)
	}

	err = syncBalanceProver.ReceiveTransfer(
		userState,
		balanceProcessor,
		blockValidityService,
		transfer.TransferDetails.TransferWitness,
		encodedSenderLastBalanceProof,
		encodedSenderBalanceTransitionProof,
	)
	if err != nil {
		fmt.Printf("prove received transfer %v\n", err.Error())
		if err.Error() == messageBalanceProcessorNotInitialized {
			return errors.New(messageBalanceProcessorNotInitialized)
		}

		return fmt.Errorf("failed to receive deposit: %+v", err.Error())
	}

	return nil
}

func applySentTransactionTransition(
	log logger.Logger,
	tx *intMaxTypes.TxDetails,
	blockValidityService block_validity_prover.BlockValidityService,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	userState UserState,
) error {
	// syncValidityProver.Sync() // sync validity proofs
	fmt.Printf("transaction hash: %d\n", tx.Hash())

	// txMerkleProof := tx.TxMerkleProof
	// transfers := tx.Transfers

	// balanceProverService.SyncBalanceProver
	txWitness, transferWitnesses, err := MakeTxWitness(blockValidityService, tx)
	if err != nil {
		const msg = "failed to send transaction: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	newSalt, err := new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		const msg = "failed to set random: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	_, err = userState.UpdateOnSendTx(*newSalt, txWitness, transferWitnesses)
	if err != nil {
		const msg = "failed to update on SendTx: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	err = syncBalanceProver.SyncSend(
		log,
		blockValidityService,
		blockSynchronizer,
		userState,
		balanceProcessor,
	)
	if err != nil {
		fmt.Printf("prove sent transaction %v\n", err.Error())
		if err.Error() == messageBalanceProcessorNotInitialized {
			return errors.New(messageBalanceProcessorNotInitialized)
		}

		return fmt.Errorf("failed to sync transaction:: %+v", err.Error())
	}

	return nil
}
