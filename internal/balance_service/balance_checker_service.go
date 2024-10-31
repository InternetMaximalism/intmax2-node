package balance_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"

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
	log logger.Logger,
	sb ServiceBlockchain,
	args []string,
	userEthPrivateKey string,
) error {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(utils.RemoveZeroX(userEthPrivateKey))
	if err != nil {
		return ErrInvalidPrivateKey
	}

	userPk, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return ErrRecoverWalletFromPrivateKey
	}

	fmt.Printf("Ethereum address: %s\n", wallet.WalletAddress.Hex())
	fmt.Printf("INTMAX address: %s\n", userPk.ToAddress().String())

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return ErrInvalidTokenType
	}

	l1Balance, err := GetTokenBalance(ctx, cfg, log, sb, *wallet.WalletAddress, *tokenInfo)
	if err != nil {
		return fmt.Errorf(ErrFailedToGetBalance, "Ethereum")
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

	tokenIndex, err := GetTokenIndexFromLiquidityContract(ctx, cfg, log, sb, *tokenInfo)
	if err != nil {
		if errors.Is(err, ErrTokenNotFoundOnIntMax) {
			return errors.New("specified token is not found in INTMAX network")
		}

		return ErrFailedToGetTokenIndex
	}

	l2Balance, err := GetUserBalance(ctx, cfg, log, userPk, tokenIndex)
	if err != nil {
		return errors.Join(fmt.Errorf(ErrFailedToGetBalance, "INTMAX"), err)
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
	log logger.Logger,
	sb ServiceBlockchain,
	owner common.Address,
	tokenInfo intMaxTypes.TokenInfo,
) (balance *big.Int, err error) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failet to get the Evm JSON RPC link")
	}

	client, err := utils.NewClient(link)
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
	log logger.Logger,
	sb ServiceBlockchain,
	tokenInfo intMaxTypes.TokenInfo,
) (uint32, error) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get the EVN JSON RPC link: %w", err)
	}

	client, err := utils.NewClient(link)
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
		return 0, ErrTokenNotFoundOnIntMax
	}

	return tokenIndex, nil
}

func GetUserBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	userPrivateKey *intMaxAcc.PrivateKey,
	tokenIndex uint32,
) (*big.Int, error) {
	storedBalanceData, err := block_synchronizer.GetBackupBalance(ctx, cfg, userPrivateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to get backup balance: %w", err)
	}
	fmt.Printf("size of StoredBalanceData: %v\n", len(storedBalanceData.EncryptedBalanceData))

	balanceData := new(block_synchronizer.BalanceData)
	if err = balanceData.Decrypt(userPrivateKey, storedBalanceData.EncryptedBalanceData); err != nil {
		if err.Error() == "empty encrypted balance data" {
			return big.NewInt(0), nil
		}

		return nil, err
	}

	amount := big.NewInt(0)
	fmt.Printf("tokenIndex: %v\n", tokenIndex)
	fmt.Printf("len(AssetLeafEntries): %v\n", len(balanceData.AssetLeafEntries))
	for _, asset := range balanceData.AssetLeafEntries {
		fmt.Printf("asset.tokenIndex: %v\n", asset.TokenIndex)
		if asset.TokenIndex == tokenIndex {
			amount = asset.Leaf.Amount.BigInt()
			fmt.Printf("amount: %v\n", amount)
			break
		}
	}

	if amount.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("balance is negative: %v", amount)
	}

	return amount, nil
}

func GetUserBalancesRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	address string,
) (*GetBalancesResponse, error) {
	const (
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/balances/%s", cfg.API.DataStoreVaultUrl, address)

	r := resty.New().R()
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to get user balances request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
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
	log logger.Logger,
	depositID string,
) (bool, error) {
	const (
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
		const msg = "failed to get deposit validity request: %w"
		return false, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return false, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"api_url":     apiUrl,
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
