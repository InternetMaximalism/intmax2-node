package types

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/pkg/utils"
	"math/big"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prodadidb/go-validation"
)

const (
	NumPublicKeyBytes   = 32
	PublicKeySenderType = "PUBLIC_KEY"

	NumAccountIDBytes   = 5
	AccountIDSenderType = "ACCOUNT_ID"

	NumOfSenders    = 128
	numFlagBytes    = 16
	numG2PointLimbs = 4
	int8Key         = 8
	int32Key        = 32
)

type PoseidonHashOut = goldenposeidon.PoseidonHashOut

// Sender represents an individual sender's details, including their public key, account ID,
// and a flag indicating if the sender has posted.
type Sender struct {
	PublicKey *accounts.PublicKey `json:"publicKey"`
	AccountID uint64              `json:"accountId"`
	IsSigned  bool                `json:"isSigned"`
}

type ColumnSender struct {
	PublicKey string `json:"publicKey"`
	AccountID uint64 `json:"accountId"`
	IsSigned  bool   `json:"isSigned"`
}

// BlockContent represents the content of a block, including sender details, transaction root,
// aggregated signature, and public key.
type BlockContent struct {
	// SenderType specifies whether senders are identified by PUBLIC_KEY or ACCOUNT_ID
	SenderType string

	// Senders is a list of senders in the block
	Senders []Sender

	// TxRoot is the root hash of the transactions in the block
	TxTreeRoot PoseidonHashOut

	// AggregatedSignature is the aggregated signature of the block
	AggregatedSignature *bn254.G2Affine

	// aggregatedPublicKey is the aggregated public key of the block
	AggregatedPublicKey *accounts.PublicKey

	MessagePoint *bn254.G2Affine
}

func NewBlockContent(
	senderType string,
	senders []Sender,
	txTreeRoot PoseidonHashOut,
	aggregatedSignature *bn254.G2Affine,
) *BlockContent {
	var bc BlockContent

	if senderType != PublicKeySenderType && senderType != AccountIDSenderType {
		return nil
	}
	bc.SenderType = senderType
	bc.Senders = make([]Sender, len(senders))
	copy(bc.Senders, senders)
	bc.TxTreeRoot.Set(&txTreeRoot)
	bc.AggregatedSignature = new(bn254.G2Affine).Set(aggregatedSignature)

	defaultPublicKey := accounts.NewDummyPublicKey()

	const numOfSenders = 128
	senderPublicKeys := make([]byte, numOfSenders*NumPublicKeyBytes)
	for i, sender := range bc.Senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
	}
	for i := len(bc.Senders); i < numOfSenders; i++ {
		senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)

	aggregatedPublicKey := new(accounts.PublicKey)
	for _, sender := range bc.Senders {
		if sender.IsSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, sender.PublicKey.WeightByHash(publicKeysHash))
		}
	}
	bc.AggregatedPublicKey = new(accounts.PublicKey).Set(aggregatedPublicKey)

	messagePoint := goldenposeidon.HashToG2(finite_field.BytesToFieldElementSlice(bc.TxTreeRoot.Marshal()))
	bc.MessagePoint = &messagePoint

	return &bc
}

