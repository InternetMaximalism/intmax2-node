package get_backup_user_state

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	getBackupUserState "intmax2-node/internal/use_cases/get_backup_user_state"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupUserState.UseCaseGetBackupUserState {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *getBackupUserState.UCGetBackupUserStateInput,
) (*getBackupUserState.UCGetBackupUserState, error) {
	const (
		hName          = "UseCase GetBackupUserState"
		userStateIDKey = "user_state_id"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupUserStateInputEmpty)
		return nil, ErrUCGetBackupUserStateInputEmpty
	}

	span.SetAttributes(
		attribute.String(userStateIDKey, input.UserStateID),
	)

	us, err := service.GetBackupUserState(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, fmt.Errorf("failed to get backup user state: %w", err)
	}

	return us, nil
}
