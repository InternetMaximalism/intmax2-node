package block_validity_prover

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/finite_field"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"math/big"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)

type mockBlockBuilder struct {
	AccountTree   *intMaxTree.AccountTree
	BlockTree     *intMaxTree.BlockHashTree
	DepositTree   *intMaxTree.KeccakMerkleTree
	DepositLeaves []*intMaxTree.DepositLeaf

	db SQLDriverApp
	// LastPostedBlockNumber         uint32
	// LastGeneratedProofBlockNumber uint32
	// ValidityProofs                map[uint32]string

	latestWitnessBlockNumber uint32
	MerkleTreeHistory        map[uint32]*MerkleTrees
	// LastSeenProcessedDepositId    uint64
	// auxInfo                  map[uint32]*mDBApp.BlockContent
	// validityWitnesses        map[uint32]*ValidityWitness
	// DepositTreeRoots           []common.Hash
	// DepositTreeHistory         map[string]*intMaxTree.KeccakMerkleTree // deposit hash -> deposit tree
}

type MockBlockBuilderMemory = mockBlockBuilder

type MockBlockBuilder interface {
	AccountBySenderAddress(_ string) (*uint256.Int, error)
	AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedInsertionProof, error)
	AppendBlockTreeLeaf(block *block_post_service.PostedBlock) error
	AppendDepositTreeLeaf(depositHash common.Hash, depositLeaf *intMaxTree.DepositLeaf) (root common.Hash, err error)
	// BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockAuxInfo(blockNumber uint32) (*AuxInfo, error)
	BlockContentByTxRoot(txRoot string) (*mDBApp.BlockContentWithProof, error)
	BlockNumberByDepositIndex(depositIndex uint32) (uint32, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
	BlockTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	ConstructSignature(txTreeRoot intMaxTypes.Bytes32, publicKeysHash intMaxTypes.Bytes32, accountIDHash intMaxTypes.Bytes32, isRegistrationBlock bool, sortedTxs []*MockTxRequest) (*SignatureContent, error)
	// CreateBlockContent(postedBlock *block_post_service.PostedBlock, blockContent *intMaxTypes.BlockContent) (*mDBApp.BlockContentWithProof, error)
	CurrentBlockTreeProof(blockNumber uint32) (*intMaxTree.MerkleProof, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
	FetchLastDepositIndex() (uint32, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
	GenerateBlock(blockContent *intMaxTypes.BlockContent, postedBlock *block_post_service.PostedBlock) (*BlockWitness, error)
	GenerateBlockWithTxTree(isRegistrationBlock bool, txs []*MockTxRequest) (*BlockWitness, *intMaxTree.TxTree, error)
	GetAccountMembershipProof(blockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error)
	GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error)
	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)
	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
	LastDepositTreeRoot() (common.Hash, error)
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	// LastValidityWitness() (*ValidityWitness, error)
	LatestIntMaxBlockNumber() uint32
	NextAccountID() (uint64, error)
	ValidityWitness(isRegistrationBlock bool, txs []*MockTxRequest) (*ValidityWitness, error)
	ProveInclusion(accountId uint64) (*AccountMerkleProof, error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accountID uint64, err error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	SetValidityProof(blockNumber uint32, proof string) error
	SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error
	UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedUpdateProof, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	ValidityProofByBlockNumber(blockNumber uint32) (*string, error)
	ValidityWitnessByBlockNumber(blockNumber uint32) (*ValidityWitness, error)
}

type UpdateWitness struct {
	ValidityProof          string                             `json:"validityProof"`
	BlockMerkleProof       intMaxTree.BlockHashMerkleProof    `json:"blockMerkleProof"`
	AccountMembershipProof *intMaxTree.IndexedMembershipProof `json:"accountMembershipProof"`
}

func (b *mockBlockBuilder) FetchUpdateWitness(
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*UpdateWitness, error) {
	fmt.Printf("FetchUpdateWitness currentBlockNumber: %d\n", currentBlockNumber)
	fmt.Printf("FetchUpdateWitness targetBlockNumber: %d\n", targetBlockNumber)
	// request validity prover
	latestValidityProof, err := b.ValidityProofByBlockNumber(currentBlockNumber)
	if err != nil {
		return nil, err
	}

	// blockMerkleProof := blockBuilder.GetBlockMerkleProof(currentBlockNumber, targetBlockNumber)
	blockMerkleProof, err := b.BlockTreeProof(currentBlockNumber, targetBlockNumber)
	if err != nil {
		return nil, err
	}

	var accountMembershipProof *intMaxTree.IndexedMembershipProof
	if isPrevAccountTree {
		fmt.Printf("is PrevAccountTree %d\n", currentBlockNumber-1)
		accountMembershipProof, err = b.GetAccountMembershipProof(currentBlockNumber-1, publicKey.BigInt())
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Printf("is not PrevAccountTree %d\n", currentBlockNumber)
		accountMembershipProof, err = b.GetAccountMembershipProof(currentBlockNumber, publicKey.BigInt())
		if err != nil {
			return nil, err
		}
	}

	return &UpdateWitness{
		ValidityProof:          *latestValidityProof,
		BlockMerkleProof:       *blockMerkleProof,
		AccountMembershipProof: accountMembershipProof,
	}, nil
}

// NewBlockHashTree is a Merkle tree that includes the genesis block in the 0th leaf from the beginning.
func NewBlockHashTree(height uint8) (*intMaxTree.BlockHashTree, error) {
	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	genesisBlockHash := intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())
	initialLeaves := []*intMaxTree.BlockHashLeaf{genesisBlockHash}

	return intMaxTree.NewBlockHashTreeWithInitialLeaves(height, initialLeaves)
}

