package block_post_service_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/block_post_service"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

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