func (bc *BlockContent) IsValid() error {
	const (
		int0Key   = 0
		int1Key   = 1
		int128Key = 128
	)

	return validation.ValidateStruct(bc,
		validation.Field(&bc.SenderType,
			validation.Required.Error(ErrBlockContentSenderTypeInvalid.Error()),
			validation.In(PublicKeySenderType, AccountIDSenderType).Error(ErrBlockContentSenderTypeInvalid.Error())),
		validation.Field(&bc.Senders,
			validation.Required.Error(ErrBlockContentSendersEmpty.Error()),
			validation.By(func(value interface{}) error {
				v, ok := value.([]Sender)
				if !ok {
					return ErrValueInvalid
				}

				if len(v) > int128Key {
					return ErrBlockContentManySenders
				}

				for i := int0Key; i < len(v)-int1Key; i++ {
					if v[i+int1Key].PublicKey.Pk.X.Cmp(&v[i].PublicKey.Pk.X) > int0Key {
						return ErrBlockContentPublicKeyNotSorted
					}
				}

				return nil
			}),
			validation.Each(validation.Required, validation.By(func(value interface{}) error {
				v, ok := value.(Sender)
				if !ok {
					return ErrValueInvalid
				}

				switch bc.SenderType {
				case PublicKeySenderType:
					if v.PublicKey == nil {
						return ErrBlockContentPublicKeyInvalid
					}

					if v.AccountID != int0Key {
						return ErrBlockContentAccIDForPubKeyInvalid
					}
				case AccountIDSenderType:
					if v.PublicKey == nil {
						return ErrBlockContentPublicKeyInvalid
					}

					if v.AccountID == int0Key && v.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) != int0Key {
						return ErrBlockContentAccIDForAccIDEmpty
					}
					if v.AccountID != int0Key && v.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) == int0Key {
						return ErrBlockContentAccIDForDefAccNotEmpty
					}
				}

				return nil
			}))),
		validation.Field(&bc.AggregatedPublicKey,
			validation.By(func(value interface{}) error {
				var isNil bool
				value, isNil = validation.Indirect(value)
				if isNil || validation.IsEmpty(value) {
					return ErrBlockContentAggPubKeyEmpty
				}

				defaultPublicKey := accounts.NewDummyPublicKey()

				const numOfSenders = 128
				senderPublicKeysBytes := make([]byte, numOfSenders*NumPublicKeyBytes)
				for key := range bc.Senders {
					senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
					copy(
						senderPublicKeysBytes[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)],
						senderPublicKey[:],
					)
				}
				for i := len(bc.Senders); i < numOfSenders; i++ {
					senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
					copy(senderPublicKeysBytes[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
				}

				publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
				aggregatedPublicKey := new(accounts.PublicKey)
				for key := range bc.Senders {
					if bc.Senders[key].IsSigned {
						aggregatedPublicKey.Add(
							aggregatedPublicKey,
							bc.Senders[key].PublicKey.WeightByHash(publicKeysHash),
						)
					}
				}

				if !aggregatedPublicKey.Equal(bc.AggregatedPublicKey) {
					return ErrBlockContentAggPubKeyInvalid
				}

				return nil
			}),
		),
		validation.Field(&bc.AggregatedSignature,
			validation.By(func(value interface{}) error {
				var isNil bool
				value, isNil = validation.Indirect(value)
				if isNil || validation.IsEmpty(value) {
					return ErrBlockContentAggSignEmpty
				}

				message := finite_field.BytesToFieldElementSlice(bc.TxTreeRoot.Marshal())
				err := accounts.VerifySignature(bc.AggregatedSignature, bc.AggregatedPublicKey, message)
				if err != nil {
					return err
				}

				return nil
			}),
		),
	)
}

func (bc *BlockContent) Marshal() []byte {
	const (
		int0Key = 0
		int1Key = 1
	)

	var data []byte
	if bc.SenderType == PublicKeySenderType {
		data = append(data, int0Key)
	} else {
		data = append(data, int1Key)
	}
	data = append(data, bc.TxTreeRoot.Marshal()...)

	// TODO: need check
	for key := range bc.Senders {
		if bc.Senders[key].IsSigned {
			data = append(data, int1Key)
		} else {
			data = append(data, int0Key)
		}
	}

	senderAccountIDs := make([]byte, len(bc.Senders)*NumAccountIDBytes)
	for key := range bc.Senders {
		var senderAccountId []byte
		if bc.SenderType == AccountIDSenderType {
			publicKeyX := bc.Senders[key].PublicKey.Pk.X.Bytes() // TODO: Use account ID
			senderAccountId = publicKeyX[:NumAccountIDBytes]
		} else {
			senderAccountId = []byte{int0Key, int0Key, int0Key, int0Key, int0Key}
		}
		copy(senderAccountIDs[NumAccountIDBytes*key:NumAccountIDBytes*(key+int1Key)], senderAccountId)
	}

	senderPublicKeys := make([]byte, len(bc.Senders)*NumPublicKeyBytes)
	for key := range bc.Senders {
		senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)], senderPublicKey[:])
	}

	data = append(data, senderAccountIDs...)
	data = append(data, senderPublicKeys...)
	data = append(data, bc.AggregatedSignature.Marshal()...)

	return data
}

