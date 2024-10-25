package scroll_eth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type scrollEth struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) ScrollEth {
	const (
		moduleKey = "module"
		gpoKey    = "scroll_eth_gpo"
	)

	return &scrollEth{
		cfg: cfg,
		log: log.WithFields(logger.Fields{moduleKey: gpoKey}),
		sb:  sb,
	}
}

// GasFee describes gas fee with scroll ETH
func (s *scrollEth) GasFee(ctx context.Context) (gasFee *big.Int, err error) {
	const (
		extraPercent        = uint64(50)
		maxBlocksFeeHistory = uint64(100)
	)

	var l1GasFee *big.Int
	l1GasFee, err = s.l1GasFee(ctx, maxBlocksFeeHistory, extraPercent)
	if err != nil {
		return nil, errors.Join(ErrL1GasFeeFail, err)
	}

	s.log.Debugf("l1GasFee value is %s", l1GasFee)

	var l2GasFee *big.Int
	l2GasFee, err = s.l2GasFee(ctx, maxBlocksFeeHistory, extraPercent)
	if err != nil {
		return nil, errors.Join(ErrL2GasFeeFail, err)
	}

	s.log.Debugf("l2GasFee value is %s", l2GasFee)

	gasFee = new(big.Int).Div(
		new(big.Int).Add(
			new(big.Int).Add(l1GasFee, l2GasFee),
			new(big.Int).SetUint64(uint64(s.cfg.GasPriceOracle.ExtraFee)),
		),
		new(big.Int).SetInt64(int64(s.cfg.GasPriceOracle.Delimiter)),
	)

	s.log.Debugf("gasFee value is %s", gasFee)

	return gasFee, nil
}

// l1GasFee returns gas fee for L1
func (s *scrollEth) l1GasFee(
	ctx context.Context,
	maxBlocksFeeHistory, extraPercent uint64,
) (l1GasFee *big.Int, err error) {
	const (
		maxGasValue      = uint64(500)
		ctrlGasUsedRatio = 0.38
	)

	var ethLink string
	ethLink, err = s.sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(errorsB.ErrEthereumNetworkChainLinkEvmJSONRPCFail)
	}

	var ethClient *ethclient.Client
	ethClient, err = utils.NewClient(ethLink)
	if err != nil {
		return nil, errors.Join(errorsB.ErrEthClientDialFail)
	}
	defer ethClient.Close()

	var fh *ethereum.FeeHistory
	fh, err = ethClient.FeeHistory(
		ctx,
		maxBlocksFeeHistory,
		nil,
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrFeeHistoryFail, err)
	}

	var gasUsedRatio float64
	for index := range fh.GasUsedRatio {
		gasUsedRatio += fh.GasUsedRatio[index]
	}

	middleGasUsedRatio := gasUsedRatio / float64(len(fh.GasUsedRatio))

	s.log.Debugf(
		"l1GasFee middleGasUsedRatio >= ctrlGasUsedRatio is %t (%v >= %v)",
		middleGasUsedRatio >= ctrlGasUsedRatio, middleGasUsedRatio, ctrlGasUsedRatio,
	)

	if middleGasUsedRatio >= ctrlGasUsedRatio {
		s.log.Debugf("l1GasFee used maxGasValue (maxGasValue = %v)", maxGasValue)

		l1GasFee, err = s.gasFee(ctx, ethClient, maxBlocksFeeHistory, extraPercent)
		if err != nil {
			return nil, errors.Join(ErrGasFeeFail, err)
		}

		return new(big.Int).Mul(l1GasFee, new(big.Int).SetUint64(maxGasValue)), nil
	}

	var sLink string
	sLink, err = s.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail)
	}

	var sClient *ethclient.Client
	sClient, err = utils.NewClient(sLink)
	if err != nil {
		return nil, errors.Join(errorsB.ErrEthClientDialFail)
	}
	defer sClient.Close()

	l1GasFee, err = s.oracleL1GasFeeScroll(ctx, sClient)
	if err != nil {
		return nil, errors.Join(ErrOracleL1GasFeeScrollFail, err)
	}

	s.log.Debugf("l1GasFee used oracleL1Scroll")

	return new(big.Int).Add(
		l1GasFee,
		s.addPercentFee(l1GasFee, extraPercent),
	), nil
}

