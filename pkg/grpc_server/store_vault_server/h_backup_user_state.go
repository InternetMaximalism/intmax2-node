package store_vault_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	postBackupUserState "intmax2-node/internal/use_cases/post_backup_user_state"
	"intmax2-node/pkg/grpc_server/utils"
	errorsDB "intmax2-node/pkg/sql_db/errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StoreVaultServer) BackupUserState(
	ctx context.Context,
	req *node.BackupUserStateRequest,
) (*node.BackupUserStateResponse, error) {
	resp := node.BackupUserStateResponse{
		Data: &node.BackupUserStateResponse_Data{},
	}

	const (
		hName      = "Handler BackupUserState"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := postBackupUserState.UCPostBackupUserStateInput{
		UserAddress:        req.UserAddress,
		BalanceProof:       req.BalanceProof,
		EncryptedUserState: req.EncryptedUserState,
		AuthSignature:      req.AuthSignature,
		BlockNumber:        int64(req.BlockNumber),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var us postBackupUserState.UCPostBackupUserState
	err = s.dbApp.Exec(spanCtx, &us, func(d interface{}, in interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		var result *postBackupUserState.UCPostBackupUserState
		result, err = s.commands.PostBackupUserState(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to create backup user state: %w"
			return fmt.Errorf(msg, err)
		}

		if v, ok := in.(*postBackupUserState.UCPostBackupUserState); ok {
			v.ID = result.ID
			v.UserAddress = result.UserAddress
			v.BalanceProof = result.BalanceProof
			v.EncryptedUserState = result.EncryptedUserState
			v.AuthSignature = result.AuthSignature
			v.BlockNumber = result.BlockNumber
			v.CreatedAt = result.CreatedAt
		} else {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to convert of backup user state"
			return fmt.Errorf(msg)
		}

		return nil
	})
	if err != nil && !errors.Is(err, errorsDB.ErrNotUnique) {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to create backup user state with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Data.Message = postBackupUserState.AlreadyExistsMsg
	if err == nil {
		resp.Success = true
		resp.Data.Message = postBackupUserState.SuccessMsg
		resp.Data.Balance = &node.BackupUserStateResponse_Data_Balance{
			UserStateId:        us.ID,
			UserAddress:        us.UserAddress,
			BalanceProof:       us.BalanceProof,
			EncryptedUserState: us.EncryptedUserState,
			BlockNumber:        uint32(us.BlockNumber),
			AuthSignature:      us.AuthSignature,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: us.CreatedAt.Unix(),
				Nanos:   int32(us.CreatedAt.Nanosecond()),
			},
		}
	}

	return &resp, utils.OK(spanCtx)
}
