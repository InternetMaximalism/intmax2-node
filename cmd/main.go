package main

import (
	"context"
	"intmax2-node/cmd/block_builder"
	"intmax2-node/cmd/deposit"
	"intmax2-node/cmd/ethereum_private_key_wallet"
	"intmax2-node/cmd/generate_account"
	"intmax2-node/cmd/intmax_private_key_wallet"
	"intmax2-node/cmd/migrator"
	"intmax2-node/cmd/mnemonic_account"
	"intmax2-node/cmd/server"
	"intmax2-node/configs"
	"intmax2-node/internal/block_builder_registry_service"
	"intmax2-node/internal/blockchain"
	"intmax2-node/internal/cli"
	"intmax2-node/internal/network_service"
	"intmax2-node/internal/open_telemetry"
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

	bc := blockchain.New(ctx, cfg)
	ns := network_service.New(cfg)
	hc := health.NewHandler()
	bbr := block_builder_registry_service.New(cfg, bc)

	wg := sync.WaitGroup{}

	err = cli.Run(
		ctx,
		server.NewServerCmd(&server.Server{
			Context: ctx,
			Cancel:  cancel,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			WG:      &wg,
			BBR:     bbr,
			SB:      bc,
			NS:      ns,
			HC:      &hc,
		}),
		migrator.NewMigratorCmd(ctx, log, dbApp),
		deposit.NewDepositCmd(&deposit.Deposit{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			DbApp:   dbApp,
			SB:      bc,
		}),
		generate_account.NewCmd(log),
		mnemonic_account.NewCmd(log),
		ethereum_private_key_wallet.NewCmd(log),
		intmax_private_key_wallet.NewCmd(log),
		block_builder.NewCmd(ctx, log, bc, bbr),
	)
	if err != nil {
		const msg = "cli: %v"
		log.Errorf(msg, err)
		return
	}

	wg.Wait()
}
