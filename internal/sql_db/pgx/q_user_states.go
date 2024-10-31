package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupUserState(
	userAddress, encryptedUserState, authSignature string,
	blockNumber int64,
) (*mDBApp.UserState, error) {
	const (
		q = ` INSERT INTO user_states (
              id ,user_address ,encrypted_user_state ,auth_signature ,block_number
              ) VALUES ($1, $2, $3, $4, $5) `
	)

	id := uuid.New().String()

	err := p.createBackupEntry(q,
		id,
		userAddress,
		encryptedUserState,
		authSignature,
		blockNumber,
	)
	if err != nil {
		return nil, err
	}

	return p.GetBackupUserState(id)
}

func (p *pgx) UpdateBackupUserState(
	id, encryptedUserState, authSignature string,
	blockNumber int64,
) (*mDBApp.UserState, error) {
	const (
		q = ` UPDATE user_states
              SET encrypted_user_state = $1 ,auth_signature = $2 ,block_number = $3 ,updated_at = $4
              WHERE id = $5 `
	)

	_, err := p.exec(p.ctx, q,
		encryptedUserState,
		authSignature,
		blockNumber,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	return p.GetBackupUserState(id)
}

func (p *pgx) GetBackupUserState(id string) (*mDBApp.UserState, error) {
	const (
		q = ` SELECT id
              ,user_address ,encrypted_user_state ,auth_signature ,block_number ,created_at ,updated_at
              FROM user_states WHERE id = $1 `
	)

	var b models.UserState
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&b.ID,
			&b.UserAddress,
			&b.EncryptedUserState,
			&b.AuthSignature,
			&b.BlockNumber,
			&b.CreatedAt,
			&b.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	userState := p.userStateToDBApp(&b)
	return &userState, nil
}

func (p *pgx) userStateToDBApp(b *models.UserState) mDBApp.UserState {
	return mDBApp.UserState{
		ID:                 b.ID,
		UserAddress:        b.UserAddress,
		EncryptedUserState: b.EncryptedUserState,
		AuthSignature:      b.AuthSignature,
		BlockNumber:        b.BlockNumber,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}
