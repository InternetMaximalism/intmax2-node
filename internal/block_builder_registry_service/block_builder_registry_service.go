package block_builder_registry_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/mnemonic_wallet"
	modelsMW "intmax2-node/internal/mnemonic_wallet/models"
	"intmax2-node/internal/open_telemetry"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type blockBuilderRegistryService struct {
	cfg *configs.Config
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	sb ServiceBlockchain,
) BlockBuilderRegistryService {
	return &blockBuilderRegistryService{
		cfg: cfg,
		sb:  sb,
	}
}

func (bbr *blockBuilderRegistryService) GetBlockBuilder(
	ctx context.Context,
) (*IBlockBuilderRegistryBlockBuilderInfo, error) {
	const (
		hName = "BlockBuilderRegistryService func:GetBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := bbr.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	var callerBBR *bindings.BlockBuilderRegistryCaller
	callerBBR, err = bindings.NewBlockBuilderRegistryCaller(
		common.HexToAddress(bbr.cfg.Blockchain.BlockBuilderRegistryContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNewBlockBuilderRegistryCallerFail, err)
	}

	var w *modelsMW.Wallet
	w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(bbr.cfg.Wallet.PrivateKeyHex)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	for {
		opts := bind.CallOpts{Context: spanCtx}
		var blockBuilderInfo IBlockBuilderRegistryBlockBuilderInfo
		blockBuilderInfo, err = callerBBR.BlockBuilders(&opts, *w.WalletAddress)
		if err != nil {
			switch {
			case
				strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr):
				continue
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(ErrProcessingFuncUpdateBlockBuilderOfBlockBuilderRegistryFail, err)
		}

		return &blockBuilderInfo, nil
	}
}

func (bbr *blockBuilderRegistryService) UpdateBlockBuilder(
	ctx context.Context,
	url string,
) error {
	const (
		hName      = "BlockBuilderRegistryService func:UpdateBlockBuilder"
		urlKey     = "url"
		valueKey   = "value"
		int0StrKey = "0"
		int1Key    = 1
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(urlKey, url),
		))
	defer span.End()

	var (
		value  *big.Int
		tryNum int
	)
	defer func() {
		if value == nil {
			span.SetAttributes(attribute.String(valueKey, int0StrKey))
		} else {
			span.SetAttributes(attribute.String(valueKey, value.String()))
		}
	}()

	link, err := bbr.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	// Check to see if you have already done a 0.1 ETH stake.
	callerBBR, err := bindings.NewBlockBuilderRegistry(
		common.HexToAddress(bbr.cfg.Blockchain.BlockBuilderRegistryContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrNewBlockBuilderRegistryCallerFail, err)
	}

	privateKey, err := crypto.HexToECDSA(bbr.cfg.Wallet.PrivateKeyHex)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrLoadPrivateKeyFail, err)
	}
	builderAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	res, err := callerBBR.BlockBuilders(&bind.CallOpts{Context: spanCtx}, builderAddress)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrGetBlockBuilderInfoFail, err)
	}

	// If the stake is more than 0.1 ETH and the URL has not changed, the update function is not executed.
	if res.StakeAmount.Cmp(&bbr.cfg.Blockchain.ScrollNetworkStakeBalance) >= 0 && res.BlockBuilderUrl == url {
		return nil
	}

	value = new(big.Int).Sub(&bbr.cfg.Blockchain.ScrollNetworkStakeBalance, res.StakeAmount)

	var transactorBBR *bindings.BlockBuilderRegistryTransactor
	transactorBBR, err = bindings.NewBlockBuilderRegistryTransactor(
		common.HexToAddress(bbr.cfg.Blockchain.BlockBuilderRegistryContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrNewBlockBuilderRegistryTransactorFail, err)
	}

	for {
		var transactOpts *bind.TransactOpts
		transactOpts, err = createTransactor(bbr.cfg)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrCreateOptionsOfTransactionFail, err)
		}
		transactOpts.Value = value

		_, err = transactorBBR.UpdateBlockBuilder(transactOpts, url)
		if err != nil {
			switch {
			case
				strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.ErrInvalidSequenceStr):
				continue
			case strings.Contains(err.Error(), errorsB.ErrInsufficientStakeAmountStr):
				value = &bbr.cfg.Blockchain.ScrollNetworkStakeBalance
				if tryNum < int1Key {
					tryNum++
					continue
				}
				errorsB.InsufficientFunds = true
			case strings.Contains(err.Error(), errorsB.ErrInsufficientFundsStr):
				errorsB.InsufficientFunds = true
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrProcessingFuncUpdateBlockBuilderOfBlockBuilderRegistryFail, err)
		}
		errorsB.InsufficientFunds = false

		return nil
	}
}