func NewMockBlockBuilder(cfg *configs.Config, db SQLDriverApp) *MockBlockBuilderMemory {
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	blockHashes := make([][32]byte, 1)
	blockHashes[0] = new(block_post_service.PostedBlock).Genesis().Hash()
	blockTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	genesisBlockHash := intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())
	_, err = blockTree.AddLeaf(0, genesisBlockHash)
	if err != nil {
		panic(err)
	}

	zeroDepositHash := new(intMaxTree.DepositLeaf).SetZero().Hash()
	depositTree, err := intMaxTree.NewKeccakMerkleTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil, zeroDepositHash)
	if err != nil {
		panic(err)
	}
	// depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()

	// validityWitness := new(ValidityWitness).Genesis()
	// validityWitnesses := make(map[uint32]*ValidityWitness)
	// validityWitnesses[0] = new(ValidityWitness).Genesis()
	// auxInfo := make(map[uint32]*mDBApp.BlockContent)

	deposits, err := db.ScanDeposits()
	if err != nil {
		panic(err)
	}

	depositLeaves := make([]*intMaxTree.DepositLeaf, 0)
	for depositIndex, deposit := range deposits {
		depositLeaf := intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		}
		if deposit.DepositIndex == nil {
			panic("deposit index should not be nil")
		}

		// depositIndex := *deposit.DepositIndex
		_, err = depositTree.AddLeaf(uint32(depositIndex), depositLeaf.Hash())
		if err != nil {
			panic(err)
		}

		depositLeaves = append(depositLeaves, &depositLeaf)
	}

	merkleTreeHistory := make(map[uint32]*MerkleTrees)
	merkleTreeHistory[0] = &MerkleTrees{
		// ValidityWitness: validityWitness,
		AccountTree:   new(intMaxTree.AccountTree).Set(accountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(blockTree),
		DepositLeaves: make([]*intMaxTree.DepositLeaf, len(depositLeaves)),
	}
	copy(merkleTreeHistory[0].DepositLeaves, depositLeaves)

	return &mockBlockBuilder{
		db:                db,
		AccountTree:       accountTree,
		BlockTree:         blockTree,
		DepositTree:       depositTree,
		DepositLeaves:     depositLeaves,
		MerkleTreeHistory: merkleTreeHistory,
		// auxInfo:                       auxInfo,
		// LastGeneratedProofBlockNumber: 0,
		// ValidityProofs:                make(map[uint32]string),
		// validityWitnesses:             validityWitnesses,
	}
}

func (b *mockBlockBuilder) Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error) {
	return b.db.Exec(ctx, input, executor)
}

type DepositLeafWithId struct {
	DepositLeaf *intMaxTree.DepositLeaf
	DepositId   uint32
}

func NewSignatureContentFromBlockContent(blockContent *intMaxTypes.BlockContent) *SignatureContent {
	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	accountIDs := make([]uint64, len(blockContent.Senders))
	senderFlagBytes := [int16Key]byte{}
	for i, sender := range blockContent.Senders {
		publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
		publicKeys[i] = *publicKey
		accountIDs[i] = sender.AccountID
		var flag uint8 = 0
		if sender.IsSigned {
			flag = 1
		}
		senderFlagBytes[i/int8Key] |= flag << (int8Key - 1 - i%int8Key)
	}

	signatureContent := SignatureContent{
		IsRegistrationBlock: isRegistrationBlock,
		TxTreeRoot:          intMaxTypes.Bytes32{},
		SenderFlag:          intMaxTypes.Bytes16{},
		PublicKeyHash:       GetPublicKeysHash(publicKeys),
		AccountIDHash:       GetAccountIDsHash(accountIDs),
		AggPublicKey:        intMaxTypes.FlattenG1Affine(blockContent.AggregatedPublicKey.Pk),
		AggSignature:        intMaxTypes.FlattenG2Affine(blockContent.AggregatedSignature),
		MessagePoint:        intMaxTypes.FlattenG2Affine(blockContent.MessagePoint),
	}
	copy(signatureContent.TxTreeRoot[:], intMaxTypes.CommonHashToUint32Slice(blockContent.TxTreeRoot))
	signatureContent.SenderFlag.FromBytes(senderFlagBytes[:])

	return &signatureContent
}

func (b *mockBlockBuilder) GenerateBlock(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*BlockWitness, error) {
	// isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	// publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	// accountIDs := make([]uint64, len(blockContent.Senders))
	// senderFlagBytes := [int16Key]byte{}
	// for i, sender := range blockContent.Senders {
	// 	publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
	// 	publicKeys[i] = *publicKey
	// 	accountIDs[i] = sender.AccountID
	// 	var flag uint8 = 0
	// 	if sender.IsSigned {
	// 		flag = 1
	// 	}
	// 	senderFlagBytes[i/int8Key] |= flag << (int8Key - 1 - i%int8Key)
	// }

	signature := NewSignatureContentFromBlockContent(blockContent)
	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	// accountIDs := make([]uint64, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
		publicKeys[i] = *publicKey
		// accountIDs[i] = sender.AccountID
	}

	prevAccountTreeRoot := b.AccountTree.GetRoot()
	prevBlockTreeRoot := b.BlockTree.GetRoot()

	if signature.IsRegistrationBlock {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, len(blockContent.Senders))
		for i, sender := range blockContent.Senders {
			accountMembershipProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
			if err != nil {
				return nil, errors.New("account membership proof error")
			}

			accountMembershipProofs[i] = *accountMembershipProof
		}

		blockWitness := &BlockWitness{
			Block:                   postedBlock,
			Signature:               *signature,
			PublicKeys:              publicKeys,
			PrevAccountTreeRoot:     prevAccountTreeRoot,
			PrevBlockTreeRoot:       prevBlockTreeRoot,
			AccountIdPacked:         nil,
			AccountMerkleProofs:     nil,
			AccountMembershipProofs: &accountMembershipProofs,
		}

		return blockWitness, nil
	}

	accountMerkleProofs := make([]AccountMerkleProof, len(blockContent.Senders))
	accountIDPackedBytes := make([]byte, numAccountIDPackedBytes)
	for i, sender := range blockContent.Senders {
		accountIDByte := make([]byte, int8Key)
		binary.BigEndian.PutUint64(accountIDByte, sender.AccountID)
		copy(accountIDPackedBytes[i/int8Key:i/int8Key+int5Key], accountIDByte[int8Key-int5Key:])
		accountMembershipProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
		if err != nil {
			return nil, errors.New("account membership proof error")
		}
		if !accountMembershipProof.IsIncluded {
			return nil, errors.New("account is not included")
		}

		accountMerkleProofs[i] = AccountMerkleProof{
			MerkleProof: accountMembershipProof.LeafProof,
			Leaf:        accountMembershipProof.Leaf,
		}
	}

	accountIDPacked := new(AccountIdPacked)
	accountIDPacked.FromBytes(accountIDPackedBytes)
	blockWitness := &BlockWitness{
		Block:                   postedBlock,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: nil,
	}

	return blockWitness, nil
}

func getBitFromUint32Slice(limbs []uint32, i int) bool {
	if i >= len(limbs)*int32Key {
		panic("out of index")
	}

	return (limbs[i/int32Key]>>(int32Key-1-i%int32Key))&1 == 1
}

