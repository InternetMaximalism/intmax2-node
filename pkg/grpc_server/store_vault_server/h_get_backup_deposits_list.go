package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	mFL "intmax2-node/internal/sql_filter/models"
	getBackupDepositsList "intmax2-node/internal/use_cases/get_backup_deposits_list"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StoreVaultServer) GetBackupDepositsList(
	ctx context.Context,
	req *node.GetBackupDepositsListRequest,
) (*node.GetBackupDepositsListResponse, error) {
	resp := node.GetBackupDepositsListResponse{}

	const (
		hName      = "Handler GetBackupDepositsList"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := getBackupDepositsList.UCGetBackupDepositsListInput{
		Recipient: req.Recipient,
		OrderBy:   mFL.OrderBy(req.OrderBy),
		Sorting:   mFL.Sorting(req.Sorting),
	}

	input.Filters = make([]*mFL.Filter, len(req.Filter))
	for i := range req.Filter {
		input.Filters[i] = &mFL.Filter{
			Relation:  mFL.Relation(req.Filter[i].Relation),
			DataField: mFL.DataField(req.Filter[i].DataField),
			Condition: mFL.Condition(req.Filter[i].Condition),
			Value:     req.Filter[i].Value,
		}
	}

	if req.Pagination != nil {
		input.Pagination = &getBackupDepositsList.UCGetBackupDepositsListPaginationInput{
			Direction: mFL.Direction(req.Pagination.Direction),
			PerPage:   req.Pagination.PerPage,
		}

		if req.Pagination.Cursor != nil {
			input.Pagination.Cursor = &getBackupDepositsList.UCGetBackupDepositsListCursorBase{
				BlockNumber:  req.Pagination.Cursor.BlockNumber,
				SortingValue: req.Pagination.Cursor.SortingValue,
			}
		}
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var list getBackupDepositsList.UCGetBackupDepositsList
	err = s.dbApp.Exec(spanCtx, &list, func(d interface{}, in interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		var results *getBackupDepositsList.UCGetBackupDepositsList
		results, err = s.commands.GetBackupDepositsList(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get backup deposits: %w"
			return fmt.Errorf(msg, err)
		}

		if v, ok := in.(*getBackupDepositsList.UCGetBackupDepositsList); ok {
			v.List = results.List
			v.Pagination = results.Pagination
		} else {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to convert of list of backup deposits"
			return fmt.Errorf(msg)
		}

		return nil
	})
	if err != nil {
		const msg = "failed to get backup deposits with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.GetBackupDepositsListResponse_Data{
		Pagination: &node.GetBackupDepositsListResponse_Pagination{
			PerPage: list.Pagination.PerPage,
		},
	}

	if list.Pagination.Cursor != nil {
		resp.Data.Pagination.Cursor = &node.GetBackupDepositsListResponse_Cursor{}
		if list.Pagination.Cursor.Prev != nil {
			resp.Data.Pagination.Cursor.Prev = &node.GetBackupDepositsListResponse_CursorBase{
				BlockNumber:  list.Pagination.Cursor.Prev.BlockNumber,
				SortingValue: list.Pagination.Cursor.Prev.SortingValue,
			}
		}
		if list.Pagination.Cursor.Next != nil {
			resp.Data.Pagination.Cursor.Next = &node.GetBackupDepositsListResponse_CursorBase{
				BlockNumber:  list.Pagination.Cursor.Next.BlockNumber,
				SortingValue: list.Pagination.Cursor.Next.SortingValue,
			}
		}
	} else if input.Pagination != nil && input.Pagination.Cursor != nil {
		resp.Data.Pagination.Cursor = &node.GetBackupDepositsListResponse_Cursor{
			Prev: &node.GetBackupDepositsListResponse_CursorBase{
				BlockNumber:  input.Pagination.Cursor.BlockNumber,
				SortingValue: input.Pagination.Cursor.SortingValue,
			},
			Next: &node.GetBackupDepositsListResponse_CursorBase{
				BlockNumber:  input.Pagination.Cursor.BlockNumber,
				SortingValue: input.Pagination.Cursor.SortingValue,
			},
		}
	}

	resp.Data.Deposits = make([]*node.GetBackupDepositsListResponse_Deposit, len(list.List))
	for key := range list.List {
		resp.Data.Deposits[key] = &node.GetBackupDepositsListResponse_Deposit{
			Id:               list.List[key].ID,
			Recipient:        list.List[key].Recipient,
			BlockNumber:      uint64(list.List[key].BlockNumber),
			EncryptedDeposit: list.List[key].EncryptedDeposit,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: list.List[key].CreatedAt.Unix(),
				Nanos:   int32(list.List[key].CreatedAt.Nanosecond()),
			},
		}
	}

	return &resp, utils.OK(spanCtx)
}
