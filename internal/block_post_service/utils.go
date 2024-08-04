package block_post_service

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"io"
	"math/big"
	"sort"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
//	 ai := NewAccountInfo(dbApp)
//		blockContent, err := FetchIntMaxBlockContentByCalldata(calldata, ai)
func FetchIntMaxBlockContentByCalldata(
	calldata []byte,
	ai AccountInfo,
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

// MakeRegistrationBlock creates a block content for registration block.
// txRoot - root of the transaction tree.
// senderPublicKeys - list of public keys for each sender.
// signatures - list of signatures for each sender. Empty string means no signature.
func MakeRegistrationBlock(
	txRoot intMaxTypes.PoseidonHashOut,
	senderPublicKeys []*intMaxAcc.PublicKey,
	signatures []string,
) (*intMaxTypes.BlockContent, error) {
	if len(senderPublicKeys) != len(signatures) {
		return nil, errors.New("length of senderPublicKeys, accountIDs, and signatures must be equal")
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, publicKey := range senderPublicKeys {
		if publicKey == nil {
			return nil, errors.New("publicKey must not be nil")
		}

		senders[i] = intMaxTypes.Sender{
			PublicKey: publicKey,
			AccountID: 0,
			IsSigned:  signatures[i] != "",
		}
	}

	defaultSender := intMaxTypes.NewDummySender()
	for i := len(senderPublicKeys); i < len(senders); i++ {
		senders[i] = defaultSender
	}

	const numPublicKeyBytes = intMaxTypes.NumPublicKeyBytes

	senderPublicKeysBytes := make([]byte, len(senders)*numPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[numPublicKeyBytes*i:numPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			weightedPublicKey := sender.PublicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggregatedSignature := new(bn254.G2Affine)
	for i, sender := range senders {
		if senders[i].IsSigned {
			if sender.IsSigned {
				signature := new(bn254.G2Affine)
				signatureBytes, err := hexutil.Decode(signatures[i])
				if err != nil {
					fmt.Printf("Failed to decode signature: %s\n", signatures[i])
					continue
				}
				err = signature.Unmarshal(signatureBytes)
				if err != nil {
					fmt.Printf("Failed to unmarshal signature: %s\n", signatures[i])
					continue
				}

				err = VerifyTxTreeSignature(
					signatureBytes, sender.PublicKey, txRoot.Marshal(), senderPublicKeys,
				)
				if err != nil {
					fmt.Printf("Failed to verify signature: %s\n", signatures[i])
					continue
				}

				aggregatedSignature.Add(aggregatedSignature, signature)
			}
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		txRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

// MakeNonRegistrationBlock creates a block content for non-registration block.
// txRoot - root of the transaction tree.
// accountIDs - list of account IDs for each sender.
// senderPublicKeys - list of public keys for each sender.
// signatures - list of signatures for each sender. Empty string means no signature.
func MakeNonRegistrationBlock(
	txRoot intMaxTypes.PoseidonHashOut,
	accountIDs []uint64,
	senderPublicKeys []*intMaxAcc.PublicKey,
	signatures []string,
) (*intMaxTypes.BlockContent, error) {
	if len(senderPublicKeys) != len(signatures) || len(senderPublicKeys) != len(accountIDs) {
		return nil, errors.New("length of senderPublicKeys, accountIDs, and signatures must be equal")
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	const maxAccountIDBits = 40

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, publicKey := range senderPublicKeys {
		if accountIDs[i] == 0 {
			return nil, errors.New("accountID must be greater than 0")
		}
		if accountIDs[i] > uint64(1)<<maxAccountIDBits {
			return nil, fmt.Errorf("accountID must be less than or equal to 2^%d", maxAccountIDBits)
		}
		if publicKey == nil {
			return nil, errors.New("publicKey must not be nil")
		}

		senders[i] = intMaxTypes.Sender{
			PublicKey: publicKey,
			AccountID: accountIDs[i],
			IsSigned:  signatures[i] != "",
		}
	}

	defaultPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(senderPublicKeys); i < len(senders); i++ {
		senders[i] = intMaxTypes.Sender{
			PublicKey: defaultPublicKey,
			AccountID: 1,
			IsSigned:  false,
		}
	}

	const numPublicKeyBytes = intMaxTypes.NumPublicKeyBytes

	senderPublicKeysBytes := make([]byte, len(senders)*numPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[numPublicKeyBytes*i:numPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			weightedPublicKey := sender.PublicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggregatedSignature := new(bn254.G2Affine)
	for i, sender := range senders {
		if senders[i].IsSigned {
			if sender.IsSigned {
				signature := new(bn254.G2Affine)
				signatureBytes, err := hexutil.Decode(signatures[i])
				if err != nil {
					fmt.Printf("Failed to decode signature: %s\n", signatures[i])
					continue
				}
				err = signature.Unmarshal(signatureBytes)
				if err != nil {
					fmt.Printf("Failed to unmarshal signature: %s\n", signatures[i])
					continue
				}

				err = VerifyTxTreeSignature(
					signatureBytes, sender.PublicKey, txRoot.Marshal(), senderPublicKeys,
				)
				if err != nil {
					fmt.Printf("Failed to verify signature: %s\n", signatures[i])
					continue
				}

				aggregatedSignature.Add(aggregatedSignature, signature)
			}
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		txRoot,
		aggregatedSignature,
	)

	return blockContent, nil
}

func VerifyTxTreeSignature(signatureBytes []byte, sender *intMaxAcc.PublicKey, txTreeRootBytes []byte, senderPublicKeys []*intMaxAcc.PublicKey) error {
	const int32Key = 32

	if len(senderPublicKeys) == 0 {
		return ErrInvalidSendersLength
	}
	if len(senderPublicKeys) > intMaxTypes.NumOfSenders {
		return ErrTooManySenderPublicKeys
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
	}
	defaultPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(senderPublicKeys); i < intMaxTypes.NumOfSenders; i++ {
		senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)

	messagePoint := finite_field.BytesToFieldElementSlice(txTreeRootBytes)

	signature := new(bn254.G2Affine)
	err := signature.Unmarshal(signatureBytes)
	if err != nil {
		return errors.Join(ErrUnmarshalSignatureFail, err)
	}

	senderWithWeight := sender.WeightByHash(publicKeysHash)
	err = intMaxAcc.VerifySignature(signature, senderWithWeight, messagePoint)
	if err != nil {
		return errors.Join(ErrInvalidSignature, err)
	}

	return nil
}

func recoverBlockContent(
	method *abi.Method,
	calldata []byte,
	ai AccountInfo,
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

		defaultAddress := intMaxAcc.NewDummyPublicKey().ToAddress().String()
		for index := range blockContent.Senders {
			address := blockContent.Senders[index].PublicKey.ToAddress().String()
			if !strings.EqualFold(address, defaultAddress) {
				err = ai.RegisterPublicKey(blockContent.Senders[index].PublicKey)
				if err != nil {
					return nil, errors.Join(ErrRegisterPublicKeyFail, err)
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
		blockContent, err = recoverNonRegistrationBlockContent(decodedInput, ai)
		if err != nil {
			return nil, errors.Join(ErrDecodeCallDataFail, err)
		}

		err = blockContent.IsValid()
		if err != nil {
			return nil, fmt.Errorf("failed to validate block content: %w", err)
		}

		defaultAddress := intMaxAcc.NewDummyPublicKey().ToAddress().String()
		for index := range blockContent.Senders {
			address := blockContent.Senders[index].PublicKey.ToAddress().String()
			if !strings.EqualFold(address, defaultAddress) {
				err = ai.RegisterPublicKey(blockContent.Senders[index].PublicKey)
				if err != nil {
					return nil, errors.Join(ErrRegisterPublicKeyFail, err)
				}
			}
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
	ai AccountInfo,
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

func updateEventBlockNumber(db SQLDriverApp, log logger.Logger, eventName string, blockNumber uint64) error {
	updatedEvent, err := db.UpsertEventBlockNumber(eventName, blockNumber)
	if err != nil {
		return err
	}
	log.Infof("Updated %s block number to %d", eventName, updatedEvent.LastProcessedBlockNumber)
	return nil
}
