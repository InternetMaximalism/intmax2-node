package post_backup_balance

import (
	"context"
	"encoding/binary"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	intMaxTypes "intmax2-node/internal/types"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"

	"io"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backupBalance.UCPostBackupBalanceInput,
) error {
	const (
		hName                    = "UseCase PostBackupBalance"
		userKey                  = "user"
		blockNumberKey           = "block_number"
		encryptedBalanceProofKey = "encrypted_balance_proof"
		encryptedBalanceDataKey  = "encrypted_balance_data"
		encryptedTxsKey          = "encrypted_txs"
		encryptedTransfersKey    = "encrypted_transfers"
		encryptedDepositsKey     = "encrypted_deposits"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupBalanceInputEmpty)
		return ErrUCPostBackupBalanceInputEmpty
	}

	span.SetAttributes(
		// attribute.String(userKey, input.DecodeUser.ToAddress().String()),
		// attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		// attribute.String(encryptedBalanceProofKey, input.EncryptedBalanceProof.Proof),
		// attribute.String(encryptedBalanceProofKey, input.EncryptedBalanceProof.EncryptedPublicInputs),
		attribute.String(encryptedBalanceDataKey, input.EncryptedBalanceData),
		attribute.StringSlice(encryptedTxsKey, input.EncryptedTxs),
		attribute.StringSlice(encryptedTransfersKey, input.EncryptedTransfers),
		attribute.StringSlice(encryptedDepositsKey, input.EncryptedDeposits),
	)

	err := service.PostBackupBalance(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return fmt.Errorf("failed to post backup balance: %w", err)
	}

	return nil
}

func WriteBalance(buf io.Writer, balance *intMaxTypes.Balance) error {
	err := binary.Write(buf, binary.LittleEndian, balance.TokenIndex)
	if err != nil {
		return err
	}
	const numAmountBytes = 32
	balanceBytes := make([]byte, numAmountBytes)
	b := balance.Amount.Bytes()
	copy(balanceBytes[32-len(b):], b)

	err = binary.Write(buf, binary.LittleEndian, balanceBytes)
	if err != nil {
		return err
	}
	return nil
}

func WriteBalances(buf io.Writer, balances []*intMaxTypes.Balance) error {
	if err := binary.Write(buf, binary.LittleEndian, int64(len(balances))); err != nil {
		return err
	}

	for _, balance := range balances {
		if err := WriteBalance(buf, balance); err != nil {
			return err
		}
	}

	return nil
}
