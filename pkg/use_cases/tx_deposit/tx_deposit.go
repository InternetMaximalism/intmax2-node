package tx_deposit

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_deposit_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/tx_deposit"

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
) tx_deposit.UseCaseTxDeposit {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, args []string, recipientAddressStr, amount, userEthPrivateKeyHex string) (err error) {
	const (
		hName     = "UseCase TxTransfer"
		senderKey = "sender"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	// The userPrivateKey is acceptable in either format:
	// it may include the '0x' prefix at the beginning,
	// or it can be provided without this prefix.
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(userEthPrivateKeyHex)
	if err != nil {
		return err
	}

	userAddress := userAccount.ToAddress()
	span.SetAttributes(
		attribute.String(senderKey, userAddress.String()),
	)

	fmt.Printf("userAddress: %s\n", userAddress.String())

	fmt.Printf("recipientAddressStr: %s\n", recipientAddressStr)
	fmt.Printf("amount: %s\n", amount)
	recipientAddress, err := intMaxAcc.NewAddressFromHex(recipientAddressStr)
	if err != nil {
		return err
	}

	if len(args) > 3 {
		return errors.New("too many arguments")
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return err
	}

	tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
	if err != nil {
		return err
	}

	d, err := service.NewDepositAnalyzerService(
		ctx, u.cfg, u.log, u.sb,
	)
	if err != nil {
		return err
	}

	if tokenInfo.TokenType == 0 {
		// ETH
		d.DepositETHWithRandomSalt(userEthPrivateKeyHex, recipientAddress, tokenIndex, amount)

		fmt.Printf("ETH deposit is successful\n")
		return nil
	}

	return errors.New("token type is not supported")
}
