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
	// AccountTree   *intMaxTree.AccountTree
	// BlockTree     *intMaxTree.BlockHashTree
	DepositTree   *intMaxTree.KeccakMerkleTree
	DepositLeaves []*intMaxTree.DepositLeaf

	db SQLDriverApp
	// LastPostedBlockNumber         uint32
	// LastGeneratedProofBlockNumber uint32
	// ValidityProofs                map[uint32]string

	// latestWitnessBlockNumber uint32
	MerkleTreeHistory MerkleTreeHistory
	// LastSeenProcessedDepositId    uint64
	// auxInfo                  map[uint32]*mDBApp.BlockContent
	// validityWitnesses        map[uint32]*ValidityWitness
	// DepositTreeRoots           []common.Hash
	// DepositTreeHistory         map[string]*intMaxTree.KeccakMerkleTree // deposit hash -> deposit tree
}

type MockBlockBuilderMemory = mockBlockBuilder

type MockBlockBuilder interface {
	AccountBySenderAddress(_ string) (*uint256.Int, error)
	// AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	// AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedInsertionProof, error)
	AppendBlockTreeLeaf(block *block_post_service.PostedBlock) (uint32, error)
	AppendDepositTreeLeaf(depositHash common.Hash, depositLeaf *intMaxTree.DepositLeaf) (root common.Hash, err error)
	// BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockAuxInfo(blockNumber uint32) (*AuxInfo, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	BlockNumberByDepositIndex(depositIndex uint32) (uint32, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error)
	BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error)
	ConstructSignature(
		txTreeRoot intMaxTypes.Bytes32,
		publicKeysHash intMaxTypes.Bytes32,
		accountIDHash intMaxTypes.Bytes32,
		isRegistrationBlock bool,
		sortedTxs []*MockTxRequest,
	) (*SignatureContent, error)
	// CreateBlockContent(postedBlock *block_post_service.PostedBlock, blockContent *intMaxTypes.BlockContent) (*mDBApp.BlockContentWithProof, error)
	CurrentBlockTreeProof(blockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
	FetchLastDepositIndex() (uint32, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
	GenerateBlock(blockContent *intMaxTypes.BlockContent, postedBlock *block_post_service.PostedBlock) (*BlockWitness, error)
	GenerateBlockWithTxTree(isRegistrationBlock bool, txs []*MockTxRequest) (*BlockWitness, *intMaxTree.TxTree, error)
	GetAccountMembershipProof(blockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error)
	// GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error)
	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)
	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
	LastDepositTreeRoot() (common.Hash, error)
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	// LastValidityWitness() (*ValidityWitness, error)
	LastWitnessGeneratedBlockNumber() uint32
	NextAccountID(blockNumber uint32) (uint64, error)
	ValidityWitness(isRegistrationBlock bool, txs []*MockTxRequest) (*ValidityWitness, error)
	// ProveInclusion(accountId uint64) (*AccountMerkleProof, error)
	ProveInclusionByPublicKey(accountId uint64) (*AccountMerkleProof, uint64, error)
	PublicKeyByAccountID(blockumber uint32, accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	// RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accountID uint64, err error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	SetValidityProof(blockNumber uint32, proof string) error
	SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error
	// UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedUpdateProof, error)
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
	merkleTrees := make(map[uint32]*MerkleTrees)

	blockHashAndSendersMap, _, err := db.ScanBlockHashAndSenders()
	if err != nil {
		panic(err)
	}

	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, []*intMaxTree.BlockHashLeaf{intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())})
	if err != nil {
		panic(err)
	}

	merkleTrees[0] = &MerkleTrees{
		AccountTree:   new(intMaxTree.AccountTree).Set(accountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(blockHashTree),
		DepositLeaves: make([]*intMaxTree.DepositLeaf, 0),
	}
	fmt.Printf("blockHashTree 0 root: %s\n", blockHashTree.GetRoot().String())

	lastProofGeneratedBlockNumber, err := db.LastBlockNumberGeneratedValidityProof()
	if err != nil {
		if err.Error() != "not found" {
			panic(err)
		}

		lastProofGeneratedBlockNumber = 0
	}

	blockHashes := make([]*intMaxTree.BlockHashLeaf, lastProofGeneratedBlockNumber+1)
	defaultPublicKey := new(intMaxAcc.Address).String()                  // zero
	dummyPublicKey := intMaxAcc.NewDummyPublicKey().ToAddress().String() // one
	for blockNumber := uint32(1); blockNumber <= lastProofGeneratedBlockNumber; blockNumber++ {
		blockHashAndSenders, ok := blockHashAndSendersMap[blockNumber]
		if !ok {
			panic(fmt.Sprintf("block number %d not found", blockNumber))
		}

		merkleTrees[blockNumber] = new(MerkleTrees)

		fmt.Printf("blockHashAndSendersMap[%d].BlockHash: %s\n", blockNumber, blockHashAndSenders.BlockHash)
		blockHashes[blockNumber] = intMaxTree.NewBlockHashLeaf(common.HexToHash("0x" + blockHashAndSenders.BlockHash))
		_, err := blockHashTree.AddLeaf(uint32(blockNumber), blockHashes[blockNumber])
		if err != nil {
			panic(err)
		}
		fmt.Printf("blockHashTree %d root: %s\n", blockNumber, blockHashTree.GetRoot().String())
		merkleTrees[blockNumber].BlockHashTree = new(intMaxTree.BlockHashTree).Set(blockHashTree)

		count := 0
		for i, sender := range blockHashAndSenders.Senders {
			if sender.PublicKey == defaultPublicKey || sender.PublicKey == dummyPublicKey {
				continue
			}
			if !sender.IsSigned {
				fmt.Printf("blockHashAndSendersMap[%d].Senders[%d] is not signed\n", blockNumber, i)
				continue
			}
			fmt.Printf("blockHashAndSendersMap[%d].Senders[%d] is valid: %s\n", blockNumber, i, sender.PublicKey)

			count++

			var senderPublicKey *intMaxAcc.PublicKey
			senderPublicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
			if err != nil {
				panic(err)
			}

			if _, ok = accountTree.GetAccountID(senderPublicKey.BigInt()); ok {
				if blockHashAndSenders.IsRegistrationBlock {
					fmt.Printf("blockHashAndSendersMap[%d].Senders[%d] is already registered\n", blockNumber, i)
					continue
				}

				_, err = accountTree.Update(senderPublicKey.BigInt(), blockNumber)
				if err != nil {
					panic(err)
				}
			} else {
				if !blockHashAndSenders.IsRegistrationBlock {
					fmt.Printf("blockHashAndSendersMap[%d].Senders[%d] is not registered\n", blockNumber, i)
					continue
				}

				_, err = accountTree.Insert(senderPublicKey.BigInt(), blockNumber)
				if err != nil {
					panic(err)
				}
			}
		}
		fmt.Printf("blockHashAndSendersMap[%d].Senders count: %d\n", blockNumber, count)
		merkleTrees[blockNumber].AccountTree = new(intMaxTree.AccountTree).Set(accountTree)
	}

	fmt.Printf("size of blockHashTree leaves: %v\n", len(blockHashTree.Leaves))
	for i, leaf := range blockHashTree.Leaves {
		fmt.Printf("blockHashes[%d]: %x\n", i, leaf.Marshal())
	}

	// validityWitness := new(ValidityWitness).Genesis()
	// validityWitnesses := make(map[uint32]*ValidityWitness)
	// validityWitnesses[0] = new(ValidityWitness).Genesis()
	// auxInfo := make(map[uint32]*mDBApp.BlockContent)

	deposits, err := db.ScanDeposits()
	if err != nil {
		panic(err)
	}

	zeroDepositHash := new(intMaxTree.DepositLeaf).SetZero().Hash()
	depositTree, err := intMaxTree.NewKeccakMerkleTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil, zeroDepositHash)
	if err != nil {
		panic(err)
	}
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	depositTreeRootHex := depositTreeRoot.Hex()[2:]

	fmt.Printf("lastProofGeneratedBlockNumber: %d\n", lastProofGeneratedBlockNumber)
	fmt.Printf("depositTreeRootHex: %s\n", depositTreeRootHex)
	blockNumber := uint32(1)
	for blockHashAndSendersMap[blockNumber].DepositTreeRoot == depositTreeRootHex && blockNumber <= lastProofGeneratedBlockNumber {
		fmt.Printf("DepositTreeRoots[%d]: %s\n", blockNumber, blockHashAndSendersMap[blockNumber].DepositTreeRoot)
		merkleTrees[blockNumber].DepositLeaves = make([]*intMaxTree.DepositLeaf, 0)
		blockNumber++
	}

	fmt.Printf("depositTreeRoots[%d]: %s\n", blockNumber, blockHashAndSendersMap[blockNumber].DepositTreeRoot)

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
		depositTreeRoot, err = depositTree.AddLeaf(uint32(depositIndex), depositLeaf.Hash())
		if err != nil {
			panic(err)
		}

		depositTreeRootHex = depositTreeRoot.Hex()[2:]
		depositLeaves = append(depositLeaves, &depositLeaf)
		for blockHashAndSendersMap[blockNumber].DepositTreeRoot == depositTreeRootHex && blockNumber <= lastProofGeneratedBlockNumber {
			fmt.Printf("depositTreeRoots[%d]: %s\n", blockNumber, blockHashAndSendersMap[blockNumber].DepositTreeRoot)
			merkleTrees[blockNumber].DepositLeaves = depositLeaves
			blockNumber++
		}
	}

	return &mockBlockBuilder{
		db: db,
		// AccountTree:   new(intMaxTree.AccountTree).Set(merkleTrees[lastProofGeneratedBlockNumber].AccountTree),
		// BlockTree:     new(intMaxTree.BlockHashTree).Set(merkleTrees[lastProofGeneratedBlockNumber].BlockHashTree),
		DepositTree: new(intMaxTree.KeccakMerkleTree).Set(depositTree),
		// DepositLeaves: merkleTrees[lastProofGeneratedBlockNumber].DepositLeaves,
		MerkleTreeHistory: *NewMerkleTreeHistory(
			lastProofGeneratedBlockNumber,
			merkleTrees,
		),
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

// mockBlockBuilder is not mutable
// TODO: Rename to GenerateBlockWitness
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

	// prevAccountTreeRoot := b.AccountTree.GetRoot()
	// prevBlockTreeRoot := b.BlockTree.GetRoot()
	var accountTree *intMaxTree.AccountTree
	err := b.CopyAccountTree(accountTree, postedBlock.BlockNumber-1)
	if err != nil {
		return nil, err
	}
	prevAccountTreeRoot := accountTree.GetRoot()
	merkleTreeHistory, ok := b.MerkleTreeHistory.MerkleTrees[postedBlock.BlockNumber-1]
	if !ok {
		return nil, fmt.Errorf("merkle tree of block number %d not found", postedBlock.BlockNumber-1)
	}

	prevBlockTreeRoot := merkleTreeHistory.BlockHashTree.GetRoot()

	if signature.IsRegistrationBlock {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, len(blockContent.Senders))
		for i, sender := range blockContent.Senders {
			// accountMembershipProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
			accountMembershipProof, _, err := accountTree.ProveMembership(sender.PublicKey.BigInt())
			if err != nil {
				return nil, errors.New("account membership proof error")
			}

			accountMembershipProofs[i] = *accountMembershipProof
		}

		blockWitness := &BlockWitness{
			Block:               postedBlock,
			Signature:           *signature,
			PublicKeys:          publicKeys,
			PrevAccountTreeRoot: prevAccountTreeRoot,
			PrevBlockTreeRoot:   prevBlockTreeRoot,
			AccountIdPacked:     nil,
			AccountMerkleProofs: AccountMerkleProofsOption{
				IsSome: false,
				Proofs: nil,
			},
			AccountMembershipProofs: AccountMembershipProofsOption{
				IsSome: true,
				Proofs: accountMembershipProofs,
			},
		}

		return blockWitness, nil
	}

	accountMerkleProofs := make([]AccountMerkleProof, len(blockContent.Senders))
	accountIDPackedBytes := make([]byte, numAccountIDPackedBytes)
	for i, sender := range blockContent.Senders {
		accountIDByte := make([]byte, int8Key)
		binary.BigEndian.PutUint64(accountIDByte, sender.AccountID)
		copy(accountIDPackedBytes[i/int8Key:i/int8Key+int5Key], accountIDByte[int8Key-int5Key:])
		accountMembershipProof, _, err := accountTree.ProveMembership(sender.PublicKey.BigInt())
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
		Block:               postedBlock,
		Signature:           *signature,
		PublicKeys:          publicKeys,
		PrevAccountTreeRoot: prevAccountTreeRoot,
		PrevBlockTreeRoot:   prevBlockTreeRoot,
		AccountIdPacked:     accountIDPacked,
		AccountMerkleProofs: AccountMerkleProofsOption{
			IsSome: true,
			Proofs: accountMerkleProofs,
		},
		AccountMembershipProofs: AccountMembershipProofsOption{
			IsSome: false,
			Proofs: nil,
		},
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

func (db *mockBlockBuilder) SetValidityWitness(_blockNumber uint32, witness *ValidityWitness, accountTree *intMaxTree.AccountTree, blockHashTree *intMaxTree.BlockHashTree) error {
	blockNumber := witness.BlockWitness.Block.BlockNumber
	if blockNumber != db.MerkleTreeHistory.lastBlockNumber+1 {
		return fmt.Errorf("new block number is not equal to the last block number + 1: %d != %d + 1", blockNumber, db.MerkleTreeHistory.lastBlockNumber)
	}

	depositTree, err := intMaxTree.NewDepositTree(int32Key)
	if err != nil {
		return err
	}

	deposits, err := db.ScanDeposits()
	if err != nil {
		panic(err)
	}
	depositLeaves := make([]*intMaxTree.DepositLeaf, len(deposits))
	for i, deposit := range deposits {
		depositLeaves[i] = &intMaxTree.DepositLeaf{
			RecipientSaltHash: deposit.RecipientSaltHash,
			TokenIndex:        deposit.TokenIndex,
			Amount:            deposit.Amount,
		}
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	if depositTreeRoot != witness.BlockWitness.Block.DepositRoot {
		// depositLeaves := db.MerkleTreeHistory.MerkleTrees[blockNumber].DepositLeaves // XXX: scan
		for i, deposit := range depositLeaves {
			depositLeaf := intMaxTree.DepositLeaf{
				RecipientSaltHash: deposit.RecipientSaltHash,
				TokenIndex:        deposit.TokenIndex,
				Amount:            deposit.Amount,
			}

			_, err = depositTree.AddLeaf(uint32(i), depositLeaf)
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

	fmt.Printf("blockNumber (SetValidityWitness): %d\n", blockNumber)
	// fmt.Printf("GetAccountMembershipProof root: %s\n", db.AccountTree.GetRoot().String())

	db.MerkleTreeHistory.PushHistory(&MerkleTrees{
		AccountTree: accountTree,
		// BlockHashTree: new(intMaxTree.BlockHashTree).Set(db.BlockTree),
		BlockHashTree: blockHashTree,
		DepositLeaves: depositTree.Leaves,
	})

	return nil
}

func (db *mockBlockBuilder) LastValidityWitness() (*ValidityWitness, error) {
	lastGeneratedProofBlockNumber, err := db.LastGeneratedProofBlockNumber()
	if err != nil {
		return nil, err
	}
	return db.ValidityWitnessByBlockNumber(lastGeneratedProofBlockNumber)
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

	fmt.Printf("auxInfo.PostedBlock.BlockNumber (ValidityWitnessByBlockNumber): %d\n", auxInfo.PostedBlock.BlockNumber)
	blockWitness, err := db.GenerateBlockWithTxTreeFromBlockContent(
		auxInfo.BlockContent,
		auxInfo.PostedBlock,
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("blockNumber (ValidityWitnessByBlockNumber): %d\n", blockNumber)
	fmt.Printf("blockWitness.Block.BlockNumber (ValidityWitnessByBlockNumber): %d\n", blockWitness.Block.BlockNumber)
	if blockNumber != blockWitness.Block.BlockNumber {
		// sanity check
		panic(fmt.Errorf("block number is not equal to block witness block number: %d != %d", blockNumber, blockWitness.Block.BlockNumber))
	}
	fmt.Printf("blockWitness.AccountMembershipProofs (validityWitnessByBlockNumber): %v\n", blockWitness.AccountMembershipProofs.IsSome)
	validityWitness, newAccountTree, newBlockHashTree, err := calculateValidityWitness(db, blockWitness)
	if err != nil {
		return nil, err
	}

	err = db.SetValidityWitness(blockWitness.Block.BlockNumber, validityWitness, newAccountTree, newBlockHashTree)
	if err != nil {
		panic(err)
	}

	return validityWitness, nil
}

// func (db *mockBlockBuilder) AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error) {
// 	return db.AccountTree.GetRoot(), nil
// }

// func (db *mockBlockBuilder) AccountTreeRoot(blockNumber uint32) (*intMaxGP.PoseidonHashOut, error) {
// 	return db.AccountTree.GetRoot(), nil
// }

func (db *mockBlockBuilder) AccountTreeRootByBlockNumber(blockNumber uint32) (*intMaxGP.PoseidonHashOut, error) {
	blockHistory, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return nil, fmt.Errorf("current block number %d not found", blockNumber)
	}

	return blockHistory.AccountTree.GetRoot(), nil
}

func (db *mockBlockBuilder) GetAccountMembershipProof(blockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error) {
	blockHistory, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return nil, fmt.Errorf("current block number %d not found", blockNumber)
	}
	proof, _, err := blockHistory.AccountTree.ProveMembership(publicKey)
	// fmt.Printf("blockNumber: %d\n", blockNumber)
	// fmt.Printf("GetAccountMembershipProof root: %s\n", root.String())
	// proof, _, err := db.AccountTree.ProveMembership(publicKey)
	if err != nil {
		return nil, errors.New("account membership proof error")
	}

	return proof, nil
}

func (db *mockBlockBuilder) ProveInclusionByPublicKey(blockNumber uint32, publicKeyX *big.Int) (*AccountMerkleProof, uint64, error) {
	merkleTreeHistory, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return nil, 0, fmt.Errorf("current block number %d not found", blockNumber)
	}

	accountTree := merkleTreeHistory.AccountTree
	accountId, ok := accountTree.GetAccountID(publicKeyX)
	if !ok {
		return nil, 0, fmt.Errorf("account id not found")
	}

	leaf := accountTree.GetLeaf(accountId)
	proof, _, err := accountTree.Prove(accountId)
	if err != nil {
		return nil, 0, err
	}

	return &AccountMerkleProof{
		MerkleProof: *proof,
		Leaf:        *leaf,
	}, accountId, nil
}

// func (db *mockBlockBuilder) ProveInclusion(accountId uint64) (*AccountMerkleProof, error) {
// 	leaf := db.AccountTree.GetLeaf(accountId)
// 	proof, _, err := db.AccountTree.Prove(accountId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &AccountMerkleProof{
// 		MerkleProof: *proof,
// 		Leaf:        *leaf,
// 	}, nil
// }

// TODO: blockNumber uint32
func (db *mockBlockBuilder) BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error) {
	if blockNumber == nil {
		panic("block number should not be nil")
		// return db.BlockTree.GetRoot(), nil
	}

	blockHistory, ok := db.MerkleTreeHistory.MerkleTrees[*blockNumber]
	if !ok {
		return nil, errors.New("block number not found")
	}

	for i, leaf := range blockHistory.BlockHashTree.Leaves {
		fmt.Printf("blockHistory.BlockHashTree.Leaves[%d] (BlockTreeRoot): %x\n", i, leaf.Marshal())
		fmt.Printf("blockHistory.BlockHashTree.Leaves[%d].Hash() (BlockTreeRoot): %x\n", i, leaf.Hash().Marshal())
	}

	return blockHistory.BlockHashTree.GetRoot(), nil
}

func (db *mockBlockBuilder) BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error) {
	if rootBlockNumber < leafBlockNumber {
		panic(fmt.Errorf("root block number should be greater than or equal to leaf block number: %d < %d", rootBlockNumber, leafBlockNumber))
		// return nil, fmt.Errorf("root block number should be greater than or equal to leaf block number: %d < %d", rootBlockNumber, leafBlockNumber)
	}

	for i := 0; i < len(db.MerkleTreeHistory.MerkleTrees); i++ {
		fmt.Printf("db.MerkleTreeHistory.MerkleTrees[%d].BlockHashTree.Root: %s\n", i, db.MerkleTreeHistory.MerkleTrees[uint32(i)].BlockHashTree.GetRoot().String())
	}

	blockHistory, ok := db.MerkleTreeHistory.MerkleTrees[rootBlockNumber]
	if !ok {
		return nil, errors.Join(ErrRootBlockNumberNotFound, fmt.Errorf("root block number %d not found (BlockTreeProof)", rootBlockNumber))
	}

	if rootBlockNumber == leafBlockNumber {
		_, numLeaves, _ := blockHistory.BlockHashTree.GetCurrentRootCountAndSiblings()
		for i := uint32(0); i < numLeaves; i++ {
			fmt.Printf("blockHistory.BlockHashTree.Leaves[%d] (BlockTreeProof): %x\n", i, blockHistory.BlockHashTree.Leaves[i].Marshal())
		}
	}

	proof, _, err := blockHistory.BlockHashTree.Prove(leafBlockNumber)
	if err != nil {
		return nil, errors.Join(ErrLeafBlockNumberNotFound, err)
	}

	return &proof, nil
}

