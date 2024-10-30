package block_validity_prover_block_status_by_block_hash

import (
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_block_status_by_block_hash_test -source=db_app.go

type SQLDriverApp interface {
	BlockContents
	RelationshipL2BatchIndexAndBlockContent
	L2BatchIndex
}

type BlockContents interface {
	BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error)
}

type RelationshipL2BatchIndexAndBlockContent interface {
	RelationshipL2BatchIndexAndBlockContentsByBlockContentID(
		blockContentID string,
	) (*mDBApp.RelationshipL2BatchIndexBlockContents, error)
}

type L2BatchIndex interface {
	L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error)
}
