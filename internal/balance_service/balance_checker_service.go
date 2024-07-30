package balance_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/deposit_service"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"intmax2-node/pkg/utils"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func SyncBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	userPrivateKey string,
) {
	fmt.Println("Not implemented")
}

func GetBalance(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	userAddress string,
) {
	tokenInfo := parseTokenInfo(args)

	tokenIndex, err := GetTokenIndex(ctx, cfg, db, sb, tokenInfo)
	if err != nil {
		fmt.Println(ErrTokenNotFound, err)
		os.Exit(1)
	}

	userPublicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(userAddress)
	if err != nil {
		fmt.Printf("fail to parse user address: %v\n", err)
		os.Exit(1)
	}

	balance, err := GetUserBalance(db, userPublicKey.ToAddress(), tokenIndex)
	if err != nil {
		fmt.Printf(ErrFailedToGetBalance+": %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Balance: %s\n", balance)
}

func GetTokenIndex(
	ctx context.Context,
	cfg *configs.Config,
	db SQLDriverApp,
	sb deposit_service.ServiceBlockchain,
	tokenInfo intMaxTypes.TokenInfo,
) (uint32, error) {
	// Check local DB for token index
	localTokenIndex, err := getLocalTokenIndex(db, tokenInfo)
	if err == nil {
		return localTokenIndex, nil
	}

	// Check liquidity contract for token index
	return getTokenIndexFromLiquidityContract(ctx, cfg, sb, tokenInfo)
}

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

// Get token index from liquidity contract
func getTokenIndexFromLiquidityContract(
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

type BalanceState struct {
	BalanceProof *intMaxTypes.Plonky2Proof
	BalanceData  map[uint32]*big.Int
	Txs          []intMaxTypes.Tx
	Transfers    []intMaxTypes.Transfer
	Deposits     []intMaxTypes.Transfer
}

func NewBalanceState(
	balanceProof *intMaxTypes.Plonky2Proof,
	balanceData map[uint32]*big.Int,
	txs []intMaxTypes.Tx,
	transfers []intMaxTypes.Transfer,
	deposits []intMaxTypes.Transfer,
) *BalanceState {
	return &BalanceState{
		BalanceProof: balanceProof,
		BalanceData:  balanceData,
		Txs:          txs,
		Transfers:    transfers,
		Deposits:     deposits,
	}
}

func (b *BalanceState) SetZero() *BalanceState {
	b.BalanceProof = nil
	b.BalanceData = make(map[uint32]*big.Int, 0)
	b.Txs = make([]intMaxTypes.Tx, 0)
	b.Transfers = make([]intMaxTypes.Transfer, 0)
	b.Deposits = make([]intMaxTypes.Transfer, 0)

	return b
}

func (b *BalanceState) SetBalance(tokenIndex uint32, amount *big.Int) {
	b.BalanceData[tokenIndex] = amount
}

func (b *BalanceState) GetBalance(tokenIndex uint32) *big.Int {
	balanceData, ok := b.BalanceData[tokenIndex]
	if !ok {
		return big.NewInt(0)
	}

	return balanceData
}

func GetUserBalance(db SQLDriverApp, userAddress intMaxAcc.Address, tokenIndex uint32) (*big.Int, error) {
	const int10Key = 10

	tokenIndexStr := strconv.FormatUint(uint64(tokenIndex), int10Key)
	balanceData, err := db.BalanceByUserAndTokenIndex(userAddress.String(), tokenIndexStr)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		return nil, ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		fmt.Printf("Balance not found for user %s and token index %d\n", userAddress.String(), tokenIndex)
		return big.NewInt(0), nil
	}

	balanceDataInt, ok := new(big.Int).SetString(balanceData.Balance, int10Key)
	if !ok {
		return nil, fmt.Errorf("failed to convert balance to int: %v", err)
	}

	return balanceDataInt, nil
}

func MakeSampleBalanceState(userAddress intMaxAcc.Address) (BalanceState, error) {
	const (
		int100Key = 100
		int200Key = 200
	)

	balanceData := make(map[uint32]*big.Int)
	balanceData[0] = big.NewInt(int100Key)
	balanceData[1] = big.NewInt(int200Key)

	balanceProof, err := intMaxTypes.MakeSamplePlonky2Proof()
	if err != nil {
		return BalanceState{}, fmt.Errorf("failed to make sample plonky2 proof: %v", err)
	}

	balanceState := BalanceState{
		BalanceProof: balanceProof,
		BalanceData:  balanceData,
		Txs:          []intMaxTypes.Tx{},
		Transfers:    []intMaxTypes.Transfer{},
		Deposits:     []intMaxTypes.Transfer{},
	}

	return balanceState, nil
}

func sendTransactionRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress, transfersHash string,
	nonce uint64,
	expiration time.Time,
	powNonce, signature string,
) (*backup_balance.UCGetBalances, error) {
	ucInput := backup_balance.UCGetBalancesInput{
		Address: senderAddress,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	schema := httpKey
	if cfg.HTTP.TLSUse {
		schema = httpsKey
	}

	apiUrl := fmt.Sprintf("%s://%s/v1/transaction", schema, cfg.HTTP.Addr())

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
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
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return nil, err
	}

	response := new(backup_balance.UCGetBalances)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