// func (db *mockBlockBuilder) CurrentBlockTreeProof(blockNumber uint32) (*intMaxTree.MerkleProof, error) {
// 	proof, _, err := db.BlockTree.Prove(blockNumber)
// 	if err != nil {
// 		return nil, errors.New("block tree proof error")
// 	}

// 	return &proof, nil
// }

func (db *mockBlockBuilder) IsSynchronizedDepositIndex(depositIndex uint32) (bool, error) {
	// lastValidityWitness, err := db.LastValidityWitness()
	lastGeneratedProofBlockNumber, err := db.LastGeneratedProofBlockNumber()
	if err != nil {
		return false, err
	}
	fmt.Printf("lastPostedBlockNumber: %d\n", lastGeneratedProofBlockNumber)

	depositLeaves := db.MerkleTreeHistory.MerkleTrees[lastGeneratedProofBlockNumber].DepositLeaves
	fmt.Printf("lastGeneratedProofBlockNumber (IsSynchronizedDepositIndex): %d\n", lastGeneratedProofBlockNumber)
	fmt.Printf("latest deposit index: %d\n", len(depositLeaves))
	fmt.Printf("depositIndex: %d\n", depositIndex)

	if depositIndex >= uint32(len(depositLeaves)) {
		return false, nil
	}

	return true, nil
}

