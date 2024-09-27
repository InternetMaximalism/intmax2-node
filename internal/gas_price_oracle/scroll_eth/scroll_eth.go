package scroll_eth

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/pkg/utils"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type scrollEth struct {
	cfg *configs.Config
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	sb ServiceBlockchain,
) ScrollEth {
	return &scrollEth{
		cfg: cfg,
		sb:  sb,
	}
}

// GasFee describes gas fee with scroll ETH
func (s *scrollEth) GasFee(ctx context.Context) (gasFee *big.Int, err error) {
	var l1GasFee *big.Int
	l1GasFee, err = s.l1GasFee(ctx)
	if err != nil {
		return nil, errors.Join(ErrL1GasFeeFail, err)
	}

	var l2GasFee *big.Int
	l2GasFee, err = s.l2GasFee(ctx)
	if err != nil {
		return nil, errors.Join(ErrL2GasFeeFail, err)
	}

	return new(big.Int).Div(
		new(big.Int).Add(l1GasFee, l2GasFee),
		new(big.Int).SetInt64(int64(s.cfg.GasPriceOracle.Delimiter)),
	), nil
}

// l1GasFee describes gas fee with scroll l1 ETH
// @see https://docs.scroll.io/en/developers/transaction-fees-on-scroll
// @see https://scrollscan.com/address/0x5300000000000000000000000000000000000002
// @see https://sepolia.scrollscan.com/address/0x5300000000000000000000000000000000000002
func (s *scrollEth) l1GasFee(ctx context.Context) (l1GasFee *big.Int, err error) {
	const (
		addressGasPriceOracleScrollL1 = "0x5300000000000000000000000000000000000002"
		zeros                         = 0
		txDataZeroGas                 = 4
		txDataNonZeroGas              = 16
		precision                     = 1e9
		txsCount                      = 128
		lenItem                       = 64
		txsAddCount                   = 3
		addPercent                    = 10
		int4Key                       = 4
		int100Key                     = 100
	)

	// nonzeros=64*128+64*3=+10%=9223
	nonZeroPart := lenItem*txsCount + lenItem*txsAddCount
	nonZeros := new(big.Int).Add(
		new(big.Int).SetInt64(int64(nonZeroPart)),
		new(big.Int).SetInt64(int64(math.Ceil(float64(nonZeroPart*addPercent/int100Key)))),
	)

	// nolint:gocritic
	// l1Gas = zeros * TX_DATA_ZERO_GAS + (nonzeros + 4) * TX_DATA_NON_ZERO_GAS
	l1Gas := new(big.Int).Add(
		new(big.Int).Mul(new(big.Int).SetInt64(zeros), new(big.Int).SetInt64(txDataZeroGas)),
		new(big.Int).Mul(
			new(big.Int).Add(nonZeros, new(big.Int).SetInt64(int4Key)),
			new(big.Int).SetInt64(txDataNonZeroGas),
		),
	)

	var ethLink string
	ethLink, err = s.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail)
	}

	var ethClient *ethclient.Client
	ethClient, err = utils.NewClient(ethLink)
	if err != nil {
		return nil, errors.Join(errorsB.ErrEthClientDialFail)
	}

	var oracle *bindings.GasPriceOracleScrollL1ETH
	oracle, err = bindings.NewGasPriceOracleScrollL1ETH(
		common.HexToAddress(addressGasPriceOracleScrollL1),
		ethClient,
	)
	if err != nil {
		return nil, errors.Join(ErrNewGasPriceOracleScrollL1ETHFail, err)
	}

	ctxScalar, cancelScalar := context.WithTimeout(ctx, s.cfg.GasPriceOracle.Timeout)
	defer func() {
		if cancelScalar != nil {
			cancelScalar()
		}
	}()

	var scalar *big.Int
	scalar, err = oracle.Scalar(&bind.CallOpts{
		Pending: false,
		Context: ctxScalar,
	})
	if err != nil {
		return nil, errors.Join(ErrScalarFail, err)
	}

	ctxOverhead, cancelOverhead := context.WithTimeout(ctx, s.cfg.GasPriceOracle.Timeout)
	defer func() {
		if cancelOverhead != nil {
			cancelOverhead()
		}
	}()

	var overhead *big.Int
	overhead, err = oracle.Overhead(&bind.CallOpts{
		Pending: false,
		Context: ctxOverhead,
	})
	if err != nil {
		return nil, errors.Join(ErrOverheadFail, err)
	}

	ctxL1BaseFee, cancelL1BaseFee := context.WithTimeout(ctx, s.cfg.GasPriceOracle.Timeout)
	defer func() {
		if cancelL1BaseFee != nil {
			cancelL1BaseFee()
		}
	}()

	var l1BaseFee *big.Int
	l1BaseFee, err = oracle.L1BaseFee(&bind.CallOpts{
		Pending: false,
		Context: ctxL1BaseFee,
	})
	if err != nil {
		return nil, errors.Join(ErrL1BaseFeeFail, err)
	}

	// nolint:gocritic
	// l1GasFee = ((l1Gas + overhead) * l1BaseFee * scalar) / PRECISION
	l1BaseFee = new(big.Int).Div(
		new(big.Int).Mul(
			new(big.Int).Add(l1Gas, overhead),
			new(big.Int).Mul(l1BaseFee, scalar),
		),
		new(big.Int).SetInt64(precision),
	)

	return l1BaseFee, nil
}

func (s *scrollEth) l2GasFee(_ context.Context) (l1GasFee *big.Int, err error) { // nolint:unparam
	const int0Key = 0
	return new(big.Int).SetInt64(int0Key), nil
}