func getSenderLeaves(publicKeys []intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) []SenderLeaf {
	senderLeaves := make([]SenderLeaf, 0)
	for i, publicKey := range publicKeys {
		senderLeaf := SenderLeaf{
			Sender:  publicKey.BigInt(),
			IsValid: getBitFromUint32Slice(senderFlag[:], i),
		}
		senderLeaves = append(senderLeaves, senderLeaf)
	}

	return senderLeaves
}

func (db *mockBlockBuilder) SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error {
	depositTree, err := intMaxTree.NewDepositTree(int32Key)
	if err != nil {
		return err
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	if depositTreeRoot != witness.BlockWitness.Block.DepositRoot {
		for i, deposit := range db.DepositLeaves {
			depositLeaf := intMaxTree.DepositLeaf{
				RecipientSaltHash: deposit.RecipientSaltHash,
				TokenIndex:        deposit.TokenIndex,
				Amount:            deposit.Amount,
			}

			_, err := depositTree.AddLeaf(uint32(i), depositLeaf)
			if err != nil {
				return err
			}

			depositTreeRoot, _, _ = depositTree.GetCurrentRootCountAndSiblings()
			fmt.Printf("SetValidityWitness depositTreeRoot: %s\n", depositTreeRoot.String())
			if depositTreeRoot == witness.BlockWitness.Block.DepositRoot {
				break
			}
		}
	}

	fmt.Printf("blockNumber: %d\n", blockNumber)
	fmt.Printf("GetAccountMembershipProof root: %s\n", db.AccountTree.GetRoot().String())

	db.latestWitnessBlockNumber = blockNumber
	// db.validityWitnesses[blockNumber] = new(ValidityWitness).Set(witness)
	db.MerkleTreeHistory[blockNumber] = &MerkleTrees{
		AccountTree:   new(intMaxTree.AccountTree).Set(db.AccountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(db.BlockTree),
		DepositLeaves: depositTree.Leaves,
		// DepositTreeRoot: depositTreeRoot,
	}

	return nil
}

func (db *mockBlockBuilder) LastValidityWitness() (*ValidityWitness, error) {
	return db.ValidityWitnessByBlockNumber(db.latestWitnessBlockNumber)
}

func (db *mockBlockBuilder) ValidityWitnessByBlockNumber(blockNumber uint32) (*ValidityWitness, error) {
	if blockNumber == 0 {
		genesisValidityWitness := new(ValidityWitness).Genesis()
		return genesisValidityWitness, nil
	}

	auxInfo, err := db.BlockAuxInfo(blockNumber)
	if err != nil {
		return nil, err
	}

	blockWitness, err := db.GenerateBlockWithTxTreeFromBlockContent(
		auxInfo.BlockContent,
		auxInfo.PostedBlock,
	)
	if err != nil {
		return nil, err
	}
	validityWitness, err := calculateValidityWitness(db, blockWitness)
	if err != nil {
		return nil, err
	}

	return validityWitness, nil
}

func (db *mockBlockBuilder) AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error) {
	return db.AccountTree.GetRoot(), nil
}

func (db *mockBlockBuilder) GetAccountMembershipProof(blockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error) {
	blockHistory, ok := db.MerkleTreeHistory[blockNumber]
	if !ok {
		return nil, fmt.Errorf("current block number %d not found", blockNumber)
	}
	proof, root, err := blockHistory.AccountTree.ProveMembership(publicKey)
	fmt.Printf("blockNumber: %d\n", blockNumber)
	fmt.Printf("GetAccountMembershipProof root: %s\n", root.String())
	// proof, _, err := db.AccountTree.ProveMembership(publicKey)
	if err != nil {
		return nil, errors.New("account membership proof error")
	}

	return proof, nil
}

func (db *mockBlockBuilder) ProveInclusion(accountId uint64) (*AccountMerkleProof, error) {
	leaf := db.AccountTree.GetLeaf(accountId)
	proof, _, err := db.AccountTree.Prove(accountId)
	if err != nil {
		return nil, err
	}

	return &AccountMerkleProof{
		MerkleProof: *proof,
		Leaf:        *leaf,
	}, nil
}

func (db *mockBlockBuilder) BlockTreeRoot() (*intMaxGP.PoseidonHashOut, error) {
	return db.BlockTree.GetRoot(), nil
}

func (db *mockBlockBuilder) BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error) {
	if rootBlockNumber < leafBlockNumber {
		return nil, errors.New("root block number should be greater than or equal to leaf block number")
	}

	blockHistory, ok := db.MerkleTreeHistory[rootBlockNumber]
	if !ok {
		return nil, errors.New("current block number %d not found (BlockTreeProof)")
	}

	proof, _, err := blockHistory.BlockHashTree.Prove(leafBlockNumber)
	if err != nil {
		return nil, errors.New("block tree proof error")
	}

	return &proof, nil
}

func (db *mockBlockBuilder) CurrentBlockTreeProof(blockNumber uint32) (*intMaxTree.MerkleProof, error) {
	proof, _, err := db.BlockTree.Prove(blockNumber)
	if err != nil {
		return nil, errors.New("block tree proof error")
	}

	return &proof, nil
}

func (db *mockBlockBuilder) IsSynchronizedDepositIndex(depositIndex uint32) (bool, error) {
	lastValidityWitness, err := db.LastValidityWitness()
	if err != nil {
		return false, err
	}
	lastBlockNumber := lastValidityWitness.BlockWitness.Block.BlockNumber
	depositLeaves := db.MerkleTreeHistory[lastBlockNumber].DepositLeaves

	if depositIndex >= uint32(len(depositLeaves)) {
		return false, nil
	}

	return true, nil
}

func (db *mockBlockBuilder) DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	depositLeaves := db.MerkleTreeHistory[blockNumber].DepositLeaves

	if depositIndex >= uint32(len(depositLeaves)) {
		return nil, common.Hash{}, errors.New("block number is out of range")
	}

	leaves := make([][32]byte, 0)
	for _, depositLeaf := range depositLeaves {
		leaves = append(leaves, [32]byte(depositLeaf.Hash()))
	}
	proof, root, err := db.DepositTree.ComputeMerkleProof(depositIndex, leaves)
	if err != nil {
		var ErrDepositTreeProof = errors.New("deposit tree proof error")
		return nil, common.Hash{}, errors.Join(ErrDepositTreeProof, err)
	}

	return proof, root, nil
}