func (db *mockBlockBuilder) DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	fmt.Printf("blockNumber (DepositTreeProof): %d\n", blockNumber)
	depositLeaves := db.MerkleTreeHistory.MerkleTrees[blockNumber].DepositLeaves

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

// The function returns the block number of the first block that was submitted with the specified
// deposit index included in the deposit tree.
func (db *mockBlockBuilder) BlockNumberByDepositIndex(depositIndex uint32) (uint32, error) {
	lastBlockNumber, err := db.db.LastPostedBlockNumber()
	if err != nil {
		return 0, err
	}
	fmt.Printf("lastPostedBlockNumber: %d\n", lastBlockNumber)

	for blockNumber := uint32(1); blockNumber <= lastBlockNumber; blockNumber++ {
		depositLeaves := db.MerkleTreeHistory.MerkleTrees[blockNumber].DepositLeaves
		fmt.Printf("size of deposit leaves: %d\n", len(depositLeaves))
		if depositIndex < uint32(len(depositLeaves)) {
			return blockNumber, nil
		}
	}

	return 0, errors.New("deposit index is out of range")
}

func (db *mockBlockBuilder) AppendBlockTreeLeaf(block *block_post_service.PostedBlock) (blockNumber uint32, err error) {
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	merkleTreeHistory, ok := db.MerkleTreeHistory.MerkleTrees[block.BlockNumber-1]
	if !ok {
		return 0, errors.New("block number not found")
	}

	blockHashTree := merkleTreeHistory.BlockHashTree

	_, blockNumber, _ = blockHashTree.GetCurrentRootCountAndSiblings()
	fmt.Printf("next block number (AppendBlockTreeLeaf): %d\n", blockNumber)
	fmt.Printf("block hashes (AppendBlockTreeLeaf): %v\n", blockHashTree.Leaves)
	if blockNumber != block.BlockNumber {
		return 0, fmt.Errorf("block number is not equal to the current block number: %d != %d", blockNumber, block.BlockNumber)
	}
	fmt.Printf("block hashes: %v", blockHashTree.Leaves)

	fmt.Printf("old block root: %s\n", blockHashTree.GetRoot().String())
	newRoot, err := blockHashTree.AddLeaf(blockNumber, blockHashLeaf)
	if err != nil {
		var ErrBlockTreeAddLeaf = errors.New("block tree add leaf error")
		return 0, errors.Join(ErrBlockTreeAddLeaf, err)
	}
	fmt.Printf("new block root (AppendBlockTreeLeaf): %s\n", newRoot.String())

	return blockNumber, nil
}