// l2GasFee returns gas fee for L2
func (s *scrollEth) l2GasFee(
	ctx context.Context,
	maxBlocksFeeHistory, extraPercent uint64,
) (l2GasFee *big.Int, err error) {
	var link string
	link, err = s.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = utils.NewClient(link)
	if err != nil {
		return nil, errors.Join(errorsB.ErrEthClientDialFail, err)
	}
	defer client.Close()

	l2GasFee, err = s.gasFee(ctx, client, maxBlocksFeeHistory, extraPercent)
	if err != nil {
		return nil, errors.Join(ErrGasFeeFail, err)
	}

	var gasValue *big.Int
	gasValue, err = s.gasValueForBlockContentWithRollupAndScrollNetwork(client)
	if err != nil {
		return nil, errors.Join(ErrGasValueForBlockContentWithRollupAndScrollNetworkFail, err)
	}

	return new(big.Int).Mul(l2GasFee, gasValue), nil
}

// gasFee returns gas fee
func (s *scrollEth) gasFee(
	ctx context.Context,
	client *ethclient.Client,
	maxBlocksFeeHistory, extraPercent uint64,
) (*big.Int, error) {
	fh, err := client.FeeHistory(
		ctx,
		maxBlocksFeeHistory,
		nil,
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrFeeHistoryFail, err)
	}

	bf := new(big.Int)
	for index := range fh.BaseFee {
		_ = bf.Add(bf, fh.BaseFee[index])
	}

	middleBF := new(big.Int).Div(bf, new(big.Int).SetInt64(int64(len(fh.BaseFee))))

	var sgTipCap *big.Int
	sgTipCap, err = client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, errors.Join(ErrL2SuggestGasTipCapFail, err)
	}

	return new(big.Int).Add(
		new(big.Int).Add(
			middleBF,
			s.addPercentFee(middleBF, extraPercent),
		),
		new(big.Int).Add(sgTipCap, s.addPercentFee(sgTipCap, extraPercent)),
	), nil
}

