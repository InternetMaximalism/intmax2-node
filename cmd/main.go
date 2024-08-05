package main

import (
	"context"
	"intmax2-node/cmd/balance_checker"
	"intmax2-node/cmd/ethereum_private_key_wallet"
	"intmax2-node/cmd/generate_account"
	"intmax2-node/cmd/intmax_private_key_wallet"
	"intmax2-node/cmd/mnemonic_account"
	"intmax2-node/cmd/transaction"
	"intmax2-node/configs"
	"intmax2-node/internal/blockchain"
	"intmax2-node/internal/cli"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	bc := blockchain.New(ctx, cfg)
	wg := sync.WaitGroup{}

	err = cli.Run(
		ctx,
		generate_account.NewCmd(log),
		mnemonic_account.NewCmd(log),
		ethereum_private_key_wallet.NewCmd(log),
		intmax_private_key_wallet.NewCmd(log),
		balance_checker.NewBalanceCmd(&balance_checker.Balance{
			Context: ctx,
			Config:  cfg,
			Log:     log,
			SB:      bc,
		}),
		transaction.NewTransactionCmd(&transaction.Transaction{
			Context: ctx,
			Config:  cfg,
			Log:     log,
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