// The rollup's calldata consists of txRoot, messagePoint, aggregatedSignature, aggregatedPublicKey,
// accountIdsHash, senderPublicKeysHash, senderFlags and senderType.
// The size of the rollup data will be 32 + 128 + 128 + 64 + 32 + 32 + 16 + 1 = 433 bytes.
func (bc *BlockContent) Rollup() []byte {
	const (
		int0Key = 0
		int1Key = 1
		int8Key = 8
	)

	var data []byte
	data = append(data, bc.TxTreeRoot.Marshal()...)
	data = append(data, bc.MessagePoint.Marshal()...)
	data = append(data, bc.AggregatedSignature.Marshal()...)
	data = append(data, bc.AggregatedPublicKey.Marshal()...)

	switch bc.SenderType {
	case PublicKeySenderType:
		senderPublicKeys := make([]byte, len(bc.Senders)*NumPublicKeyBytes)
		for key := range bc.Senders {
			senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)], senderPublicKey[:])
		}
		data = append(data, senderPublicKeys...)
	case AccountIDSenderType:
		senderAccountIDs := make([]byte, len(bc.Senders)*NumAccountIDBytes)
		for key := range bc.Senders {
			var senderAccountId []byte
			if bc.SenderType == AccountIDSenderType {
				publicKeyX := bc.Senders[key].PublicKey.Pk.X.Bytes() // TODO: Use account ID
				senderAccountId = publicKeyX[:NumAccountIDBytes]
			} else {
				senderAccountId = []byte{int0Key, int0Key, int0Key, int0Key, int0Key}
			}
			copy(senderAccountIDs[NumAccountIDBytes*key:NumAccountIDBytes*(key+int1Key)], senderAccountId)
		}
		data = append(data, senderAccountIDs...)
	}

	numFlagBytes := (len(bc.Senders) + int8Key - 1) / int8Key
	senderFlags := make([]byte, numFlagBytes)
	for key := range bc.Senders {
		var isPosted uint8
		if bc.Senders[key].IsSigned {
			isPosted = int1Key
		} else {
			isPosted = int0Key
		}
		senderFlags[key/int8Key] |= isPosted << (uint(key) % int8Key)
	}
	data = append(data, senderFlags...)

	if bc.SenderType == PublicKeySenderType {
		data = append(data, int0Key)
	} else {
		data = append(data, int1Key)
	}

	// TODO: accountIDsHash *common.Hash

	// TODO: senderPublicKeysHash *common.Hash

	return data
}

func (bc *BlockContent) Hash() common.Hash {
	return crypto.Keccak256Hash(bc.Marshal())
}

type PostedBlock struct {
	// The previous block hash.
	PrevBlockHash common.Hash
	// The block number, which is the latest block number in the Rollup contract plus 1.
	BlockNumber uint32
	// The deposit root at the time of block posting (written in the Rollup contract).
	DepositRoot common.Hash
	// The hash value that the Block Builder must provide to the Rollup contract when posting a new block.
	ContentHash common.Hash
}

func NewPostedBlock(prevBlockHash, depositRoot common.Hash, blockNumber uint32, contentHash common.Hash) *PostedBlock {
	return &PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   blockNumber,
		DepositRoot:   depositRoot,
		ContentHash:   contentHash,
	}
}

func (pb *PostedBlock) Marshal() []byte {
	const int4Key = 4

	data := make([]byte, 0)

	data = append(data, pb.PrevBlockHash.Bytes()...)
	blockNumberBytes := [int4Key]byte{}
	binary.BigEndian.PutUint32(blockNumberBytes[:], pb.BlockNumber)
	data = append(data, blockNumberBytes[:]...)
	data = append(data, pb.DepositRoot.Bytes()...)
	data = append(data, pb.ContentHash.Bytes()...)

	return data
}

func (pb *PostedBlock) Hash() common.Hash {
	return crypto.Keccak256Hash(pb.Marshal())
}

type PostRegistrationBlockInput struct {
	TxTreeRoot          [32]byte
	SenderFlags         [16]byte
	AggregatedPublicKey [2][32]byte
	AggregatedSignature [4][32]byte
	MessagePoint        [4][32]byte
	SenderPublicKeys    []*big.Int
}

