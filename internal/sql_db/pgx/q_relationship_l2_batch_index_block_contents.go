package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

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

func (p *pgx) RelationshipL2BatchIndexAndBlockContentsByBlockContentID(
	blockContentID string,
) (*mDBApp.RelationshipL2BatchIndexBlockContents, error) {
	const (
		q = ` SELECT l2_batch_index, block_contents_id, created_at
              FROM relationship_l2_batch_index_block_contents
              WHERE block_contents_id = $1 `
	)

	var tmp models.RelationshipL2BatchIndexBlockContents
	err := errPgx.Err(p.queryRow(p.ctx, q, blockContentID).
		Scan(
			&tmp.L2BatchIndex,
			&tmp.BlockContentsID,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	relDBApp := p.relationshipL2BatchIndexAndBlockContentsToDBApp(&tmp)

	return relDBApp, nil
}

func (p *pgx) relationshipL2BatchIndexAndBlockContentsToDBApp(
	tmp *models.RelationshipL2BatchIndexBlockContents,
) *mDBApp.RelationshipL2BatchIndexBlockContents {
	m := mDBApp.RelationshipL2BatchIndexBlockContents{
		L2BatchIndex:    tmp.L2BatchIndex,
		BlockContentsID: tmp.BlockContentsID,
		CreatedAt:       tmp.CreatedAt,
	}

	return &m
}
