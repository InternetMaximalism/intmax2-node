package blockchain

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	modelsMW "intmax2-node/internal/mnemonic_wallet/models"
	"intmax2-node/internal/open_telemetry"
	"math/big"
	"syscall"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/term"
)

type serviceBlockchain struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
}

func New(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) ServiceBlockchain {
	return &serviceBlockchain{
		ctx: ctx,
		cfg: cfg,
		log: log,
	}
}

func (sb *serviceBlockchain) Check(_ context.Context) (res health.Health) {
	const (
		insufficientFundsKey = "insufficient_funds"
	)

	res.AddInfo(insufficientFundsKey, errorsB.InsufficientFunds)
	res.Up()
	if errorsB.InsufficientFunds {
		res.Down()
	}

	return res
}

func (sb *serviceBlockchain) CheckScrollPrivateKey(ctx context.Context) (err error) {
	const (
		hName = "ServiceBlockchain func:CheckScrollPrivateKey"

		minus1Key = -1
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err = sb.SetupScrollNetworkChainID(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrSetupScrollNetworkChainIDFail, err)
	}

	var pk string
	pk, err = sb.recognizingScrollPrivateKey(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrRecognizingScrollPrivateKeyFail, err)
	}

	var w *modelsMW.Wallet
	w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(pk)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	var bal *big.Int
	bal, err = sb.walletBalance(spanCtx, *w.WalletAddress)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrGettingWalletBalanceErrorOccurred, err)
	}

	if minus1Key == bal.Cmp(&sb.cfg.Blockchain.ScrollNetworkMinBalance) {
		open_telemetry.MarkSpanError(spanCtx, errorsB.ErrWalletInsufficientFundsForNodeStart)
		return errorsB.ErrWalletInsufficientFundsForNodeStart
	}

	return nil
}

func (sb *serviceBlockchain) CheckEthereumPrivateKey(ctx context.Context) (err error) {
	const (
		hName = "ServiceBlockchain func:CheckEthereumPrivateKey"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err = sb.SetupEthereumNetworkChainID(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrSetupEthereumNetworkChainIDFail, err)
	}

	_, err = sb.recognizingEthereumPrivateKey(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrRecognizingEthereumPrivateKeyFail, err)
	}

	return nil
}

func (sb *serviceBlockchain) recognizingScrollPrivateKey(
	ctx context.Context,
) (string, error) {
	const (
		hName    = "ServiceBlockchain func:recognizingScrollPrivateKey"
		emptyKey = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	w, err := mnemonic_wallet.New().WalletFromMnemonic(
		sb.cfg.Wallet.MnemonicValue,
		sb.cfg.Wallet.MnemonicPassword,
		sb.cfg.Wallet.MnemonicDerivationPath,
	)
	if err == nil {
		sb.cfg.Blockchain.BuilderPrivateKeyHex = w.PrivateKey
	} else {
		_, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(
			sb.cfg.Blockchain.BuilderPrivateKeyHex,
		)
		if err != nil {
			const enterMSG = "Enter private key:"
			fmt.Printf(enterMSG)
			var (
				text   string
				bytePK []byte
			)
			bytePK, err = term.ReadPassword(syscall.Stdin)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return emptyKey, errors.Join(errorsB.ErrStdinProcessingFail, err)
			}
			text = string(bytePK)
			_, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(
				text,
			)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return emptyKey, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
			}
			sb.cfg.Blockchain.BuilderPrivateKeyHex = text
			fmt.Println(emptyKey)
		}
	}

	return sb.cfg.Blockchain.BuilderPrivateKeyHex, nil
}

func (sb *serviceBlockchain) recognizingEthereumPrivateKey(
	ctx context.Context,
) (string, error) {
	const (
		hName    = "ServiceBlockchain func:recognizingEthereumPrivateKey"
		emptyKey = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	_, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(
		sb.cfg.Blockchain.BuilderPrivateKeyHex,
	)
	if err != nil {
		const enterMSG = "Enter private key:"
		fmt.Printf(enterMSG)
		var (
			text   string
			bytePK []byte
		)
		bytePK, err = term.ReadPassword(syscall.Stdin)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return emptyKey, errors.Join(errorsB.ErrStdinProcessingFail, err)
		}
		text = string(bytePK)
		_, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(
			text,
		)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return emptyKey, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
		}
		sb.cfg.Blockchain.BuilderPrivateKeyHex = text
		fmt.Println(emptyKey)
	}

	return sb.cfg.Blockchain.BuilderPrivateKeyHex, nil
}

func (sb *serviceBlockchain) ethClient(
	ctx context.Context,
) (c *ethclient.Client, cancel func(), err error) {
	const (
		hName      = "ServiceBlockchain func:ethClient"
		chainIDKey = "chain_id"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(chainIDKey, sb.cfg.Blockchain.ScrollNetworkChainID),
		))
	defer span.End()

	var cID uint256.Int
	err = cID.Scan(sb.cfg.Blockchain.ScrollNetworkChainID)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, nil, errors.Join(errorsB.ErrParseChainIDFail, err)
	}

	var link string
	link, err = sb.ScrollNetworkChainLinkEvmJSONRPC(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, nil, errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	c, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, nil, errors.Join(errorsB.ErrCreateNewClientOfRPCEthFail, err)
	}
	cancel = func() {
		c.Close()
	}

	return c, cancel, nil
}
