package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func (p *pgx) CreateEthereumCounterparty(
	address string,
) (*mDBApp.EthereumCounterparty, error) {
	const (
		q = ` INSERT INTO ethereum_counterparties (address) VALUES ($1)
              ON CONFLICT (address) DO nothing `
	)

	_, err := p.exec(p.ctx, q, address)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var ethereumCounterpartyDBApp *mDBApp.EthereumCounterparty
	ethereumCounterpartyDBApp, err = p.EthereumCounterpartyByAddress(address)
	if err != nil {
		return nil, err
	}

	return ethereumCounterpartyDBApp, nil
}

func (p *pgx) EthereumCounterpartyByAddress(address string) (*mDBApp.EthereumCounterparty, error) {
	const (
		q = ` SELECT id ,address ,created_at
              FROM ethereum_counterparties
              WHERE address = $1 `
	)

	var ethereumCounterparty models.EthereumCounterparty
	err := errPgx.Err(p.queryRow(p.ctx, q, address).
		Scan(
			&ethereumCounterparty.ID,
			&ethereumCounterparty.Address,
			&ethereumCounterparty.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	ethereumCounterpartyDBApp := p.ethereumCounterpartyToDBApp(&ethereumCounterparty)

	return &ethereumCounterpartyDBApp, nil
}

func (p *pgx) ethereumCounterpartyToDBApp(sender *models.EthereumCounterparty) mDBApp.EthereumCounterparty {
	return mDBApp.EthereumCounterparty{
		ID:        sender.ID,
		Address:   sender.Address,
		CreatedAt: sender.CreatedAt,
	}
}