type PostNonRegistrationBlockInput struct {
	TxTreeRoot          [32]byte
	SenderFlags         [16]byte
	AggregatedPublicKey [2][32]byte
	AggregatedSignature [4][32]byte
	MessagePoint        [4][32]byte
	PublicKeysHash      [32]byte
	SenderAccountIds    []byte
}

// MakePostRegistrationBlockInput creates a PostRegistrationBlockInput from a BlockContent.
// The input is used to post a block on the smart contract:
//
//	rollup, err := bindings.NewRollup(rollupContractAddress, client)
//	input, err := MakePostRegistrationBlockInput(blockContent)
//	rollup.PostRegistrationBlock(
//		opts,
//		input.TxTreeRoot,
//		input.SenderFlags,
//		input.AggregatedPublicKey,
//		input.AggregatedSignature,
//		input.MessagePoint,
//		input.SenderPublicKeys)
func MakePostRegistrationBlockInput(blockContent *BlockContent) (*PostRegistrationBlockInput, error) {
	if len(blockContent.Senders) != NumOfSenders {
		return nil, errors.New("invalid number of senders")
	}

	const int3Key = 3

	txTreeRoot := [numHashBytes]byte{}
	copy(txTreeRoot[:], blockContent.TxTreeRoot.Marshal())

	senderFlags := [numFlagBytes]byte{}
	senderPublicKeys := make([]*big.Int, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		if sender.IsSigned {
			senderFlags[i/int8Key] |= 1 << (i % int8Key)
		}

		senderPublicKeys[i] = new(big.Int).Set(sender.PublicKey.BigInt())
	}

	// Follow the ordering of the coordinates in the smart contract.
	aggregatedPublicKey := [2][int32Key]byte{}
	aggregatedPublicKey[0] = blockContent.AggregatedPublicKey.Pk.X.Bytes()
	aggregatedPublicKey[1] = blockContent.AggregatedPublicKey.Pk.Y.Bytes()

	aggregatedSignature := [numG2PointLimbs][int32Key]byte{}
	aggregatedSignature[0] = blockContent.AggregatedSignature.X.A1.Bytes()
	aggregatedSignature[1] = blockContent.AggregatedSignature.X.A0.Bytes()
	aggregatedSignature[2] = blockContent.AggregatedSignature.Y.A1.Bytes()
	aggregatedSignature[int3Key] = blockContent.AggregatedSignature.Y.A0.Bytes()

	messagePoint := [numG2PointLimbs][int32Key]byte{}
	messagePoint[0] = blockContent.MessagePoint.X.A1.Bytes()
	messagePoint[1] = blockContent.MessagePoint.X.A0.Bytes()
	messagePoint[2] = blockContent.MessagePoint.Y.A1.Bytes()
	messagePoint[int3Key] = blockContent.MessagePoint.Y.A0.Bytes()

	return &PostRegistrationBlockInput{
		TxTreeRoot:          txTreeRoot,
		SenderFlags:         senderFlags,
		AggregatedPublicKey: aggregatedPublicKey,
		AggregatedSignature: aggregatedSignature,
		MessagePoint:        messagePoint,
		SenderPublicKeys:    senderPublicKeys,
	}, nil
}

func MakePostNonRegistrationBlockInput(blockContent *BlockContent) (*PostNonRegistrationBlockInput, error) {
	b, err := MakePostRegistrationBlockInput(blockContent)
	if err != nil {
		return nil, err
	}

	if len(blockContent.Senders) != NumOfSenders {
		return nil, errors.New("invalid number of senders")
	}

	senderPublicKeys := make([][]byte, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		address := sender.PublicKey.ToAddress()
		senderPublicKeys[i] = address[:]
	}

	publicKeysHash := [NumPublicKeyBytes]byte{}
	copy(publicKeysHash[:], crypto.Keccak256(senderPublicKeys...))

	senderAccountIds, err := MakeAccountIds(blockContent)
	if err != nil {
		return nil, err
	}

	return &PostNonRegistrationBlockInput{
		TxTreeRoot:          b.TxTreeRoot,
		SenderFlags:         b.SenderFlags,
		AggregatedPublicKey: b.AggregatedPublicKey,
		AggregatedSignature: b.AggregatedSignature,
		MessagePoint:        b.MessagePoint,
		PublicKeysHash:      publicKeysHash,
		SenderAccountIds:    senderAccountIds,
	}, nil
}

