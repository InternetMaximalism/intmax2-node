package block_post_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"io"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	scrollNetworkRpcUrl   = "https://sepolia-rpc.scroll.io"
	senderPublicKeysIndex = 5
	numOfSenders          = intMaxTypes.NumOfSenders

	postRegistrationBlockMethod    = "postRegistrationBlock"
	postNonRegistrationBlockMethod = "postNonRegistrationBlock"

	int0Key  = 0
	int1Key  = 1
	int2Key  = 2
	int3Key  = 3
	int4Key  = 4
	int5Key  = 5
	int6Key  = 6
	int8Key  = 8
	int16Key = 16
	int32Key = 32
)

type BlockPostService struct {
	ctx context.Context
	// cfg       *configs.Config
	// log       logger.Logger
	ethClient    *ethclient.Client
	scrollClient *ethclient.Client
	liquidity    *bindings.Liquidity
	rollup       *bindings.Rollup
}

func NewBlockPostService(ctx context.Context, cfg *configs.Config) (*BlockPostService, error) {
	ethClient, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, errors.Join(ErrNewEthereumClientFail, err)
	}
	defer ethClient.Close()

	var scrollClient *ethclient.Client
	scrollClient, err = utils.NewClient(scrollNetworkRpcUrl)
	if err != nil {
		return nil, errors.Join(ErrNewScrollClientFail, err)
	}
	defer scrollClient.Close()

	var liquidity *bindings.Liquidity
	liquidity, err = bindings.NewLiquidity(
		common.HexToAddress(cfg.Blockchain.LiquidityContractAddress),
		ethClient,
	)
	if err != nil {
		return nil, errors.Join(ErrInstantiateLiquidityContractFail, err)
	}

	var rollup *bindings.Rollup
	rollup, err = bindings.NewRollup(
		common.HexToAddress(cfg.Blockchain.RollupContractAddress),
		scrollClient,
	)
	if err != nil {
		return nil, errors.Join(ErrInstantiateRollupContractFail, err)
	}

	return &BlockPostService{
		ctx:          ctx,
		ethClient:    ethClient,
		scrollClient: scrollClient,
		liquidity:    liquidity,
		rollup:       rollup,
	}, nil
}

func (d *BlockPostService) FetchNewPostedBlocks(startBlock uint64) ([]*bindings.RollupBlockPosted, *big.Int, error) {
	nextBlock := startBlock + int1Key
	iterator, err := d.rollup.FilterBlockPosted(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: d.ctx,
	}, [][int32Key]byte{}, []common.Address{})
	if err != nil {
		return nil, nil, errors.Join(ErrFilterLogsFail, err)
	}

	defer func() {
		_ = iterator.Close()
	}()

	var events []*bindings.RollupBlockPosted
	maxBlockNumber := new(big.Int)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
		if event.BlockNumber.Cmp(maxBlockNumber) > int0Key {
			maxBlockNumber.Set(event.BlockNumber)
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, nil, errors.Join(ErrEncounteredWhileIterating, err)
	}

	return events, maxBlockNumber, nil
}

func (d *BlockPostService) FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error) {
	tx, isPending, err := d.scrollClient.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, errors.Join(ErrTransactionByHashNotFound, err)
	}

	if isPending {
		return nil, ErrTransactionIsStillPending
	}

	calldata := tx.Data()

	return calldata, nil
}

