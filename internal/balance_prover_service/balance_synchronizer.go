package balance_prover_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet/models"
	intMaxTree "intmax2-node/internal/tree"
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

func (s *balanceSynchronizer) Sync(blockValidityProver block_validity_prover.BlockValidityProver, blockBuilderWallet *models.Wallet) error {
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
			fmt.Printf("deposits in userAllData: %+v\n", len(userAllData.Deposits))

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

			result, err := block_validity_prover.BlockAuxInfo(blockValidityProver.BlockBuilder(), blockNumber)
			if err != nil {
				if err.Error() == "block content by block number error" {
					time.Sleep(1 * time.Second)
					// return errors.New("block content by block number error")
					continue
				}

				const msg = "failed to fetch new posted blocks: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			err = blockValidityProver.SyncBlockProverWithAuxInfo(result.BlockContent, result.PostedBlock)
			if err != nil {
				const msg = "failed to sync block prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}
			err = balanceProverService.SyncBalanceProver.SyncNoSend(
				syncValidityProver,
				mockWallet,
				balanceProverService.BalanceProcessor,
				// blockValidityProver.BlockBuilder(),
			)
			if err != nil {
				const msg = "failed to sync balance prover: %+v"
				s.log.Fatalf(msg, err.Error())
			}

			for _, deposit := range userAllData.Deposits {
				fmt.Printf("deposit ID: %d\n", deposit.DepositID)
				_, depositIndex, err := blockValidityProver.BlockBuilder().GetDepositLeafAndIndexByHash(deposit.DepositHash)
				if err != nil {
					const msg = "failed to get Deposit Index by Hash: %+v"
					s.log.Warnf(msg, err.Error())
					// return errors.New("failed to get Deposit Index by Hash")
					continue
				}
				if depositIndex == nil {
					const msg = "failed to get Deposit Index by Hash: %+v"
					s.log.Warnf(msg, "depositIndex is nil")
					// return errors.New("block content by block number error")
					continue
				}

				IsSynchronizedDepositIndex, err := blockValidityProver.BlockBuilder().IsSynchronizedDepositIndex(*depositIndex)
				if err != nil {
					const msg = "failed to check IsSynchronizedDepositIndex: %+v"
					s.log.Warnf(msg, err.Error())
					// return errors.New("failed to check IsSynchronizedDepositIndex")
					continue
				}
				if !IsSynchronizedDepositIndex {
					const msg = "deposit index %d is not synchronized"
					s.log.Warnf(msg, *depositIndex)
					// return errors.New("deposit index is not synchronized")
					continue
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
					s.log.Fatalf(msg, err.Error())
				}
			}

			// for _, transfer := range userAllData.Transfers {
			// 	fmt.Printf("transfer hash: %d\n", transfer.Hash())
			// 	_, depositIndex, err := blockValidityProver.BlockBuilder().GetDepositLeafAndIndexByHash(transfer.Hash())
			// 	if err != nil {
			// 		const msg = "failed to get Deposit Index by Hash: %+v"
			// 		s.log.Warnf(msg, err.Error())
			// 		// return errors.New("failed to get Deposit Index by Hash")
			// 		continue
			// 	}
			// 	if depositIndex == nil {
			// 		const msg = "failed to get Deposit Index by Hash: %+v"
			// 		s.log.Warnf(msg, "depositIndex is nil")
			// 		// return errors.New("block content by block number error")
			// 		continue
			// 	}

			// 	IsSynchronizedDepositIndex, err := blockValidityProver.BlockBuilder().IsSynchronizedDepositIndex(*depositIndex)
			// 	if err != nil {
			// 		const msg = "failed to check IsSynchronizedDepositIndex: %+v"
			// 		s.log.Warnf(msg, err.Error())
			// 		// return errors.New("failed to check IsSynchronizedDepositIndex")
			// 		continue
			// 	}
			// 	if !IsSynchronizedDepositIndex {
			// 		const msg = "deposit index %d is not synchronized"
			// 		s.log.Warnf(msg, *depositIndex)
			// 		// return errors.New("deposit index is not synchronized")
			// 		continue
			// 	}

			// 	fmt.Printf("deposit index: %d\n", *depositIndex)

			// 	depositCase := DepositCase{
			// 		Deposit: intMaxTree.DepositLeaf{
			// 			RecipientSaltHash: deposit.RecipientSaltHash,
			// 			TokenIndex:        deposit.TokenIndex,
			// 			Amount:            deposit.Amount,
			// 		},
			// 		DepositIndex: *depositIndex,
			// 		DepositSalt:  *deposit.Salt,
			// 	}
			// 	mockWallet.AddDepositCase(deposit.DepositID, &depositCase)
			// 	err = balanceProverService.SyncBalanceProver.ReceiveTransfer(
			// 		mockWallet,
			// 		balanceProverService.BalanceProcessor,
			// 		blockValidityProver.BlockBuilder(),
			// 		deposit.DepositID,
			// 	)
			// 	if err != nil {
			// 		const msg = "failed to receive deposit: %+v"
			// 		s.log.Fatalf(msg, err.Error())
			// 	}
			// }

			// for _, tx := range userAllData.Transactions {
			// 	syncValidityProver.Sync() // sync validity proofs
			// 	err = balanceProverService.SyncBalanceProver.SyncSend(
			// 		syncValidityProver,
			// 		mockWallet,
			// 		balanceProverService.BalanceProcessor,
			// 		blockValidityProver.BlockBuilder(),
			// 	)
			// 	if err != nil {
			// 		const msg = "failed to send transaction: %+v"
			// 		s.log.Fatalf(msg, err.Error())
			// 	}
			// }
		}

		blockNumber++
	}
}
