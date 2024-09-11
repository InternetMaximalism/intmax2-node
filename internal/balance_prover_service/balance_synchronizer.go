package balance_prover_service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"time"
)

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
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProcessor BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	intMaxPrivateKey *intMaxAcc.PrivateKey,
) error {
	mockWallet, err := NewMockWallet(intMaxPrivateKey)
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
			userAllData, err := NewBalanceTransitionData(s.ctx, s.cfg, intMaxPrivateKey)
			if err != nil {
				const msg = "failed to start Balance Prover Service: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			sortedValidUserData, err := userAllData.SortValidUserData(
				s.log,
				blockValidityProver,
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

			err = blockValidityProver.SyncBlockProverWithBlockNumber(blockNumber)
			if err != nil {
				if err.Error() == "block content by block number error" {
					time.Sleep(1 * time.Second)
					// return errors.New("block content by block number error")
					continue
				}

				const msg = "failed to sync block prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			// err = balanceProverService.SyncBalanceProver.SyncNoSend(
			// 	syncValidityProver,
			// 	mockWallet,
			// 	balanceProverService.BalanceProcessor,
			// )
			// if err != nil {
			// 	const msg = "failed to sync balance prover: %+v"
			// 	s.log.Fatalf(msg, err.Error())
			// }

			for _, transition := range sortedValidUserData {
				fmt.Printf("valid transition: %v\n", transition)

				switch transition := transition.(type) {
				case ValidSentTx:
					fmt.Printf("valid sent transaction: %v\n", transition.TxHash)
					err := applySentTransactionTransition(
						s.log,
						transition.Tx,
						blockValidityProver,
						balanceProcessor,
						syncBalanceProver,
						mockWallet,
					)

					if err != nil {
						const msg = "failed to send transaction: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case ValidReceivedDeposit:
					fmt.Printf("valid received deposit: %v\n", transition.DepositHash)
					transitionBlockNumber := transition.BlockNumber()
					fmt.Printf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						s.log,
						blockValidityProver,
						mockWallet,
						balanceProcessor,
					)
					if err != nil {
						const msg = "failed to sync balance prover: %+v"
						s.log.Fatalf(msg, err.Error())
					}

					err := applyReceivedDepositTransition(
						transition.Deposit,
						blockValidityProver,
						balanceProcessor,
						syncBalanceProver,
						mockWallet,
					)
					if err != nil {
						const msg = "failed to receive deposit: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case ValidReceivedTransfer:
					fmt.Printf("valid received transfer: %v\n", transition.TransferHash)
					transitionBlockNumber := transition.BlockNumber()
					fmt.Printf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						s.log,
						blockValidityProver,
						mockWallet,
						balanceProcessor,
					)
					if err != nil {
						const msg = "failed to sync balance prover: %+v"
						s.log.Fatalf(msg, err.Error())
					}

					err := applyReceivedTransferTransition(
						transition.Transfer,
						blockValidityProver,
						balanceProcessor,
						syncBalanceProver,
						mockWallet,
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
	deposit *DepositDetails,
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProcessor BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	mockWallet *MockWallet,
) error {
	if deposit == nil {
		return errors.New("deposit is not found")
	}

	fmt.Printf("deposit ID: %d\n", deposit.DepositID)
	_, depositIndex, err := blockValidityProver.GetDepositLeafAndIndexByHash(deposit.DepositHash)
	if err != nil {
		const msg = "failed to get Deposit Leaf and Index by Hash: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if depositIndex == nil {
		const msg = "failed to get Deposit Index by Hash: %+v"
		return fmt.Errorf(msg, "depositIndex is nil")
	}

	IsSynchronizedDepositIndex, err := blockValidityProver.IsSynchronizedDepositIndex(*depositIndex)
	if err != nil {
		const msg = "failed to check IsSynchronizedDepositIndex: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	if !IsSynchronizedDepositIndex {
		const msg = "deposit index %d is not synchronized"
		return fmt.Errorf(msg, *depositIndex)
	}

	fmt.Printf("deposit index: %d\n", *depositIndex)

	depositCase := DepositCase{
		Deposit: intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		},
		DepositIndex: *depositIndex,
		DepositID:    deposit.DepositID,
		DepositSalt:  *deposit.Salt,
	}
	mockWallet.AddDepositCase(*depositIndex, &depositCase)
	fmt.Printf("start to prove deposit\n")
	err = syncBalanceProver.ReceiveDeposit(
		mockWallet,
		balanceProcessor,
		blockValidityProver,
		*depositIndex,
	)
	fmt.Printf("prove deposit %v\n", err.Error())
	if err != nil {
		if err.Error() == "balance processor not initialized" {
			return errors.New("balance processor not initialized")
		}

		return fmt.Errorf("failed to receive deposit: %+v", err.Error())
	}

	return nil
}

func applyReceivedTransferTransition(
	transfer *intMaxTypes.TransferDetailsWithProofBody,
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProcessor BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	mockWallet *MockWallet,
) error {
	fmt.Printf("transfer hash: %d\n", transfer.TransferDetails.TransferWitness.Transfer.Hash())

	senderBalanceProofBody, err := base64.StdEncoding.DecodeString(transfer.SenderBalanceProofBody)
	if err != nil {
		var ErrDecodeSenderBalanceProofBody = errors.New("failed to decode sender balance proof body")
		return errors.Join(ErrDecodeSenderBalanceProofBody, err)
	}

	reader := bufio.NewReader(bytes.NewReader(transfer.TransferDetails.SenderBalancePublicInputs))
	senderBalancePublicInputs, err := intMaxTypes.DecodePublicInputs(reader, uint32(len(transfer.TransferDetails.SenderBalancePublicInputs))/int8Key)
	if err != nil {
		var ErrDecodeSenderBalancePublicInputs = errors.New("failed to decode sender balance public inputs")
		return errors.Join(ErrDecodeSenderBalancePublicInputs, err)
	}

	senderBalanceProof := intMaxTypes.Plonky2Proof{
		Proof:        senderBalanceProofBody,
		PublicInputs: senderBalancePublicInputs,
	}

	encodedSenderBalanceProof, err := senderBalanceProof.Base64String()
	if err != nil {
		var ErrEncodeSenderBalanceProof = errors.New("failed to encode sender balance proof")
		return errors.Join(ErrEncodeSenderBalanceProof, err)
	}

	syncBalanceProver.ReceiveTransfer(
		mockWallet,
		balanceProcessor,
		blockValidityProver,
		transfer.TransferDetails.TransferWitness,
		encodedSenderBalanceProof,
	)

	return nil
}

func applySentTransactionTransition(
	log logger.Logger,
	tx *intMaxTypes.TxDetails,
	blockValidityProver block_validity_prover.BlockValidityProver,
	// blockSynchronizer block_validity_prover.BlockSynchronizer,
	balanceProcessor BalanceProcessor,
	syncBalanceProver *SyncBalanceProver,
	mockWallet *MockWallet,
	// syncValidityProver *syncValidityProver, // XXX
) error {
	// syncValidityProver.Sync() // sync validity proofs
	fmt.Printf("transaction hash: %d\n", tx.Hash())

	// txMerkleProof := tx.TxMerkleProof
	transfers := tx.Transfers

	// balanceProverService.SyncBalanceProver
	txWitness, transferWitnesses, err := mockWallet.SendTx(blockValidityProver, transfers)
	if err != nil {
		const msg = "failed to send transaction: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	newSalt, err := new(Salt).SetRandom()
	if err != nil {
		const msg = "failed to set random: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	_, err = mockWallet.UpdateOnSendTx(*newSalt, txWitness, transferWitnesses)
	if err != nil {
		const msg = "failed to update on SendTx: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	err = syncBalanceProver.SyncSend(
		log,
		blockValidityProver,
		mockWallet,
		balanceProcessor,
	)
	if err != nil {
		const msg = "failed to sync transaction: %+v"
		panic(fmt.Sprintf(msg, err.Error()))
	}

	return nil
}
