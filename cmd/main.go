package main

import (
	"context"
	"intmax2-node/cmd/block_builder"
	"intmax2-node/cmd/deposit"
	"intmax2-node/cmd/ethereum_private_key_wallet"
	"intmax2-node/cmd/generate_account"
	"intmax2-node/cmd/intmax_private_key_wallet"
	"intmax2-node/cmd/messenger"
	"intmax2-node/cmd/migrator"
	"intmax2-node/cmd/mnemonic_account"
	"intmax2-node/cmd/server"
	"intmax2-node/cmd/store_vault_server"
	"intmax2-node/cmd/sync_balance"
	"intmax2-node/cmd/transaction"
	"intmax2-node/cmd/withdrawal"
	"intmax2-node/cmd/withdrawal_server"
	"intmax2-node/configs"
	"intmax2-node/internal/block_builder_registry_service"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/blockchain"
	"intmax2-node/internal/cli"
	"intmax2-node/internal/deposit_synchronizer"
	"intmax2-node/internal/network_service"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pow"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/logger"
	"intmax2-node/pkg/sql_db/db_app"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dimiro1/health"
)

func main() {
	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	const int1 = 1
	done := make(chan os.Signal, int1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer close(done)

	go func() {
		<-done
		const msg = "SIGTERM detected"
		log.Errorf(msg)
		if cancel != nil {
			cancel()
		}
	}()

	err := open_telemetry.Init(cfg.OpenTelemetry.Enable)
	if err != nil {
		const msg = "open_telemetry init: %v"
		log.Errorf(msg, err)
		return
	}

	var dbApp db_app.SQLDb
	dbApp, err = db_app.New(ctx, log, &cfg.SQLDb)
	if err != nil {
		const msg = "db application init: %v"
		log.Errorf(msg, err)
		return
	}

	w := worker.New(cfg, log, dbApp)
	depositSynchronizer := deposit_synchronizer.New(cfg, log, dbApp)
	blockValidityProver := block_validity_prover.New(cfg, log, dbApp)
	bc := blockchain.New(ctx, cfg)
	ns := network_service.New(cfg)
	hc := health.NewHandler()
	bbr := block_builder_registry_service.New(cfg, bc)

	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	pwNonce := pow.NewPoWNonce(pw, pWorker)

	wg := sync.WaitGroup{}

	err = cli.Run(
		ctx,
		server.NewServerCmd(&server.Server{
			Context:             ctx,
			Cancel:              cancel,
			Config:              cfg,
			Log:                 log,
			DbApp:               dbApp,
			WG:                  &wg,
			BBR:                 bbr,
			SB:                  bc,
			NS:                  ns,
			HC:                  &hc,
			PoW:                 pwNonce,
			Worker:              w,
			DepositSynchronizer: depositSynchronizer,
			BlockValidityProver: blockValidityProver,
		}),
		migrator.NewMigratorCmd(ctx, log, dbApp),
		deposit.NewDepositCmd(&deposit.Deposit{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
		withdrawal.NewWithdrawCmd(&withdrawal.Withdrawal{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
		withdrawal_server.NewServerCmd(&withdrawal_server.WithdrawalServer{
			Context: ctx,
			Cancel:  cancel,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			WG:      &wg,
			HC:      &hc,
		}),
		store_vault_server.NewServerCmd(&store_vault_server.StoreVaultServer{
			Context: ctx,
			Cancel:  cancel,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			WG:      &wg,
			HC:      &hc,
		}),
		generate_account.NewCmd(log),
		mnemonic_account.NewCmd(log),
		ethereum_private_key_wallet.NewCmd(log),
		intmax_private_key_wallet.NewCmd(log),
		sync_balance.NewBalanceCmd(&sync_balance.Balance{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
		transaction.NewTransactionCmd(&transaction.Transaction{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
		block_builder.NewCmd(ctx, log, bc, bbr),
		messenger.NewMessengerCmd(&messenger.Messenger{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
	)
	if err != nil {
		const msg = "cli: %v"
		log.Errorf(msg, err)
		return
	}

	wg.Wait()
}
