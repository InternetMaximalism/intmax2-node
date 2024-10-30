package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	getBackupUserState "intmax2-node/internal/use_cases/get_backup_user_state"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupUserState(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *getBackupUserState.UCGetBackupUserStateInput,
) (*getBackupUserState.UCGetBackupUserState, error) {
	us, err := db.GetBackupUserState(input.UserStateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup user state from db: %w", err)
	}

	var bpDB *mDBApp.BalanceProof
	bpDB, err = db.GetBalanceProofByUserStateID(us.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance proof to db: %w", err)
	}

	var bp intMaxTypes.Plonky2Proof
	err = bp.UnmarshalJSON(bpDB.BalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Plonky2Proof with BalanceProof: %w", err)
	}

	return &getBackupUserState.UCGetBackupUserState{
		ID:                 us.ID,
		UserAddress:        us.UserAddress,
		BalanceProof:       bp.ProofBase64String(),
		EncryptedUserState: us.EncryptedUserState,
		AuthSignature:      us.AuthSignature,
		BlockNumber:        us.BlockNumber,
		CreatedAt:          us.CreatedAt,
	}, nil
}
