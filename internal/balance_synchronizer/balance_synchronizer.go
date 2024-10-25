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
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
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
	return s.userState.PrivateState().TransactionCount
}

func (s *balanceSynchronizer) LastBalanceProof() *intMaxTypes.Plonky2Proof {
	return &intMaxTypes.Plonky2Proof{
		Proof:        s.syncBalanceProver.lastBalanceProofBody,
		PublicInputs: s.syncBalanceProver.balanceProofPublicInputs,
	}
}

func (s *balanceSynchronizer) ProveSendTransition(
	spentTokenWitness *balance_prover_service.SpentTokenWitness,
) (string, error) {
	publicKey := s.userState.PublicKey()
	lastBalanceProof := s.LastBalanceProof().ProofBase64String()
	return s.balanceProcessor.ProveSendTransition(
		publicKey,
		spentTokenWitness,
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
	s.log.Infof("start syncProcessing")
	balanceTransitionData, err := balance_prover_service.FetchBalanceTransitionData(s.ctx, s.cfg, s.log, intMaxPrivateKey)
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	lastUpdatedBlockNumber, err := (func() (lastUpdatedBlockNumber uint32, err error) {
		lastBalanceProof := s.syncBalanceProver.LastBalanceProof()
		if lastBalanceProof == nil {
			return 0, nil
		}

		lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*lastBalanceProof)
		if err != nil {
			return 0, errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
		}

		lastBalancePublicInputs, err := new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
		if err != nil {
			return 0, errors.Join(ErrBalancePublicInputsFromPublicInputs, err)
		}

		return lastBalancePublicInputs.PublicState.BlockNumber, nil
	})()
	if err != nil {
		return err
	}

	sortedValidUserData, err := balanceTransitionData.SortValidUserData(
		s.log,
		s.blockValidityService,
		lastUpdatedBlockNumber,
	)
	if err != nil {
		const msg = "failed to sort valid user data: %+v"
		s.log.Fatalf(msg, err.Error())
	}
	fmt.Printf("size of sortedValidUserData: %v\n", len(sortedValidUserData))
	for _, transition := range sortedValidUserData {
		fmt.Printf("[transition] type: %s, block number: %d\n", transition.Type(), transition.BlockNumber())
	}

	storedBalanceData, err := block_synchronizer.GetBackupBalance(s.ctx, s.cfg, s.userState.PublicKey())
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	err = s.syncBalanceProver.SetEncryptedBalanceData(s.userState, storedBalanceData)
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	validityProverInfo, err := s.blockValidityService.FetchValidityProverInfo()
	if err != nil {
		const msg = "failed to fetch validity prover info: %+v"
		s.log.Fatalf(msg, err.Error())
		return errors.Join(ErrLatestSynchronizedBlockNumberFail, err)
	}

	latestSynchronizedBlockNumber := validityProverInfo.BlockNumber
	lastUpdatedBlockNumber = s.userState.PublicState().BlockNumber
	if latestSynchronizedBlockNumber <= lastUpdatedBlockNumber {
		return ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber
	}

	if len(sortedValidUserData) == 0 {
		return ErrNoValidUserData
	}

	for _, transition := range sortedValidUserData {
		fmt.Printf("wallet private state commitment (before): %s\n", s.userState.PrivateState().Commitment().String())
		fmt.Printf("valid transition: %+v\n", transition)
		fmt.Printf("valid transition block number: %v\n", transition.BlockNumber())

		switch transition := transition.(type) {
		case balance_prover_service.ValidSentTx:
			fmt.Printf("valid sent transaction: %s (nonce: %d)\n", transition.TxHash.String(), transition.Tx.Nonce)
			err = s.validSentTx(&transition)
			if err != nil {
				// When synchronizing transactions that one has sent, they must not fail except for specific reasons.
				// This is because, for a balance proof to be valid,
				// it needs to reflect all transfers that have been reflected into valid blocks.
				return fmt.Errorf("failed to send transaction: %w", err)
			}
		case balance_prover_service.ValidReceivedDeposit:
			err = s.validReceivedDeposit(&transition)
			if err != nil {
				if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
					errors.Is(err, ErrApplyReceivedDepositTransitionFail) || errors.Is(err, ErrNullifierAlreadyExists) {
					const msg = "failed to receive deposit: %+v"
					s.log.Warnf(msg, err.Error())
					continue
				}
				// else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
				// 	// continue
				// 	return err
				// }

				const msg = "failed to sync balance prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}
		case balance_prover_service.ValidReceivedTransfer:
			err = s.validReceivedTransfer(&transition)
			if err != nil {
				if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
					errors.Is(err, ErrApplyReceivedTransferTransitionFail) || errors.Is(err, ErrNullifierAlreadyExists) {
					const msg = "failed to receive transfer: %+v"
					s.log.Warnf(msg, err.Error())
					continue
				}
				// else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
				// 	// continue
				// 	return err
				// }

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
	transitionBlockNumber := transition.BlockNumber()
	fmt.Printf("transitionBlockNumber: %d\n", transitionBlockNumber)
	err := applySentTransactionTransition(
		s.log,
		transition.Tx,
		s.blockValidityService,
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
	fmt.Printf("transitionBlockNumber: %d\n", transitionBlockNumber)

	// nullifier already exists
	nullifierBytes := intMaxTypes.Bytes32{}
	nullifierBytes.FromBytes(transition.DepositHash[:])
	isIncluded, err := s.userState.IsIncludedInNullifierTree(nullifierBytes)
	if err != nil {
		const msg = "failed to check nullifier: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if isIncluded {
		fmt.Printf("WARNING: (validReceiveDeposit) nullifier %s already exists\n", transition.DepositHash.String())
		var ErrNullifierAlreadyExists = errors.New("nullifier already exists")
		return errors.Join(ErrNullifierAlreadyExists, err)
	}

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

	// nullifier already exists
	nullifierBytes := intMaxTypes.Bytes32{}
	nullifierBytes.FromBytes(transition.TransferHash.Marshal())
	isIncluded, err := s.userState.IsIncludedInNullifierTree(nullifierBytes)
	if err != nil {
		const msg = "failed to check nullifier: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if isIncluded {
		fmt.Printf("WARNING: nullifier %x already exists\n", transition.TransferHash.Marshal())
		var ErrNullifierAlreadyExists = errors.New("nullifier already exists")
		return errors.Join(ErrNullifierAlreadyExists, err)
	}

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
		return fmt.Errorf("failed to get deposit leaf and index by hash: %w", err)
	}
	if depositInfo.DepositIndex == nil {
		var ErrDepositIndexIsNil = errors.New("deposit index is nil")
		return fmt.Errorf("failed to get deposit Index by hash: %w", ErrDepositIndexIsNil)
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
	fmt.Printf("(applyReceivedDepositTransition.AddDepositCase) depositIndex: %+v\n", depositIndex)
	fmt.Printf("(applyReceivedDepositTransition.AddDepositCase) depositCase: %+v\n", depositCase)
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
		if err.Error() == ErrNullifierAlreadyExists.Error() {
			_ = userState.DeleteDepositCase(depositIndex)
			fmt.Printf("WARNING: (applyReceivedReceiveDepositTransition) nullifier %s already exists\n", depositCase.Deposit.Nullifier().String())
			return nil
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

	senderEnoughBalanceProofResponse, err := block_synchronizer.GetBackupSenderBalanceProofs(
		syncBalanceProver.ctx,
		syncBalanceProver.cfg,
		syncBalanceProver.log,
		[]string{transfer.TransferDetails.SenderEnoughBalanceProofBodyHash},
	)
	if err != nil {
		var ErrGetSenderEnoughBalanceProofBodyFail = errors.New("failed to get sender enough balance proof body")
		return errors.Join(ErrGetSenderEnoughBalanceProofBodyFail, err)
	}

	senderLastBalanceProof, err := recoverSenderLastBalanceProof(senderEnoughBalanceProofResponse.Proofs, transfer.TransferDetails)
	if err != nil {
		return fmt.Errorf("failed to recover sender last balance proof: %w", err)
	}

	encodedSenderLastBalanceProof, err := senderLastBalanceProof.Base64String()
	if err != nil {
		var ErrEncodeSenderBalanceProof = errors.New("failed to encode sender balance proof")
		return errors.Join(ErrEncodeSenderBalanceProof, err)
	}

	senderBalanceTransitionProof, err := recoverSenderBalanceTransitionProof(senderEnoughBalanceProofResponse.Proofs, transfer.TransferDetails)
	if err != nil {
		return fmt.Errorf("failed to recover sender balance transition proof: %w", err)
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

func recoverSenderLastBalanceProof(proofs []block_synchronizer.BackupBalanceProofData, transferDetails *intMaxTypes.TransferDetails) (*intMaxTypes.Plonky2Proof, error) {
	if len(proofs) == 0 {
		var ErrSenderEnoughBalanceProofBodyNotFound = errors.New("sender enough balance proof body not found")
		return nil, ErrSenderEnoughBalanceProofBodyNotFound
	}

	senderEnoughBalanceProofBody := proofs[0]

	senderLastBalanceProofBody, err := base64.StdEncoding.DecodeString(senderEnoughBalanceProofBody.LastBalanceProofBody)
	if err != nil {
		var ErrDecodeSenderBalanceProofBody = errors.New("failed to decode sender balance proof body")
		return nil, errors.Join(ErrDecodeSenderBalanceProofBody, err)
	}

	reader := bufio.NewReader(bytes.NewReader(transferDetails.SenderLastBalancePublicInputs))
	senderLastBalancePublicInputs, err := intMaxTypes.DecodePublicInputs(reader, uint32(len(transferDetails.SenderLastBalancePublicInputs))/int8Key)
	if err != nil {
		var ErrDecodeSenderBalancePublicInputs = errors.New("failed to decode sender balance public inputs")
		return nil, errors.Join(ErrDecodeSenderBalancePublicInputs, err)
	}

	senderLastBalanceProof := intMaxTypes.Plonky2Proof{
		Proof:        senderLastBalanceProofBody,
		PublicInputs: senderLastBalancePublicInputs,
	}

	return &senderLastBalanceProof, nil
}

func recoverSenderBalanceTransitionProof(proofs []block_synchronizer.BackupBalanceProofData, transferDetails *intMaxTypes.TransferDetails) (*intMaxTypes.Plonky2Proof, error) {
	if len(proofs) == 0 {
		var ErrSenderEnoughBalanceProofBodyNotFound = errors.New("sender enough balance proof body not found")
		return nil, ErrSenderEnoughBalanceProofBodyNotFound
	}

	senderBalanceTransitionProofBody, err := base64.StdEncoding.DecodeString(proofs[0].BalanceTransitionProofBody)
	if err != nil {
		var ErrDecodeSenderBalanceProofBody = errors.New("failed to decode sender balance proof body")
		return nil, errors.Join(ErrDecodeSenderBalanceProofBody, err)
	}

	reader := bufio.NewReader(bytes.NewReader(transferDetails.SenderBalanceTransitionPublicInputs))
	senderBalanceTransitionPublicInputs, err := intMaxTypes.DecodePublicInputs(reader, uint32(len(transferDetails.SenderBalanceTransitionPublicInputs))/int8Key)
	if err != nil {
		var ErrDecodeSenderBalancePublicInputs = errors.New("failed to decode sender balance public inputs")
		return nil, errors.Join(ErrDecodeSenderBalancePublicInputs, err)
	}

	senderBalanceTransitionProof := intMaxTypes.Plonky2Proof{
		Proof:        senderBalanceTransitionProofBody,
		PublicInputs: senderBalanceTransitionPublicInputs,
	}

	return &senderBalanceTransitionProof, nil
}

func applySentTransactionTransition(
	log logger.Logger,
	tx *intMaxTypes.TxDetails,
	blockValidityService block_validity_prover.BlockValidityService,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	userState UserState,
) error {
	log.Infof("applySentTransactionTransition: transaction hash: %s", tx.Hash().String())

	txWitness, transferWitnesses, err := MakeTxWitness(blockValidityService, tx)
	if err != nil {
		return fmt.Errorf("failed to make tx witness: %w", err)
	}

	newSalt, err := new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		return fmt.Errorf("failed to set random salt: %w", err)
	}

	// Update user state, including salt, nonce, nullifier, and balance
	sendWitness, err := userState.UpdateOnSendTx(*newSalt, txWitness, transferWitnesses)
	if err != nil {
		return fmt.Errorf("failed to update on send tx: %w", err)
	}

	// For debugging, print transfers in sendWitness
	for _, transfer := range sendWitness.SpentTokenWitness.Transfers {
		if transfer.Amount.Cmp(big.NewInt(0)) != 0 {
			log.Debugf("(sendWitness) transfer: %+v\n", transfer)
		}
	}

	newBalancePisBlockNumber := sendWitness.GetIncludedBlockNumber()
	prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
	log.Debugf("(sendWitness): Transition from block %d to block %d", prevBalancePisBlockNumber, newBalancePisBlockNumber)
	updateWitness, err := blockValidityService.FetchUpdateWitness(
		userState.PublicKey(),
		newBalancePisBlockNumber,
		prevBalancePisBlockNumber,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch update witness: %w", err)
	}

	_validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(updateWitness.ValidityProof)
	if err != nil {
		return fmt.Errorf("failed to create validity proof with pis: %w", err)
	}
	updateWitnessValidityPis := new(block_validity_prover.ValidityPublicInputs).FromPublicInputs(_validityProofWithPis.PublicInputs)

	sendWitnessValidityPis := sendWitness.TxWitness.ValidityPis
	if !updateWitnessValidityPis.Equal(&sendWitnessValidityPis) {
		log.Errorf("update witness %+v is not equal to send witness %+v", updateWitnessValidityPis, sendWitnessValidityPis)
		return errors.New("update witness validity proof is not equal to send witness validity proof")
	}

	if updateWitnessValidityPis.IsValidBlock {
		log.Debugf("Block %d is valid", updateWitnessValidityPis.PublicState.BlockNumber)
	} else {
		log.Debugf("Block %d is invalid", updateWitnessValidityPis.PublicState.BlockNumber)
	}

	balanceProof, err := balanceProcessor.ProveSend(
		userState.PublicKey(),
		sendWitness,
		updateWitness,
		syncBalanceProver.LastBalanceProof(),
	)
	if err != nil {
		return fmt.Errorf("failed to prove send: %w", err)
	}

	// wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)

	err = syncBalanceProver.UploadLastBalanceProof(newBalancePisBlockNumber, balanceProof.Proof, userState)
	if err != nil {
		return fmt.Errorf("failed to upload last balance proof in SyncSend: %w", err)
	}
	return nil
}