func (bbr *blockBuilderRegistryService) StopBlockBuilder(
	ctx context.Context,
) (err error) {
	const (
		hName = "BlockBuilderRegistryService func:StopBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := bbr.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	var transactorBBR *bindings.BlockBuilderRegistryTransactor
	transactorBBR, err = bindings.NewBlockBuilderRegistryTransactor(
		common.HexToAddress(bbr.cfg.Blockchain.BlockBuilderRegistryContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrNewBlockBuilderRegistryTransactorFail, err)
	}

	for {
		var transactOpts *bind.TransactOpts
		transactOpts, err = createTransactor(bbr.cfg)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrCreateOptionsOfTransactionFail, err)
		}

		_, err = transactorBBR.StopBlockBuilder(transactOpts)
		if err != nil {
			switch {
			case
				strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.ErrInvalidSequenceStr):
				<-time.After(time.Second)
				continue
			case strings.Contains(err.Error(), errorsB.ErrBlockBuilderNotFoundStr):
				const mask = "%s"
				err = fmt.Errorf(mask, errorsB.ErrBlockBuilderNotFoundStr)
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrProcessingFuncStopOfBlockBuilderRegistryFail, err)
		}

		return nil
	}
}

func (bbr *blockBuilderRegistryService) UnStakeBlockBuilder(
	ctx context.Context,
) (err error) {
	const (
		hName = "BlockBuilderRegistryService func:UnStakeBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := bbr.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	var transactorBBR *bindings.BlockBuilderRegistryTransactor
	transactorBBR, err = bindings.NewBlockBuilderRegistryTransactor(
		common.HexToAddress(bbr.cfg.Blockchain.BlockBuilderRegistryContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrNewBlockBuilderRegistryTransactorFail, err)
	}

	for {
		var transactOpts *bind.TransactOpts
		transactOpts, err = createTransactor(bbr.cfg)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrCreateOptionsOfTransactionFail, err)
		}

		_, err = transactorBBR.Unstake(transactOpts)
		if err != nil {
			switch {
			case
				strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr),
				strings.Contains(err.Error(), errorsB.ErrInvalidSequenceStr):
				<-time.After(time.Second)
				continue
			case strings.Contains(err.Error(), errorsB.ErrBlockBuilderNotFoundStr):
				const mask = "%s"
				err = fmt.Errorf(mask, errorsB.ErrBlockBuilderNotFoundStr)
			case strings.Contains(err.Error(), errorsB.ErrCantUnStakeBlockBuilderStr):
				const mask = "%s"
				err = fmt.Errorf(mask, errorsB.ErrCantUnStakeBlockBuilderStr)
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(ErrProcessingFuncUnStakeOfBlockBuilderRegistryFail, err)
		}

		return nil
	}
}

func createTransactor(cfg *configs.Config) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(cfg.Wallet.PrivateKeyHex)
	if err != nil {
		return nil, errors.Join(ErrLoadPrivateKeyFail, err)
	}

	const (
		int10Key = 10
		int64Key = 64
	)
	var chainID int64
	chainID, err = strconv.ParseInt(cfg.Blockchain.ScrollNetworkChainID, int10Key, int64Key)
	if err != nil {
		return nil, errors.Join(ErrParseStrToIntFail, err)
	}

	var transactOpts *bind.TransactOpts
	transactOpts, err = bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, errors.Join(ErrCreateTransactorFail, err)
	}

	return transactOpts, nil
}