// FetchIntMaxBlockContentByCalldata fetches the block content by transaction hash.
//
// Example:
//
//	ctx := context.Background()
//	cfg := configs.Config{}
//	var startScrollBlockNumber uint64 = 0
//	d, err := block_post_service.NewBlockPostService(ctx, &cfg)
//	events, lastIntMaxBlockNumber, err := d.FetchNewPostedBlocks(startScrollBlockNumber)
//	calldata, err := d.FetchScrollCalldataByHash(events[0].Raw.TxHash)
//	blockContent, err := FetchIntMaxBlockContentByCalldata(calldata)
func FetchIntMaxBlockContentByCalldata(calldata []byte, accountInfoMap AccountInfoMap) (*intMaxTypes.BlockContent, error) {
	// Parse calldata
	rollupABI := io.Reader(strings.NewReader(bindings.RollupABI))
	parsedABI, err := abi.JSON(rollupABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	method, err := parsedABI.MethodById(calldata[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}

	blockContent, err := recoverBlockContent(method, calldata, accountInfoMap)
	if err != nil {
		return nil, errors.Join(ErrDecodeCallDataFail, err)
	}

	return blockContent, nil
}

type AccountInfoMap struct {
	LastAccountID uint64
	AccountMap    map[uint64]*intMaxAcc.PublicKey
	PublicKeyMap  map[string]uint64
}

func NewAccountInfoMap() AccountInfoMap {
	return AccountInfoMap{
		AccountMap:    make(map[uint64]*intMaxAcc.PublicKey),
		PublicKeyMap:  make(map[string]uint64),
		LastAccountID: int0Key,
	}
}

func recoverBlockContent(
	method *abi.Method,
	calldata []byte,
	accountInfoMap AccountInfoMap,
) (*intMaxTypes.BlockContent, error) {
	switch method.Name {
	case postRegistrationBlockMethod:
		decodedInput, err := decodePostRegistrationBlockCalldata(method, calldata)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		var blockContent *intMaxTypes.BlockContent
		blockContent, err = recoverRegistrationBlockContent(decodedInput)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		dummyAddress := intMaxAcc.NewDummyPublicKey().ToAddress().String()
		for _, sender := range blockContent.Senders {
			if accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] == int0Key {
				accountInfoMap.LastAccountID += int1Key
				accountInfoMap.AccountMap[accountInfoMap.LastAccountID] = sender.PublicKey
				accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] = accountInfoMap.LastAccountID
			} else {
				// TODO: error handling
				address := sender.PublicKey.ToAddress().String()
				if address != dummyAddress {
					fmt.Printf("Account already exists %v\n", address)
				}
			}
		}

		return blockContent, nil
	case postNonRegistrationBlockMethod:
		decodedInput, err := decodePostNonRegistrationBlockCalldata(method, calldata)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		var blockContent *intMaxTypes.BlockContent
		blockContent, err = recoverNonRegistrationBlockContent(decodedInput, accountInfoMap)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		/**
		// TODO
		// for _, sender := range blockContent.Senders {
		// 	if accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] == int0Key {
		// 		accountInfoMap.LastAccountID += int1Key
		// 		accountInfoMap.AccountMap[accountInfoMap.LastAccountID] = sender.PublicKey
		// 		accountInfoMap.PublicKeyMap[sender.PublicKey.ToAddress().String()] = accountInfoMap.LastAccountID
		// 	} else {
		// 		// TODO: error handling
		// 		if
		// 		fmt.Println("Account already exists")
		// 	}
		// }
		*/

		return blockContent, nil
	default:
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}
}

func decodePostRegistrationBlockCalldata(
	method *abi.Method,
	calldata []byte,
) (*intMaxTypes.PostRegistrationBlockInput, error) {
	if method.Name != postRegistrationBlockMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[int4Key:])
	if err != nil {
		return nil, errors.Join(ErrUnpackCalldataFail, err)
	}

	decodedInput := intMaxTypes.PostRegistrationBlockInput{
		TxTreeRoot:          args[int0Key].([int32Key]byte),
		SenderFlags:         args[int1Key].([int16Key]byte),
		AggregatedPublicKey: args[int2Key].([int2Key][int32Key]byte),
		AggregatedSignature: args[int3Key].([int4Key][int32Key]byte),
		MessagePoint:        args[int4Key].([int4Key][int32Key]byte),
		SenderPublicKeys:    args[int5Key].([]*big.Int),
	}

	return &decodedInput, nil
}

func decodePostNonRegistrationBlockCalldata(
	method *abi.Method, calldata []byte,
) (*intMaxTypes.PostNonRegistrationBlockInput, error) {
	if method.Name != postNonRegistrationBlockMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(calldata[int4Key:])
	if err != nil {
		return nil, errors.Join(ErrUnpackCalldataFail, err)
	}

	decodedInput := intMaxTypes.PostNonRegistrationBlockInput{
		TxTreeRoot:          args[int0Key].([int32Key]byte),
		SenderFlags:         args[int1Key].([int16Key]byte),
		AggregatedPublicKey: args[int2Key].([int2Key][int32Key]byte),
		AggregatedSignature: args[int3Key].([int4Key][int32Key]byte),
		MessagePoint:        args[int4Key].([int4Key][int32Key]byte),
		PublicKeysHash:      args[int5Key].([int32Key]byte),
		SenderAccountIds:    args[int6Key].([]byte),
	}

	return &decodedInput, nil
}

