package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateRelationshipL2BatchIndexAndBlockContentID(
	batchIndex *uint256.Int,
	blockContentID string,
) (err error) {
	const (
		q = ` INSERT INTO relationship_l2_batch_index_block_contents
              (l2_batch_index, block_contents_id) VALUES ($1, $2)
              ON CONFLICT (l2_batch_index, block_contents_id)
              DO nothing `
	)

	bi, _ := batchIndex.Value()

	_, err = p.exec(p.ctx, q, bi, blockContentID)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}
