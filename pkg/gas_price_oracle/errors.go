package gas_price_oracle

import "errors"

// ErrGPODriverNameInvalid error: the gas price oracle driver name must be valid.
var ErrGPODriverNameInvalid = errors.New("the gas price oracle driver name must be valid")

// ErrNewCtrlProcessingJobsFail error: failed to create new ctrl-processing-job row.
var ErrNewCtrlProcessingJobsFail = errors.New("failed to create new ctrl-processing-job row")

// ErrCtrlProcessingJobsFail error: failed to get ctrl-processing-job row.
var ErrCtrlProcessingJobsFail = errors.New("failed to get ctrl-processing-job row")

// ErrNewGasPriceOracleFail error: failed to create new the gas price oracle.
var ErrNewGasPriceOracleFail = errors.New("failed to create new the gas price oracle")

// ErrGasFeeFail error: failed to get gas fee.
var ErrGasFeeFail = errors.New("failed to get gas fee")

// ErrCreateGasPriceOracleFail error: failed to create the gas price oracle row.
var ErrCreateGasPriceOracleFail = errors.New("failed to create the gas price oracle row")

// ErrGasPriceOracleRowFail error: failed to get the gas price oracle row.
var ErrGasPriceOracleRowFail = errors.New("failed to get the gas price oracle row")

// ErrValueToBigIntFail error: failed to convert the value to big int.
var ErrValueToBigIntFail = errors.New("failed to convert the value to big int")

// ErrUpdValueFail error: failed to update value.
var ErrUpdValueFail = errors.New("failed to update value")