// func (db *mockBlockBuilder) AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint32) (*intMaxTree.IndexedInsertionProof, error) {
// 	proof, err := db.AccountTree.Insert(sender, lastBlockNumber)
// 	if err != nil {
// 		// invalid block
// 		var ErrAccountTreeInsert = errors.New("account tree insert error")
// 		return nil, errors.Join(ErrAccountTreeInsert, err)
// 	}

// 	return proof, nil
// }

func (db *mockBlockBuilder) CopyAccountTree(dst *intMaxTree.AccountTree, blockNumber uint32) error {
	src, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return errors.New("block number not found")
	}

	dst.Set(src.AccountTree)

	return nil
}

func (db *mockBlockBuilder) CopyBlockHashTree(dst *intMaxTree.BlockHashTree, blockNumber uint32) error {
	src, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return errors.New("block number not found")
	}

	dst.Set(src.BlockHashTree)

	return nil
}

// func (db *mockBlockBuilder) CopyDepositTree(depositTree *intMaxTree.DepositTree, blockNumber uint32) (err error) {
// 	merkleTreeHistory, ok := db.MerkleTreeHistory.MerkleTrees[blockNumber]
// 	if !ok {
// 		return errors.New("block number not found")
// 	}

// 	depositTree, err = intMaxTree.NewDepositTreeWithInitialLeaves(intMaxTree.DEPOSIT_TREE_HEIGHT, merkleTreeHistory.DepositLeaves)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (db *mockBlockBuilder) UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint32) (*intMaxTree.IndexedUpdateProof, error) {
// 	proof, err := db.AccountTree.Update(sender, lastBlockNumber)
// 	if err != nil {
// 		var ErrAccountTreeUpdate = errors.New("account tree update error")
// 		return nil, errors.Join(ErrAccountTreeUpdate, err)
// 	}

// 	return proof, nil
// }

func (db *mockBlockBuilder) GetAccountTreeLeafByAccountId(blockNumber uint32, sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error) {
	accountTree := new(intMaxTree.AccountTree)
	err := db.CopyAccountTree(accountTree, blockNumber)
	if err != nil {
		var ErrCopyAccountTree = errors.New("copy account tree error")
		return nil, ErrCopyAccountTree
	}
	accountID, ok := accountTree.GetAccountID(sender)
	if !ok {
		return nil, ErrAccountTreeGetAccountID
	}
	prevLeaf := accountTree.GetLeaf(accountID)

	return prevLeaf, nil
}

// func (db *mockBlockBuilder) GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error) {
// 	accountID, ok := db.AccountTree.GetAccountID(sender)
// 	if !ok {
// 		return nil, ErrAccountTreeGetAccountID
// 	}
// 	prevLeaf := db.AccountTree.GetLeaf(accountID)

