package block_post_service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_post_service"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountInfoMap(t *testing.T) {
	ctx := context.Background()
	cfg := configs.Config{}
	cfg.Blockchain.EthereumNetworkRpcUrl = "https://eth-sepolia.g.alchemy.com/v2/OE-Ocf1AKEHq5UlRKEXGYJ6mc7dJamOV"
	cfg.Blockchain.LiquidityContractAddress = "0x86f08b1DcDe0673A5562384eAFD5df6EE83cd73a"
	cfg.Blockchain.RollupContractAddress = "0x33a463381140C97B3bFd41eAADAE38ee98fe410c"
	cfg.SQLDb.DNSConnection = "postgresql://postgres:pass@intmax2-node-postgres:5432/state?sslmode=disable"
	cfg.Blockchain.RollupContractDeployedBlockNumber = 5843354

	d, err := block_post_service.NewBlockPostService(ctx, &cfg, nil)
	require.NoError(t, err)

	events, _, err := d.FetchNewPostedBlocks(cfg.Blockchain.RollupContractDeployedBlockNumber)
	require.NoError(t, err)

	accountInfoMap := block_post_service.NewAccountInfoMap()
	for i, event := range events {
		calldata, err := d.FetchScrollCalldataByHash(event.Raw.TxHash)
		require.NoError(t, err)

		// fmt.Println("calldata:", hexutil.Encode(calldata))

		_, err = block_post_service.FetchIntMaxBlockContentByCalldata(calldata, accountInfoMap)
		if err != nil {
			if errors.Is(err, block_post_service.ErrUnknownAccountID) {
				fmt.Printf("block %d is ErrUnknownAccountID\n", i)
				continue
			}
			if errors.Is(err, block_post_service.ErrCannotDecodeAddress) {
				fmt.Printf("block %d is ErrCannotDecodeAddress\n", i)
				continue
			}
			assert.NoError(t, err)
			continue
		}

		fmt.Printf("block %d is valid\n", i)
	}
}

func TestFetchNewPostedBlocks(t *testing.T) {
	calldataJson, err := readPostedBlockEventsJson("../../pkg/data/posted_block_calldata.json")
	assert.NoError(t, err)

	accountInfoMap := block_post_service.NewAccountInfoMap()
	for i, calldata := range calldataJson {
		_, err = block_post_service.FetchIntMaxBlockContentByCalldata(calldata, accountInfoMap)
		if err != nil {
			if errors.Is(err, block_post_service.ErrUnknownAccountID) {
				fmt.Printf("block %d is ErrUnknownAccountID\n", i)
				continue
			}
			if errors.Is(err, block_post_service.ErrCannotDecodeAddress) {
				fmt.Printf("block %d is ErrCannotDecodeAddress\n", i)
				continue
			}
		}
		assert.NoError(t, err)

		if err == nil {
			fmt.Printf("block %d is valid\n", i)
		}
	}

	// t.Log("lastIntMaxBlockNumber:", lastIntMaxBlockNumber)

	// _, err = block_post_service.FetchIntMaxBlockContentByCalldata(calldataJson[2])
	// assert.NoError(t, err)
}

var ErrCannotOpenBinaryFile = errors.New("cannot open binary file")
var ErrCannotGetFileInformation = errors.New("cannot get file information")
var ErrCannotReadBinaryFile = errors.New("cannot read binary file")
var ErrCannotParseJson = errors.New("cannot parse JSON")
var ErrCannotDecodeHex = errors.New("cannot decode hex")

func readPostedBlockEventsJson(filePath string) ([][]byte, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(ErrCannotOpenBinaryFile, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Join(ErrCannotGetFileInformation, err)
	}

	// Create buffer
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	// Read file content
	_, err = file.Read(buffer)
	if err != nil {
		return nil, errors.Join(ErrCannotReadBinaryFile, err)
	}

	// Parse JSON
	var encodedPostedBlockCalldata []string
	err = json.Unmarshal(buffer, &encodedPostedBlockCalldata)
	if err != nil {
		return nil, errors.Join(ErrCannotParseJson, err)
	}

	postedBlockCalldata := make([][]byte, len(encodedPostedBlockCalldata))
	for i, encodedCalldata := range encodedPostedBlockCalldata {
		postedBlockCalldata[i], err = hexutil.Decode(encodedCalldata)
		if err != nil {
			return nil, errors.Join(ErrCannotDecodeHex, err)
		}
	}

	return postedBlockCalldata, nil
}
