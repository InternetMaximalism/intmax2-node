package models

import "errors"

// ErrHexToECDSAFail error: failed to convert HEX to ECDSA.
var ErrHexToECDSAFail = errors.New("failed to convert HEX to ECDSA")
