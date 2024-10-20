package pgx

import (
	"encoding/json"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateL2BatchIndex(batchIndex *uint256.Int) (err error) {
	const (
		q = ` INSERT INTO l2_batch_index
              (l2_batch_index) VALUES ($1)
              ON CONFLICT (l2_batch_index)
              DO nothing `
	)

	bi, _ := batchIndex.Value()

	_, err = p.exec(p.ctx, q, bi)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error) {
	const (
		q = ` SELECT l2_batch_index, options, l1_verified_batch_tx_hash, created_at
              FROM l2_batch_index
              WHERE l2_batch_index = $1 `
	)

	biV, _ := batchIndex.Value()

	var bi models.L2BatchIndex
	err := errPgx.Err(p.queryRow(p.ctx, q, biV).
		Scan(
			&bi.L2BatchIndex,
			&bi.Options,
			&bi.L1VerifiedBatchTxHash,
			&bi.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	biDBApp := p.l2BatchIndexToDBApp(&bi)

	return &biDBApp, nil
}

func (p *pgx) UpdOptionsOfBatchIndex(batchIndex *uint256.Int, options json.RawMessage) (err error) {
	if options == nil {
		return nil
	}

	const (
		q = ` UPDATE l2_batch_index SET options = $1 WHERE l2_batch_index = $2 `
	)

	options = json.RawMessage(strings.ReplaceAll(string(options), findU0, replaceU0))

	bi, _ := batchIndex.Value()

	_, err = p.exec(p.ctx, q, options, bi)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgx) UpdL1VerifiedBatchTxHashOfBatchIndex(batchIndex *uint256.Int, hash string) (err error) {
	const (
		q = ` UPDATE l2_batch_index SET l1_verified_batch_tx_hash = $1 WHERE l2_batch_index = $2 `
	)

	bi, _ := batchIndex.Value()

	_, err = p.exec(p.ctx, q, hash, bi)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgx) l2BatchIndexToDBApp(bi *models.L2BatchIndex) mDBApp.L2BatchIndex {
	m := mDBApp.L2BatchIndex{
		L2BatchIndex: bi.L2BatchIndex,
		Options:      bi.Options,
		CreatedAt:    bi.CreatedAt,
	}

	if m.Options == nil {
		m.Options = json.RawMessage(`{}`)
	}
	if len(m.Options) > 0 {
		m.Options = json.RawMessage(strings.ReplaceAll(string(m.Options), replaceU0, findU0))
	}

	const emptyKey = ""
	l1VerifiedBatchTxHash := strings.TrimSpace(bi.L1VerifiedBatchTxHash.String)
	if !strings.EqualFold(l1VerifiedBatchTxHash, emptyKey) {
		m.L1VerifiedBatchTxHash = l1VerifiedBatchTxHash
	}

	return m
}