func MakeAccountIds(blockContent *BlockContent) ([]byte, error) {
	if blockContent.SenderType != AccountIDSenderType {
		return nil, ErrBlockContentSenderTypeInvalid
	}

	accountIds := make([]uint64, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		accountIds[i] = sender.AccountID
	}

	return MarshalAccountIds(accountIds)
}

func MarshalAccountIds(accountIds []uint64) ([]byte, error) {
	accountIdsBytes := make([]byte, len(accountIds)*NumAccountIDBytes)
	for i, accountID := range accountIds {
		if accountID >= 1<<(NumAccountIDBytes*int8Key) {
			return nil, errors.New("invalid account ID")
		}
		// account ID is big-endian
		for j := 0; j < NumAccountIDBytes; j++ {
			reverseIndex := NumAccountIDBytes - (j + 1)
			accountIdsBytes[i*NumAccountIDBytes+j] = byte(accountID >> uint(int8Key*reverseIndex))
		}
	}

	return accountIdsBytes, nil
}

func UnmarshalAccountIds(accountIdsBytes []byte) ([]uint64, error) {
	const (
		int0Key = 0
		int8Key = 8
	)

	if len(accountIdsBytes)%NumAccountIDBytes != int0Key {
		return nil, fmt.Errorf("length of account IDs bytes is not a multiple of 5")
	}

	accountIds := make([]uint64, len(accountIdsBytes)/NumAccountIDBytes)
	for i := int0Key; i < len(accountIds); i++ {
		bytes := make([]byte, int8Key)
		copy(bytes[int8Key-NumAccountIDBytes:], accountIdsBytes[i*NumAccountIDBytes:(i+1)*NumAccountIDBytes])
		accountIds[i] = binary.BigEndian.Uint64(bytes)
	}

	return accountIds, nil
}

type RollupContractConfig struct {
	// NetworkChainID is the chain ID of the network
	NetworkChainID string

	// NetworkRpcUrl is the URL of the network RPC endpoint
	NetworkRpcUrl string

	// RollupContractAddressHex is the address of the Rollup contract
	RollupContractAddressHex string

	// EthereumPrivateKeyHex is the private key used to sign transactions
	EthereumPrivateKeyHex string
}

// NewRollupContractConfigFromEnv creates a new RollupContractConfig from the environment variables.
func NewRollupContractConfigFromEnv(cfg *configs.Config, networkRpcUrl string) *RollupContractConfig {
	return &RollupContractConfig{
		NetworkRpcUrl:            networkRpcUrl,
		RollupContractAddressHex: cfg.Blockchain.RollupContractAddress,
		EthereumPrivateKeyHex:    cfg.Blockchain.EthereumPrivateKeyHex,
		NetworkChainID:           cfg.Blockchain.ScrollNetworkChainID,
	}
}

