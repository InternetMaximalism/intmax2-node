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

const int8Key = 8

type balanceSynchronizer struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
	sb  block_validity_prover.ServiceBlockchain
	db  block_validity_prover.SQLDriverApp
}

func NewSynchronizer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	db block_validity_prover.SQLDriverApp,
) *balanceSynchronizer {
	return &balanceSynchronizer{
		ctx: ctx,
		cfg: cfg,
		log: log,
		sb:  sb,
		db:  db,
	}
}

func (s *balanceSynchronizer) Sync(
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	blockValidityService block_validity_prover.BlockValidityService,
	balanceProcessor balance_prover_service.BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	intMaxPrivateKey *intMaxAcc.PrivateKey,
) error {
	userWalletState, err := NewMockWallet(intMaxPrivateKey)
	if err != nil {
		const msg = "failed to get Mock Wallet: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	timeout := 1 * time.Second
	ticker := time.NewTicker(timeout)
	blockNumber := uint32(1)
	for {
		select {
		case <-s.ctx.Done():
			ticker.Stop()
			s.log.Warnf("Received cancel signal from context, stopping...")
			return nil
		case <-ticker.C:
			userAllData, err := balance_prover_service.NewBalanceTransitionData(s.ctx, s.cfg, intMaxPrivateKey)
			if err != nil {
				const msg = "failed to start Balance Prover Service: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			sortedValidUserData, err := userAllData.SortValidUserData(
				s.log,
				blockValidityService,
				blockSynchronizer,
			)
			if err != nil {
				const msg = "failed to sort valid user data: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			fmt.Printf("size of sortedValidUserData: %v\n", len(sortedValidUserData))
			for _, transition := range sortedValidUserData {
				fmt.Printf("transition block number: %d\n", transition.BlockNumber())
			}

			latestSynchronizedBlockNumber, err := blockValidityService.LatestSynchronizedBlockNumber()
			if err != nil {
				const msg = "failed to sync block prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			if latestSynchronizedBlockNumber <= blockNumber {
				// return errors.New("block content by block number error")
				time.Sleep(1 * time.Second)
				continue
			}

			// err = balanceProverService.SyncBalanceProver.SyncNoSend(
			// 	syncValidityProver,
			// 	userWalletState,
			// 	balanceProverService.BalanceProcessor,
			// )
			// if err != nil {
			// 	const msg = "failed to sync balance prover: %+v"
			// 	s.log.Fatalf(msg, err.Error())
			// }

			for _, transition := range sortedValidUserData {
				fmt.Printf("valid transition: %v\n", transition)

				switch transition := transition.(type) {
				case balance_prover_service.ValidSentTx:
					fmt.Printf("valid sent transaction: %v\n", transition.TxHash)
					err := applySentTransactionTransition(
						s.log,
						transition.Tx,
						blockValidityService,
						blockSynchronizer,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)

					if err != nil {
						const msg = "failed to send transaction: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case balance_prover_service.ValidReceivedDeposit:
					fmt.Printf("valid received deposit: %v\n", transition.DepositHash)
					transitionBlockNumber := transition.BlockNumber()
					fmt.Printf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						s.log,
						blockValidityService,
						blockSynchronizer,
						userWalletState,
						balanceProcessor,
					)
					if err != nil {
						var ErrNoValidityProofByBlockNumber = errors.New("no validity proof by block number")
						if err.Error() == ErrNoValidityProofByBlockNumber.Error() {
							const msg = "failed to receive deposit: %+v"
							s.log.Warnf(msg, err.Error())
							continue
						}

						const msg = "failed to sync balance prover: %+v"
						s.log.Fatalf(msg, err.Error())
					}

					err := applyReceivedDepositTransition(
						transition.Deposit,
						blockValidityService,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)
					if err != nil {
						const msg = "failed to receive deposit: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case balance_prover_service.ValidReceivedTransfer:
					fmt.Printf("valid received transfer: %v\n", transition.TransferHash)
					transitionBlockNumber := transition.BlockNumber()
					fmt.Printf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						s.log,
						blockValidityService,
						blockSynchronizer,
						userWalletState,
						balanceProcessor,
					)
					if err != nil {
						const msg = "failed to sync balance prover: %+v"
						s.log.Fatalf(msg, err.Error())
					}

					err := applyReceivedTransferTransition(
						transition.Transfer,
						blockValidityService,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)
					if err != nil {
						const msg = "failed to receive transfer: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				default:
					fmt.Printf("unknown transition: %v\n", transition)
				}
			}
		}

		blockNumber++
	}
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
	_, depositIndex, err := blockValidityService.GetDepositLeafAndIndexByHash(deposit.DepositHash)
	if err != nil {
		const msg = "failed to get Deposit Leaf and Index by Hash: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if depositIndex == nil {
		const msg = "failed to get Deposit Index by Hash: %+v"
		return fmt.Errorf(msg, "depositIndex is nil")
	}

	IsSynchronizedDepositIndex, err := blockValidityService.IsSynchronizedDepositIndex(*depositIndex)
	if err != nil {
		const msg = "failed to check IsSynchronizedDepositIndex: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if !IsSynchronizedDepositIndex {
		const msg = "deposit index %d is not synchronized"
		return fmt.Errorf(msg, *depositIndex)
	}

	fmt.Printf("deposit index: %d\n", *depositIndex)

	depositCase := balance_prover_service.DepositCase{
		Deposit: intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		},
		DepositIndex: *depositIndex,
		DepositID:    deposit.DepositID,
		DepositSalt:  *deposit.Salt,
	}
	userState.AddDepositCase(*depositIndex, &depositCase)
	fmt.Printf("start to prove deposit\n")
	err = syncBalanceProver.ReceiveDeposit(
		userState,
		balanceProcessor,
		blockValidityService,
		*depositIndex,
	)
	if err != nil {
		fmt.Printf("prove deposit %v\n", err.Error())
		if err.Error() == "balance processor not initialized" {
			return errors.New("balance processor not initialized")
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

	syncBalanceProver.ReceiveTransfer(
		userState,
		balanceProcessor,
		blockValidityService,
		transfer.TransferDetails.TransferWitness,
		encodedSenderLastBalanceProof,
		encodedSenderBalanceTransitionProof,
	)

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
	// syncValidityProver *syncValidityProver, // XXX
) error {
	// syncValidityProver.Sync() // sync validity proofs
	fmt.Printf("transaction hash: %d\n", tx.Hash())

	// txMerkleProof := tx.TxMerkleProof
	transfers := tx.Transfers

	// balanceProverService.SyncBalanceProver
	txWitness, transferWitnesses, err := userState.SendTx(blockValidityService, transfers)
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
		const msg = "failed to sync transaction: %+v"
		panic(fmt.Sprintf(msg, err.Error()))
	}

	return nil
}
