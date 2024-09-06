package models

import (
	"time"

	"github.com/holiman/uint256"
)

type GasPriceOracle struct {
	GasPriceOracleName string
	Value              *uint256.Int
	CreatedAt          time.Time
}
