package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateSignature(signature, proposalBlockID string) (*mDBApp.Signature, error) {
	s := models.Signature{
		SignatureID:     uuid.New().String(),
		Signature:       signature,
		ProposalBlockID: proposalBlockID,
		CreatedAt:       time.Now().UTC(),
	}

	const (
		q = ` INSERT INTO signatures
              (signature_id ,signature ,proposal_block_id ,created_at)
              VALUES ($1, $2, $3, $4) `
	)

	_, err := p.exec(p.ctx, q,
		s.SignatureID, s.Signature, s.ProposalBlockID, s.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var sDBApp *mDBApp.Signature
	sDBApp, err = p.SignatureByID(s.SignatureID)
	if err != nil {
		return nil, err
	}

	return sDBApp, nil
}

func (p *pgx) SignatureByID(signatureID string) (*mDBApp.Signature, error) {
	const (
		q = `SELECT signature_id, signature, proposal_block_id, created_at
             FROM signatures WHERE signature_id = $1`
	)

	var tmp models.Signature
	err := errPgx.Err(p.queryRow(p.ctx, q, signatureID).
		Scan(
			&tmp.SignatureID,
			&tmp.Signature,
			&tmp.ProposalBlockID,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	sDBApp := p.sToDBApp(&tmp)

	return &sDBApp, nil
}

func (p *pgx) sToDBApp(s *models.Signature) mDBApp.Signature {
	m := mDBApp.Signature{
		SignatureID:     s.SignatureID,
		Signature:       s.Signature,
		ProposalBlockID: s.ProposalBlockID,
		CreatedAt:       s.CreatedAt,
	}

	return m
}