func recoverRegistrationBlockContent(
	decodedInput *intMaxTypes.PostRegistrationBlockInput,
) (_ *intMaxTypes.BlockContent, err error) {
	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, key := range decodedInput.SenderPublicKeys {
		senderPublicKeys[i], err = intMaxAcc.NewPublicKeyFromAddressInt(key)
		if err != nil {
			return nil, errors.Join(ErrCannotDecodeAddress, err)
		}
	}
	for i := len(decodedInput.SenderPublicKeys); i < numOfSenders; i++ {
		senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
	}

	senderFlags := make([]bool, numOfSenders)
	for i := int0Key; i < numOfSenders; i++ {
		byteIndex := i / int8Key
		bitIndex := i % int8Key
		senderFlags[i] = (decodedInput.SenderFlags[byteIndex] & (int1Key << bitIndex)) != int0Key
	}

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		if senderFlags[i] {
			senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+int1Key)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, isSigned := range senderFlags {
		if isSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, senderPublicKeys[i].WeightByHash(publicKeysHash))
		}
	}

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, sender := range senderPublicKeys {
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: int0Key,
			IsSigned:  senderFlags[i],
		}
	}

	txTreeRoot := new(intMaxTypes.PoseidonHashOut)
	if err = txTreeRoot.Unmarshal(decodedInput.TxTreeRoot[:]); err != nil {
		return nil, errors.Join(ErrSetTxRootFail, err)
	}

	// Recover aggregatedSignature from decodedInput.AggregatedSignature
	aggregatedSignature := new(bn254.G2Affine)
	aggregatedSignature.X.A1.SetBytes(decodedInput.AggregatedSignature[int0Key][:])
	aggregatedSignature.X.A0.SetBytes(decodedInput.AggregatedSignature[int1Key][:])
	aggregatedSignature.Y.A1.SetBytes(decodedInput.AggregatedSignature[int2Key][:])
	aggregatedSignature.Y.A0.SetBytes(decodedInput.AggregatedSignature[int3Key][:])

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		*txTreeRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

func recoverNonRegistrationBlockContent(
	decodedInput *intMaxTypes.PostNonRegistrationBlockInput,
	accountInfoMap AccountInfoMap,
) (*intMaxTypes.BlockContent, error) {
	senderAccountIds, err := intMaxTypes.UnmarshalAccountIds(decodedInput.SenderAccountIds)
	if err != nil {
		return nil, errors.Join(ErrRecoverAccountIDsFromBytesFail, err)
	}

	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, accountId := range senderAccountIds {
		if accountId == int0Key {
			senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
			continue
		}

		key, ok := accountInfoMap.AccountMap[accountId]
		if !ok {
			return nil, errors.Join(ErrUnknownAccountID, fmt.Errorf("%d", accountId))
		}

		senderPublicKeys[i] = key
	}
	for i := len(senderAccountIds); i < numOfSenders; i++ {
		senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
	}

	senderFlags := make([]bool, numOfSenders)
	for i := int0Key; i < numOfSenders; i++ {
		byteIndex := i / int8Key
		bitIndex := i % int8Key
		senderFlags[i] = (decodedInput.SenderFlags[byteIndex] & (int1Key << bitIndex)) != int0Key
	}

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		if senderFlags[i] {
			senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+int1Key)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, isSigned := range senderFlags {
		if isSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, senderPublicKeys[i].WeightByHash(publicKeysHash))
		}
	}

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, sender := range senderPublicKeys {
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: int0Key,
			IsSigned:  senderFlags[i],
		}
	}

	txTreeRoot := new(intMaxTypes.PoseidonHashOut)
	if err = txTreeRoot.Unmarshal(decodedInput.TxTreeRoot[:]); err != nil {
		return nil, errors.Join(ErrSetTxRootFail, err)
	}

	// Recover aggregatedSignature from decodedInput.AggregatedSignature
	aggregatedSignature := new(bn254.G2Affine)
	aggregatedSignature.X.A1.SetBytes(decodedInput.AggregatedSignature[int0Key][:])
	aggregatedSignature.X.A0.SetBytes(decodedInput.AggregatedSignature[int1Key][:])
	aggregatedSignature.Y.A1.SetBytes(decodedInput.AggregatedSignature[int2Key][:])
	aggregatedSignature.Y.A0.SetBytes(decodedInput.AggregatedSignature[int3Key][:])

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		*txTreeRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

// func (d *BlockPostService) FetchNewDeposits(startBlock uint64) ([]*bindings.LiquidityDeposited, *big.Int, map[uint32]bool, error) {
// 	nextBlock := startBlock + 1
// 	iterator, err := d.liquidity.FilterDeposited(&bind.FilterOpts{
// 		Start:   nextBlock,
// 		End:     nil,
// 		Context: d.ctx,
// 	}, []*big.Int{}, []common.Address{})
// 	if err != nil {
// 		return nil, nil, nil, errors.Join(ErrFilterLogsFail, err)
// 	}

// 	defer iterator.Close()

// 	var events []*bindings.LiquidityDeposited
// 	maxDepositIndex := new(big.Int)
// 	tokenIndexMap := make(map[uint32]bool)

// 	for iterator.Next() {
// 		event := iterator.Event
// 		events = append(events, event)
// 		tokenIndexMap[event.TokenIndex] = true
// 		if event.DepositId.Cmp(maxDepositIndex) > 0 {
// 			maxDepositIndex.Set(event.DepositId)
// 		}
// 	}

// 	if err = iterator.Error(); err != nil {
// 		return nil, nil, nil, errors.Join(ErrEncounteredWhileIterating, err)
// 	}

// 	return events, maxDepositIndex, tokenIndexMap, nil
// }