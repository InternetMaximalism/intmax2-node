package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/grpc_server/utils"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
	for i, deposit := range deposits {
		result[i] = &node.BackupDeposit{
			Recipient:        deposit.Recipient,
			EncryptedDeposit: deposit.EncryptedDeposit,
			BlockNumber:      deposit.BlockNumber,
			CreatedAt:        deposit.CreatedAt.Format(time.RFC3339),
		}
	}
	return result
}

func convertToTransfers(transfers []*backupBalance.BackupTransfer) []*node.BackupTransfer {
	result := make([]*node.BackupTransfer, len(transfers))
	for i, transfer := range transfers {
		result[i] = &node.BackupTransfer{
			Recipient:         transfer.Recipient,
			EncryptedTransfer: transfer.EncryptedTransfer,
			BlockNumber:       transfer.BlockNumber,
			CreatedAt:         transfer.CreatedAt.Format(time.RFC3339),
		}
	}
	return result
}

func convertToTransactions(transactions []*backupBalance.BackupTransaction) []*node.BackupTransaction {
	result := make([]*node.BackupTransaction, len(transactions))
	for i, transaction := range transactions {
		result[i] = &node.BackupTransaction{
			Sender:      transaction.Sender,
			EncryptedTx: transaction.EncryptedTx,
			BlockNumber: string(transaction.BlockNumber),
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		}
	}
	return result
}
