package tx_withdrawal_transfers_list

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_withdrawal_service"
	txWithdrawalTransfersList "intmax2-node/internal/use_cases/tx_withdrawal_transfers_list"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txWithdrawalTransfersList.UseCaseTxWithdrawalTransfersList {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *txWithdrawalTransfersList.UCTxWithdrawalTransfersListInput,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		hName         = "UseCase TxWithdrawalTransfersList"
		recipientKey  = "recipient"
		inputValueKey = "input_value"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrInputValueEmpty)
		return nil, ErrInputValueEmpty
	}

	inBytes, err := json.Marshal(&input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrMarshalJSONFail, err)
	}

	span.SetAttributes(attribute.String(inputValueKey, string(inBytes)))

	if userEthPrivateKey == "" {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyUserPrivateKey)
		return nil, ErrEmptyUserPrivateKey
	}

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, fmt.Errorf("fail to parse user private key: %v", err)
	}

	span.SetAttributes(
		attribute.String(recipientKey, wallet.WalletAddress.String()),
	)

	return service.TransfersList(spanCtx, u.cfg, &service.GetTransfersListInput{
		Sorting: input.Sorting,
		Pagination: &service.GetTransfersListPagination{
			Direction: input.Pagination.Direction,
			Limit:     input.Pagination.Limit,
			Cursor: &service.GetTransfersListPaginationCursor{
				BlockNumber:  input.Pagination.Cursor.BlockNumber,
				SortingValue: input.Pagination.Cursor.SortingValue,
			},
		},
		Filter: &service.GetTransfersListFilter{
			Name:      input.Filter.Name,
			Condition: input.Filter.Condition,
			Value:     input.Filter.Value,
		},
	}, userEthPrivateKey)
}
