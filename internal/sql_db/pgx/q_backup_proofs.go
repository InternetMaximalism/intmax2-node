package pgx

import (
	"database/sql"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupSenderProof(
	lastBalanceProofBody, balanceTransitionProofBody []byte,
	enoughBalanceProofBodyHash string,
) (*mDBApp.BackupSenderProof, error) {
	const query = `
	    INSERT INTO backup_sender_proofs
		id, enough_balance_proof_body_hash, last_balance_proof_body, balance_transition_proof_body, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query,
		id, enoughBalanceProofBodyHash,
		lastBalanceProofBody, balanceTransitionProofBody, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupTransferProof(id)
}

func (p *pgx) GetBackupTransferProof(id string) (*mDBApp.BackupSenderProof, error) {
	const query = `
		SELECT id, enough_balance_proof_body_hash, last_balance_proof_body, balance_transition_proof_body, created_at
		FROM backup_proofs WHERE id = $1 `

	var b models.BackupSenderProof
	err := errPgx.Err(p.queryRow(p.ctx, query, id).
		Scan(
			&b.ID,
			&b.EnoughBalanceProofBodyHash,
			&b.LastBalanceProofBody,
			&b.BalanceTransitionProofBody,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}
	transfer := p.backupSenderProofToDBApp(&b)
	return &transfer, nil
}

func (p *pgx) GetBackupSenderProofsByHashes(enoughBalanceProofBodyHashes []string) ([]*mDBApp.BackupSenderProof, error) {
	const query = `
		SELECT id, enough_balance_proof_body_hash, last_balance_proof_body, balance_transition_proof_body, created_at
		FROM backup_proofs WHERE enough_balance_proof_body_hash = ANY($1)`

	var proofs []*mDBApp.BackupSenderProof
	err := p.getBackupEntries(query, enoughBalanceProofBodyHashes, func(rows *sql.Rows) error {
		var b models.BackupSenderProof
		err := rows.Scan(&b.ID, &b.EnoughBalanceProofBodyHash, &b.LastBalanceProofBody, &b.BalanceTransitionProofBody, &b.CreatedAt)
		if err != nil {
			return err
		}
		proof := p.backupSenderProofToDBApp(&b)
		proofs = append(proofs, &proof)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proofs, nil
}

func (p *pgx) backupSenderProofToDBApp(b *models.BackupSenderProof) mDBApp.BackupSenderProof {
	return mDBApp.BackupSenderProof{
		ID:                         b.ID,
		EnoughBalanceProofBodyHash: b.EnoughBalanceProofBodyHash,
		LastBalanceProofBody:       b.LastBalanceProofBody,
		BalanceTransitionProofBody: b.BalanceTransitionProofBody,
		CreatedAt:                  b.CreatedAt,
	}
}
