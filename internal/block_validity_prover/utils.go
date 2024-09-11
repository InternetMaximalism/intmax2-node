package block_validity_prover

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
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
	calldata []byte,
	postedBlock *block_post_service.PostedBlock,
	ai block_post_service.AccountInfo,
) (*intMaxTypes.BlockContent, error) {
	// Parse calldata
	rollupABI := io.Reader(strings.NewReader(bindings.RollupMetaData.ABI))
	parsedABI, err := abi.JSON(rollupABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	method, err := parsedABI.MethodById(calldata[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to parse calldata: %w", err)
	}

	blockContent, err := recoverBlockContent(method, calldata, ai)
	if err != nil {
		return nil, errors.Join(ErrDecodeCallDataFail, err)
	}

	return blockContent, nil
}

func recoverBlockContent(
	method *abi.Method,
	calldata []byte,
	ai block_post_service.AccountInfo,
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

		err = blockContent.IsValid()
		if err != nil {
			return nil, fmt.Errorf("failed to validate block content: %w", err)
		}

		return blockContent, nil
	case postNonRegistrationBlockMethod:
		decodedInput, err := decodePostNonRegistrationBlockCalldata(method, calldata)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		var blockContent *intMaxTypes.BlockContent
		blockContent, err = recoverNonRegistrationBlockContent(decodedInput, ai)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		err = blockContent.IsValid()
		if err != nil {
			return nil, fmt.Errorf("failed to validate block content: %w", err)
		}

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
	ai block_post_service.AccountInfo,
) (*intMaxTypes.BlockContent, error) {
	senderAccountIds, err := intMaxTypes.UnmarshalAccountIds(decodedInput.SenderAccountIds)
	if err != nil {
		return nil, errors.Join(ErrRecoverAccountIDsFromBytesFail, err)
	}

	senderPublicKeys := make([]*intMaxAcc.PublicKey, numOfSenders)
	for i, accountId := range senderAccountIds {
		if accountId == 0 {
			senderPublicKeys[i] = intMaxAcc.NewDummyPublicKey()
			continue
		}

		var pk *intMaxAcc.PublicKey
		pk, err = ai.PublicKeyByAccountID(accountId)
		if err != nil {
			return nil, errors.Join(ErrUnknownAccountID, fmt.Errorf("%d", accountId))
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
	for i, sender := range senderPublicKeys {
		senders[i] = intMaxTypes.Sender{
			PublicKey: sender,
			AccountID: senderAccountIds[i],
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
