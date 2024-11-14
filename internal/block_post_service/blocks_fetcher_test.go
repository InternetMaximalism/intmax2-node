package block_post_service_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/intmax_block_content"
	"intmax2-node/pkg/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAccountInfoMap(t *testing.T) {
	const int2Key = 2
	assert.NoError(t, configs.LoadDotEnv(int2Key))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	cfg.Blockchain.EthereumNetworkRpcUrl = "https://eth-sepolia.g.alchemy.com/v2/OE-Ocf1AKEHq5UlRKEXGYJ6mc7dJamOV"
	cfg.Blockchain.LiquidityContractAddress = "0x86f08b1DcDe0673A5562384eAFD5df6EE83cd73a"
	cfg.Blockchain.RollupContractAddress = "0x33a463381140C97B3bFd41eAADAE38ee98fe410c"
	cfg.SQLDb.DNSConnection = "postgresql://postgres:pass@intmax2-node-postgres:5432/state?sslmode=disable"
	cfg.Blockchain.RollupContractDeployedBlockNumber = 5843354

	dbApp := NewMockSQLDriverApp(ctrl)
	dbApp.EXPECT().DelAllAccounts().AnyTimes()
	dbApp.EXPECT().ResetSequenceByAccounts().AnyTimes()
	dbApp.EXPECT().UpsertEventBlockNumbersErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	dbApp.EXPECT().EventBlockNumberByEventName(gomock.Any()).AnyTimes()
	dbApp.EXPECT().UpsertEventBlockNumber(gomock.Any(), gomock.Any()).Return(&mDBApp.EventBlockNumber{
		EventName:                mDBApp.BlockPostedEvent,
		LastProcessedBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
	}, nil).AnyTimes()
	dbApp.EXPECT().SenderByAddress(gomock.Any()).Return(&mDBApp.Sender{
		ID:        uuid.New().String(),
		Address:   "0x",
		PublicKey: "0x",
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()
	dbApp.EXPECT().AccountBySenderID(gomock.Any()).Return(&mDBApp.Account{
		ID:        uuid.New().String(),
		AccountID: new(uint256.Int).SetUint64(uint64(1)),
		SenderID:  uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()
	dbApp.EXPECT().UpdateBlockStatus(gomock.Any().String(), gomock.Any().String(), uint32(1)).Return(nil).AnyTimes()
	dbApp.EXPECT().GetUnprocessedBlocks().Return(nil, nil).AnyTimes()

	lg := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	bps, err := block_synchronizer.NewBlockSynchronizer(ctx, cfg, lg)
	assert.NoError(t, err)

	assert.NoError(t, block_post_service.ProcessingPostedBlocks(ctx, cfg, lg, dbApp, bps))
}

func TestFetchNewPostedBlocks(t *testing.T) {
	calldataJson, err := readPostedBlockEventsJson("../../pkg/data/posted_block_calldata.json")
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbApp := NewMockSQLDriverApp(ctrl)
	dbApp.EXPECT().AccountByAccountID(gomock.Any()).Return(&mDBApp.Account{
		ID:        uuid.New().String(),
		AccountID: new(uint256.Int).SetUint64(uint64(1)),
		SenderID:  uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()
	dbApp.EXPECT().SenderByID(gomock.Any()).Return(&mDBApp.Sender{
		ID:        uuid.New().String(),
		Address:   "0x",
		PublicKey: "0x",
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()
	dbApp.EXPECT().SenderByAddress(gomock.Any()).Return(&mDBApp.Sender{
		ID:        uuid.New().String(),
		Address:   "0x",
		PublicKey: "0x",
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()
	dbApp.EXPECT().AccountBySenderID(gomock.Any()).Return(&mDBApp.Account{
		ID:        uuid.New().String(),
		AccountID: new(uint256.Int).SetUint64(uint64(1)),
		SenderID:  uuid.New().String(),
		CreatedAt: time.Now().UTC(),
	}, nil).AnyTimes()

	prevBlockHash := common.Hash{}
	depositRoot := common.Hash{}
	signatureHash := common.Hash{}
	postedBlock := intmax_block_content.NewPostedBlock(prevBlockHash, depositRoot, uint32(0), signatureHash)

	accountInfoMap := block_post_service.NewAccountInfo(dbApp)
	for i, calldata := range calldataJson {
		_, err = intmax_block_content.FetchIntMaxBlockContentByCalldata(calldata, postedBlock, accountInfoMap)
		if err != nil {
			if errors.Is(err, block_validity_prover.ErrUnknownAccountID) {
				fmt.Printf("block %d is ErrUnknownAccountID\n", i)
				continue
			}
			if errors.Is(err, block_validity_prover.ErrCannotDecodeAddress) {
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