// TODO: refactor
func (db *mockBlockBuilder) BlockNumberByDepositIndex(depositIndex uint32) (uint32, error) {
	lastValidityWitness, err := db.LastValidityWitness()
	if err != nil {
		return 0, err
	}

	blockNumber := uint32(1)
	fmt.Printf("lastValidityWitness.BlockWitness.Block.BlockNumber: %d\n", lastValidityWitness.BlockWitness.Block.BlockNumber)
	for ; blockNumber <= lastValidityWitness.BlockWitness.Block.BlockNumber; blockNumber++ {
		depositLeaves := db.MerkleTreeHistory[blockNumber].DepositLeaves
		fmt.Printf("latest deposit index: %d\n", len(depositLeaves))
		if depositIndex >= uint32(len(depositLeaves)) {
			return 0, errors.New("deposit index is out of range")
		}
	}

	return blockNumber, nil
}

func (db *mockBlockBuilder) AppendBlockTreeLeaf(block *block_post_service.PostedBlock) error {
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	_, count, _ := db.BlockTree.GetCurrentRootCountAndSiblings()
	if count != block.BlockNumber {
		return errors.New("block number is not equal to the current block number")
	}

	fmt.Printf("old block root: %s\n", db.BlockTree.GetRoot().String())
	newRoot, err := db.BlockTree.AddLeaf(count, blockHashLeaf)
	if err != nil {
		var ErrBlockTreeAddLeaf = errors.New("block tree add leaf error")
		return errors.Join(ErrBlockTreeAddLeaf, err)
	}
	fmt.Printf("new block root: %s\n", newRoot.String())

	return nil
}

func (db *mockBlockBuilder) AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedInsertionProof, error) {
	proof, err := db.AccountTree.Insert(sender, lastBlockNumber)
	if err != nil {
		// invalid block
		var ErrAccountTreeInsert = errors.New("account tree insert error")
		return nil, errors.Join(ErrAccountTreeInsert, err)
	}

	return proof, nil
}

func (db *mockBlockBuilder) UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedUpdateProof, error) {
	proof, err := db.AccountTree.Update(sender, lastBlockNumber)
	if err != nil {
		var ErrAccountTreeUpdate = errors.New("account tree update error")
		return nil, errors.Join(ErrAccountTreeUpdate, err)
	}

	return proof, nil
}

func (db *mockBlockBuilder) GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error) {
	accountID, ok := db.AccountTree.GetAccountID(sender)
	if !ok {
		return nil, ErrAccountTreeGetAccountID
	}
	prevLeaf := db.AccountTree.GetLeaf(accountID)

	return prevLeaf, nil
}

func (db *mockBlockBuilder) ConstructSignature(
	txTreeRoot intMaxTypes.Bytes32,
	publicKeysHash intMaxTypes.Bytes32,
	accountIDHash intMaxTypes.Bytes32,
	isRegistrationBlock bool,
	sortedTxs []*MockTxRequest,
) (*SignatureContent, error) {
	senderFlagBytes := [int16Key]byte{}
	for i, tx := range sortedTxs {
		var flag uint8 = 0
		if tx.WillReturnSignature {
			flag = 1
		}
		senderFlagBytes[i/int8Key] |= flag << (int8Key - 1 - i%int8Key)
	}
	senderFlag := intMaxTypes.Bytes16{}
	senderFlag.FromBytes(senderFlagBytes[:])
	// fmt.Printf("senderFlag: %v\n", senderFlag)

	flattenTxTreeRoot := finite_field.BytesToFieldElementSlice(txTreeRoot.Bytes())

	signatures := make([]*bn254.G2Affine, len(sortedTxs))
	for i, keyPair := range sortedTxs {
		signature, err := keyPair.Sender.WeightByHash(publicKeysHash.Bytes()).Sign(flattenTxTreeRoot)
		if err != nil {
			return nil, err
		}
		signatures[i] = signature
	}

	messagePoint := intMaxGP.HashToG2(flattenTxTreeRoot)

	aggregatedSignature := new(bn254.G2Affine)
	for _, signature := range signatures {
		aggregatedSignature.Add(aggregatedSignature, signature)
	}

	fmt.Printf("publicKeysHash: %v\n", hexutil.Encode(publicKeysHash.Bytes()))
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, keyPair := range sortedTxs {
		weightedPublicKey := keyPair.Sender.Public().WeightByHash(publicKeysHash.Bytes())
		aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		fmt.Printf("weightedPublicKey: %v\n", weightedPublicKey.BigInt().String())
		fmt.Printf("aggregatedPublicKey: %v\n", aggregatedPublicKey.BigInt().String())
	}

	if aggregatedPublicKey.Pk == nil {
		aggregatedPublicKey.Pk = new(bn254.G1Affine)
		aggregatedPublicKey.Pk.X.SetZero()
		aggregatedPublicKey.Pk.Y.SetZero()
	}

	err := intMaxAcc.VerifySignature(aggregatedSignature, aggregatedPublicKey, flattenTxTreeRoot)
	if err != nil {
		// debug assertion
		return nil, fmt.Errorf("fail to verify aggregatedPublicKey: %s", aggregatedPublicKey.BigInt().String())
	}

	return &SignatureContent{
		IsRegistrationBlock: isRegistrationBlock,
		TxTreeRoot:          txTreeRoot,
		SenderFlag:          senderFlag,
		PublicKeyHash:       publicKeysHash,
		AccountIDHash:       accountIDHash,
		AggPublicKey:        intMaxTypes.FlattenG1Affine(aggregatedPublicKey.Pk),
		AggSignature:        intMaxTypes.FlattenG2Affine(aggregatedSignature),
		MessagePoint:        intMaxTypes.FlattenG2Affine(&messagePoint),
	}, nil
}

