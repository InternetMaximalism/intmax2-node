package tx_deposit

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	balanceService "intmax2-node/internal/balance_service"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	txDepositService "intmax2-node/internal/tx_deposit_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/tx_deposit"
	"math/big"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

const (
	int3Key  = 3
	int10Key = 10
)

var (
	ErrBackupDeposit         = errors.New("failed to backup deposit")
	ErrInvalidArguments      = errors.New("invalid arguments")
	ErrUnsupportedToken      = errors.New("unsupported token type")
	ErrEmptyUserPrivateKey   = errors.New("user private key is empty")
	ErrEmptyRecipientAddress = errors.New("recipient address is empty")
	ErrEmptyAmount           = errors.New("amount is empty")
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
	ds  *txDepositService.TxDepositService
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
		hName     = "UseCase TxDepoxit"
		senderKey = "sender"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if userEthPrivateKeyHex == "" {
		return ErrEmptyUserPrivateKey
	}

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

	if recipientAddressStr == "" {
		return ErrEmptyRecipientAddress
	}

	if amount == "" {
		return ErrEmptyAmount
	}

	recipientAddress, err := intMaxAcc.NewAddressFromHex(recipientAddressStr)
	if err != nil {
		return err
	}

	if len(args) > int3Key {
		return ErrInvalidArguments
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return err
	}

	u.ds, err = txDepositService.NewTxDepositService(
		ctx, u.cfg, u.log, u.sb,
	)
	if err != nil {
		return err
	}

	amountInt, ok := new(big.Int).SetString(amount, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %s", amount)
	}

	return u.processDeposit(ctx, tokenInfo, userEthPrivateKeyHex, recipientAddress, amount, amountInt)
}

func (u *uc) processDeposit(ctx context.Context, tokenInfo *intMaxTypes.TokenInfo, privateKey string, recipient intMaxAcc.Address, amountStr string, amountInt *big.Int) error {
	const (
		ethTokenTypeEnum = iota
		erc20TokenTypeEnum
		erc721TokenTypeEnum
		erc1155TokenTypeEnum
	)
	var (
		depositID uint64
		salt      *goldenposeidon.PoseidonHashOut
		err       error
		tokenType string
	)

	switch tokenInfo.TokenType {
	case ethTokenTypeEnum:
		depositID, salt, err = u.ds.DepositETHWithRandomSalt(privateKey, recipient, amountStr)
		tokenType = "ETH"
	case erc20TokenTypeEnum:
		depositID, salt, err = u.ds.DepositERC20WithRandomSalt(privateKey, recipient, tokenInfo.TokenAddress, amountStr)
		tokenType = "ERC20"
	case erc721TokenTypeEnum:
		depositID, salt, err = u.ds.DepositERC721WithRandomSalt(privateKey, recipient, tokenInfo.TokenAddress, tokenInfo.TokenID)
		amountInt = big.NewInt(1) // ERC721 always has an amount of 1
		tokenType = "ERC721"
	case erc1155TokenTypeEnum:
		return errors.New("ERC1155 is not supported yet")
	default:
		return ErrUnsupportedToken
	}

	if err != nil {
		if strings.Contains(err.Error(), "failed to send ERC20.Approve transaction:") {
			return fmt.Errorf("the specified ERC20 is not owned by the account or insufficient balance")
		}
		if strings.Contains(err.Error(), "failed to send ERC721.Approve transaction:") {
			return fmt.Errorf("the specified ERC721 is not owned by the account")
		}
		return fmt.Errorf("%s deposit is failed: %w", tokenType, err)
	}

	u.log.Infof("%s deposit is successful", tokenType)

	tokenIndex, err := balanceService.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
	if err != nil {
		return err
	}

	if err = u.ds.BackupDeposit(recipient, tokenIndex, amountInt, salt, depositID); err != nil {
		return errors.Join(ErrBackupDeposit, err)
	}

	return nil
}
