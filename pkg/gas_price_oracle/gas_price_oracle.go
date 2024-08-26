package gas_price_oracle

import (
	"intmax2-node/configs"
	"intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/gas_price_oracle/scroll_eth"
)

func NewGasPriceOracle(
	cfg *configs.Config,
	gpo string,
	sb ServiceBlockchain,
) (GasPriceOracle, error) {
	if gpo == gas_price_oracle.ScrollEthGPO {
		return scroll_eth.New(cfg, sb), nil
	}

	return nil, ErrGPODriverNameInvalid
}
