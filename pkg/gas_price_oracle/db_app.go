package gas_price_oracle

import (
	"context"
	"encoding/json"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=gas_price_oracle_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	CtrlProcessingJobs
	GasPriceOracleApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
	CtrlProcessingJobs(name string) (*mDBApp.CtrlProcessingJobs, error)
}

type GasPriceOracleApp interface {
	CreateGasPriceOracle(name string, value *uint256.Int) error
	GasPriceOracle(name string) (*mDBApp.GasPriceOracle, error)
}
