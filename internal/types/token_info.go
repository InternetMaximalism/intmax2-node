package types

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	ethTokenType         = "eth"
	erc20TokenType       = "erc20"
	erc721TokenType      = "erc721"
	erc1155TokenType     = "erc1155"
	ethTokenTypeEnum     = 0
	erc20TokenTypeEnum   = 1
	erc721TokenTypeEnum  = 2
	erc1155TokenTypeEnum = 3
)

type TokenInfo struct {
	TokenType    uint8
	TokenAddress common.Address
	TokenID      *big.Int
}

func NewTokenInfo(tokenType uint8, tokenAddress common.Address, tokenID *big.Int) *TokenInfo {
	return &TokenInfo{TokenType: tokenType, TokenAddress: tokenAddress, TokenID: tokenID}
}

func (ti *TokenInfo) ParseFromStrings(args []string) (*TokenInfo, error) {
	if len(args) < 1 {
		return nil, ErrTokenTypeRequired
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
			return nil, ErrInvalidETHArgs
		}
		tokenType = ethTokenTypeEnum
	case erc20TokenType:
		if len(args) != int2Key {
			return nil, ErrInvalidERC20Args
		}
		tokenType = erc20TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			return nil, ErrInvalidERC20Args
		}
		tokenAddress = common.BytesToAddress(tokenAddressBytes)
	case erc721TokenType:
		if len(args) != int3Key {
			return nil, ErrInvalidERC721Args
		}
		tokenType = erc721TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			return nil, ErrInvalidERC721Args
		}
		tokenAddress = common.BytesToAddress(tokenAddressBytes)
		tokenIDStr := args[2]
		tokenID, ok = new(big.Int).SetString(tokenIDStr, int10Key)
		if !ok {
			return nil, ErrInvalidERC721Args
		}
	case erc1155TokenType:
		if len(args) != int3Key {
			return nil, ErrInvalidERC1155Args
		}
		tokenType = erc1155TokenTypeEnum
		tokenAddressBytes, err := hexutil.Decode(args[1])
		if err != nil {
			return nil, ErrInvalidERC1155Args
		}
		tokenAddress = common.BytesToAddress(tokenAddressBytes)
		tokenIDStr := args[2]
		tokenID, ok = new(big.Int).SetString(tokenIDStr, int10Key)
		if !ok {
			return nil, ErrInvalidERC1155Args
		}
	default:
		return nil, ErrInvalidTokenType
	}

	return &TokenInfo{TokenType: tokenType, TokenAddress: tokenAddress, TokenID: tokenID}, nil
}