func (db *mockBlockBuilder) GenerateBlockWithTxTree(
	isRegistrationBlock bool,
	txs []*MockTxRequest,
) (*BlockWitness, *intMaxTree.TxTree, error) {
	fmt.Println("-----------GenerateBlockWithTxTree------------------")
	const numOfSenders = 128
	if len(txs) > numOfSenders {
		panic("too many txs")
	}

	// sort and pad txs
	sortedTxs := make([]*MockTxRequest, len(txs))
	copy(sortedTxs, txs)
	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[j].Sender.PublicKey.BigInt().Cmp(sortedTxs[i].Sender.PublicKey.BigInt()) == 1
	})

	publicKeys := make([]intMaxTypes.Uint256, len(sortedTxs))
	for i, tx := range sortedTxs {
		publicKeys[i] = *new(intMaxTypes.Uint256).FromBigInt(tx.Sender.Public().BigInt())
	}

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, *new(intMaxTypes.Uint256).FromBigInt(dummyPublicKey.BigInt()))
	}

	lastValidityWitness, err := db.LastValidityWitness()
	if err != nil {
		panic(err)
	}
	postedBlock := lastValidityWitness.BlockWitness.Block

	var accountIDPacked *AccountIdPacked
	var accountMerkleProofs []AccountMerkleProof
	var accountMembershipProofs []intMaxTree.IndexedMembershipProof
	accountIDHash := intMaxTypes.Bytes32{}
	if isRegistrationBlock {
		accountMembershipProofs = make([]intMaxTree.IndexedMembershipProof, len(publicKeys))
		fmt.Printf("size of publicKeys: %d\n", len(publicKeys))
		for i, publicKey := range publicKeys {
			// for pubkey in pubkeys.iter() {
			// 	let is_dummy = pubkey.is_dummy_pubkey();
			// 	assert!(
			// 		self.account_tree.index(*pubkey).is_none() || is_dummy,
			// 		"account already exists"
			// 	);
			// 	let proof = self.account_tree.prove_membership(*pubkey);
			// 	account_membership_proofs.push(proof);
			// }

			isDummy := publicKey.BigInt().Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			fmt.Printf("isDummy: %v, ", isDummy)

			leaf, err := db.GetAccountTreeLeaf(publicKey.BigInt())
			if err != nil {
				if err.Error() != ErrAccountTreeGetAccountID.Error() {
					return nil, nil, errors.Join(errors.New("account tree leaf error"), err)
				}
			}

			if !isDummy && leaf != nil {
				return nil, nil, errors.New("account already exists")
			}

			proof, err := db.GetAccountMembershipProof(postedBlock.BlockNumber, publicKey.BigInt())
			if err != nil {
				return nil, nil, errors.Join(errors.New("account membership proof error"), err)
			}

			accountMembershipProofs[i] = *proof
		}
	} else {
		accountIDs := make([]uint64, len(publicKeys))
		accountMerkleProofs = make([]AccountMerkleProof, len(publicKeys))
		for i, publicKey := range publicKeys {
			accountID, ok := db.AccountTree.GetAccountID(publicKey.BigInt())
			if !ok {
				return nil, nil, errors.New("account id not found")
			}
			proof, err := db.ProveInclusion(accountID)
			if err != nil {
				return nil, nil, errors.New("account inclusion proof error")
			}

			accountIDs[i] = accountID
			accountMerkleProofs[i] = AccountMerkleProof{
				MerkleProof: proof.MerkleProof,
				Leaf:        proof.Leaf,
			}
		}

		accountIDPacked = new(AccountIdPacked).Pack(accountIDs)
		accountIDHash = GetAccountIDsHash(accountIDs)
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	txTree, err := intMaxTree.NewTxTree(uint8(intMaxTree.TX_TREE_HEIGHT), nil, zeroTx.Hash())
	if err != nil {
		panic(err)
	}

	for _, tx := range txs {
		_, index, _ := txTree.GetCurrentRootCountAndSiblings()
		txTree.AddLeaf(index, tx.Tx)
	}

	txTreeRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()
	signature, err := db.ConstructSignature(
		*new(intMaxTypes.Bytes32).FromPoseidonHashOut(&txTreeRoot),
		GetPublicKeysHash(publicKeys),
		accountIDHash,
		isRegistrationBlock,
		sortedTxs,
	)
	if err != nil {
		panic(err)
	}

	depositRoot, _, _ := db.DepositTree.GetCurrentRootCountAndSiblings()
	block := &block_post_service.PostedBlock{
		PrevBlockHash: postedBlock.Hash(),
		DepositRoot:   depositRoot,
		SignatureHash: signature.Hash(),
		BlockNumber:   postedBlock.BlockNumber + 1,
	}

	prevAccountTreeRoot := db.AccountTree.GetRoot()
	prevBlockTreeRoot := db.BlockTree.GetRoot()
	blockWitness := &BlockWitness{
		Block:                   block,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: &accountMembershipProofs,
	}

	validationPis := blockWitness.MainValidationPublicInputs()
	fmt.Printf("validationPis: %v\n", validationPis)
	if !validationPis.IsValid && len(txs) > 0 {
		panic("should be valid block")
	}

	return blockWitness, txTree, nil
}

func (db *mockBlockBuilder) GenerateBlockWithTxTreeFromBlockContent(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*BlockWitness, error) {
	const numOfSenders = 128
	if len(blockContent.Senders) > numOfSenders {
		// panic("too many txs")
		return nil, errors.New("too many txs")
	}

	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		publicKeys[i].FromBigInt(sender.PublicKey.BigInt())
	}

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, *new(intMaxTypes.Uint256).FromBigInt(dummyPublicKey.BigInt()))
	}

	blockNumber := postedBlock.BlockNumber

	var accountIDPacked *AccountIdPacked
	var accountMerkleProofs []AccountMerkleProof
	var accountMembershipProofs []intMaxTree.IndexedMembershipProof
	isRegistrationBlock := blockContent.SenderType == "PUBLIC_KEY"
	if isRegistrationBlock {
		accountMembershipProofs = make([]intMaxTree.IndexedMembershipProof, len(publicKeys))
		fmt.Printf("size of publicKeys: %d\n", len(publicKeys))
		for i, publicKey := range publicKeys {
			isDummy := publicKey.BigInt().Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			fmt.Printf("isDummy: %v, ", isDummy)

			leaf, err := db.GetAccountTreeLeaf(publicKey.BigInt())
			if err != nil {
				if err.Error() != ErrAccountTreeGetAccountID.Error() {
					return nil, errors.Join(errors.New("account tree leaf error"), err)
				}
			}

			if !isDummy && leaf != nil {
				return nil, errors.New("account already exists")
			}

			proof, err := db.GetAccountMembershipProof(blockNumber, publicKey.BigInt())
			if err != nil {
				return nil, errors.Join(errors.New("account membership proof error"), err)
			}

			accountMembershipProofs[i] = *proof
		}
	} else {
		accountIDs := make([]uint64, len(publicKeys))
		accountMerkleProofs = make([]AccountMerkleProof, len(publicKeys))
		for i, publicKey := range publicKeys {
			accountID, ok := db.AccountTree.GetAccountID(publicKey.BigInt())
			if !ok {
				return nil, errors.New("account id not found")
			}
			proof, err := db.ProveInclusion(accountID)
			if err != nil {
				return nil, errors.New("account inclusion proof error")
			}

			accountIDs[i] = accountID
			accountMerkleProofs[i] = AccountMerkleProof{
				MerkleProof: proof.MerkleProof,
				Leaf:        proof.Leaf,
			}
		}

		accountIDPacked = new(AccountIdPacked).Pack(accountIDs)
		// accountIDHash = GetAccountIDsHash(accountIDs)
	}

	txTreeRoot := intMaxTypes.Bytes32{}
	txTreeRoot.FromBytes(blockContent.TxTreeRoot[:])
	signature := NewSignatureContentFromBlockContent(blockContent)

	depositRoot, _, _ := db.DepositTree.GetCurrentRootCountAndSiblings()
	block := &block_post_service.PostedBlock{
		PrevBlockHash: postedBlock.Hash(),
		DepositRoot:   depositRoot,
		SignatureHash: signature.Hash(),
		BlockNumber:   blockNumber + 1,
	}

	prevAccountTreeRoot := db.AccountTree.GetRoot()
	prevBlockTreeRoot := db.BlockTree.GetRoot()
	blockWitness := &BlockWitness{
		Block:                   block,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: &accountMembershipProofs,
	}

	validationPis := blockWitness.MainValidationPublicInputs()
	fmt.Printf("validationPis: %v\n", validationPis)
	if !validationPis.IsValid && len(blockContent.Senders) > 0 {
		// Despite non-empty block, the block is not valid.
		panic("the block should be valid if it is not an empty block")
	}

	return blockWitness, nil
}

