package configs

import "time"

const gasPriceOracleDelimiterDef = 10

type GasPriceOracle struct {
	ExtraFee  int           `env:"GAS_PRICE_ORACLE_EXTRA_FEE" envDefault:"0"`
	Delimiter int           `env:"GAS_PRICE_ORACLE_DELIMITER" envDefault:"10"`
	Timeout   time.Duration `env:"GAS_PRICE_ORACLE_TIMEOUT,required" envDefault:"30s"`
}
