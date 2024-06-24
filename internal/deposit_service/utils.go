package deposit_service

import (
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func newClient(url string) (*ethclient.Client, error) {
	ethClient, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("error connecting to rpc service: %+v", err)
		return nil, err
	}
	return ethClient, nil
}

func createTransactor(cfg *configs.Config) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(cfg.Blockchain.PRIVATE_KEY)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	chainID, err := strconv.ParseInt(cfg.Blockchain.EthreumNetworkChainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	return transactOpts, nil
}

func fetchBlockNumberByDepositIndex(liquidity *bindings.Liquidity, depositIndex uint64) (uint64, error) {
	if depositIndex == 0 {
		return 0, nil
	}

	iterator, err := liquidity.FilterDeposited(&bind.FilterOpts{
		Start: 0,
		End:   nil,
	}, [][32]byte{}, []uint64{depositIndex}, []common.Address{})
	if err != nil {
		return 0, fmt.Errorf("failed to filter logs: %v", err)
	}

	var blockNumber uint64
	for iterator.Next() {
		if iterator.Error() != nil {
			return 0, fmt.Errorf("error encountered: %v", iterator.Error())
		}
		blockNumber = iterator.Event.Raw.BlockNumber
	}

	return blockNumber, nil
}