type MockTxRequest struct {
	Sender              *intMaxAcc.PrivateKey
	AccountID           uint64
	Tx                  *intMaxTypes.Tx
	WillReturnSignature bool
}

func (db *mockBlockBuilder) UpdateValidityWitness(
	blockContent *intMaxTypes.BlockContent,
	prevValidityWitness *ValidityWitness,
) (*ValidityWitness, error) {
	blockWitness, err := db.GenerateBlockWithTxTreeFromBlockContent(
		blockContent,
		prevValidityWitness.BlockWitness.Block,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("blockWitness.Block.BlockNumber (PostBlock): %d\n", blockWitness.Block.BlockNumber)
	latestIntMaxBlockNumber := db.LatestIntMaxBlockNumber()
	if blockWitness.Block.BlockNumber != latestIntMaxBlockNumber+1 {
		fmt.Printf("latestIntMaxBlockNumber: %d\n", latestIntMaxBlockNumber)
		return nil, errors.New("block number is not equal to the last block number + 1")
	}
	validityWitness, err := calculateValidityWitnessWithConsistencyCheck(
		db,
		blockWitness,
		prevValidityWitness,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("blockWitness.Block.BlockNumber: %d\n", blockWitness.Block.BlockNumber)
	fmt.Printf("validityWitness.BlockWitness.Block.BlockNumber: %d\n", validityWitness.BlockWitness.Block.BlockNumber)
	// encodedBlockWitness, err := json.Marshal(validityWitness.BlockWitness)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("validityWitness.BlockWitness after generateValidityWitness: %s\n", encodedBlockWitness)

	fmt.Printf("SenderFlag 1: %v\n", validityWitness.BlockWitness.Signature.SenderFlag)
	validityPis := validityWitness.ValidityPublicInputs()
	encodedValidityPis, err := json.Marshal(validityPis)
	if err != nil {
		panic(err)
	}
	fmt.Printf("validityPis (PostBlock): %s\n", encodedValidityPis)

	fmt.Printf("SetValidityWitness SenderFlag: %v\n", validityWitness.BlockWitness.Signature.SenderFlag)
	db.SetValidityWitness(blockWitness.Block.BlockNumber, validityWitness)

	fmt.Printf("post block #%d\n", validityWitness.BlockWitness.Block.BlockNumber)
	return validityWitness, nil
}

func calculateValidityWitnessWithConsistencyCheck(db BlockBuilderStorage, blockWitness *BlockWitness, prevValidityWitness *ValidityWitness) (*ValidityWitness, error) {
	fmt.Printf("---------------------- generateValidityWitness ----------------------\n")
	latestIntMaxBlockNumber := db.LatestIntMaxBlockNumber()
	if blockWitness.Block.BlockNumber != latestIntMaxBlockNumber+1 {
		fmt.Printf("blockWitness.Block.BlockNumber (generateValidityWitness): %d\n", blockWitness.Block.BlockNumber)
		fmt.Printf("latestIntMaxBlockNumber (generateValidityWitness): %d\n", latestIntMaxBlockNumber)
		return nil, errors.New("block number is not equal to the last block number + 1")
	}

	prevPis := prevValidityWitness.ValidityPublicInputs()
	accountTreeRoot, err := db.AccountTreeRoot()
	if err != nil {
		return nil, errors.New("account tree root error")
	}
	if !prevPis.PublicState.AccountTreeRoot.Equal(accountTreeRoot) {
		fmt.Printf("prevPis.PublicState.AccountTreeRoot is not the same with accountTreeRoot, %s != %s", prevPis.PublicState.AccountTreeRoot.String(), accountTreeRoot.String())
		return nil, errors.New("account tree root is not equal to the last account tree root")
	}

	prevBlockTreeRoot, err := db.BlockTreeRoot()
	if err != nil {
		return nil, errors.New("block tree root error")
	}
	if prevPis.IsValidBlock {
		fmt.Printf("block number %d is valid\n", prevPis.PublicState.BlockNumber+1)
	} else {
		fmt.Printf("block number %d is invalid\n", prevPis.PublicState.BlockNumber+1)
	}
	fmt.Printf("prevBlockTreeRoot: %s\n", prevBlockTreeRoot.String())
	if !prevPis.PublicState.BlockTreeRoot.Equal(prevBlockTreeRoot) {
		fmt.Printf("prevPis.PublicState.BlockTreeRoot is not the same with blockTreeRoot, %s != %s", prevPis.PublicState.BlockTreeRoot.String(), prevBlockTreeRoot.String())
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}

	blockMerkleProof, err := db.CurrentBlockTreeProof(blockWitness.Block.BlockNumber)
	if err != nil {
		var ErrBlockTreeProve = errors.New("block tree prove error")
		return nil, errors.Join(ErrBlockTreeProve, err)
	}

	return calculateValidityWitnessWithMerkleProofs(db, blockWitness, prevBlockTreeRoot, blockMerkleProof)
}

func calculateValidityWitness(db BlockBuilderStorage, blockWitness *BlockWitness) (*ValidityWitness, error) {
	fmt.Printf("---------------------- generateValidityWitness ----------------------\n")
	prevBlockTreeRoot, err := db.BlockTreeRoot()
	if err != nil {
		return nil, errors.New("block tree root error")
	}

	blockMerkleProof, err := db.CurrentBlockTreeProof(blockWitness.Block.BlockNumber)
	if err != nil {
		var ErrBlockTreeProve = errors.New("block tree prove error")
		return nil, errors.Join(ErrBlockTreeProve, err)
	}

	return calculateValidityWitnessWithMerkleProofs(db, blockWitness, prevBlockTreeRoot, blockMerkleProof)
}

func calculateValidityWitnessWithMerkleProofs(
	db BlockBuilderStorage,
	blockWitness *BlockWitness,
	prevBlockTreeRoot *intMaxGP.PoseidonHashOut,
	blockMerkleProof *intMaxTree.MerkleProof,
) (*ValidityWitness, error) {
	// debug
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	err := blockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		prevBlockTreeRoot,
	)
	if err != nil {
		panic("old block merkle proof is invalid")
	}

	err = db.AppendBlockTreeLeaf(blockWitness.Block)
	if err != nil {
		return nil, errors.New("append block tree leaf error")
	}

	// debug
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(blockWitness.Block.Hash())
	newBlockTreeRoot, err := db.BlockTreeRoot()
	if err != nil {
		return nil, errors.New("block tree root error")
	}
	err = blockMerkleProof.Verify(
		blockHashLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		newBlockTreeRoot,
	)
	if err != nil {
		panic("new block merkle proof is invalid")
	}

	senderLeaves := getSenderLeaves(blockWitness.PublicKeys, blockWitness.Signature.SenderFlag)

	blockPis := blockWitness.MainValidationPublicInputs()

	accountRegistrationProofsWitness := AccountRegistrationProofs{
		IsValid: false,
		Proofs:  nil,
	}
	if blockPis.IsValid && blockPis.IsRegistrationBlock {
		accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0)
		for _, senderLeaf := range senderLeaves {
			lastBlockNumber := uint32(0)
			if senderLeaf.IsValid {
				lastBlockNumber = blockPis.BlockNumber
			}

			var proof *intMaxTree.IndexedInsertionProof
			isDummy := senderLeaf.Sender.Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			if isDummy {
				proof = intMaxTree.NewDummyAccountRegistrationProof(intMaxTree.ACCOUNT_TREE_HEIGHT)
			} else {
				proof, err = db.AppendAccountTreeLeaf(senderLeaf.Sender, uint64(lastBlockNumber))
				if err != nil {
					var ErrAppendAccountTreeLeaf = errors.New("append account tree leaf error")
					return nil, errors.Join(ErrAppendAccountTreeLeaf, err)
				}
			}

			accountRegistrationProofs = append(accountRegistrationProofs, *proof)
		}

		accountRegistrationProofsWitness = AccountRegistrationProofs{
			IsValid: true,
			Proofs:  accountRegistrationProofs,
		}
	}

	accountUpdateProofsWitness := AccountUpdateProofs{
		IsValid: false,
		Proofs:  nil,
	}
	if blockPis.IsValid && !blockPis.IsRegistrationBlock {
		accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0, len(senderLeaves))
		for _, senderLeaf := range senderLeaves {
			prevLeaf, err := db.GetAccountTreeLeaf(senderLeaf.Sender)
			if err != nil {
				var ErrAccountTreeLeaf = errors.New("account tree leaf error")
				return nil, errors.Join(ErrAccountTreeLeaf, err)
			}

			prevLastBlockNumber := uint32(prevLeaf.Value)
			lastBlockNumber := prevLastBlockNumber
			if senderLeaf.IsValid {
				lastBlockNumber = blockPis.BlockNumber
			}
			var proof *intMaxTree.IndexedUpdateProof
			proof, err = db.UpdateAccountTreeLeaf(senderLeaf.Sender, uint64(lastBlockNumber))
			if err != nil {
				var ErrUpdateAccountTreeLeaf = errors.New("update account tree leaf error")
				return nil, errors.Join(ErrUpdateAccountTreeLeaf, err)
			}
			accountUpdateProofs = append(accountUpdateProofs, *proof)
		}

		accountUpdateProofsWitness = AccountUpdateProofs{
			IsValid: true,
			Proofs:  accountUpdateProofs,
		}
	}

	fmt.Printf("validity_witness prev_account_tree_root: %v\n", blockWitness.PrevAccountTreeRoot.String())
	fmt.Printf("validity_witness accountRegistrationProofsWitness: %v\n", accountRegistrationProofsWitness)
	return &ValidityWitness{
		BlockWitness: blockWitness,
		ValidityTransitionWitness: &ValidityTransitionWitness{
			SenderLeaves:              senderLeaves,
			BlockMerkleProof:          *blockMerkleProof,
			AccountRegistrationProofs: accountRegistrationProofsWitness,
			AccountUpdateProofs:       accountUpdateProofsWitness,
		},
	}, nil
}