// 	return prevLeaf, nil
// }

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

			// isDummy := publicKey.BigInt().Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			// fmt.Printf("isDummy: %v, ", isDummy)

			// var leaf *intMaxTree.IndexedMerkleLeaf
			// leaf, err = db.GetAccountTreeLeaf(publicKey.BigInt())
			// if err != nil {
			// 	if err.Error() != ErrAccountTreeGetAccountID.Error() {
			// 		return nil, nil, errors.Join(errors.New("account tree leaf error"), err)
			// 	}
			// }

			// if !isDummy && leaf != nil {
			// 	return nil, nil, errors.New("account already exists")
			// }

			var proof *intMaxTree.IndexedMembershipProof
			proof, err = db.GetAccountMembershipProof(postedBlock.BlockNumber, publicKey.BigInt())
			if err != nil {
				return nil, nil, errors.Join(ErrAccountMembershipProof, err)
			}

			accountMembershipProofs[i] = *proof
		}
	} else {
		accountIDs := make([]uint64, len(publicKeys))
		accountMerkleProofs = make([]AccountMerkleProof, len(publicKeys))
		for i, publicKey := range publicKeys {
			// accountID, ok := db.AccountTree.GetAccountID(publicKey.BigInt())
			// if !ok {
			// 	return nil, nil, errors.New("account id not found")
			// }
			var proof *AccountMerkleProof
			var accountID uint64
			proof, accountID, err = db.ProveInclusionByPublicKey(postedBlock.BlockNumber, publicKey.BigInt())
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

	merkleTreeHistory := db.MerkleTreeHistory.MerkleTrees[postedBlock.BlockNumber+1]
	prevAccountTreeRoot := merkleTreeHistory.AccountTree.GetRoot()
	prevBlockHashTreeRoot := merkleTreeHistory.BlockHashTree.GetRoot()
	blockWitness := &BlockWitness{
		Block:               block,
		Signature:           *signature,
		PublicKeys:          publicKeys,
		PrevAccountTreeRoot: prevAccountTreeRoot,
		PrevBlockTreeRoot:   prevBlockHashTreeRoot,
		AccountIdPacked:     accountIDPacked,
		AccountMerkleProofs: AccountMerkleProofsOption{
			IsSome: true,
			Proofs: accountMerkleProofs,
		},
		AccountMembershipProofs: AccountMembershipProofsOption{
			IsSome: true,
			Proofs: accountMembershipProofs,
		},
	}

	validationPis, invalidReason := blockWitness.MainValidationPublicInputs()
	fmt.Printf("validationPis: %v\n", validationPis)
	isValid := validationPis.IsValid

	if !isValid && len(txs) > 0 {
		panic(fmt.Errorf("should be valid block: %s", invalidReason))
	}

	return blockWitness, txTree, nil
}

func IsValidBlockSenders(blockContent *intMaxTypes.BlockContent, prevAccountTree *intMaxTree.AccountTree) error {
	// isValid := blockContent.IsValid()
	// if isValid != nil {
	// 	return isValid
	// }

	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType
	if isRegistrationBlock {
		for _, sender := range blockContent.Senders {
			isDummy := sender.PublicKey.Equal(intMaxAcc.NewDummyPublicKey())
			if !isDummy {
				if _, ok := prevAccountTree.GetAccountID(sender.PublicKey.BigInt()); ok {
					return ErrAccountAlreadyExists
				}
			}
		}

		return nil
	}

	for _, sender := range blockContent.Senders {
		isDummy := sender.PublicKey.Equal(intMaxAcc.NewDummyPublicKey())
		if !isDummy {
			if _, ok := prevAccountTree.GetAccountID(sender.PublicKey.BigInt()); !ok {
				return ErrAccountTreeGetAccountID
			}
		}
	}

	return nil
}

// func IsValidWithAccountInfo(blockContent *intMaxTypes.BlockContent, accountInfo block_post_service.AccountInfo) error {
// 	if inValidReason := blockContent.IsValid(); inValidReason != nil {
// 		return inValidReason
// 	}

// 	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType
// 	if isRegistrationBlock {
// 		for _, sender := range blockContent.Senders {
// 			isDummy := sender.PublicKey.Equal(intMaxAcc.NewDummyPublicKey())
// 			if !isDummy {
// 				// if _, ok := prevAccountTree.GetAccountID(sender.PublicKey.BigInt()); ok {
// 				// 	return ErrAccountAlreadyExists
// 				// }
// 				accountId, err := accountInfo.AccountBySenderAddress(sender.PublicKey.ToAddress().String())
// 				if err == nil && accountId.Uint64() != 0 {
// 					return ErrAccountAlreadyExists
// 				}
// 			}
// 		}

// 		return nil
// 	}

// 	for _, sender := range blockContent.Senders {
// 		isDummy := sender.PublicKey.Equal(intMaxAcc.NewDummyPublicKey())
// 		if !isDummy {
// 			// if _, ok := prevAccountTree.GetAccountID(sender.PublicKey.BigInt()); !ok {
// 			// 	return ErrAccountTreeGetAccountID
// 			// }
// 			accountId, err := accountInfo.AccountBySenderAddress(sender.PublicKey.ToAddress().String())
// 			if err != nil || accountId.Uint64() == 0 {
// 				return ErrAccountAlreadyExists
// 			}
// 		}
// 	}

// 	return nil
// }

func (db *mockBlockBuilder) GenerateBlockWithTxTreeFromBlockContentAndPrevBlock(
	blockContent *intMaxTypes.BlockContent,
	prevPostedBlock *block_post_service.PostedBlock,
) (*BlockWitness, error) {
	depositRoot, _, _ := db.DepositTree.GetCurrentRootCountAndSiblings()
	signature := NewSignatureContentFromBlockContent(blockContent)

	// if depositRoot != postedBlock.DepositRoot {
	// 	return nil, fmt.Errorf("deposit root is not equal to the last deposit root: %s != %s", depositRoot.Hex(), postedBlock.DepositRoot.Hex())
	// }

	// if signature.Hash() != postedBlock.SignatureHash {
	// 	return nil, fmt.Errorf("signature hash is not equal to the last signature hash: %s != %s", signature.Hash().Hex(), postedBlock.SignatureHash.Hex())
	// }

	postedBlock := &block_post_service.PostedBlock{
		PrevBlockHash: prevPostedBlock.Hash(),
		DepositRoot:   depositRoot,
		SignatureHash: signature.Hash(),
		BlockNumber:   prevPostedBlock.BlockNumber + 1,
	}

	return db.GenerateBlockWithTxTreeFromBlockContent(
		blockContent,
		postedBlock,
	)
}

func (db *mockBlockBuilder) GenerateBlockWithTxTreeFromBlockContent(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*BlockWitness, error) {
	prevBlockNumber := postedBlock.BlockNumber - 1

	// DEBUG
	// for i := 1; i <= int(prevBlockNumber+1); i++ {
	// 	// account root
	// 	accountTreeRoot, err := db.AccountTreeRootByBlockNumber(uint32(i))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("accountTreeRoot[%d]: %s\n", i, accountTreeRoot.String())
	// }

	const numOfSenders = 128
	if len(blockContent.Senders) > numOfSenders {
		// panic("too many txs")
		return nil, errors.New("too many txs")
	}

	// TODO: If not sorted
	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		publicKeys[i].FromBigInt(sender.PublicKey.BigInt())
	}

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, *new(intMaxTypes.Uint256).FromBigInt(dummyPublicKey.BigInt()))
	}

	prevAccountTree := new(intMaxTree.AccountTree)              // prev account tree
	err := db.CopyAccountTree(prevAccountTree, prevBlockNumber) // only reference
	if err != nil {
		var ErrCopyAccountTree = errors.New("copy account tree error")
		return nil, ErrCopyAccountTree
	}

	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	var accountIDPacked *AccountIdPacked
	// var accountMerkleProofs []AccountMerkleProof
	// var accountMembershipProofs []intMaxTree.IndexedMembershipProof
	accountMembershipProofs := AccountMembershipProofsOption{
		IsSome: false,
		Proofs: make([]intMaxTree.IndexedMembershipProof, 0, len(publicKeys)),
	}
	accountMerkleProofs := AccountMerkleProofsOption{
		IsSome: false,
		Proofs: make([]AccountMerkleProof, 0, len(publicKeys)),
	}
	// if invalidReason := IsValidBlockSenders(blockContent, prevAccountTree); invalidReason != nil {
	// isValid := blockContent.IsValid() == nil
	// mainValidationPublicInputs := blockWitness.MainValidationPublicInputs()

	// prevAccountTreeRoot := prevAccountTree.GetRoot()
	// tmpAccountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(prevAccountTreeRoot)
	// if !isValid {
	// 	fmt.Printf("block content %d is invalid (GenerateBlockWithTxTreeFromBlockContent)\n", postedBlock.BlockNumber)
	// } else {
	fmt.Printf("block content %d is valid (GenerateBlockWithTxTreeFromBlockContent)\n", postedBlock.BlockNumber)
	if isRegistrationBlock {
		fmt.Printf("size of publicKeys: %d\n", len(publicKeys))
		// accountMembershipProofs = make([]intMaxTree.IndexedMembershipProof, len(publicKeys))
		accountMembershipProofs.IsSome = true
		for _, publicKey := range publicKeys {
			isDummy := publicKey.BigInt().Cmp(dummyPublicKey.BigInt()) == 0
			_, ok := prevAccountTree.GetAccountID(publicKey.BigInt())
			if ok && !isDummy {
				// If it fails here, the block is not valid.
				fmt.Printf("WARNING: public key %s is invalid\n", publicKey.BigInt())
			}

			proof, prevAccountTreeRoot, err := prevAccountTree.ProveMembership(publicKey.BigInt())
			if err != nil {
				return nil, errors.New("account membership proof error")
			}
			// var lastSeenBlockNumber uint32 = 0
			// if blockContent.Senders[i].IsSigned {
			// 	lastSeenBlockNumber = postedBlock.BlockNumber
			// }
			// _, err = prevAccountTree.Insert(publicKey.BigInt(), lastSeenBlockNumber)
			// if err == nil {
			// 	return nil, err
			// }
			// tmpAccountTreeRoot = prevAccountTree.GetRoot()

			err = proof.Verify(publicKey.BigInt(), prevAccountTreeRoot)
			if err != nil {
				fmt.Printf("length of account leaves: %d\n", len(prevAccountTree.Leaves()))
				for key, leaf := range prevAccountTree.Leaves() {
					fmt.Printf("leaves[%d]: %+v\n", key, leaf)
				}
				for i, sibling := range proof.LeafProof.Siblings {
					fmt.Printf("sibling[%d]: %v\n", i, sibling)
				}
				fmt.Printf("leaf index: %d\n", proof.LeafIndex)
				fmt.Printf("leaf: %+v\n", proof.Leaf)
				fmt.Printf("prevAccountTreeRoot: %s\n", prevAccountTreeRoot.String())

				panic(fmt.Errorf("account membership proof verification error: %w", err))
			}

			// if !isDummy {
			// 	fmt.Printf("length of account leaves: %d\n", len(currentAccountTree.Leaves()))
			// 	for i, sibling := range proof.LeafProof.Siblings {
			// 		fmt.Printf("sibling[%d]: %v\n", i, sibling)
			// 	}
			// 	fmt.Printf("leaf index: %d\n", proof.LeafIndex)
			// 	fmt.Printf("leaf: %+v\n", proof.Leaf)
			// 	fmt.Printf("accountTreeRoot: %s\n", accountTreeRoot.String())
			// }

			accountMembershipProofs.Proofs = append(accountMembershipProofs.Proofs, *proof)
		}
	} else {
		accountIDs := make([]uint64, len(publicKeys))
		// accountMerkleProofs = make([]AccountMerkleProof, len(publicKeys))
		accountMerkleProofs.IsSome = true
		for i, publicKey := range publicKeys {
			isDummy := publicKey.BigInt().Cmp(dummyPublicKey.BigInt()) == 0
			accountID, ok := prevAccountTree.GetAccountID(publicKey.BigInt())
			if !ok && !isDummy {
				// If it fails here, the block is not valid.
				fmt.Printf("WARNING: public key %s is invalid\n", publicKey.BigInt())
			}

			// proof, err := db.ProveInclusion(accountID)
			prevLeaf := prevAccountTree.GetLeaf(accountID)
			merkleProof, _, err := prevAccountTree.Prove(accountID)
			if err != nil {
				return nil, errors.New("account inclusion proof error")
			}

			accountIDs[i] = accountID
			accountMerkleProofs.Proofs = append(accountMerkleProofs.Proofs, AccountMerkleProof{
				MerkleProof: *merkleProof,
				Leaf:        *prevLeaf,
			})
		}

		accountIDPacked = new(AccountIdPacked).Pack(accountIDs)
		// accountIDHash = GetAccountIDsHash(accountIDs)
	}
	// }

	txTreeRoot := intMaxTypes.Bytes32{}
	txTreeRoot.FromBytes(blockContent.TxTreeRoot[:])
	signature := NewSignatureContentFromBlockContent(blockContent)

	prevAccountTreeRoot := db.MerkleTreeHistory.MerkleTrees[prevBlockNumber].AccountTree.GetRoot()
	prevBlockTreeRoot := db.MerkleTreeHistory.MerkleTrees[prevBlockNumber].BlockHashTree.GetRoot()
	blockWitness := &BlockWitness{
		Block:                   postedBlock,
		Signature:               *signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     accountMerkleProofs,
		AccountMembershipProofs: accountMembershipProofs,
	}

	// validationPis := blockWitness.MainValidationPublicInputs()
	// fmt.Printf("validationPis: %v\n", validationPis)
	// if !validationPis.IsValid && len(blockContent.Senders) > 0 {
	// 	// Despite non-empty block, the block is not valid.
	// 	return nil, ErrBlockShouldBeValid
	// }

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
	fmt.Printf("---------------------- UpdateValidityWitness ----------------------\n")
	blockWitness, err := db.GenerateBlockWithTxTreeFromBlockContentAndPrevBlock(
		blockContent,
		prevValidityWitness.BlockWitness.Block,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("blockWitness.Block.BlockNumber (UpdateValidityWitness): %d\n", blockWitness.Block.BlockNumber)
	// latestIntMaxBlockNumber := db.LastWitnessGeneratedBlockNumber()
	if blockWitness.Block.BlockNumber != prevValidityWitness.BlockWitness.Block.BlockNumber+1 {
		fmt.Printf("latestIntMaxBlockNumber: %d\n", prevValidityWitness.BlockWitness.Block.BlockNumber)
		return nil, errors.New("block number is not equal to the last block number + 1")
	}
	prevPis := prevValidityWitness.ValidityPublicInputs()
	validityWitness, err := updateValidityWitnessWithConsistencyCheck(
		db,
		blockWitness,
		prevPis,
	)
	if err != nil {
		if errors.Is(err, ErrRootBlockNumberNotFound) {
			return nil, ErrRootBlockNumberNotFound
		}

		panic(err)
	}

	return validityWitness, nil
}

