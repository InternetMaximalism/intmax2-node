package blockchain

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/mnemonic_wallet"
	modelsMW "intmax2-node/internal/mnemonic_wallet/models"
	"intmax2-node/internal/open_telemetry"
	"math/big"
	"os"
	"strings"
	"syscall"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/term"
)

const (
	TransactorChainIDCtx = "transactor_chain_id"
)

type serviceBlockchain struct {
	ctx context.Context
	cfg *configs.Config
}

func New(
	ctx context.Context,
	cfg *configs.Config,
) ServiceBlockchain {
	return &serviceBlockchain{
		ctx: ctx,
		cfg: cfg,
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

func (sb *serviceBlockchain) CheckPrivateKey(ctx context.Context) (err error) {
	const (
		hName = "ServiceBlockchain func:CheckPrivateKey"

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
	pk, err = sb.recognizingPrivateKey(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrRecognizingPrivateKeyFail, err)
	}

	var w *modelsMW.Wallet
	w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(pk)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	var bal *big.Int
	bal, err = sb.WalletBalance(spanCtx, *w.WalletAddress)
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

func (sb *serviceBlockchain) recognizingPrivateKey(
	ctx context.Context,
) (string, error) {
	const (
		hName    = "ServiceBlockchain func:recognizingPrivateKey"
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
		sb.cfg.Wallet.PrivateKeyHex = w.PrivateKey
	} else {
		_, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(
			sb.cfg.Wallet.PrivateKeyHex,
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
			sb.cfg.Wallet.PrivateKeyHex = text
			fmt.Println(emptyKey)
		}
	}

	return sb.cfg.Wallet.PrivateKeyHex, nil
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

func (sb *serviceBlockchain) callContract( // nolint:unused
	ctx context.Context,
	contractAddress common.Address,
	contractAbiPath, method string,
	args ...any,
) (resp []interface{}, err error) {
	const (
		hName              = "ServiceBlockchain func:callContract"
		chainIDKey         = "chain_id"
		contractAddressKey = "contract_address"
		contractAbiPathKey = "contract_abi_path"
		methodKey          = "method"
		argsKey            = "args"
		abiKey             = "abi"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(chainIDKey, sb.cfg.Blockchain.ScrollNetworkChainID),
			attribute.String(contractAddressKey, contractAddress.Hex()),
			attribute.String(contractAbiPathKey, contractAbiPath),
			attribute.String(methodKey, method),
			attribute.StringSlice(argsKey, func() (ret []string) {
				for key := range args {
					ret = append(ret, fmt.Sprintf("%+v", args[key]))
				}
				return ret
			}()),
		))
	defer span.End()

	var templateBytes []byte
	templateBytes, err = os.ReadFile(contractAbiPath)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrReadContractTemplateFile, err)
	}

	var abiJSON abi.ABI
	abiJSON, err = abi.JSON(strings.NewReader(gjson.GetBytes(templateBytes, abiKey).String()))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrGetAbiFail, err)
	}

	var (
		c      *ethclient.Client
		cancel func()
	)
	c, cancel, err = sb.ethClient(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrCreateEthClientFail, err)
	}
	defer cancel()

	err = bind.NewBoundContract(
		contractAddress, abiJSON, c, nil, nil,
	).Call(&bind.CallOpts{Context: spanCtx}, &resp, method, args...)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrCallContractFromBlockchainFail, err)
	}

	return resp, nil
}

func (sb *serviceBlockchain) contractTransactor(
	ctx context.Context,
	contractAddress common.Address,
	contractAbiPath string,
	value *big.Int,
	method string, args ...any,
) (resp *types.Transaction, err error) {
	const (
		hName              = "ServiceBlockchain func:contractTransactor"
		chainIDKey         = "chain_id"
		contractAddressKey = "contract_address"
		contractAbiPathKey = "contract_abi_path"
		methodKey          = "method"
		argsKey            = "args"
		abiKey             = "abi"
		intKey             = 0
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(chainIDKey, sb.cfg.Blockchain.ScrollNetworkChainID),
			attribute.String(contractAddressKey, contractAddress.Hex()),
			attribute.String(contractAbiPathKey, contractAbiPath),
			attribute.String(methodKey, method),
			attribute.StringSlice(argsKey, func() (ret []string) {
				for key := range args {
					ret = append(ret, fmt.Sprintf("%+v", args[key]))
				}
				return ret
			}()),
		))
	defer span.End()

	var templateBytes []byte
	templateBytes, err = os.ReadFile(contractAbiPath)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrReadContractTemplateFile, err)
	}

	var abiJSON abi.ABI
	abiJSON, err = abi.JSON(strings.NewReader(gjson.GetBytes(templateBytes, abiKey).String()))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrGetAbiFail, err)
	}

	var (
		c      *ethclient.Client
		cancel func()
	)
	c, cancel, err = sb.ethClient(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrCreateEthClientFail, err)
	}
	defer cancel()

	var pk string
	pk, err = sb.recognizingPrivateKey(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrRecognizingPrivateKeyFail, err)
	}

	var w *modelsMW.Wallet
	w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(pk)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	var nonce uint64
	nonce, err = c.PendingNonceAt(spanCtx, *w.WalletAddress)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrPendingNonceAtFail, err)
	}

	var chainID *big.Int
	chainID, err = c.ChainID(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrChainIDFormCtxFail, err)
	}

	var gasPrice *big.Int
	gasPrice, err = c.SuggestGasPrice(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrSuggestGasPriceFail, err)
	}

	var txOpts *bind.TransactOpts
	txOpts, err = bind.NewKeyedTransactorWithChainID(w.PK, chainID)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrNewKeyedTransactorWithChainIDFail, err)
	}

	if value == nil {
		value = big.NewInt(intKey) /** in wei */
	}

	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.Value = value /** in wei */
	txOpts.GasLimit = uint64(intKey)
	txOpts.GasPrice = gasPrice /** big.NewInt(1082420000000000) */
	txOpts.Context = context.WithValue(txOpts.Context, TransactorChainIDCtx, chainID)

	resp, err = bind.NewBoundContract(
		contractAddress, abiJSON, nil, c, nil,
	).Transact(txOpts, method, args...)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrApplyBoundContractTransactorFail, err)
	}

	return resp, nil
}
