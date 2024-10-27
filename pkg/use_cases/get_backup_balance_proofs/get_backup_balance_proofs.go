package get_backup_balance_proofs

import (
	"context"
	"encoding/base64"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_balance_proof"
	"intmax2-node/pkg/sql_db/db_app/models"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_balance_proof.UseCaseGetBackupBalanceProofs {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance_proof.UCGetBackupBalanceProofsInput,
) (*node.GetBackupBalanceProofsResponse_Data, error) {
	const (
		hName          = "UseCase GetBackupBalances"
		userKey        = "user"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupBalanceProofsInputEmpty)
		return nil, ErrUCGetBackupBalanceProofsInputEmpty
	}

	proofs, err := service.GetBackupSenderProofs(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		var ErrGetBackupBalances = errors.New("failed to get backup balances")
		return nil, errors.Join(ErrGetBackupBalances, err)
	}

	data := node.GetBackupBalanceProofsResponse_Data{
		Proofs: generateBackupProofs(proofs),
	}

	return &data, nil
}

func generateBackupProofs(proofs []*models.BackupSenderProof) []*node.GetBackupBalanceProofsResponse_Proof {
	results := make([]*node.GetBackupBalanceProofsResponse_Proof, 0, len(proofs))
	for _, proof := range proofs {
		lastBalanceProofBody := base64.StdEncoding.EncodeToString(proof.LastBalanceProofBody)
		balanceTransitionProofBody := base64.StdEncoding.EncodeToString(proof.BalanceTransitionProofBody)
		backupBalance := &node.GetBackupBalanceProofsResponse_Proof{
			Id:                         proof.ID,
			EnoughBalanceProofBodyHash: proof.EnoughBalanceProofBodyHash,
			LastBalanceProofBody:       lastBalanceProofBody,
			BalanceTransitionProofBody: balanceTransitionProofBody,
		}

		results = append(results, backupBalance)
	}
	return results
}