func updateValidityWitnessWithConsistencyCheck(db BlockBuilderStorage, blockWitness *BlockWitness, prevPis *ValidityPublicInputs) (*ValidityWitness, error) {
	fmt.Printf("---------------------- updateValidityWitnessWithConsistencyCheck ----------------------\n")
	// latestIntMaxBlockNumber := db.LastWitnessGeneratedBlockNumber()
	// prevPis := prevValidityWitness.ValidityPublicInputs()
	// blockWitness.Block.BlockNumber != latestIntMaxBlockNumber+1
	if blockWitness.Block.BlockNumber > prevPis.PublicState.BlockNumber+1 {
		fmt.Printf("blockWitness.Block.BlockNumber (generateValidityWitness): %d\n", blockWitness.Block.BlockNumber)
		fmt.Printf("prevPis.PublicState.BlockNumber (generateValidityWitness): %d\n", prevPis.PublicState.BlockNumber)
		return nil, errors.New("block number is not greater than the last block number")
	}

	// accountTreeRoot, err := db.AccountTreeRoot()
	// if err != nil {
	// 	return nil, errors.New("account tree root error")
	// }
	// if !prevPis.PublicState.AccountTreeRoot.Equal(accountTreeRoot) {
	// 	fmt.Printf("prevPis.PublicState.AccountTreeRoot is not the same with accountTreeRoot, %s != %s", prevPis.PublicState.AccountTreeRoot.String(), accountTreeRoot.String())
	// 	return nil, errors.New("account tree root is not equal to the last account tree root")
	// }

	prevBlockTreeRoot, err := db.BlockTreeRoot(&prevPis.PublicState.BlockNumber)
	if err != nil {
		return nil, errors.New("block tree root error")
	}
	if prevPis.IsValidBlock {
		fmt.Printf("block number %d is valid (updateValidityWitness)\n", prevPis.PublicState.BlockNumber+1)
	} else {
		fmt.Printf("block number %d is invalid (updateValidityWitness)\n", prevPis.PublicState.BlockNumber+1)
	}
	fmt.Printf("prevBlockTreeRoot (update): %s\n", prevBlockTreeRoot.String())
	if !prevPis.PublicState.BlockTreeRoot.Equal(prevBlockTreeRoot) {
		fmt.Printf("prevPis.PublicState.BlockTreeRoot is not the same with blockTreeRoot, %s != %s", prevPis.PublicState.BlockTreeRoot.String(), prevBlockTreeRoot.String())
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}
	// blockMerkleProof, err := db.BlockTreeProof(prevPis.PublicState.BlockNumber+1, blockWitness.Block.BlockNumber)
	// if err != nil {
	// 	var ErrBlockTreeProve = errors.New("block tree prove error")
	// 	return nil, errors.Join(ErrBlockTreeProve, err)
	// }

	// newBlockHashTree := new(intMaxTree.BlockHashTree).Set(
	// 	db.MerkleTreeHistory.MerkleTrees[db.MerkleTreeHistory.lastBlockNumber].BlockHashTree,
	// addedBlockNumber, err := db.AppendBlockTreeLeaf(blockWitness.Block)
	// if err != nil {
	// 	return nil, fmt.Errorf("append block tree leaf error: %w", err)
	// }

	// if addedBlockNumber != blockWitness.Block.BlockNumber {
	// 	return nil, errors.New("block number is not equal to the added block number")
	// }

	validityWitness, newAccountTree, newBlockHashTree, err := calculateValidityWitnessWithMerkleProofs(db, blockWitness)
	if err != nil {
		// if errors.Is(err, ErrRootBlockNumberNotFound) {
		// 	return nil, ErrRootBlockNumberNotFound
		// }

		return nil, fmt.Errorf("failed to calculate validity witness: %w", err)
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
	fmt.Printf("validityPis (updateValidityWitnessWithConsistencyCheck): %s\n", encodedValidityPis)

	fmt.Printf("SetValidityWitness SenderFlag: %v\n", validityWitness.BlockWitness.Signature.SenderFlag)
	err = db.SetValidityWitness(blockWitness.Block.BlockNumber, validityWitness, newAccountTree, newBlockHashTree)
	if err != nil {
		panic(err)
	}

	return validityWitness, nil
}

func calculateValidityWitness(db BlockBuilderStorage, blockWitness *BlockWitness) (validityWitness *ValidityWitness, newAccountTree *intMaxTree.AccountTree, newBlockHashTree *intMaxTree.BlockHashTree, err error) {
	fmt.Printf("---------------------- calculateValidityWitness ----------------------\n")
	fmt.Printf("blockWitness.AccountMembershipProofs: %v\n", blockWitness.AccountMembershipProofs.IsSome)
	mainValidationPublicInputs, invalidReason := blockWitness.MainValidationPublicInputs()
	fmt.Printf("mainValidationPublicInputs.IsValid: %v\n", mainValidationPublicInputs.IsValid)
	if !mainValidationPublicInputs.IsValid {
		fmt.Printf("WARNING: invalid reason (calculateValidityWitness): %s\n", invalidReason)
	}
	fmt.Printf("mainValidationPublicInputs.BlockNumber: %d\n", mainValidationPublicInputs.BlockNumber)
	// prevBlockNumber := mainValidationPublicInputs.BlockNumber - 1
	// prevBlockTreeRoot, err := db.BlockTreeRoot(&prevBlockNumber)
	// if err != nil {
	// 	return nil, nil, nil, fmt.Errorf("block tree root error: %w", err)
	// }

	// blockMerkleProof, err := db.BlockTreeProof(prevBlockNumber, blockWitness.Block.BlockNumber)
	// if err != nil {
	// 	var ErrBlockTreeProve = errors.New("block tree prove error")
	// 	return nil, errors.Join(ErrBlockTreeProve, err)
	// }

	return calculateValidityWitnessWithMerkleProofs(db, blockWitness)
}

func calculateValidityWitnessWithMerkleProofs(
	db BlockBuilderStorage,
	blockWitness *BlockWitness,
	// _prevBlockTreeRoot *intMaxGP.PoseidonHashOut,
) (validityWitness *ValidityWitness, newAccountTree *intMaxTree.AccountTree, newBlockHashTree *intMaxTree.BlockHashTree, err error) {
	// Only simulate the account tree.
	prevAccountTree := new(intMaxTree.AccountTree)
	err = db.CopyAccountTree(prevAccountTree, blockWitness.Block.BlockNumber-1)
	if err != nil {
		return nil, nil, nil, errors.New("copy account tree error")
	}

	prevBlockHashTree := new(intMaxTree.BlockHashTree)
	err = db.CopyBlockHashTree(prevBlockHashTree, blockWitness.Block.BlockNumber-1)
	if err != nil {
		return nil, nil, nil, errors.New("copy block hash tree error")
	}

	return calculateValidityWitnessWithMerkleProofsInner(blockWitness, prevBlockHashTree, prevAccountTree)
}

func calculateValidityWitnessWithMerkleProofsInner(
	blockWitness *BlockWitness,
	prevBlockHashTree *intMaxTree.BlockHashTree,
	prevAccountTree *intMaxTree.AccountTree,
) (*ValidityWitness, *intMaxTree.AccountTree, *intMaxTree.BlockHashTree, error) {
	fmt.Printf("BlockTreeProof (calculateValidityWitnessWithMerkleProofs): %d\n", blockWitness.Block.BlockNumber)
	// blockMerkleProof, err := db.BlockTreeProof(blockWitness.Block.BlockNumber, blockWitness.Block.BlockNumber)
	// if err != nil {
	// 	return nil, nil, nil, errors.Join(ErrLeafBlockNumberNotFound, err)
	// }
	blockMerkleProof, prevBlockTreeRoot, err := prevBlockHashTree.Prove(blockWitness.Block.BlockNumber)
	if err != nil {
		// if errors.Is(err, ErrRootBlockNumberNotFound) {
		// 	return nil, nil, ErrRootBlockNumberNotFound
		// }

		return nil, nil, nil, fmt.Errorf("block tree prove error: %w", err)
	}

	// debug
	// Verify that the Merkle proof for the block hash tree is correct in its old state.
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	err = blockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		prevBlockTreeRoot,
	)
	if err != nil {
		panic(fmt.Errorf("old block merkle proof is invalid: %w", err))
	}

	blockHashLeaf := intMaxTree.NewBlockHashLeaf(blockWitness.Block.Hash())
	newBlockTreeRoot, err := prevBlockHashTree.AddLeaf(blockWitness.Block.BlockNumber, blockHashLeaf)
	// newBlockTreeRoot, err := db.BlockTreeRoot(&blockWitness.Block.BlockNumber)
	if err != nil {
		return nil, nil, nil, errors.New("block tree root error")
	}
	err = blockMerkleProof.Verify(
		blockHashLeaf.Hash(),
		int(blockWitness.Block.BlockNumber),
		newBlockTreeRoot,
	)
	if err != nil {
		fmt.Printf("blockHashLeaf.Hash(): %s\n", blockHashLeaf.Hash().String())
		fmt.Printf("blockWitness.Block.BlockNumber: %d\n", blockWitness.Block.BlockNumber)
		fmt.Printf("prevBlockTreeRoot: %s\n", prevBlockTreeRoot.String())
		fmt.Printf("newBlockTreeRoot: %s\n", newBlockTreeRoot.String())
		for i, sibling := range blockMerkleProof.Siblings {
			fmt.Printf("sibling[%d]: %s\n", i, sibling.String())
		}
		panic("new block merkle proof is invalid")
	}

	senderLeaves := getSenderLeaves(blockWitness.PublicKeys, blockWitness.Signature.SenderFlag)

	fmt.Printf("blockWitness accountMembershipProof1: %v\n", blockWitness.AccountMembershipProofs.IsSome)
	blockPis, invalidReason := blockWitness.MainValidationPublicInputs()

	accountRegistrationProofsWitness := AccountRegistrationProofsOption{
		IsSome: false,
		Proofs: nil,
	}
	fmt.Printf("blockNumber: %d\n", blockPis.BlockNumber)
	fmt.Printf("(calculateValidityWitnessWithMerkleProofs) blockPis.IsValid: %v\n", blockPis.IsValid)
	if !blockPis.IsValid {
		fmt.Printf("WARNING: invalid reason (calculateValidityWitnessWithMerkleProofs): %s\n", invalidReason)
	}
	if blockPis.IsValid && blockPis.IsRegistrationBlock {
		accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0, len(senderLeaves))
		for _, senderLeaf := range senderLeaves {
			lastBlockNumber := blockPis.BlockNumber
			if !senderLeaf.IsValid {
				lastBlockNumber = 0
			}

			var proof *intMaxTree.IndexedInsertionProof
			isDummy := senderLeaf.Sender.Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			if isDummy {
				proof = intMaxTree.NewDummyAccountRegistrationProof(intMaxTree.ACCOUNT_TREE_HEIGHT)
			} else if _, ok := prevAccountTree.GetAccountID(senderLeaf.Sender); ok {
				proof = intMaxTree.NewDummyAccountRegistrationProof(intMaxTree.ACCOUNT_TREE_HEIGHT)
			} else {
				proof, err = prevAccountTree.Insert(senderLeaf.Sender, lastBlockNumber)
				if err != nil {
					var ErrAppendAccountTreeLeaf = errors.New("append account tree leaf error")
					return nil, nil, nil, errors.Join(ErrAppendAccountTreeLeaf, err)
				}
			}

			accountRegistrationProofs = append(accountRegistrationProofs, *proof)
		}

		accountRegistrationProofsWitness = AccountRegistrationProofsOption{
			IsSome: true,
			Proofs: accountRegistrationProofs,
		}
	}

	accountUpdateProofsWitness := AccountUpdateProofsOption{
		IsSome: false,
		Proofs: nil,
	}
	if blockPis.IsValid && !blockPis.IsRegistrationBlock {
		accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0, len(senderLeaves))
		for _, senderLeaf := range senderLeaves {
			var prevLeaf *intMaxTree.IndexedMerkleLeaf
			var proof *intMaxTree.IndexedUpdateProof
			// prevLeaf, err = db.GetAccountTreeLeaf(senderLeaf.Sender)
			accountID, ok := prevAccountTree.GetAccountID(senderLeaf.Sender)
			if !ok {
				return nil, nil, nil, ErrAccountTreeGetAccountID
			}
			prevLeaf = prevAccountTree.GetLeaf(accountID)
			if err != nil {
				fmt.Printf("WARNING: sender %d is already exist\n", senderLeaf.Sender)
				// var ErrAccountTreeLeaf = errors.New("account tree leaf error")
				// return nil, nil, errors.Join(ErrAccountTreeLeaf, err)
				proof, err = prevAccountTree.Update(big.NewInt(0), 0)
				if err != nil {
					var ErrAccountTreeUpdate = errors.New("account tree update error")
					return nil, nil, nil, errors.Join(ErrAccountTreeUpdate, err)
				}
			} else {
				lastBlockNumber := blockPis.BlockNumber
				if !senderLeaf.IsValid {
					lastBlockNumber = uint32(prevLeaf.Value)
				}
				proof, err = prevAccountTree.Update(senderLeaf.Sender, lastBlockNumber)
				if err != nil {
					var ErrAccountTreeUpdate = errors.New("account tree update error")
					return nil, nil, nil, errors.Join(ErrAccountTreeUpdate, err)
				}
			}

			accountUpdateProofs = append(accountUpdateProofs, *proof)
		}

		accountUpdateProofsWitness = AccountUpdateProofsOption{
			IsSome: true,
			Proofs: accountUpdateProofs,
		}
	}

	fmt.Printf("validity_witness prev_account_tree_root: %v\n", blockWitness.PrevAccountTreeRoot.String())
	fmt.Printf("validity_witness accountRegistrationProofsWitness: %v\n", accountRegistrationProofsWitness)
	return &ValidityWitness{
		BlockWitness: blockWitness,
		ValidityTransitionWitness: &ValidityTransitionWitness{
			SenderLeaves:              senderLeaves,
			BlockMerkleProof:          blockMerkleProof,
			AccountRegistrationProofs: accountRegistrationProofsWitness,
			AccountUpdateProofs:       accountUpdateProofsWitness,
		},
	}, prevAccountTree, prevBlockHashTree, nil
}