// PostRegistrationBlock posts a registration block on the Rollup contract.
// It returns the transaction hash if the block is successfully posted.
func PostRegistrationBlock(cfg *RollupContractConfig, blockContent *BlockContent) (*types.Transaction, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	input, err := MakePostRegistrationBlockInput(blockContent)
	if err != nil {
		return nil, fmt.Errorf("failed to make post registration block input: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.EthereumPrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	const (
		int10Key = 10
		int64Key = 64
	)
	chainID, err := strconv.ParseInt(cfg.NetworkChainID, int10Key, int64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Check recover block content
	err = blockContent.IsValid()
	if err != nil {
		return nil, fmt.Errorf("block content is invalid: %w", err)
	}

	return rollup.PostRegistrationBlock(
		transactOpts,
		input.TxTreeRoot,
		input.SenderFlags,
		input.AggregatedPublicKey,
		input.AggregatedSignature,
		input.MessagePoint,
		input.SenderPublicKeys,
	)
}

// PostNonRegistrationBlock posts a non-registration block on the Rollup contract.
// It returns the transaction hash if the block is successfully posted.
func PostNonRegistrationBlock(cfg *RollupContractConfig, blockContent *BlockContent) (*types.Transaction, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	input, err := MakePostNonRegistrationBlockInput(blockContent)
	if err != nil {
		return nil, fmt.Errorf("failed to make post registration block input: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.EthereumPrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	const (
		int10Key = 10
		int64Key = 64
	)
	chainID, err := strconv.ParseInt(cfg.NetworkChainID, int10Key, int64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Output calldata
	fmt.Printf("Tx tree root: %x\n", input.TxTreeRoot)
	fmt.Printf("Sender flags: %x\n", input.SenderFlags)
	fmt.Printf("Aggregated public key: %x\n", input.AggregatedPublicKey)
	fmt.Printf("Aggregated signature: %x\n", input.AggregatedSignature)
	fmt.Printf("Message point: %x\n", input.MessagePoint)
	fmt.Printf("Public keys hash: %x\n", input.PublicKeysHash)
	fmt.Printf("Sender account IDs: %x\n", input.SenderAccountIds)

	return rollup.PostNonRegistrationBlock(
		transactOpts,
		input.TxTreeRoot,
		input.SenderFlags,
		input.AggregatedPublicKey,
		input.AggregatedSignature,
		input.MessagePoint,
		input.PublicKeysHash,
		input.SenderAccountIds,
	)
}

func FetchDepositRoot(cfg *RollupContractConfig, ctx context.Context) ([int32Key]byte, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	opts := bind.CallOpts{
		Pending: false,
		Context: ctx,
	}
	depositRoot, err := rollup.DepositTreeRoot(&opts)

	return depositRoot, err
}

func FetchLatestIntMaxBlockNumber(cfg *RollupContractConfig, ctx context.Context) (uint32, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return 0, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	opts := bind.CallOpts{
		Pending: false,
		Context: ctx,
	}
	latestBlockNumber, err := rollup.GetLatestBlockNumber(&opts)

	return latestBlockNumber, err
}

func FetchIntMaxBlock(cfg *RollupContractConfig, ctx context.Context, blockNumber uint32) (common.Hash, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	opts := bind.CallOpts{
		Pending: false,
		Context: ctx,
	}
	blockHash, err := rollup.GetBlockHash(&opts, blockNumber)

	return blockHash, err
}

func FetchPostedBlocks(cfg *RollupContractConfig, ctx context.Context, startBlock uint64, prevBlockHash [][32]byte, blockBuilder []common.Address) ([]*bindings.RollupBlockPosted, *big.Int, error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	nextBlock := startBlock + 1
	iterator, err := rollup.FilterBlockPosted(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: ctx,
	}, prevBlockHash, blockBuilder)
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
		if event.BlockNumber.Cmp(maxBlockNumber) > 0 {
			maxBlockNumber.Set(event.BlockNumber)
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, nil, errors.Join(ErrEncounteredWhileIterating, err)
	}

	return events, maxBlockNumber, nil
}

func FetchLatestIntMaxBlock(cfg *RollupContractConfig, ctx context.Context) (*bindings.RollupBlockPosted, error) {
	latestBlockNumber, err := FetchLatestIntMaxBlockNumber(cfg, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest block number: %w", err)
	}
	if latestBlockNumber == 0 {
		defaultDepositTreeRoot := [32]byte{}
		decodedDefaultDepositTreeRoot, err := hexutil.Decode("0xb6155ab566bbd2e341525fd88c43b4d69572bf4afe7df45cd74d6901a172e41c")
		if err != nil {
			return nil, fmt.Errorf("failed to decode default deposit tree root: %w", err)
		}

		copy(defaultDepositTreeRoot[:], decodedDefaultDepositTreeRoot)
		return &bindings.RollupBlockPosted{
			PrevBlockHash:   [32]byte{},
			BlockBuilder:    common.Address{},
			BlockNumber:     big.NewInt(0),
			DepositTreeRoot: defaultDepositTreeRoot,
			SignatureHash:   [32]byte{},
		}, nil
	}

	latestPrevBlockHash, err := FetchIntMaxBlock(cfg, ctx, latestBlockNumber-1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest block: %w", err)
	}

	blocks, _, err := FetchPostedBlocks(cfg, ctx, 0, [][int32Key]byte{latestPrevBlockHash}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posted blocks: %w", err)
	}
	if len(blocks) == 0 {
		return nil, errors.New("no posted blocks found")
	}

	return blocks[0], nil
}
