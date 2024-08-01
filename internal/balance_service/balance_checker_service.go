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
	"intmax2-node/internal/deposit_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTypes "intmax2-node/internal/types"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"intmax2-node/pkg/utils"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
)

const (
	tokenTypeKey     = "token-type"
	ethTokenType     = "eth"
	erc20TokenType   = "erc20"
	erc721TokenType  = "erc721"
	erc1155TokenType = "erc1155"
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

func parseTokenInfo(args []string) intMaxTypes.TokenInfo {
	if len(args) < 1 {
		fmt.Println(ErrTokenTypeRequired)
		os.Exit(1)
	}

	tokenTypeStr := strings.ToLower(args[0])
	var (
		ok           bool
		tokenType    uint8
		tokenAddress = common.Address{}
		tokenID      = big.NewInt(0)
	)

	const (
		int2Key  = 2
		int3Key  = 3
		int10Key = 10
	)

	switch tokenTypeStr {
	case ethTokenType:
		if len(args) != 1 {
			fmt.Println(ErrETHBalanceCheckArgs)
			os.Exit(1)
		}
		tokenType = ethTokenTypeEnum
	case erc20TokenType:
		if len(args) != int2Key {
			fmt.Println(ErrERC20BalanceCheckArgs)
			os.Exit(1)
		}
		tokenType = erc20TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			fmt.Println(ErrERC721BalanceCheckArgs)
		}
		tokenAddress = common.Address(tokenAddressBytes)
	case erc721TokenType:
		if len(args) != int3Key {
			fmt.Println(ErrERC721BalanceCheckArgs)
			os.Exit(1)
		}
		tokenType = erc721TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			fmt.Println(ErrERC721BalanceCheckArgs)
		}
		tokenAddress = common.Address(tokenAddressBytes)
		tokenIDStr := args[2]
		tokenID, ok = new(big.Int).SetString(tokenIDStr, int10Key)
		if !ok {
			fmt.Println(ErrERC721BalanceCheckArgs)
			os.Exit(1)
		}
	case erc1155TokenType:
		if len(args) != int3Key {
			fmt.Println(ErrERC1155BalanceCheckArgs)
			os.Exit(1)
		}
		tokenType = erc1155TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			fmt.Println(ErrERC721BalanceCheckArgs)
		}
		tokenAddress = common.Address(tokenAddressBytes)
		tokenIDStr := args[2]
		tokenID, ok = new(big.Int).SetString(tokenIDStr, int10Key)
		if !ok {
			fmt.Println(ErrERC721BalanceCheckArgs)
			os.Exit(1)
		}
	default:
		fmt.Println(ErrInvalidTokenType)
	}

	return intMaxTypes.TokenInfo{TokenType: tokenType, TokenAddress: tokenAddress, TokenID: tokenID}
}

func GetBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	sb ServiceBlockchain,
	args []string,
	userEthPrivateKey string,
) {
	fmt.Printf("GetBalance args: %v\n", args)
	tokenInfo := parseTokenInfo(args)

	tokenIndex, err := GetTokenIndex(ctx, cfg, sb, tokenInfo)
	if err != nil {
		fmt.Println(ErrTokenNotFound, err)
		os.Exit(1)
	}

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(removeZeroX(userEthPrivateKey))
	if err != nil {
		fmt.Printf("fail to parse user address: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wallet: %v\n", wallet)

	userPk, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		fmt.Printf("fail to parse user address: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("User address: %s\n", userPk.ToAddress().String())

	balance, err := GetUserBalance(ctx, cfg, lg, userPk, tokenIndex)
	if err != nil {
		fmt.Printf(ErrFailedToGetBalance+": %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Balance: %s\n", balance)
}

func GetTokenIndex(
	ctx context.Context,
	cfg *configs.Config,
	sb deposit_service.ServiceBlockchain,
	tokenInfo intMaxTypes.TokenInfo,
) (uint32, error) {
	// Check local DB for token index
	// localTokenIndex, err := getLocalTokenIndex(db, tokenInfo)
	// if err == nil {
	// 	return localTokenIndex, nil
	// }

	// Check liquidity contract for token index
	return GetTokenIndexFromLiquidityContract(ctx, cfg, sb, tokenInfo)
}

/*
func getLocalTokenIndex(db SQLDriverApp, tokenInfo intMaxTypes.TokenInfo) (uint32, error) {
	tokenAddressStr := tokenInfo.TokenAddress.String()
	tokenIDStr := fmt.Sprintf("%d", tokenInfo.TokenID)

	token, err := db.TokenByTokenInfo(tokenAddressStr, tokenIDStr)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(ErrFetchTokenByTokenAddressAndTokenIDWithDBApp, err.Error()))
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		return 0, errors.Join(errors.New(ErrTokenNotFound), err)
	}

	const (
		int10Key = 10
		int32Key = 32
	)
	tokenIndex, err := strconv.ParseUint(token.TokenIndex, int10Key, int32Key)
	if err != nil {
		return 0, fmt.Errorf("failed to convert token index to int: %v", err)
	}

	return uint32(tokenIndex), err
}
*/

// Get token index from liquidity contract
func GetTokenIndexFromLiquidityContract(
	ctx context.Context,
	cfg *configs.Config,
	sb deposit_service.ServiceBlockchain,
	tokenInfo intMaxTypes.TokenInfo,
) (uint32, error) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var client *ethclient.Client
	client, err = utils.NewClient(link)
	if err != nil {
		log.Fatalf(err.Error())
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
	}

	ok, tokenIndex, err := liquidity.GetTokenIndex(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	}, tokenInfo.TokenType, tokenInfo.TokenAddress, tokenInfo.TokenID)
	if err != nil {
		return 0, fmt.Errorf("failed to get token index from liquidity contract: %v", err)
	}
	if !ok {
		return 0, errors.New(ErrTokenNotFound)
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
	userAllData, err := getUserBalancesRawRequest(ctx, cfg, lg, userPrivateKey.ToAddress().String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
	}
	balanceData, err := CalculateBalance(userAllData, tokenIndex, *userPrivateKey)
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

func getUserBalancesRawRequest(
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

	apiUrl := fmt.Sprintf("%s/v1/balances/%s", cfg.HTTP.DataStoreVaultUrl, address)

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

type GetBalancesResponse struct {
	// The list of deposits
	Deposits []*BackupDeposit `json:"deposits,omitempty"`
	// The list of transfers
	Transfers []*BackupTransfer `json:"transfers,omitempty"`
	// The list of transactions
	Transactions []*BackupTransaction `json:"transactions,omitempty"`
}

type BackupDeposit struct {
	Recipient        string `json:"recipient,omitempty"`
	EncryptedDeposit string `json:"encryptedDeposit,omitempty"`
	BlockNumber      string `json:"blockNumber,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
}

type BackupTransfer struct {
	EncryptedTransfer string `json:"encryptedTransfer,omitempty"`
	Recipient         string `json:"recipient,omitempty"`
	BlockNumber       string `json:"blockNumber,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
}

type BackupTransaction struct {
	Sender      string `json:"sender,omitempty"`
	EncryptedTx string `json:"encryptedTx,omitempty"`
	BlockNumber string `json:"blockNumber,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

func CalculateBalance(userAllData *GetBalancesResponse, tokenIndex uint32, userPrivateKey intMaxAcc.PrivateKey) (*intMaxTypes.Balance, error) {
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

func removeZeroX(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}
