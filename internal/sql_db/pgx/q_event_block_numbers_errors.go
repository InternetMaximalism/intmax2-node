package pgx

import (
	"encoding/json"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/holiman/uint256"
)

func (p *pgx) UpsertEventBlockNumbersErrors(
	eventName string,
	blockNumber *uint256.Int,
	options []byte,
	updErr error,
) error {
	const (
		q = ` INSERT INTO event_block_numbers_errors
              (event_name ,block_number ,options ,created_at ,updated_at)
              VALUES ($1 ,$2 ,$3 ,$4 ,$5)
              ON CONFLICT (event_name, block_number)
              DO UPDATE SET
              options = EXCLUDED.options
              ,updated_at = EXCLUDED.updated_at `

		qUpd = ` UPDATE event_block_numbers_errors SET 
              errors = jsonb_set(errors, '{%d}', '{"error":%q}'),
              updated_at = $1 WHERE event_name = $2 AND block_number = $3 `
	)

	tm := time.Now().UTC()
	bn, _ := blockNumber.Value()

	_, err := p.exec(p.ctx, q, eventName, bn, options, tm, tm)
	if err != nil {
		return errPgx.Err(err)
	}

	_, err = p.exec(p.ctx, fmt.Sprintf(qUpd,
		tm.Unix(),
		strings.ReplaceAll(strings.ReplaceAll(updErr.Error(), `'`, "`"), `"`, "`"),
	), tm, eventName, bn)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) EventBlockNumbersErrors(
	eventName string,
	blockNumber *uint256.Int,
) (*mDBApp.EventBlockNumbersErrors, error) {
	const (
		q = ` SELECT id ,event_name ,block_number ,options, errors, created_at, updated_at
              FROM event_block_numbers_errors
              WHERE event_name = $1 AND block_number = $2 `
	)

	bn, _ := blockNumber.Value()

	var eBnErrors models.EventBlockNumbersErrors
	err := errPgx.Err(p.queryRow(p.ctx, q, eventName, bn).
		Scan(
			&eBnErrors.ID,
			&eBnErrors.EventName,
			&eBnErrors.BlockNumber,
			&eBnErrors.Options,
			&eBnErrors.Errors,
			&eBnErrors.CreatedAt,
			&eBnErrors.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	eBnErrorsDBApp := p.eBnErrorsToDBApp(&eBnErrors)

	return &eBnErrorsDBApp, nil
}

func (p *pgx) eBnErrorsToDBApp(eBnErrors *models.EventBlockNumbersErrors) mDBApp.EventBlockNumbersErrors {
	m := mDBApp.EventBlockNumbersErrors{
		ID:          eBnErrors.ID,
		EventName:   eBnErrors.EventName,
		BlockNumber: &eBnErrors.BlockNumber,
		Options:     eBnErrors.Options,
		Errors:      eBnErrors.Errors,
		CreatedAt:   eBnErrors.CreatedAt,
		UpdatedAt:   eBnErrors.UpdatedAt,
	}

	if m.Options == nil {
		m.Options = json.RawMessage(`{}`)
	}

	if m.Errors == nil {
		m.Errors = json.RawMessage(`{}`)
	}

	return m
}
