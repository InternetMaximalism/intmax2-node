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
	"math/big"

	"go.opentelemetry.io/otel/attribute"
)

const (
	int3Key  = 3
	int10Key = 10
)

var ErrBackupDeposit = errors.New("failed to backup deposit")

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

	if len(args) > int3Key {
		return errors.New("too many arguments")
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return err
	}

	// tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
	// if err != nil {
	// 	if err.Error() != "token not found on INTMAX network" {
	// 		return err
	// 	}

	// 	fmt.Println("AAA: Token not found on INTMAX network")
	// 	return err
	// }

	d, err := service.NewDepositAnalyzerService(
		ctx, u.cfg, u.log, u.sb,
	)
	if err != nil {
		return err
	}

	const (
		ethTokenTypeEnum = iota
		erc20TokenTypeEnum
		erc721TokenTypeEnum
		erc1155TokenTypeEnum
	)

	amountInt, ok := new(big.Int).SetString(amount, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %s", amount)
	}

	var tokenIndex uint32
	if tokenInfo.TokenType == ethTokenTypeEnum {
		// ETH
		depositID, salt, innerErr := d.DepositETHWithRandomSalt(userEthPrivateKeyHex, recipientAddress, amount)
		if innerErr != nil {
			return innerErr
		}

		u.log.Infof("ETH deposit is successful")

		tokenIndex, err = balance_service.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
		if err != nil {
			return err
		}

		err = d.BackupDeposit(recipientAddress, tokenIndex, amountInt, salt, depositID)
		if err != nil {
			return errors.Join(ErrBackupDeposit, err)
		}
	} else if tokenInfo.TokenType == erc20TokenTypeEnum {
		// ERC20
		depositID, salt, innerErr := d.DepositERC20WithRandomSalt(userEthPrivateKeyHex, recipientAddress, tokenInfo.TokenAddress, amount)
		if innerErr != nil {
			return innerErr
		}

		u.log.Infof("ERC20 deposit is successful")

		tokenIndex, err = balance_service.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
		if err != nil {
			return err
		}

		err = d.BackupDeposit(recipientAddress, tokenIndex, amountInt, salt, depositID)
		if err != nil {
			return errors.Join(ErrBackupDeposit, err)
		}
	} else if tokenInfo.TokenType == erc721TokenTypeEnum {
		// ERC721
		depositID, salt, innerErr := d.DepositERC721WithRandomSalt(userEthPrivateKeyHex, recipientAddress, tokenInfo.TokenAddress, tokenInfo.TokenID)
		if innerErr != nil {
			return innerErr
		}

		u.log.Infof("ERC721 deposit is successful")

		tokenIndex, err = balance_service.GetTokenIndexFromLiquidityContract(ctx, u.cfg, u.sb, *tokenInfo)
		if err != nil {
			return err
		}

		err = d.BackupDeposit(recipientAddress, tokenIndex, big.NewInt(1), salt, depositID)
		if err != nil {
			return errors.Join(ErrBackupDeposit, err)
		}
	} else if tokenInfo.TokenType == erc1155TokenTypeEnum {
		// ERC1155
		return errors.New("ERC1155 is not supported yet")
	} else {
		return errors.New("token type is not supported")
	}

	return nil
}
