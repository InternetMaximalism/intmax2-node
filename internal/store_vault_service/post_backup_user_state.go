package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	bps "intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	backupUserState "intmax2-node/internal/use_cases/post_backup_user_state"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func PostBackupUserState(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupUserState.UCPostBackupUserStateInput,
) (*backupUserState.UCPostBackupUserState, error) {
	bp, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(input.BalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to get compressed Plonky2Proof from Base64String: %w", err)
	}

	var bpi *bps.BalancePublicInputs
	bpi, err = new(bps.BalancePublicInputs).FromPublicInputs(bp.PublicInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to get BalancePublicInputs from PublicInputs: %w", err)
	}

	var bytesBP []byte
	bytesBP, err = bp.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Plonky2Proof with BalanceProof: %w", err)
	}

	var us *mDBApp.UserState
	us, err = db.CreateBackupUserState(
		input.UserAddress, input.EncryptedUserState,
		input.AuthSignature, input.BlockNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup user state to db: %w", err)
	}

	var bpDB *mDBApp.BalanceProof
	bpDB, err = db.CreateBalanceProof(
		us.ID, input.UserAddress, bpi.PrivateCommitment.String(), input.BlockNumber, bytesBP,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance proof to db: %w", err)
	}

	err = bp.UnmarshalJSON(bpDB.BalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Plonky2Proof with BalanceProof: %w", err)
	}

	return &backupUserState.UCPostBackupUserState{
		ID:                 us.ID,
		UserAddress:        us.UserAddress,
		BalanceProof:       bp.ProofBase64String(),
		EncryptedUserState: us.EncryptedUserState,
		AuthSignature:      us.AuthSignature,
		BlockNumber:        us.BlockNumber,
		CreatedAt:          us.CreatedAt,
	}, nil
}
