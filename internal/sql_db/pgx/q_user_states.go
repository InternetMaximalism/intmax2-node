package pgx

import (
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupUserState(
	userAddress string,
	encryptedUserState []byte,
	authSignature string,
) (*mDBApp.UserState, error) {
	const query = ` INSERT INTO user_states
	(id, user_address, encrypted_user_state, auth_signature, created_at, modified_at)
	VALUES ($1, $2, $3, $4, $5, $6) `

	id := uuid.New().String()
	createdAt := time.Now().UTC()
	modifiedAt := time.Now().UTC()

	err := p.createBackupEntry(
		query,
		id,
		userAddress,
		encryptedUserState,
		authSignature,
		createdAt,
		modifiedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create spent proof: %w", err)
	}

	return p.GetBackupUserState(id)
}

func (p *pgx) UpdateBackupUserState(
	id string,
	encryptedUserState []byte,
	authSignature string,
) (*mDBApp.UserState, error) {
	const query = ` UPDATE user_states
	SET encrypted_user_state = $1, auth_signature = $2, modified_at = $3
	WHERE id = $4 `

	modifiedAt := time.Now().UTC()

	_, err := p.exec(p.ctx, query,
		encryptedUserState,
		authSignature,
		modifiedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update spent proof: %w", errPgx.Err(err))
	}
	return p.GetBackupUserState(id)
}

func (p *pgx) GetBackupUserState(id string) (*mDBApp.UserState, error) {
	const query = ` SELECT id, user_address, encrypted_user_state, auth_signature, created_at, modified_at
	FROM user_states WHERE id = $1 `

	var b models.UserState
	err := errPgx.Err(p.queryRow(p.ctx, query, id).
		Scan(
			&b.ID,
			&b.UserAddress,
			&b.EncryptedUserState,
			&b.AuthSignature,
			&b.CreatedAt,
			&b.ModifiedAt,
		))
	if err != nil {
		return nil, fmt.Errorf("failed to get spent proof: %w", err)
	}
	spentProof := p.userStateToDBApp(&b)
	return &spentProof, nil
}

func (p *pgx) userStateToDBApp(
	b *models.UserState,
) mDBApp.UserState {
	return mDBApp.UserState{
		ID:                 b.ID,
		UserAddress:        b.UserAddress,
		EncryptedUserState: b.EncryptedUserState,
		AuthSignature:      b.AuthSignature,
		CreatedAt:          b.CreatedAt,
		ModifiedAt:         b.ModifiedAt,
	}
}
