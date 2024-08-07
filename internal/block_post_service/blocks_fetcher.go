package block_post_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

type errStartBlocksFetcher struct {
	BlockNumber *uint256.Int
	Options     []byte
	UpdErr      error
}

const fetcherChannelLen = 1024

var ErrChanStartBlocksFetcher = make(chan errStartBlocksFetcher, fetcherChannelLen)

func StartBlocksFetcher(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	dbApp SQLDriverApp,
	tickerEventWatcher *time.Ticker,
) (err error) {
	err = dbApp.CreateCtrlEventBlockNumbersJobs(mDBApp.BlockPostedEvent)
	if err != nil {
		return errors.Join(ErrNewCtrlEventBlockNumbersJobsFail, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case upd := <-ErrChanStartBlocksFetcher:
				_ = dbApp.Exec(ctx, nil, func(d interface{}, input interface{}) (err error) {
					q := d.(SQLDriverApp)

					err = q.UpsertEventBlockNumbersErrors(
						mDBApp.BlockPostedEvent, upd.BlockNumber, upd.Options, upd.UpdErr,
					)
					if err != nil {
						const msg = "block %q is ErrUpsertEventBlockNumbersErrors"
						lg.WithError(err).Errorf(msg, upd.BlockNumber.String())
					}

					return nil
				})
			}
		}
	}()

	f := func() {
		err = dbApp.Exec(ctx, nil, func(d interface{}, input interface{}) (err error) {
			q := d.(SQLDriverApp)

			_, err = q.CtrlEventBlockNumbersJobs(mDBApp.BlockPostedEvent)
			if err != nil && errors.Is(err, errorsDB.ErrNotFound) {
				return errors.Join(ErrCtrlEventBlockNumbersJobsFail, err)
			}
			if errors.Is(err, errorsDB.ErrNotFound) {
				return nil
			}

			err = ProcessingPostedBlocks(ctx, cfg, lg, q)
			if err != nil {
				return errors.Join(ErrProcessingBlocksFail, err)
			}

			return nil
		})
		if err != nil {
			const msg = "start of blocks fetcher error occurred"
			lg.WithError(err).Errorf(msg)
		}
	}
	f()

	if tickerEventWatcher == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			<-tickerEventWatcher.C
			f()
		}
	}
}

func ProcessingPostedBlocks(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	dbApp SQLDriverApp,
) (err error) {
	var bps BlockPostService
	bps, err = NewBlockPostService(ctx, cfg, lg)
	if err != nil {
		return errors.Join(ErrNewBlockPostServiceFail, err)
	}

	var (
		bn  uint64
		ebn *mDBApp.EventBlockNumber
	)

	ebn, err = dbApp.EventBlockNumberByEventName(mDBApp.BlockPostedEvent)
	switch {
	case err != nil:
		if !errors.Is(err, errorsDB.ErrNotFound) {
			const msg = "error fetching event block number: %w"
			return fmt.Errorf(msg, err)
		}

		fallthrough
	default:
		if ebn == nil {
			ebn = &mDBApp.EventBlockNumber{
				EventName:                mDBApp.BlockPostedEvent,
				LastProcessedBlockNumber: 0,
			}
		}
		bn = ebn.LastProcessedBlockNumber
	}

	if ebn.LastProcessedBlockNumber == 0 {
		bn = cfg.Blockchain.RollupContractDeployedBlockNumber

		err = dbApp.DelAllAccounts()
		if err != nil {
			return errors.Join(ErrDelAllAccountsFail, err)
		}

		err = dbApp.ResetSequenceByAccounts()
		if err != nil {
			return errors.Join(ErrResetSequenceByAccountsFail, err)
		}
	}

	var (
		events []*bindings.RollupBlockPosted
		nextBN *big.Int
	)

	events, nextBN, err = bps.FetchNewPostedBlocks(bn)
	if err != nil {
		return errors.Join(ErrFetchNewPostedBlocksFail, err)
	}

	if len(events) == 0 {
		return nil
	}

	ai := NewAccountInfo(dbApp)
	for key := range events {
		var blN uint256.Int
		_ = blN.SetFromBig(new(big.Int).SetUint64(events[key].Raw.BlockNumber))

		var cd []byte
		cd, err = bps.FetchScrollCalldataByHash(events[key].Raw.TxHash)
		if err != nil {
			return errors.Join(ErrFetchScrollCalldataByHashFail, err)
		}

		_, err = FetchIntMaxBlockContentByCalldata(cd, ai)
		if err != nil {
			err = errors.Join(ErrFetchIntMaxBlockContentByCalldataFail, err)
			switch {
			case errors.Is(err, ErrUnknownAccountID):
				const msg = "block %q is ErrUnknownAccountID"
				lg.WithError(err).Errorf(msg, blN.String())
			case errors.Is(err, ErrCannotDecodeAddress):
				const msg = "block %q is ErrCannotDecodeAddress"
				lg.WithError(err).Errorf(msg, blN.String())
			default:
				const msg = "block %q processing error occurred"
				lg.WithError(err).Errorf(msg, blN.String())
			}

			bytesEvent, errEvent := json.Marshal(events[key])
			if errEvent != nil {
				const msg = "block %q is ErrMarshalContent: %w"
				return fmt.Errorf(msg, blN.String(), errEvent)
			}

			ErrChanStartBlocksFetcher <- errStartBlocksFetcher{
				BlockNumber: &blN,
				Options:     bytesEvent,
				UpdErr:      err,
			}

			const msg = "processing of block %q error occurred"
			lg.Debugf(msg, blN.String())

			continue
		}

		const msg = "block %q is valid"
		lg.Debugf(msg, blN.String())
	}

	err = updateEventBlockNumber(dbApp, lg, mDBApp.BlockPostedEvent, nextBN.Uint64())
	if err != nil {
		const msg = "failed to update event block number: %v"
		return fmt.Errorf(msg, err.Error())
	}

	const msg = "next block number %q"
	lg.Debugf(msg, nextBN.String())

	return nil
}
