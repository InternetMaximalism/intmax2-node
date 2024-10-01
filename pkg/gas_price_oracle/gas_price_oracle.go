package gas_price_oracle

import (
	"intmax2-node/configs"
	"intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/gas_price_oracle/scroll_eth"
	"intmax2-node/internal/logger"
)

func NewGasPriceOracle(
	cfg *configs.Config,
	log logger.Logger,
	gpo string,
	sb ServiceBlockchain,
) (GasPriceOracle, error) {
	if gpo == gas_price_oracle.ScrollEthGPO {
		return scroll_eth.New(cfg, log, sb), nil
	}

	return nil, ErrGPODriverNameInvalid
}
