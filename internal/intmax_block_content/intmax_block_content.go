package intmax_block_content

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	intMaxTypes "intmax2-node/internal/types"
	"io"
	"math/big"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	defaultAccountID = 0
	dummyAccountID   = 1

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

	numOfSenders = intMaxTypes.NumOfSenders

	postRegistrationBlockMethod    = "postRegistrationBlock"
	postNonRegistrationBlockMethod = "postNonRegistrationBlock"
)

// FetchIntMaxBlockContentByCalldata fetches the block content by transaction hash.
// accountInfo is mutable and will be updated with new account information.
//
// Example:
//
//		ctx := context.Background()
//		cfg := configs.Config{}
//		var startScrollBlockNumber uint64 = 0
//		d, err := block_post_service.NewBlockPostService(ctx, &cfg)
//		events, lastIntMaxBlockNumber, err := d.FetchNewPostedBlocks(startScrollBlockNumber)
//		calldata, err := d.FetchScrollCalldataByHash(events[0].Raw.TxHash)
//	    ai := NewAccountInfo(dbApp)
//		blockContent, err := FetchIntMaxBlockContentByCalldata(calldata, ai)
func FetchIntMaxBlockContentByCalldata(
	callData []byte,
	postedBlock *PostedBlock,
	ai AccountInfo,
) (*intMaxTypes.BlockContent, error) {
	parsedABI, err := abi.JSON(io.Reader(strings.NewReader(bindings.RollupMetaData.ABI)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	var method *abi.Method
	method, err = parsedABI.MethodById(callData[:int4Key])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}

	var blockContent *intMaxTypes.BlockContent
	blockContent, err = recoverBlockContent(method, callData, ai, postedBlock.BlockNumber)
	if err != nil {
		return nil, errors.Join(ErrDecodeCallDataFail, err)
	}

	return blockContent, nil
}

func recoverBlockContent(
	method *abi.Method,
	callData []byte,
	ai AccountInfo,
	intMaxBlockNumber uint32,
) (*intMaxTypes.BlockContent, error) {
	switch method.Name {
	case postRegistrationBlockMethod:
		decodedInput, err := decodePostRegistrationBlockCalldata(method, callData)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		var blockContent *intMaxTypes.BlockContent
		blockContent, err = recoverRegistrationBlockContent(decodedInput)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		return blockContent, nil
	case postNonRegistrationBlockMethod:
		decodedInput, err := decodePostNonRegistrationBlockCalldata(method, callData)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		var blockContent *intMaxTypes.BlockContent
		blockContent, err = recoverNonRegistrationBlockContent(decodedInput, ai, intMaxBlockNumber)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		return blockContent, nil
	default:
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}
}

func decodePostRegistrationBlockCalldata(
	method *abi.Method,
	callData []byte,
) (*intMaxTypes.PostRegistrationBlockInput, error) {
	if method.Name != postRegistrationBlockMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(callData[int4Key:])
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
	method *abi.Method, callData []byte,
) (*intMaxTypes.PostNonRegistrationBlockInput, error) {
	if method.Name != postNonRegistrationBlockMethod {
		return nil, fmt.Errorf(ErrMethodNameInvalidStr, method.Name)
	}

	args, err := method.Inputs.Unpack(callData[int4Key:])
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

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, sender := range senderPublicKeys {
		var accountID uint64 = defaultAccountID
		if sender.Equal(dummyPublicKey) {
			accountID = dummyAccountID
		}
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: accountID,
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

	txRootBytes := [int32Key]byte{}
	copy(txRootBytes[:], txTreeRoot.Marshal())

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		txRootBytes,
		aggregatedSignature,
	)

	return blockContent, nil
}

func recoverNonRegistrationBlockContent(
	decodedInput *intMaxTypes.PostNonRegistrationBlockInput,
	ai AccountInfo,
	blockNumber uint32,
) (*intMaxTypes.BlockContent, error) {
	senderAccountIds, err := intMaxTypes.UnmarshalAccountIds(decodedInput.SenderAccountIds)
	if err != nil {
		return nil, errors.Join(ErrRecoverAccountIDsFromBytesFail, err)
	}

	if blockNumber == 0 {
		panic("block number 0 is not published")
	}

	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, accountId := range senderAccountIds {
		if accountId == 0 {
			panic("account ID is 0")
			// senderAddresses[i], err = intMaxAcc.NewAddressFromAddressInt(big.NewInt(0))
			// continue
		}

		var pk *intMaxAcc.PublicKey
		pk, err = ai.PublicKeyByAccountID(blockNumber-1, accountId)
		if err != nil {
			return nil, errors.Join(ErrUnknownAccountID, fmt.Errorf("account %d is invalid: %w", accountId, err))
		}

		senderPublicKeys[i] = pk
	}
	for i := len(senderAccountIds); i < numOfSenders; i++ {
		senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
	}

	senderFlags := make([]bool, numOfSenders)
	for i := 0; i < numOfSenders; i++ {
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
	for i, senderAccountId := range senderAccountIds {
		senders[i] = intMaxTypes.Sender{
			PublicKey: senderPublicKeys[i],
			AccountID: senderAccountId,
			IsSigned:  senderFlags[i],
		}
	}
	for i := len(senderAccountIds); i < numOfSenders; i++ {
		senders[i] = intMaxTypes.Sender{
			PublicKey: intMaxAcc.NewDummyPublicKey(),
			AccountID: dummyAccountID,
			IsSigned:  false,
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

	txRootBytes := [int32Key]byte{}
	copy(txRootBytes[:], txTreeRoot.Marshal())

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		txRootBytes,
		aggregatedSignature,
	)

	return blockContent, nil
}