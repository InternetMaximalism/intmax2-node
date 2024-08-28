package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateGasPriceOracle(name string, value *uint256.Int) error {
	const (
		q = ` INSERT INTO gas_price_oracle
              (gas_price_oracle_name, value, created_at) VALUES ($1, $2, NOW())
              ON CONFLICT (gas_price_oracle_name)
              DO UPDATE SET value = EXCLUDED.value, created_at = EXCLUDED.created_at`
	)

	v, _ := value.Value()

	_, err := p.exec(p.ctx, q, name, v)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) GasPriceOracle(name string) (*mDBApp.GasPriceOracle, error) {
	const (
		q = ` SELECT gas_price_oracle_name, value, created_at FROM gas_price_oracle WHERE gas_price_oracle_name = $1 `
	)

	var o models.GasPriceOracle
	err := errPgx.Err(p.queryRow(p.ctx, q, name).
		Scan(
			&o.GasPriceOracleName,
			&o.Value,
			&o.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	oDBApp := p.gasPriceOracleToDBApp(&o)

	return &oDBApp, nil
}

func (p *pgx) gasPriceOracleToDBApp(o *models.GasPriceOracle) mDBApp.GasPriceOracle {
	return mDBApp.GasPriceOracle{
		GasPriceOracleName: o.GasPriceOracleName,
		Value:              o.Value,
		CreatedAt:          o.CreatedAt,
	}
}
