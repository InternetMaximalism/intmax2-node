package l2_batch_index

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/holiman/uint256"
	"github.com/tidwall/gjson"
)

type l2BatchIndex struct {
	cfg *configs.Config
	db  SQLDriverApp
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	db SQLDriverApp,
	sb ServiceBlockchain,
) L2IndexIndex {
	return &l2BatchIndex{
		cfg: cfg,
		db:  db,
		sb:  sb,
	}
}

func (l2bi *l2BatchIndex) Start(ctx context.Context) (err error) {
	var startL2BlockNumber, startL2BatchIndex bool
	tickerL2BlockNumber := time.NewTicker(time.Second)
	tickerL2BatchIndex := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			tickerL2BlockNumber.Stop()
			tickerL2BatchIndex.Stop()
			return nil
		case <-tickerL2BlockNumber.C:
			if startL2BlockNumber {
				continue
			}

			startL2BlockNumber = true
			err = l2bi.db.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
				defer func() {
					<-time.After(time.Second)
					startL2BlockNumber = false
				}()

				q, _ := d.(SQLDriverApp)

				var ctrlJob *mDBApp.CtrlProcessingJobs
				ctrlJob, err = q.CtrlProcessingJobsByMaskName(L2BlockNumberJobMask)
				if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
					return errors.Join(ErrCtrlProcessingJobsByMaskNameFail, err)
				} else if errors.Is(err, errorsDB.ErrNotFound) {
					return nil
				}

				const maskL2BlockNumber = "l2_block_number"
				var bn uint256.Int
				err = bn.Scan(gjson.GetBytes(ctrlJob.Options, maskL2BlockNumber).String())
				if err != nil {
					return errors.Join(ErrJSONUnmarshalFail, err)
				}

				err = l2bi.processingL2BlockNumber(ctx, q, &bn)
				switch {
				case err == nil:
				case errors.Is(err, ErrNeedRepeatAction):
					updAt := time.Now().UTC().Add(l2bi.cfg.L2BatchIndex.L2BlockNumberTimeout)
					err = q.UpdatedAtOfCtrlProcessingJobByName(ctrlJob.ProcessingJobName, updAt)
					if err != nil {
						return errors.Join(ErrUpdateAtOfCtrlProcessingJobByNameFail, err)
					}
					return nil
				default:
					return errors.Join(ErrProcessingL2BlockNumberFail, err)
				}

				err = q.DeleteCtrlProcessingJobByName(ctrlJob.ProcessingJobName)
				if err != nil {
					return errors.Join(ErrDeleteCtrlProcessingJobByNameFail, err)
				}

				return nil
			})
			if err != nil {
				return errors.Join(ErrApplyTickerL2BlockNumberWithDBAppFail, err)
			}
		case <-tickerL2BatchIndex.C:
			if startL2BatchIndex {
				continue
			}

			startL2BatchIndex = true
			err = l2bi.db.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
				defer func() {
					<-time.After(time.Second)
					startL2BatchIndex = false
				}()

				q, _ := d.(SQLDriverApp)

				var ctrlJob *mDBApp.CtrlProcessingJobs
				ctrlJob, err = q.CtrlProcessingJobsByMaskName(L2BatchIndexJobMask)
				if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
					return errors.Join(ErrCtrlProcessingJobsByMaskNameFail, err)
				} else if errors.Is(err, errorsDB.ErrNotFound) {
					return nil
				}

				const maskL2BatchIndex = "batch_index"
				var bi uint256.Int
				err = bi.Scan(gjson.GetBytes(ctrlJob.Options, maskL2BatchIndex).String())
				if err != nil {
					return errors.Join(ErrJSONUnmarshalFail, err)
				}

				err = l2bi.processingL2BatchIndex(ctx, q, &bi)
				switch {
				case err == nil:
				case errors.Is(err, ErrNeedRepeatAction):
					updAt := time.Now().UTC().Add(l2bi.cfg.L2BatchIndex.L2BatchIndexTimeout)
					err = q.UpdatedAtOfCtrlProcessingJobByName(ctrlJob.ProcessingJobName, updAt)
					if err != nil {
						return errors.Join(ErrUpdateAtOfCtrlProcessingJobByNameFail, err)
					}
					return nil
				default:
					return errors.Join(ErrProcessingL2BatchIndexFail, err)
				}

				err = q.DeleteCtrlProcessingJobByName(ctrlJob.ProcessingJobName)
				if err != nil {
					return errors.Join(ErrDeleteCtrlProcessingJobByNameFail, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}
	}
}

