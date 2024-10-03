package gas_price_oracle

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"

	"github.com/holiman/uint256"
)

type storage struct {
	cfg   *configs.Config
	log   logger.Logger
	dbApp SQLDriverApp
	sb    ServiceBlockchain
}

func NewStoreGPO(
	cfg *configs.Config,
	log logger.Logger,
	dbApp SQLDriverApp,
	sb ServiceBlockchain,
) Storage {
	return &storage{
		cfg:   cfg,
		log:   log,
		dbApp: dbApp,
		sb:    sb,
	}
}

func (s *storage) Init(ctx context.Context) (err error) {
	err = s.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
		q := d.(SQLDriverApp)

		err = q.CreateCtrlProcessingJobs(gas_price_oracle.ScrollEthGPO)
		if err != nil {
			return errors.Join(ErrNewCtrlProcessingJobsFail, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) Value(ctx context.Context, name string) (*big.Int, error) {
	var v big.Int
	err := s.dbApp.Exec(ctx, &v, func(d interface{}, in interface{}) (err error) {
		q := d.(SQLDriverApp)

		var vDBApp *mDBApp.GasPriceOracle
		vDBApp, err = q.GasPriceOracle(name)
		if err != nil {
			return errors.Join(ErrGasPriceOracleRowFail, err)
		}
		vIn, ok := in.(*big.Int)
		if !ok {
			return ErrValueToBigIntFail
		}
		*vIn = *vDBApp.Value.ToBig()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *storage) UpdValue(ctx context.Context, name string) (err error) {
	err = s.dbApp.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
		q := d.(SQLDriverApp)

		_, err = q.CtrlProcessingJobs(name)
		if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
			return errors.Join(ErrCtrlProcessingJobsFail, err)
		}
		if errors.Is(err, errorsDB.ErrNotFound) {
			return nil
		}

		var oracle GasPriceOracle
		oracle, err = NewGasPriceOracle(s.cfg, s.log, name, s.sb)
		if err != nil {
			return errors.Join(ErrNewGasPriceOracleFail, err)
		}

		var gasFee *big.Int
		gasFee, err = oracle.GasFee(ctx)
		if err != nil {
			return errors.Join(ErrGasFeeFail, err)
		}

		var gf uint256.Int
		_ = gf.SetFromBig(gasFee)

		err = q.CreateGasPriceOracle(name, &gf)
		if err != nil {
			return errors.Join(ErrCreateGasPriceOracleFail, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) UpdValues(ctx context.Context, name ...string) (err error) {
	for key := range name {
		err = s.UpdValue(ctx, name[key])
		if err != nil {
			return errors.Join(ErrUpdValueFail, err)
		}
	}

	return nil
}
