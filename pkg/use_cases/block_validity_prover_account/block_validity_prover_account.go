package block_validity_prover_account

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverAccount "intmax2-node/internal/use_cases/block_validity_prover_account"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"go.opentelemetry.io/otel/attribute"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) ucBlockValidityProverAccount.UseCaseBlockValidityProverAccount {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverAccount.UCBlockValidityProverAccountInput,
) (*ucBlockValidityProverAccount.UCBlockValidityProverAccount, error) {
	const (
		hName            = "UseCase BlockValidityProverAccount"
		senderAddressKey = "sender_address"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverAccountInputEmpty)
		return nil, ErrUCBlockValidityProverAccountInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderAddressKey, input.Address),
	)

	address, err := intMaxAcc.NewAddressFromHex(input.Address)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNewAddressFromHexFail, err)
	}

	_, err = address.Public()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrPublicKeyFromIntMaxAccFail, err)
	}

	var sender *mDBApp.BlockSender
	sender, err = u.db.BlockSenderByAddress(address.String())
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrSenderByAddressFail, err)
	}

	var acc *mDBApp.BlockAccount
	acc, err = u.db.BlockAccountBySenderID(sender.ID)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrAccountBySenderIDFail, err)
	}

	return &ucBlockValidityProverAccount.UCBlockValidityProverAccount{
		AccountID: acc.AccountID,
	}, nil
}
