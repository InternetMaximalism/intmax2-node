package pgx

import (
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateSpentProof(
	balanceProof []byte,
) (*mDBApp.SpentProof, error) {
	const query = ` INSERT INTO spent_proofs
	(id, balance_proof, created_at)
	VALUES ($1, $2, $3) `

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(
		query,
		id,
		balanceProof,
		createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create spent proof: %w", err)
	}

	return p.GetSpentProof(id)
}

func (p *pgx) GetSpentProof(id string) (*mDBApp.SpentProof, error) {
	const query = ` SELECT id, balance_proof, created_at
	FROM spent_proofs WHERE id = $1 `

	var b models.SpentProof
	err := errPgx.Err(p.queryRow(p.ctx, query, id).
		Scan(
			&b.ID,
			&b.SpentProof,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, fmt.Errorf("failed to get spent proof: %w", err)
	}
	spentProof := p.spentProofToDBApp(&b)
	return &spentProof, nil
}

func (p *pgx) spentProofToDBApp(
	b *models.SpentProof,
) mDBApp.SpentProof {
	return mDBApp.SpentProof{
		ID:         b.ID,
		SpentProof: b.SpentProof,
		CreatedAt:  b.CreatedAt,
	}
}