func (b *mockBlockBuilder) LastWitnessGeneratedBlockNumber() uint32 {
	return b.MerkleTreeHistory.lastBlockNumber
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
		return nil, ErrBlockContentByBlockNumber
	}

	encodedValidityProof := base64.StdEncoding.EncodeToString(blockContent.ValidityProof)

	return &encodedValidityProof, nil
}

func (b *mockBlockBuilder) LastGeneratedProofBlockNumber() (uint32, error) {
	lastValidityProof, err := b.db.LastBlockValidityProof()
	if err != nil {
		if err.Error() == "not found" {
			return 0, nil
		}

		return 0, err
	}

	return lastValidityProof.BlockNumber, nil
}

func (b *mockBlockBuilder) SetValidityProof(blockHash common.Hash, proof string) error {
	validityProof, err := base64.StdEncoding.DecodeString(proof)
	if err != nil {
		return err
	}

	_, err = b.db.CreateValidityProof(blockHash, validityProof)
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
		return nil, errors.Join(ErrBlockContentByBlockNumber, err)
	}

	return blockAuxInfoFromBlockContent(auxInfo)
}

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

func (b *mockBlockBuilder) BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error) {
	return b.db.BlockContentByTxRoot(txRoot)
}

func (b *mockBlockBuilder) NextAccountID(blockNumber uint32) (uint64, error) {
	merkleTreeHistory, ok := b.MerkleTreeHistory.MerkleTrees[blockNumber]
	if !ok {
		return 0, errors.New("merkle tree history not found")
	}

	accountTree := merkleTreeHistory.AccountTree
	return uint64(accountTree.Count()), nil
}

func (b *mockBlockBuilder) ScanDeposits() ([]*mDBApp.Deposit, error) {
	return b.db.ScanDeposits()
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
