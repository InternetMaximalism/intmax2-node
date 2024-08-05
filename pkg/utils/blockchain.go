package utils

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewClient(url string) (*ethclient.Client, error) {
	ethClient, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("error connecting to rpc service: %+v", err)
		return nil, err
	}
	return ethClient, nil
}

func CreateTransactor(ethereumPrivateKeyHex, networkChainID string) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(ethereumPrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	const (
		int10Key = 10
		int64Key = 64
	)
	chainID, err := strconv.ParseInt(networkChainID, int10Key, int64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	if transactOpts == nil {
		return nil, fmt.Errorf("transactOpts is nil")
	}

	return transactOpts, nil
}

func StringToBigInt(s string) (*big.Int, error) {
	const int10Key = 10
	i := new(big.Int)
	if _, success := i.SetString(s, int10Key); !success {
		return nil, fmt.Errorf("failed to convert string to big.Int: %s", s)
	}
	return i, nil
}

func IsValidEthereumPrivateKey(key string) error {
	key = strings.TrimPrefix(key, "0x")
	_, err := crypto.HexToECDSA(key)
	if err != nil {
		return fmt.Errorf("invalid ethereum private key: %w", err)
	}
	return nil
}