// oracleL1GasFeeScroll describes gas fee with scroll l1 ETH
// @see https://docs.scroll.io/en/developers/transaction-fees-on-scroll
// @see https://scrollscan.com/address/0x5300000000000000000000000000000000000002
// @see https://sepolia.scrollscan.com/address/0x5300000000000000000000000000000000000002
func (s *scrollEth) oracleL1GasFeeScroll(
	ctx context.Context,
	sClient *ethclient.Client,
) (l1GasFee *big.Int, err error) {
	const (
		addressGasPriceOracleScrollL1 = "0x5300000000000000000000000000000000000002"
		zeros                         = 0
		txDataZeroGas                 = 4
		txDataNonZeroGas              = 16
		precision                     = 1e9
		delimiterForCorrection        = 10
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

	var oracle *bindings.GasPriceOracleScrollL1ETH
	oracle, err = bindings.NewGasPriceOracleScrollL1ETH(
		common.HexToAddress(addressGasPriceOracleScrollL1),
		sClient,
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
	// oracleL1GasFeeScroll = ((l1Gas + overhead) * l1BaseFee * scalar) / PRECISION / delimiterForCorrection
	l1BaseFee = new(big.Int).Div(new(big.Int).Div(
		new(big.Int).Mul(
			new(big.Int).Add(l1Gas, overhead),
			new(big.Int).Mul(l1BaseFee, scalar),
		),
		new(big.Int).SetInt64(precision),
	), new(big.Int).SetUint64(uint64(delimiterForCorrection)))

	return l1BaseFee, nil
}

func (s *scrollEth) addPercentFee(input *big.Int, extraPercent uint64) (value *big.Int) {
	const (
		base100Percent = 100
	)

	value = new(big.Int).Mul(input, new(big.Int).SetUint64(extraPercent))
	return new(big.Int).Div(value, new(big.Int).SetInt64(base100Percent))
}

func (s *scrollEth) gasValueForBlockContentWithRollupAndScrollNetwork(
	client *ethclient.Client,
) (value *big.Int, err error) {
	const (
		int10Key         = 10
		int64Key         = 64
		maxGasValue      = uint64(500000)
		errExecRevertStr = "execution reverted"
	)

	var rollup *bindings.Rollup
	rollup, err = bindings.NewRollup(
		common.HexToAddress(s.cfg.Blockchain.RollupContractAddress), client)
	if err != nil {
		return nil, errors.Join(ErrNewRollupFail, err)
	}

	var privateKey *ecdsa.PrivateKey
	privateKey, err = crypto.HexToECDSA(s.cfg.Blockchain.BuilderPrivateKeyHex)
	if err != nil {
		return nil, errors.Join(ErrLoadPkBlockBuilderFail, err)
	}

	var chainID int64
	chainID, err = strconv.ParseInt(
		s.cfg.Blockchain.ScrollNetworkChainID, int10Key, int64Key,
	)
	if err != nil {
		return nil, errors.Join(ErrParseScrollNetworkChainIDFail, err)
	}

	var transactOpts *bind.TransactOpts
	transactOpts, err = bind.NewKeyedTransactorWithChainID(
		privateKey,
		big.NewInt(chainID),
	)
	if err != nil {
		return nil, errors.Join(ErrNewKeyedTransactorWithChainIDFail, err)
	}

	transactOpts.NoSend = true

	var input *intMaxTypes.PostRegistrationBlockInput
	input, err = s.blockContent()
	if err != nil {
		return nil, errors.Join(ErrBlockContentFail, err)
	}

	var tx *types.Transaction
	tx, err = rollup.PostRegistrationBlock(
		transactOpts,
		input.TxTreeRoot,
		input.SenderFlags,
		input.AggregatedPublicKey,
		input.AggregatedSignature,
		input.MessagePoint,
		input.SenderPublicKeys,
	)
	if err != nil {
		if strings.Contains(err.Error(), errExecRevertStr) {
			s.log.Debugf("l2GasFee used maxGasValue (maxGasValue = %v)", maxGasValue)
			return new(big.Int).SetUint64(maxGasValue), nil
		}
		return nil, errors.Join(ErrPostRegistrationBlockFail, err)
	}

	gas := tx.Gas()
	if gas > maxGasValue {
		s.log.Debugf("l2GasFee used gasValue = %v", gas)

		return new(big.Int).SetUint64(gas), nil
	}

	s.log.Debugf("l2GasFee used maxGasValue (maxGasValue = %v)", maxGasValue)

	return new(big.Int).SetUint64(maxGasValue), nil
}

func (s *scrollEth) blockContent() (*intMaxTypes.PostRegistrationBlockInput, error) {
	const (
		int0Key   = 0
		int1Key   = 1
		int2Key   = 2
		int32Key  = 32
		int100Key = 100
		int128Key = 128
	)

	bigInt1 := big.NewInt(int1Key)

	keyPairs := make([]*accounts.PrivateKey, int100Key)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(
			rand.Reader,
			new(big.Int).Sub(fr.Modulus(), bigInt1),
		)
		if err != nil {
			return nil, errors.Join(ErrRandIntFail, err)
		}

		privateKey.Add(privateKey, bigInt1)
		keyPairs[i], err = accounts.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		if err != nil {
			return nil, errors.Join(ErrNewPrivateKeyWithReCalcPubKeyIfPkNegatesFail, err)
		}
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > int0Key
	})

	senders := make([]intMaxTypes.Sender, int128Key)
	for i, keyPair := range keyPairs {
		senders[i] = intMaxTypes.Sender{
			PublicKey: keyPair.Public(),
			AccountID: uint64(i) + int2Key,
			IsSigned:  true,
		}
	}

	defaultSender := intMaxTypes.NewDummySender()
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = defaultSender
	}

	txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
	if err != nil {
		return nil, errors.Join(ErrSetRandomTxRootFail, err)
	}

	senderPublicKeys := make([]byte, int128Key*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[int32Key*i:int32Key*(i+int1Key)], senderPublicKey[:])
	}
	for i := len(senders); i < int128Key; i++ {
		senderPublicKey := defaultSender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[int32Key*i:int32Key*(i+int1Key)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)
	aggregatedPublicKey := new(accounts.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			aggregatedPublicKey.Add(
				aggregatedPublicKey,
				sender.PublicKey.WeightByHash(publicKeysHash),
			)
		}
	}

	message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

	aggregatedSignature := new(bn254.G2Affine)
	for i, keyPair := range keyPairs {
		if senders[i].IsSigned {
			var signature *bn254.G2Affine
			signature, err = keyPair.WeightByHash(publicKeysHash).Sign(message)
			if err != nil {
				return nil, errors.Join(ErrSignKeyPairForWeightByHashFail, err)
			}
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	txRootBytes := [32]byte{}
	copy(txRootBytes[:], txRoot.Marshal())

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		txRootBytes,
		aggregatedSignature,
	)

	return intMaxTypes.MakePostRegistrationBlockInput(
		blockContent,
	)
}
