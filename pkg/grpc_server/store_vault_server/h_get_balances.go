package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/grpc_server/utils"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const int10Key = 10

func (s *StoreVaultServer) GetBalances(ctx context.Context, req *node.GetBalancesRequest) (*node.GetBalancesResponse, error) {
	resp := node.GetBalancesResponse{}

	const (
		hName      = "Handler GetBalances"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupBalance.UCGetBalancesInput{
		Address: req.Address,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBalances(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get balances request: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Deposits = convertToDeposits(results.Deposits)
		resp.Transfers = convertToTransfers(results.Transfers)
		resp.Transactions = convertToTransactions(results.Transactions)
		return nil
	})
	if err != nil {
		const msg = "failed to get balances with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	return &resp, utils.OK(spanCtx)
}

func convertToDeposits(deposits []*backupBalance.BackupDeposit) []*node.BackupDeposit {
	result := make([]*node.BackupDeposit, len(deposits))
	for key := range deposits {
		result[key] = &node.BackupDeposit{
			Recipient:        deposits[key].Recipient,
			EncryptedDeposit: deposits[key].EncryptedDeposit,
			BlockNumber:      deposits[key].BlockNumber,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: deposits[key].CreatedAt.Unix(),
				Nanos:   int32(deposits[key].CreatedAt.Nanosecond()),
			},
		}
	}
	return result
}

func convertToTransfers(transfers []*backupBalance.BackupTransfer) []*node.BackupTransfer {
	result := make([]*node.BackupTransfer, len(transfers))
	for key := range transfers {
		result[key] = &node.BackupTransfer{
			Recipient:         transfers[key].Recipient,
			EncryptedTransfer: transfers[key].EncryptedTransfer,
			BlockNumber:       transfers[key].BlockNumber,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transfers[key].CreatedAt.Unix(),
				Nanos:   int32(transfers[key].CreatedAt.Nanosecond()),
			},
		}
	}
	return result
}

func convertToTransactions(transactions []*backupBalance.BackupTransaction) []*node.BackupTransaction {
	result := make([]*node.BackupTransaction, len(transactions))
	for key := range transactions {
		result[key] = &node.BackupTransaction{
			Sender:      transactions[key].Sender,
			EncryptedTx: transactions[key].EncryptedTx,
			BlockNumber: strconv.FormatUint(transactions[key].BlockNumber, int10Key),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transactions[key].CreatedAt.Unix(),
				Nanos:   int32(transactions[key].CreatedAt.Nanosecond()),
			},
		}
	}
	return result
}
