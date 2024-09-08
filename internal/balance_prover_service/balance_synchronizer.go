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
	"intmax2-node/internal/mnemonic_wallet/models"
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

func (s *balanceSynchronizer) Sync(syncValidityProver *syncValidityProver, blockBuilderWallet *models.Wallet) error {
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
			balanceProverService := NewBalanceProverService(s.ctx, s.cfg, s.log, blockBuilderWallet)
			userAllData, err := balanceProverService.DecodedUserData()
			if err != nil {
				const msg = "failed to start Balance Prover Service: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			sortedValidUserData, err := userAllData.SortValidUserData(syncValidityProver)
			if err != nil {
				const msg = "failed to sort valid user data: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			intMaxPrivateKey, err := intMaxAcc.NewPrivateKeyFromString(blockBuilderWallet.IntMaxPrivateKey)
			if err != nil {
				const msg = "failed to get IntMax Private Key: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			mockWallet, err := NewMockWallet(intMaxPrivateKey)
			if err != nil {
				const msg = "failed to get Mock Wallet: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			syncValidityProver, err := NewSyncValidityProver(
				s.ctx, s.cfg, s.log, s.sb, s.db,
			)
			if err != nil {
				const msg = "failed to get Sync Validity Prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			result, err := block_validity_prover.BlockAuxInfo(syncValidityProver.ValidityProver.BlockBuilder(), blockNumber)
			if err != nil {
				if err.Error() == "block content by block number error" {
					time.Sleep(1 * time.Second)
					// return errors.New("block content by block number error")
					continue
				}

				const msg = "failed to fetch new posted blocks: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			err = syncValidityProver.ValidityProver.SyncBlockProverWithAuxInfo(result.BlockContent, result.PostedBlock)
			if err != nil {
				const msg = "failed to sync block prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			err = balanceProverService.SyncBalanceProver.SyncNoSend(
				syncValidityProver,
				mockWallet,
				balanceProverService.BalanceProcessor,
			)
			if err != nil {
				const msg = "failed to sync balance prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			blockValidityProver := syncValidityProver.ValidityProver

			for _, transition := range sortedValidUserData {
				fmt.Printf("valid transition: %v\n", transition)

				switch transition := transition.(type) {
				case ValidSentTx:
					fmt.Printf("valid sent transaction: %v\n", transition.TxHash)
					err := s.applySentTransactionTransition(
						transition.Tx,
						blockValidityProver,
						balanceProverService,
						mockWallet,
						syncValidityProver,
					)

					if err != nil {
						const msg = "failed to send transaction: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case ValidReceivedDeposit:
					fmt.Printf("valid received deposit: %v\n", transition.DepositHash)
					err := s.applyReceivedDepositTransition(
						transition.Deposit,
						blockValidityProver,
						balanceProverService,
						mockWallet,
					)
					if err != nil {
						const msg = "failed to receive deposit: %+v"
						s.log.Warnf(msg, err.Error())
						continue
					}
				case ValidReceivedTransfer:
					fmt.Printf("valid received transfer: %v\n", transition.TransferHash)
					err := s.applyReceivedTransferTransition(
						transition.Transfer,
						blockValidityProver,
						balanceProverService,
						mockWallet,
						balanceProverService.SyncBalanceProver,
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

func (s balanceSynchronizer) applyReceivedDepositTransition(
	deposit *DepositDetails,
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProverService *balanceProverService,
	mockWallet *MockWallet,
) error {
	fmt.Printf("deposit ID: %d\n", deposit.DepositID)
	fmt.Printf("deposit hash: %d\n", deposit.DepositHash)
	_, depositIndex, err := blockValidityProver.BlockBuilder().GetDepositLeafAndIndexByHash(deposit.DepositHash)
	if err != nil {
		const msg = "failed to get Deposit Index by Hash: %+v"
		s.log.Warnf(msg, err.Error())
		return nil
	}
	if depositIndex == nil {
		const msg = "failed to get Deposit Index by Hash: %+v"
		s.log.Warnf(msg, "depositIndex is nil")
		return nil
	}

	IsSynchronizedDepositIndex, err := blockValidityProver.BlockBuilder().IsSynchronizedDepositIndex(*depositIndex)
	if err != nil {
		const msg = "failed to check IsSynchronizedDepositIndex: %+v"
		s.log.Warnf(msg, err.Error())
		return nil
	}
	if !IsSynchronizedDepositIndex {
		const msg = "deposit index %d is not synchronized"
		s.log.Warnf(msg, *depositIndex)
		return nil
	}

	fmt.Printf("deposit index: %d\n", *depositIndex)

	depositCase := DepositCase{
		Deposit: intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		},
		DepositIndex: *depositIndex,
		DepositSalt:  *deposit.Salt,
	}
	mockWallet.AddDepositCase(deposit.DepositID, &depositCase)
	err = balanceProverService.SyncBalanceProver.ReceiveDeposit(
		mockWallet,
		balanceProverService.BalanceProcessor,
		blockValidityProver.BlockBuilder(),
		deposit.DepositID,
	)
	if err != nil {
		const msg = "failed to receive deposit: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	return nil
}

func (s balanceSynchronizer) applyReceivedTransferTransition(
	transfer *intMaxTypes.TransferDetailsWithProofBody,
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProverService *BalanceProverService,
	mockWallet *MockWallet,
	syncBalanceProver *SyncBalanceProver,
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
		balanceProverService.BalanceProcessor,
		blockValidityProver.BlockBuilder(),
		transfer.TransferDetails.TransferWitness,
		encodedSenderBalanceProof,
	)

	return nil
}

func (s *balanceSynchronizer) applySentTransactionTransition(
	tx *intMaxTypes.TxDetails,
	blockValidityProver block_validity_prover.BlockValidityProver,
	balanceProverService *balanceProverService,
	mockWallet *MockWallet,
	syncValidityProver *syncValidityProver,
) error {
	// syncValidityProver.Sync() // sync validity proofs
	fmt.Printf("transaction hash: %d\n", tx.Hash())

	// txMerkleProof := tx.TxMerkleProof
	transfers := tx.Transfers

	// balanceProverService.SyncBalanceProver
	txWitness, transferWitnesses, err := mockWallet.SendTx(blockValidityProver.BlockBuilder(), transfers)
	if err != nil {
		const msg = "failed to send transaction: %+v"
		s.log.Warnf(msg, err.Error())
		return nil
	}
	newSalt, err := new(Salt).SetRandom()
	if err != nil {
		const msg = "failed to set random: %+v"
		s.log.Warnf(msg, err.Error())
		return nil
	}

	_, err = mockWallet.UpdateOnSendTx(*newSalt, txWitness, transferWitnesses)
	if err != nil {
		const msg = "failed to update on SendTx: %+v"
		s.log.Warnf(msg, err.Error())
		return nil
	}

	err = balanceProverService.SyncBalanceProver.SyncSend(
		syncValidityProver,
		mockWallet,
		balanceProverService.BalanceProcessor,
	)
	if err != nil {
		const msg = "failed to sync transaction: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	return nil
}
