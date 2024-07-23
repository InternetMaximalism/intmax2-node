package balance_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/deposit_service"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"intmax2-node/pkg/utils"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	tokenTypeKey            = "token-type"
	ethTokenType            = "eth"
	erc20TokenType          = "erc20"
	erc721TokenType         = "erc721"
	erc1155TokenType        = "erc1155"
	tokenTypeDescription    = "token type flag. input one of the following four values: \"ETH\", \"ERC20\", \"ERC721\" or \"ERC1155\". The default value is ETH. use as --token \"ETH\""
	tokenAddressKey         = "token-address"
	defaultAddress          = "0x0000000000000000000000000000000000000000"
	tokenAddressDescription = "token address flag. The default value is zero address. use as --token-address \"0x0000000000000000000000000000000000000000\""
	ethTokenTypeEnum        = 0
	erc20TokenTypeEnum      = 1
	erc721TokenTypeEnum     = 2
	erc1155TokenTypeEnum    = 3
)

type TokenInfo struct {
	TokenType    uint8
	TokenAddress common.Address
	TokenID      *big.Int
}

type TokenIndexMap = map[TokenInfo]uint32

func parseTokenInfo(args []string) TokenInfo {
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

	switch tokenTypeStr {
	case ethTokenType:
		if len(args) != 1 {
			fmt.Println(ErrETHBalanceCheckArgs)
			os.Exit(1)
		}
		tokenType = ethTokenTypeEnum
	case erc20TokenType:
		if len(args) != 2 {
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
		if len(args) != 3 {
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
		tokenID, ok = new(big.Int).SetString(tokenIDStr, 10)
		if !ok {
			fmt.Println(ErrERC721BalanceCheckArgs)
			os.Exit(1)
		}
	case erc1155TokenType:
		if len(args) != 3 {
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
		tokenID, ok = new(big.Int).SetString(tokenIDStr, 10)
		if !ok {
			fmt.Println(ErrERC721BalanceCheckArgs)
			os.Exit(1)
		}
	default:
		fmt.Println(ErrInvalidTokenType)
	}

	return TokenInfo{TokenType: tokenType, TokenAddress: tokenAddress, TokenID: tokenID}
}

func SyncBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	userAddress string,
) {
	fmt.Println("Not implemented")
}

func GetBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	userAddress string,
) {
	// userAddress := "0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3"
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
	tokenInfo TokenInfo,
) (uint32, error) {
	// Check local DB for token index
	localTokenIndex, err := getLocalTokenIndex(db, tokenInfo)
	if err == nil {
		return localTokenIndex, nil
	}

	// Check liquidity contract for token index
	return getTokenIndexFromLiquidityContract(ctx, cfg, sb, tokenInfo)
}

func GetTokenInfoMap(ctx context.Context, liquidity *bindings.Liquidity, tokenIndexMap map[uint32]bool) (map[uint32]common.Address, error) {
	var tokenIndices []uint32
	for tokenIndex := range tokenIndexMap {
		tokenIndices = append(tokenIndices, tokenIndex)
	}

	tokenInfoMap := make(map[uint32]common.Address)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(tokenIndices))

	for _, tokenIndex := range tokenIndices {
		wg.Add(1)
		go func(tokenIndex uint32) {
			defer wg.Done()
			tokenInfo, err := liquidity.GetTokenInfo(&bind.CallOpts{
				Pending: false,
				Context: ctx,
			}, tokenIndex)
			if err != nil {
				errChan <- fmt.Errorf("failed to get token info for index %d: %w", tokenIndex, err)
				return
			}
			mu.Lock()
			tokenInfoMap[tokenIndex] = tokenInfo.TokenAddress
			mu.Unlock()
		}(tokenIndex)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return tokenInfoMap, nil
}

func getLocalTokenIndex(db SQLDriverApp, tokenInfo TokenInfo) (uint32, error) {
	tokenAddressStr := tokenInfo.TokenAddress.String()
	tokenIDStr := fmt.Sprintf("%d", tokenInfo.TokenID)

	token, err := db.TokenByTokenInfo(tokenAddressStr, tokenIDStr)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(ErrFetchTokenByTokenAddressAndTokenIDWithDBApp, err.Error()))
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		return 0, errors.Join(errors.New(ErrTokenNotFound), err)
	}

	tokenIndex, err := strconv.ParseUint(token.TokenIndex, 10, 32)
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
	tokenInfo TokenInfo,
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

	tokenIndex, err := liquidity.GetTokenIndex(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	}, tokenInfo.TokenType, tokenInfo.TokenAddress, tokenInfo.TokenID)
	if err != nil {
		return 0, fmt.Errorf("failed to get token index from liquidity contract: %v", err)
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
	tokenIndexStr := strconv.FormatUint(uint64(tokenIndex), 10)
	balanceData, err := db.BalanceByUserAndTokenIndex(userAddress.String(), tokenIndexStr)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(ErrFetchTokenByTokenAddressAndTokenIDWithDBApp, err.Error()))
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		fmt.Printf("Balance not found for user %s and token index %d\n", userAddress.String(), tokenIndex)
		return big.NewInt(0), nil
	}

	balanceDataInt, ok := new(big.Int).SetString(balanceData.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert balance to int: %v", err)
	}

	return balanceDataInt, nil
}

func MakeSampleBalanceState(userAddress intMaxAcc.Address) (BalanceState, error) {
	balanceData := make(map[uint32]*big.Int)
	balanceData[0] = big.NewInt(100)
	balanceData[1] = big.NewInt(200)

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
