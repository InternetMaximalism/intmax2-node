package pgx

import (
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	const (
		q = ` INSERT INTO event_block_numbers
              (id, event_name, last_processed_block_number ,created_at)
              VALUES ($1, $2, $3, $4)
              ON CONFLICT (event_name)
              DO UPDATE SET
			    last_processed_block_number = EXCLUDED.last_processed_block_number
              RETURNING *`
	)

	_, err := p.exec(p.ctx, q, id, eventName, blockNumber, now)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var eDBApp *mDBApp.EventBlockNumber
	eDBApp, err = p.EventBlockNumberByEventName(eventName)
	if err != nil {
		return nil, err
	}

	return eDBApp, nil
}

func (p *pgx) EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error) {
	placeholder := make([]string, len(eventNames))
	for i := range eventNames {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholderStr := strings.Join(placeholder, ", ")

	q := fmt.Sprintf(`
        SELECT event_name, last_processed_block_number
        FROM event_block_numbers
        WHERE event_name IN (%s)
    `, placeholderStr)

	args := make([]interface{}, len(eventNames))
	for i, name := range eventNames {
		args[i] = name
	}

	rows, err := p.query(p.ctx, q, args...)
	if err != nil {
		return nil, errPgx.Err(err)
	}
	defer rows.Close()

	var results []*mDBApp.EventBlockNumber
	for rows.Next() {
		var e models.EventBlockNumber
		err = rows.Scan(
			&e.EventName,
			&e.LastProcessedBlockNumber,
		)
		if err != nil {
			return nil, errPgx.Err(err)
		}
		eDBApp := p.eToDBApp(&e)
		results = append(results, &eDBApp)
	}

	if err = rows.Err(); err != nil {
		return nil, errPgx.Err(err)
	}

	return results, nil
}

func (p *pgx) EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error) {
	const q = `
	    SELECT event_name, last_processed_block_number
	    FROM event_block_numbers
	    WHERE event_name = $1
    `

	var e models.EventBlockNumber
	err := errPgx.Err(p.queryRow(p.ctx, q, eventName).
		Scan(
			&e.EventName,
			&e.LastProcessedBlockNumber,
		))
	if err != nil {
		return nil, err
	}

	eDBApp := p.eToDBApp(&e)
	return &eDBApp, nil
}

func (p *pgx) eToDBApp(e *models.EventBlockNumber) mDBApp.EventBlockNumber {
	return mDBApp.EventBlockNumber{
		EventName:                e.EventName,
		LastProcessedBlockNumber: e.LastProcessedBlockNumber,
	}
}