func (l2bi *l2BatchIndex) processingL2BlockNumber(
	ctx context.Context,
	db SQLDriverApp,
	bn *uint256.Int,
) (err error) {
	const (
		int0Key = 0
	)

	var bcID string
	bcID, err = db.BlockContentIDByL2BlockNumber(bn.String())
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		return errors.Join(ErrBlockContentIDByL2BlockNumberFail, err)
	} else if errors.Is(err, errorsDB.ErrNotFound) {
		return nil
	}

	var sLink string
	sLink, err = l2bi.sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(sLink)
	if err != nil {
		return errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	_, err = client.BlockByNumber(ctx, bn.ToBig())
	if err != nil {
		return errors.Join(ErrBlockByNumberWithClientOfRPCEthFail, err)
	}

	var rollupExplorerLink string
	rollupExplorerLink, err = l2bi.sb.ScrollNetworkChainLinkRollupExplorer(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkRollupExplorerFail, err)
	}

	const (
		maskApiUrl  = "%s/api/search?keyword=%s"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		maskApiUrl,
		rollupExplorerLink,
		url.QueryEscape(bn.String()),
	)

	r := resty.New().R()

	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get batch_index for block_number request from rollup explorer: %w"
		return fmt.Errorf(msg, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrNeedRepeatAction
	}

	const maskBatchIndexJSON = "batch_index"
	respBI := gjson.GetBytes(resp.Body(), maskBatchIndexJSON).Int()
	if respBI <= int0Key {
		return errors.Join(ErrNeedRepeatAction, err)
	}

	var bi uint256.Int
	_ = bi.SetUint64(uint64(respBI))

	err = db.CreateL2BatchIndex(&bi)
	if err != nil {
		return errors.Join(ErrCreateL2BatchIndexWithDBFail, err)
	}

	err = db.CreateRelationshipL2BatchIndexAndBlockContentID(&bi, bcID)
	if err != nil {
		return errors.Join(ErrCreateRelationshipL2BatchIndexAndBlockContentIDWithDBFail, err)
	}

	const (
		maskL2BatchIndexJob = "%s%s"
		l2BatchIndexJobJSON = `{"batch_index":%q}`
	)
	err = db.CreateCtrlProcessingJobs(
		fmt.Sprintf(maskL2BatchIndexJob, L2BatchIndexJobMask, bi.String()),
		json.RawMessage(fmt.Sprintf(l2BatchIndexJobJSON, bi.String())),
	)
	if err != nil {
		return errors.Join(ErrCreateCtrlProcessingJobsFail, err)
	}

	return nil
}

func (l2bi *l2BatchIndex) processingL2BatchIndex(
	ctx context.Context,
	db SQLDriverApp,
	bi *uint256.Int,
) (err error) {
	const (
		emptyKey = ""
	)

	_, err = db.L2BatchIndex(bi)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		return errors.Join(ErrL2BatchIndexFail, err)
	} else if errors.Is(err, errorsDB.ErrNotFound) {
		return nil
	}

	var rollupExplorerLink string
	rollupExplorerLink, err = l2bi.sb.ScrollNetworkChainLinkRollupExplorer(ctx)
	if err != nil {
		return errors.Join(ErrScrollNetworkChainLinkRollupExplorerFail, err)
	}

	const (
		maskApiUrl  = "%s/api/batch?index=%s"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf(
		maskApiUrl,
		rollupExplorerLink,
		url.QueryEscape(bi.String()),
	)

	r := resty.New().R()

	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get batch_index info request from rollup explorer: %w"
		return fmt.Errorf(msg, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrNeedRepeatAction
	}

	err = db.UpdOptionsOfBatchIndex(bi, resp.Body())
	if err != nil {
		return errors.Join(ErrUpdOptionsOfBatchIndexWithDBFail, err)
	}

	const maskBatchFinalizeTxHashJSON = "batch.finalize_tx_hash"
	respBatchFinalizeTxHash := strings.TrimSpace(gjson.GetBytes(resp.Body(), maskBatchFinalizeTxHashJSON).String())
	if strings.EqualFold(respBatchFinalizeTxHash, emptyKey) {
		return errors.Join(ErrNeedRepeatAction, err)
	}

	err = db.UpdL1VerifiedBatchTxHashOfBatchIndex(bi, respBatchFinalizeTxHash)
	if err != nil {
		return errors.Join(ErrUpdL1VerifiedBatchTxHashOfBatchIndexWithDBFail, err)
	}

	return nil
}
