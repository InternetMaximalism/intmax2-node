package balance_service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTypes "intmax2-node/internal/types"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"intmax2-node/pkg/utils"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-resty/resty/v2"
)

const (
	//nolint:gosec
	tokenTypeDescription    = "token type flag. input one of the following four values: \"ETH\", \"ERC20\", \"ERC721\" or \"ERC1155\". The default value is ETH. use as --token \"ETH\""
	tokenAddressKey         = "token-address"
	defaultAddress          = "0x0000000000000000000000000000000000000000"
	tokenAddressDescription = "token address flag. The default value is zero address. use as --token-address \"0x0000000000000000000000000000000000000000\""
	ethTokenTypeEnum        = 0
	erc20TokenTypeEnum      = 1
	erc721TokenTypeEnum     = 2
	erc1155TokenTypeEnum    = 3
)

func GetBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	sb ServiceBlockchain,
	args []string,
	userEthPrivateKey string,
) error {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(utils.RemoveZeroX(userEthPrivateKey))
	if err != nil {
		return errors.Join(ErrInvalidPrivateKey, err)
	}

	userPk, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return errors.Join(ErrRecoverWalletFromPrivateKey, err)
	}

	fmt.Printf("Ethereum address: %s\n", wallet.WalletAddress.Hex())
	fmt.Printf("INTMAX address: %s\n", userPk.ToAddress().String())

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		fmt.Println(ErrInvalidTokenType)
		return nil
	}

	l1Balance, err := GetTokenBalance(ctx, cfg, lg, *wallet.WalletAddress, *tokenInfo)
	if err != nil {
		fmt.Printf(ErrFailedToGetBalance, "Ethereum")
		return nil
	}

	switch tokenInfo.TokenType {
	case ethTokenTypeEnum:
		fmt.Println("ETH Balance")
	case erc20TokenTypeEnum:
		fmt.Printf("Token Balance (Address: %s)\n", tokenInfo.TokenAddress.String())
	case erc721TokenTypeEnum:
		fmt.Printf("Token Balance (Address: %s, ID: %s)\n", tokenInfo.TokenAddress.String(), tokenInfo.TokenID)
	case erc1155TokenTypeEnum:
		fmt.Printf("Token Balance (Address: %s, ID: %s)\n", tokenInfo.TokenAddress.String(), tokenInfo.TokenID)
	default:
		return ErrInvalidTokenType
	}

	switch tokenInfo.TokenType {
	case erc721TokenTypeEnum:
		if l1Balance.Cmp(big.NewInt(0)) == 0 {
			fmt.Println("You don't own this token on Ethereum")
		} else {
			fmt.Println("You own this token on Ethereum")
		}
	default:
		fmt.Printf("Balance on Ethereum: %s\n", l1Balance)
	}

	tokenIndex, err := GetTokenIndexFromLiquidityContract(ctx, cfg, sb, *tokenInfo)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			fmt.Println("Specified token is not found in INTMAX network")
			return nil
		}

		return errors.Join(ErrFailedToGetTokenIndex, err)
	}

	l2Balance, err := GetUserBalance(ctx, cfg, lg, userPk, tokenIndex)
	if err != nil {
		fmt.Printf(ErrFailedToGetBalance, "INTMAX")
		return nil
	}

	switch tokenInfo.TokenType {
	case erc721TokenTypeEnum:
		if l2Balance.Cmp(big.NewInt(0)) == 0 {
			fmt.Println("You don't own this token on INTMAX network")
		} else {
			fmt.Println("You own this token on INTMAX network")
		}
	default:
		fmt.Printf("Balance on INTMAX network: %s\n", l2Balance)
	}

	return nil
}

func GetTokenBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	owner common.Address,
	tokenInfo intMaxTypes.TokenInfo,
) (balance *big.Int, err error) {
	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	switch tokenInfo.TokenType {
	case ethTokenTypeEnum:
		balance, err = client.BalanceAt(ctx, owner, nil)
		if err != nil {
			return nil, errors.Join(ErrFailedToGetETHBalance, err)
		}
	case erc20TokenTypeEnum:
		var erc20Contract *bindings.Erc20
		erc20Contract, err = bindings.NewErc20(tokenInfo.TokenAddress, client)
		if err != nil {
			log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
		}

		balance, err = erc20Contract.BalanceOf(&bind.CallOpts{
			Pending: false,
			Context: ctx,
		}, owner)
		if err != nil {
			return nil, errors.Join(ErrFailedToGetERC20Balance, err)
		}
	case erc721TokenTypeEnum:
		var erc721Contract *bindings.Erc721
		erc721Contract, err = bindings.NewErc721(tokenInfo.TokenAddress, client)
		if err != nil {
			log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
		}

		var actualOwner common.Address
		actualOwner, err = erc721Contract.OwnerOf(&bind.CallOpts{
			Pending: false,
			Context: ctx,
		}, tokenInfo.TokenID)
		if err != nil {
			return nil, errors.Join(ErrFailedToGetERC721Owner, err)
		}
		if actualOwner == owner {
			balance = big.NewInt(1)
		} else {
			balance = big.NewInt(0)
		}
	case erc1155TokenTypeEnum:
		var erc1155Contract *bindings.Erc1155
		erc1155Contract, err = bindings.NewErc1155(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
		if err != nil {
			log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
		}

		balance, err = erc1155Contract.BalanceOf(&bind.CallOpts{
			Pending: false,
			Context: ctx,
		}, owner, tokenInfo.TokenID)
		if err != nil {
			return nil, errors.Join(ErrFailedToGetERC1155Balance, err)
		}
	default:
		return nil, ErrInvalidTokenType
	}

	return balance, nil
}

func GetTokenIndexFromLiquidityContract(
	ctx context.Context,
	cfg *configs.Config,
	sb ServiceBlockchain,
	tokenInfo intMaxTypes.TokenInfo,
) (uint32, error) {
	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to create new client: %w", err)
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		const msg = "failed to instantiate a Liquidity contract"
		log.Fatalf("%s: %v", msg, err.Error())
	}

	ok, tokenIndex, err := liquidity.GetTokenIndex(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	}, tokenInfo.TokenType, tokenInfo.TokenAddress, tokenInfo.TokenID)
	if err != nil {
		return 0, fmt.Errorf("failed to get token index from liquidity contract: %v", err)
	}
	if !ok {
		return 0, ErrTokenNotFound
	}

	return tokenIndex, nil
}

func GetUserBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	userPrivateKey *intMaxAcc.PrivateKey,
	tokenIndex uint32,
) (*big.Int, error) {
	userAllData, err := GetUserBalancesRawRequest(ctx, cfg, lg, userPrivateKey.ToAddress().String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
	}
	balanceData, err := CalculateBalance(ctx, cfg, lg, userAllData, tokenIndex, *userPrivateKey)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		return nil, ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		fmt.Printf("Balance not found for user %s and token index %d\n", userPrivateKey.ToAddress().String(), tokenIndex)
		return big.NewInt(0), nil
	}
	if balanceData.Amount.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("balance is negative: %v", balanceData.Amount)
	}

	return balanceData.Amount, nil
}

func GetUserBalancesRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	address string,
) (*GetBalancesResponse, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/balances/%s", cfg.API.DataStoreVaultUrl, address)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		lg.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return nil, err
	}

	response := new(GetBalancesResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

func GetDepositValidityRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	depositID string,
) (bool, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	// GetVerifyDepositConfirmation
	apiUrl := fmt.Sprintf("%s/v1/deposits/%s/verify-confirmation", cfg.API.DataStoreVaultUrl, depositID)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return false, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return false, fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		lg.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return false, err
	}

	response := new(GetVerifyDepositConfirmationResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return false, fmt.Errorf("failed to get verify deposit confirmation response: %v", response)
	}

	return response.Data.Confirmed, nil
}

func CalculateBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	userAllData *GetBalancesResponse,
	tokenIndex uint32,
	userPrivateKey intMaxAcc.PrivateKey,
) (*intMaxTypes.Balance, error) {
	balance := big.NewInt(0)
	for _, deposit := range userAllData.Deposits {
		encryptedDepositBytes, err := base64.StdEncoding.DecodeString(deposit.EncryptedDeposit)
		if err != nil {
			log.Printf("failed to decode deposit: %v", err)
			continue
		}

		encodedDeposit, err := userPrivateKey.DecryptECIES(encryptedDepositBytes)
		if err != nil {
			log.Printf("failed to decrypt deposit: %v", err)
			continue
		}

		var decodedDeposit intMaxTypes.Deposit
		err = decodedDeposit.Unmarshal(encodedDeposit)
		if err != nil {
			log.Printf("failed to unmarshal deposit: %v", err)
			continue
		}

		// Request data store vault if deposit is valid
		depositID := deposit.BlockNumber
		ok, err := GetDepositValidityRawRequest(
			ctx,
			cfg,
			lg,
			depositID,
		)
		if err != nil {
			return nil, errors.Join(ErrDepositValidity, err)
		}
		if !ok {
			continue
		}

		if decodedDeposit.TokenIndex == tokenIndex {
			balance = new(big.Int).Add(balance, decodedDeposit.Amount)
		}
	}

	for _, transfer := range userAllData.Transfers {
		encryptedTransferBytes, err := base64.StdEncoding.DecodeString(transfer.EncryptedTransfer)
		if err != nil {
			log.Printf("failed to decode transfer: %v", err)
			continue
		}
		encodedTransfer, err := userPrivateKey.DecryptECIES(encryptedTransferBytes)
		if err != nil {
			log.Printf("failed to decrypt transfer: %v", err)
			continue
		}
		var decodedTransfer intMaxTypes.Transfer
		err = decodedTransfer.Unmarshal(encodedTransfer)
		if err != nil {
			log.Printf("failed to unmarshal transfer: %v", err)
			continue
		}
		if tokenIndex == decodedTransfer.TokenIndex {
			balance = new(big.Int).Add(balance, decodedTransfer.Amount)
		}
	}

	for _, transaction := range userAllData.Transactions {
		encryptedTxBytes, err := base64.StdEncoding.DecodeString(transaction.EncryptedTx)
		if err != nil {
			log.Printf("failed to decode transaction: %v", err)
			continue
		}
		encodedTx, err := userPrivateKey.DecryptECIES(encryptedTxBytes)
		if err != nil {
			log.Printf("failed to decrypt transaction: %v", err)
			continue
		}
		var decodedTx intMaxTypes.TxDetails
		err = decodedTx.Unmarshal(encodedTx)
		if err != nil {
			log.Printf("failed to unmarshal transaction: %v", err)
			continue
		}
		for _, transfer := range decodedTx.Transfers {
			if tokenIndex == transfer.TokenIndex {
				balance = new(big.Int).Sub(balance, transfer.Amount)
			}
		}
	}

	return &intMaxTypes.Balance{
		TokenIndex: tokenIndex,
		Amount:     balance,
	}, nil
}