func (b *mockBlockBuilder) LatestIntMaxBlockNumber() uint32 {
	return b.latestWitnessBlockNumber
}

func (b *mockBlockBuilder) LastSeenBlockPostedEventBlockNumber() (uint64, error) {
	event, err := b.db.EventBlockNumberByEventNameForValidityProver("BlockPosted")
	if err != nil {
		return 0, err
	}

	return event.LastProcessedBlockNumber, err
}

func (b *mockBlockBuilder) SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error {
	_, err := b.db.UpsertEventBlockNumberForValidityProver("BlockPosted", blockNumber)

	return err
}

func (b *mockBlockBuilder) ValidityProofByBlockNumber(blockNumber uint32) (*string, error) {
	if blockNumber == 0 {
		return nil, ErrGenesisValidityProof
	}

	// if blockNumber >= uint32(len(b.ValidityProofs)) {
	// 	fmt.Printf("len(b.ValidityProofs) (GetValidityProof): %d\n", len(b.ValidityProofs))
	// 	return nil, ErrNoValidityProofByBlockNumber
	// }

	blockContent, err := b.db.BlockContentByBlockNumber(blockNumber)
	if err != nil {
		fmt.Printf("blockNumber (GetValidityProof): %d\n", blockNumber)
		return nil, errors.New("block content by block number error")
	}

	encodedValidityProof := base64.StdEncoding.EncodeToString(blockContent.ValidityProof)

	return &encodedValidityProof, nil
}

func (b *mockBlockBuilder) LastGeneratedProofBlockNumber() (uint32, error) {
	lastValidityProof, err := b.db.LastBlockValidityProof()
	if err != nil {
		return 0, err
	}

	return lastValidityProof.BlockNumber, nil
}

