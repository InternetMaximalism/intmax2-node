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
