package scroll_eth

import "errors"

// ErrNewGasPriceOracleScrollL1ETHFail error: failed to create new the gas price oracle of scroll L1 Eth.
var ErrNewGasPriceOracleScrollL1ETHFail = errors.New(
	"failed to create new the gas price oracle of scroll L1 Eth",
)

// ErrScalarFail error: failed to get scalar of the gas price oracle of the scroll network.
var ErrScalarFail = errors.New(
	"failed to get scalar of the gas price oracle of the scroll network",
)

// ErrOverheadFail error: failed to get overhead of the gas price oracle of the scroll network.
var ErrOverheadFail = errors.New(
	"failed to get overhead of the gas price oracle of the scroll network",
)

// ErrL1BaseFeeFail error: failed to get the L1 base fee of the gas price oracle of the scroll network.
var ErrL1BaseFeeFail = errors.New(
	"failed to get the L1 base fee of the gas price oracle of the scroll network",
)

// ErrL1GasFeeFail error: failed to get the L1 gas fee of the gas price oracle of the scroll network.
var ErrL1GasFeeFail = errors.New(
	"failed to get the L1 gas fee of the gas price oracle of the scroll network",
)

// ErrL2GasFeeFail error: failed to get the L2 gas fee of the gas price oracle of the scroll network.
var ErrL2GasFeeFail = errors.New(
	"failed to get the L2 gas fee of the gas price oracle of the scroll network",
)

// ErrGasFeeFail error: failed to get the gas fee.
var ErrGasFeeFail = errors.New("failed to get the gas fee")

// ErrFeeHistoryFail error: failed to get the gas fee history.
var ErrFeeHistoryFail = errors.New("failed to get the gas fee history")

// ErrL2SuggestGasTipCapFail error: failed to get the L2 SuggestGasTipCap of the gas price oracle of the scroll network.
var ErrL2SuggestGasTipCapFail = errors.New(
	"failed to get the L2 SuggestGasTipCap of the gas price oracle of the scroll network",
)

// ErrRandIntFail error: failed to get rand integer.
var ErrRandIntFail = errors.New("failed to get rand integer")

// ErrNewPrivateKeyWithReCalcPubKeyIfPkNegatesFail error: failed to create new private key with re-calc pubKey if pk negates.
var ErrNewPrivateKeyWithReCalcPubKeyIfPkNegatesFail = errors.New(
	"failed to create new private key with re-calc pubKey if pk negates",
)

// ErrSetRandomTxRootFail error: failed to set random tx root.
var ErrSetRandomTxRootFail = errors.New("failed to set random tx root")

// ErrSignKeyPairForWeightByHashFail error: failed to sign key pair for weight by hash.
var ErrSignKeyPairForWeightByHashFail = errors.New(
	"failed to sign key pair for weight by hash",
)

// ErrNewRollupFail error: failed to instantiate a Rollup contract.
var ErrNewRollupFail = errors.New("failed to instantiate a Rollup contract")

// ErrLoadPkBlockBuilderFail error: failed to load private key of block builder.
var ErrLoadPkBlockBuilderFail = errors.New("failed to load private key of block builder")

// ErrParseScrollNetworkChainIDFail error: failed to parse the scroll network chain ID.
var ErrParseScrollNetworkChainIDFail = errors.New(
	"failed to parse the scroll network chain ID",
)

// ErrNewKeyedTransactorWithChainIDFail error: failed to create new keyed transactor with chain ID.
var ErrNewKeyedTransactorWithChainIDFail = errors.New(
	"failed to create new keyed transactor with chain ID",
)

// ErrBlockContentFail error: failed to get block content.
var ErrBlockContentFail = errors.New("failed to get block content")

// ErrPostRegistrationBlockFail error: failed to post registration block.
var ErrPostRegistrationBlockFail = errors.New("failed to post registration block")

// ErrGasValueForBlockContentWithRollupAndScrollNetworkFail error: failed to get the gas value for block content with rollup and scroll network.
var ErrGasValueForBlockContentWithRollupAndScrollNetworkFail = errors.New(
	"failed to get the gas value for block content with rollup and scroll network",
)

// ErrOracleL1GasFeeScrollFail error: failed to get the L1 gas fee from the scroll oracle.
var ErrOracleL1GasFeeScrollFail = errors.New("failed to get the L1 gas fee from the scroll oracle")