func (b *mockBlockBuilder) SetValidityProof(blockHash common.Hash, proof string) error {
	validityProof, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		return err
	}

	_, err = b.db.CreateValidityProof(blockHash.Hex(), validityProof)
	if err != nil {
		return err
	}

	return err
}

// func (b *mockBlockBuilder) blockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error) {
// 	return b.db.BlockContentByBlockNumber(blockNumber)
// }

func (b *mockBlockBuilder) BlockAuxInfo(blockNumber uint32) (*AuxInfo, error) {
	auxInfo, err := b.db.BlockContentByBlockNumber(blockNumber)
	if err != nil {
		return nil, errors.New("block content by block number error")
	}

	return blockAuxInfoFromBlockContent(auxInfo)
}

// func BlockAuxInfo(db BlockBuilderStorage, blockNumber uint32) (*AuxInfo, error) {
// 	auxInfo, err := db.BlockContentByBlockNumber(blockNumber)
// 	if err != nil {
// 		return nil, errors.New("block content by block number error")
// 	}

// 	return BlockAuxInfoFromBlockContent(auxInfo)
// }

func blockAuxInfoFromBlockContent(auxInfo *mDBApp.BlockContentWithProof) (*AuxInfo, error) {
	decodedAggregatedPublicKeyPoint, err := hexutil.Decode("0x" + auxInfo.AggregatedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("aggregated public key hex decode error: %w", err)
	}
	aggregatedPublicKeyPoint := new(bn254.G1Affine)
	err = aggregatedPublicKeyPoint.Unmarshal(decodedAggregatedPublicKeyPoint)
	if err != nil {
		return nil, fmt.Errorf("aggregated public key unmarshal error: %w", err)
	}
	aggregatedPublicKey, err := intMaxAcc.NewPublicKey(aggregatedPublicKeyPoint)
	if err != nil {
		return nil, fmt.Errorf("aggregated public key error: %w", err)
	}

	decodedAggregatedSignature, err := hexutil.Decode("0x" + auxInfo.AggregatedSignature)
	if err != nil {
		return nil, fmt.Errorf("aggregated signature hex decode error: %w", err)
	}
	aggregatedSignature := new(bn254.G2Affine)
	err = aggregatedSignature.Unmarshal([]byte(decodedAggregatedSignature))
	if err != nil {
		return nil, fmt.Errorf("aggregated signature unmarshal error: %w", err)
	}

	decodedMessagePoint, err := hexutil.Decode("0x" + auxInfo.MessagePoint)
	if err != nil {
		return nil, fmt.Errorf("aggregated message point hex decode error: %w", err)
	}
	messagePoint := new(bn254.G2Affine)
	err = messagePoint.Unmarshal([]byte(decodedMessagePoint))
	if err != nil {
		return nil, fmt.Errorf("message point unmarshal error: %w", err)
	}

	var columnSenders []intMaxTypes.ColumnSender
	err = json.Unmarshal([]byte(auxInfo.Senders), &columnSenders)
	if err != nil {
		return nil, fmt.Errorf("senders unmarshal error: %w", err)
	}
	senders := make([]intMaxTypes.Sender, len(columnSenders))
	for i, sender := range columnSenders {
		publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("public key unmarshal decode error: %w", err)
		}

		senders[i] = intMaxTypes.Sender{
			AccountID: sender.AccountID,
			PublicKey: publicKey,
			IsSigned:  sender.IsSigned,
		}
	}

	var senderType string
	if auxInfo.IsRegistrationBlock {
		senderType = intMaxTypes.PublicKeySenderType
	} else {
		senderType = intMaxTypes.AccountIDSenderType
	}

	blockContent := intMaxTypes.BlockContent{
		TxTreeRoot:          common.HexToHash("0x" + auxInfo.TxRoot),
		AggregatedPublicKey: aggregatedPublicKey,
		AggregatedSignature: aggregatedSignature,
		MessagePoint:        messagePoint,
		Senders:             senders,
		SenderType:          senderType,
	}

	postedBlock := block_post_service.PostedBlock{
		BlockNumber:   auxInfo.BlockNumber,
		PrevBlockHash: common.HexToHash("0x" + auxInfo.PrevBlockHash),
		DepositRoot:   common.HexToHash("0x" + auxInfo.DepositRoot),
		SignatureHash: common.HexToHash("0x" + auxInfo.SignatureHash), // TODO: Calculate from blockContent
	}

	if blockHash := postedBlock.Hash(); blockHash.Hex() != "0x"+auxInfo.BlockHash {
		fmt.Printf("postedBlock: %v\n", postedBlock)
		fmt.Printf("blockHash: %s != %s\n", blockHash.Hex(), auxInfo.BlockHash)
		panic("block hash mismatch")
	}

	return &AuxInfo{
		PostedBlock:  &postedBlock,
		BlockContent: &blockContent,
	}, nil
}

func (b *mockBlockBuilder) CreateBlockContent(
	postedBlock *block_post_service.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
) (*mDBApp.BlockContentWithProof, error) {
	return b.db.CreateBlockContent(
		postedBlock,
		blockContent,
	)
}

func (b *mockBlockBuilder) BlockContentByTxRoot(txRoot string) (*mDBApp.BlockContentWithProof, error) {
	return b.db.BlockContentByTxRoot(txRoot)
}

func (b *mockBlockBuilder) NextAccountID() (uint64, error) {
	return uint64(b.AccountTree.Count()), nil
}

func (b *mockBlockBuilder) EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error) {
	return b.db.EventBlockNumberByEventNameForValidityProver("DepositsProcessed")
}

func (b *mockBlockBuilder) UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error) {
	return b.db.UpsertEventBlockNumberForValidityProver(eventName, blockNumber)
}

func (b *mockBlockBuilder) GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error) {
	fmt.Printf("GetDepositIndexByHash deposit hash: %s\n", depositHash.String())
	deposit, err := b.db.DepositByDepositHash(depositHash)
	if err != nil {
		return nil, new(uint32), err
	}

	depositLeaf := intMaxTree.DepositLeaf{
		RecipientSaltHash: deposit.RecipientSaltHash,
		TokenIndex:        deposit.TokenIndex,
		Amount:            deposit.Amount,
	}

	return &DepositLeafWithId{
		DepositId:   deposit.DepositID,
		DepositLeaf: &depositLeaf,
	}, deposit.DepositIndex, nil
}

func (b *mockBlockBuilder) UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error {
	err := b.db.UpdateDepositIndexByDepositHash(depositHash, depositIndex)
	if err != nil {
		return err
	}

	return nil
}
