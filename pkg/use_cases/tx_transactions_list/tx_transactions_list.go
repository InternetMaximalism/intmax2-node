package tx_transactions_list

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_transfer_service"
	txTransactionsList "intmax2-node/internal/use_cases/tx_transactions_list"

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
) txTransactionsList.UseCaseTxTransactionsList {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *txTransactionsList.UCTxTransactionsListInput,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		hName         = "UseCase TxTransactionsList"
		senderKey     = "sender"
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

	// The userPrivateKey is acceptable in either format:
	// it may include the '0x' prefix at the beginning,
	// or it can be provided without this prefix.
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	userAddress := userAccount.ToAddress()

	span.SetAttributes(
		attribute.String(senderKey, userAddress.String()),
	)

	return service.TransactionsList(spanCtx, u.cfg, &service.GetTransactionsListInput{
		Sorting: input.Sorting,
		Pagination: &service.GetTransactionsListPagination{
			Direction: input.Pagination.Direction,
			Limit:     input.Pagination.Limit,
			Cursor: &service.GetTransactionsListPaginationCursor{
				BlockNumber:  input.Pagination.Cursor.BlockNumber,
				SortingValue: input.Pagination.Cursor.SortingValue,
			},
		},
		Filter: &service.GetTransactionsListFilter{
			Name:      input.Filter.Name,
			Condition: input.Filter.Condition,
			Value:     input.Filter.Value,
		},
	}, userEthPrivateKey)
}
