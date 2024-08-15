package block_validity_prover

import (
	"encoding/binary"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type PostedBlock struct {
	// The previous block hash.
	PrevBlockHash common.Hash
	// The block number, which is the latest block number in the Rollup contract plus 1.
	BlockNumber uint32
	// The deposit root at the time of block posting (written in the Rollup contract).
	DepositRoot common.Hash
	// The hash value that the Block Builder must provide to the Rollup contract when posting a new block.
	SignatureHash common.Hash
}

func NewPostedBlock(prevBlockHash, depositRoot common.Hash, blockNumber uint32, signatureHash common.Hash) *PostedBlock {
	return &PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   blockNumber,
		DepositRoot:   depositRoot,
		SignatureHash: signatureHash,
	}
}

func (pb *PostedBlock) Set(other *PostedBlock) *PostedBlock {
	copy(pb.PrevBlockHash[:], other.PrevBlockHash[:])
	copy(pb.DepositRoot[:], other.DepositRoot[:])
	copy(pb.SignatureHash[:], other.SignatureHash[:])
	pb.BlockNumber = other.BlockNumber

	return pb
}

func (pb *PostedBlock) Equals(other *PostedBlock) bool {
	return pb.PrevBlockHash == other.PrevBlockHash &&
		pb.DepositRoot == other.DepositRoot &&
		pb.SignatureHash == other.SignatureHash &&
		pb.BlockNumber == other.BlockNumber
}

func (pb *PostedBlock) Genesis() *PostedBlock {
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()

	return NewPostedBlock(common.Hash{}, depositTreeRoot, 0, common.Hash{})
}

func (pb *PostedBlock) Marshal() []byte {
	const int4Key = 4

	data := make([]byte, 0)

	data = append(data, pb.PrevBlockHash.Bytes()...)
	data = append(data, pb.DepositRoot.Bytes()...)
	data = append(data, pb.SignatureHash.Bytes()...)
	blockNumberBytes := [int4Key]byte{}
	binary.BigEndian.PutUint32(blockNumberBytes[:], pb.BlockNumber)
	data = append(data, blockNumberBytes[:]...)

	return data
}

func CommonHashToUint32Slice(h common.Hash) []uint32 {
	b := intMaxTypes.Bytes32{}
	b.FromBytes(h[:])

	return b[:]
}

func (pb *PostedBlock) Uint32Slice() []uint32 {
	var buf []uint32
	buf = append(buf, CommonHashToUint32Slice(pb.PrevBlockHash)...)
	buf = append(buf, CommonHashToUint32Slice(pb.DepositRoot)...)
	buf = append(buf, CommonHashToUint32Slice(pb.SignatureHash)...)
	buf = append(buf, pb.BlockNumber)

	return buf
}

func (pb *PostedBlock) Hash() common.Hash {
	return crypto.Keccak256Hash(intMaxTypes.Uint32SliceToBytes(pb.Uint32Slice()))
}

type PublicState struct {
	BlockTreeRoot       intMaxGP.PoseidonHashOut
	PrevAccountTreeRoot intMaxGP.PoseidonHashOut
	AccountTreeRoot     intMaxGP.PoseidonHashOut
	DepositTreeRoot     common.Hash
	BlockHash           common.Hash
	BlockNumber         uint32
}

func (ps *PublicState) Genesis() *PublicState {
	blockTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	genesisBlockHash := new(PostedBlock).Genesis().Hash()
	blockTreeRoot, err := blockTree.AddLeaf(0, intMaxTree.NewBlockHashLeaf(genesisBlockHash))
	if err != nil {
		panic(err)
	}

	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	prevAccountTreeRoot := accountTree.GetRoot()
	accountTreeRoot := accountTree.GetRoot()
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	return &PublicState{
		BlockTreeRoot:       *blockTreeRoot,
		PrevAccountTreeRoot: prevAccountTreeRoot,
		AccountTreeRoot:     accountTreeRoot,
		DepositTreeRoot:     depositTreeRoot,
		BlockHash:           genesisBlockHash,
		BlockNumber:         1,
	}
}

type ValidityPublicInputs struct {
	PublicState    *PublicState
	TxTreeRoot     intMaxTypes.Bytes32
	SenderTreeRoot intMaxGP.PoseidonHashOut
	IsValidBlock   bool
}

func (vpi *ValidityPublicInputs) Genesis() *ValidityPublicInputs {
	txTreeRoot := intMaxTypes.Bytes32{}
	senderTreeRoot := new(intMaxGP.PoseidonHashOut).SetZero()
	isValidBlock := false

	return &ValidityPublicInputs{
		PublicState:    new(PublicState).Genesis(),
		TxTreeRoot:     txTreeRoot,
		SenderTreeRoot: *senderTreeRoot,
		IsValidBlock:   isValidBlock,
	}
}

type SenderLeaf struct {
	Sender  *big.Int
	IsValid bool
}

type AccountRegistrationProofs struct {
	Proofs  []intMaxTree.IndexedInsertionProof
	IsValid bool
}

type AccountUpdateProofs struct {
	Proofs  []intMaxTree.IndexedUpdateProof
	IsValid bool
}

type ValidityTransitionWitness struct {
	SenderLeaves              []SenderLeaf
	BlockMerkleProof          intMaxTree.MerkleProof
	AccountRegistrationProofs AccountRegistrationProofs
	AccountUpdateProofs       AccountUpdateProofs
}

func (vtw *ValidityTransitionWitness) Genesis() *ValidityTransitionWitness {
	senderLeaves := make([]SenderLeaf, 0)
	accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0)
	accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0)
	blockHashTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	blockMerkleProof, _, err := blockHashTree.Prove(0)
	if err != nil {
		panic(err)
	}

	return &ValidityTransitionWitness{
		SenderLeaves:     senderLeaves,
		BlockMerkleProof: blockMerkleProof,
		AccountRegistrationProofs: AccountRegistrationProofs{
			IsValid: false,
			Proofs:  accountRegistrationProofs,
		},
		AccountUpdateProofs: AccountUpdateProofs{
			IsValid: false,
			Proofs:  accountUpdateProofs,
		},
	}
}

type AccountMerkleProof struct {
	MerkleProof intMaxTree.IndexedMerkleProof
	Leaf        intMaxTree.IndexedMerkleLeaf
}

type IndexedMembershipProof struct {
	IsIncluded bool
	LeafProof  intMaxTree.IndexedMerkleProof
	LeafIndex  uint
	Leaf       intMaxTree.IndexedMerkleLeaf
}

const accountIdPackedLen = 160

type AccountIdPacked = [accountIdPackedLen]uint32

type SignatureContent struct {
	IsRegistrationBlock bool
	TxTreeRoot          intMaxTypes.Bytes32
	SenderFlag          intMaxTypes.Bytes16
	PublicKeyHash       intMaxTypes.Bytes32
	AccountIDHash       intMaxTypes.Bytes32
	AggPublicKey        intMaxTypes.FlatG1
	AggSignature        intMaxTypes.FlatG2
	MessagePoint        intMaxTypes.FlatG2
}

func (sc *SignatureContent) Set(other SignatureContent) *SignatureContent {
	sc.IsRegistrationBlock = other.IsRegistrationBlock
	sc.TxTreeRoot = other.TxTreeRoot
	sc.SenderFlag = other.SenderFlag
	sc.PublicKeyHash = other.PublicKeyHash
	sc.AccountIDHash = other.AccountIDHash
	sc.AggPublicKey = other.AggPublicKey
	sc.AggSignature = other.AggSignature
	sc.MessagePoint = other.MessagePoint

	return sc
}

type BlockWitness struct {
	Block                   PostedBlock
	Signature               SignatureContent
	PublicKeys              []intMaxTypes.Uint256
	PrevAccountTreeRoot     intMaxTree.PoseidonHashOut
	PrevBlockTreeRoot       intMaxTree.PoseidonHashOut
	AccountIdPacked         *AccountIdPacked          // in account id case
	AccountMerkleProofs     *[]AccountMerkleProof     // in account id case
	AccountMembershipProofs *[]IndexedMembershipProof // in pubkey case
}

func (bw *BlockWitness) Genesis() *BlockWitness {
	blockHashTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}
	prevBlockTreeRoot, _, _ := blockHashTree.GetCurrentRootCountAndSiblings()
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	prevAccountTreeRoot := accountTree.GetRoot()

	return &BlockWitness{
		Block:                   *new(PostedBlock).Genesis(),
		Signature:               SignatureContent{},
		PublicKeys:              make([]intMaxTypes.Uint256, 0),
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         nil,
		AccountMerkleProofs:     nil,
		AccountMembershipProofs: nil,
	}
}

type MainValidationPublicInputs struct {
	PrevBlockHash       common.Hash
	BlockHash           common.Hash
	DepositTreeRoot     common.Hash
	AccountTreeRoot     intMaxGP.PoseidonHashOut
	TxTreeRoot          intMaxTypes.Bytes32
	SenderTreeRoot      intMaxGP.PoseidonHashOut
	BlockNumber         uint32
	IsRegistrationBlock bool
	IsValid             bool
}

// TODO
func GetPublicKeysHash(publicKeys *[]intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	return intMaxTypes.Bytes32{}
}

type AccountExclusionValue struct {
	IsValid bool
}

// TODO
func NewAccountExclusionValue(
	accountTreeRoot intMaxTree.PoseidonHashOut,
	accountMembershipProofs []IndexedMembershipProof,
	publicKeys []intMaxTypes.Uint256,
) *AccountExclusionValue {
	return &AccountExclusionValue{
		IsValid: true,
	}
}

type AccountInclusionValue struct {
	IsValid bool
}

// TODO
func NewAccountInclusionValue(
	accountTreeRoot intMaxTree.PoseidonHashOut,
	accountIdPacked AccountIdPacked,
	accountMerkleProofs []AccountMerkleProof,
	publicKeys []intMaxTypes.Uint256,
) *AccountInclusionValue {
	return &AccountInclusionValue{
		IsValid: true,
	}
}

type FormatValidationValue struct {
	IsValid bool
}

// TODO
func NewFormatValidationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *FormatValidationValue {
	return &FormatValidationValue{
		IsValid: true,
	}
}

type AggregationValue struct {
	IsValid bool
}

// TODO
func NewAggregationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *AggregationValue {
	return &AggregationValue{
		IsValid: true,
	}
}

// TODO
func GetSenderTreeRoot(publicKeys *[]intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) intMaxGP.PoseidonHashOut {
	return intMaxGP.PoseidonHashOut{}
}

func (w *BlockWitness) ToMainValidationPublicInputs() *MainValidationPublicInputs {
	if new(PostedBlock).Genesis().Equals(&w.Block) {
		validityPis := new(ValidityPublicInputs).Genesis()
		return &MainValidationPublicInputs{
			PrevBlockHash:       new(PostedBlock).Genesis().PrevBlockHash,
			BlockHash:           validityPis.PublicState.BlockHash,
			DepositTreeRoot:     validityPis.PublicState.DepositTreeRoot,
			AccountTreeRoot:     validityPis.PublicState.AccountTreeRoot,
			TxTreeRoot:          validityPis.TxTreeRoot,
			SenderTreeRoot:      validityPis.SenderTreeRoot,
			BlockNumber:         validityPis.PublicState.BlockNumber,
			IsRegistrationBlock: false, // genesis block is not a registration block
			IsValid:             validityPis.IsValidBlock,
		}
	}

	result := true
	block := new(PostedBlock).Set(&w.Block)
	signature := new(SignatureContent).Set(w.Signature)
	publicKeys := make([]intMaxTypes.Uint256, len(w.PublicKeys))
	copy(publicKeys, w.PublicKeys)

	accountTreeRoot := w.PrevAccountTreeRoot

	publicKeysHash := GetPublicKeysHash(&publicKeys)
	isRegistrationBlock := signature.IsRegistrationBlock
	isPubkeyEq := signature.PublicKeyHash == publicKeysHash
	if isRegistrationBlock {
		if !isPubkeyEq {
			panic("pubkey hash mismatch")
		}
	} else {
		result = result && isPubkeyEq
	}
	if isRegistrationBlock {
		if w.AccountMembershipProofs == nil {
			panic("account membership proofs should be given")
		}

		// Account exclusion verification
		accountExclusionValue := NewAccountExclusionValue(
			accountTreeRoot,
			*w.AccountMembershipProofs,
			publicKeys,
		)
		result = result && accountExclusionValue.IsValid
	} else {
		if w.AccountIdPacked != nil {
			panic("account id packed should be given")
		}

		if w.AccountMerkleProofs == nil {
			panic("account merkle proofs should be given")
		}

		// Account inclusion verification
		accountInclusionValue := NewAccountInclusionValue(
			accountTreeRoot,
			*w.AccountIdPacked,
			*w.AccountMerkleProofs,
			publicKeys,
		)
		result = result && accountInclusionValue.IsValid
	}

	// Format validation
	formatValidationValue :=
		NewFormatValidationValue(publicKeys, signature)
	result = result && formatValidationValue.IsValid

	if result {
		aggregationValue := NewAggregationValue(publicKeys, signature)
		result = result && aggregationValue.IsValid
	}

	prev_block_hash := block.PrevBlockHash
	blockHash := block.Hash()
	senderTreeRoot := GetSenderTreeRoot(&publicKeys, signature.SenderFlag)

	txTreeRoot := signature.TxTreeRoot

	return &MainValidationPublicInputs{
		PrevBlockHash:       prev_block_hash,
		BlockHash:           blockHash,
		DepositTreeRoot:     block.DepositRoot,
		AccountTreeRoot:     accountTreeRoot,
		TxTreeRoot:          txTreeRoot,
		SenderTreeRoot:      senderTreeRoot,
		BlockNumber:         block.BlockNumber,
		IsRegistrationBlock: isRegistrationBlock,
		IsValid:             result,
	}
}

type ValidityWitness struct {
	BlockWitness              *BlockWitness
	ValidityTransitionWitness *ValidityTransitionWitness
}

func (vw *ValidityWitness) Genesis() *ValidityWitness {
	return &ValidityWitness{
		BlockWitness:              new(BlockWitness).Genesis(),
		ValidityTransitionWitness: new(ValidityTransitionWitness).Genesis(),
	}
}

func (vw *ValidityWitness) ToValidityPublicInputs() *ValidityPublicInputs {
	prevBlockTreeRoot := vw.BlockWitness.PrevBlockTreeRoot

	// Check transition block tree root
	block := vw.BlockWitness.Block
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	err := vw.ValidityTransitionWitness.BlockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(block.BlockNumber),
		&prevBlockTreeRoot,
	)

	if err != nil {
		panic("Block merkle proof is invalid")
	}
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	blockTreeRoot := vw.ValidityTransitionWitness.BlockMerkleProof.GetRoot(blockHashLeaf.Hash(), int(block.BlockNumber))

	mainValidationPis := vw.BlockWitness.ToMainValidationPublicInputs()

	// transition account tree root
	prevAccountTreeRoot := vw.BlockWitness.PrevAccountTreeRoot
	accountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(&prevAccountTreeRoot)
	if mainValidationPis.IsValid && mainValidationPis.IsRegistrationBlock {
		accountRegistrationProofs := vw.ValidityTransitionWitness.AccountRegistrationProofs
		if !accountRegistrationProofs.IsValid {
			panic("account_registration_proofs should be given")
		}
		for i, senderLeaf := range vw.ValidityTransitionWitness.SenderLeaves {
			accountRegistrationProof := accountRegistrationProofs.Proofs[i]
			var lastBlockNumber uint32 = 0
			if senderLeaf.IsValid {
				lastBlockNumber = block.BlockNumber
			}

			dummyPublicKey := intMaxAcc.NewDummyPublicKey()
			isDummy := senderLeaf.Sender.Cmp(dummyPublicKey.BigInt()) == 0
			accountTreeRoot, err = accountRegistrationProof.ConditionalGetNewRoot(
				!isDummy,
				senderLeaf.Sender,
				uint64(lastBlockNumber),
				accountTreeRoot,
			)
			if err != nil {
				panic("Invalid account registoration proof")
			}
		}
	}
	if mainValidationPis.IsValid && !mainValidationPis.IsRegistrationBlock {
		accountUpdateProofs := vw.ValidityTransitionWitness.AccountUpdateProofs
		if !accountUpdateProofs.IsValid {
			panic("account_update_proofs should be given")
		}
		for i, senderLeaf := range vw.ValidityTransitionWitness.SenderLeaves {
			accountUpdateProof := accountUpdateProofs.Proofs[i]
			prevLastBlockNumber := uint32(accountUpdateProof.PrevLeaf.Value)
			lastBlockNumber := prevLastBlockNumber
			if senderLeaf.IsValid {
				lastBlockNumber = block.BlockNumber
			}
			accountTreeRoot, err = accountUpdateProof.GetNewRoot(
				senderLeaf.Sender,
				uint64(prevLastBlockNumber),
				uint64(lastBlockNumber),
				accountTreeRoot,
			)

			if err != nil {
				panic("Invalid account update proof")
			}
		}
	}

	return &ValidityPublicInputs{
		PublicState: &PublicState{
			BlockTreeRoot:       *blockTreeRoot,
			PrevAccountTreeRoot: prevAccountTreeRoot,
			AccountTreeRoot:     *accountTreeRoot,
			DepositTreeRoot:     block.DepositRoot,
			BlockHash:           mainValidationPis.BlockHash,
			BlockNumber:         block.BlockNumber,
		},
		TxTreeRoot:     mainValidationPis.TxTreeRoot,
		SenderTreeRoot: mainValidationPis.SenderTreeRoot,
		IsValidBlock:   mainValidationPis.IsValid,
	}
}

type AuxInfo struct {
	TxTree          *intMaxTree.TxTree
	ValidityWitness *ValidityWitness
	AccountTree     *intMaxTree.AccountTree
	BlockTree       *intMaxTree.BlockHashTree
}

type MockBlockBuilder struct {
	LastBlockNumber     uint32
	AccountTree         *intMaxTree.AccountTree   // current account tree
	BlockTree           *intMaxTree.BlockHashTree // current block hash tree
	DepositTree         *intMaxTree.DepositTree   // current deposit tree
	LastValidityWitness *ValidityWitness
	AuxInfo             map[uint32]AuxInfo
}

type MockTxRequest struct {
	Tx                  intMaxTypes.Tx
	Sender              intMaxAcc.PrivateKey
	WillReturnSignature bool
}

func (b *MockBlockBuilder) generateBlock(
	isRegistrationBlock bool,
	txs []MockTxRequest,
) (*BlockWitness, *intMaxTree.TxTree) {
	return nil, nil
}

func NewMockBlockBuilder() *MockBlockBuilder {
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	blockHashes := make([][32]byte, 1)
	blockHashes[0] = new(PostedBlock).Genesis().Hash()
	blockTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, blockHashes)
	if err != nil {
		panic(err)
	}

	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	validityWitness := new(ValidityWitness).Genesis()
	zeroHash := new(intMaxGP.PoseidonHashOut).SetZero()
	txTree, err := intMaxTree.NewTxTree(intMaxTree.TX_TREE_HEIGHT, nil, zeroHash)
	if err != nil {
		panic(err)
	}

	auxInfo := make(map[uint32]AuxInfo)
	auxInfo[0] =
		AuxInfo{
			TxTree:          txTree,
			ValidityWitness: validityWitness, // clone()
			AccountTree:     accountTree,     // clone()
			BlockTree:       blockTree,       // clone()
		}
	return &MockBlockBuilder{
		LastBlockNumber:     0,
		LastValidityWitness: validityWitness,
		AccountTree:         accountTree,
		BlockTree:           blockTree,
		DepositTree:         depositTree,
		AuxInfo:             auxInfo,
	}
}

func (b *MockBlockBuilder) generateValidityWitness(blockWitness *BlockWitness) *ValidityWitness {
	if blockWitness.Block.BlockNumber != b.LastBlockNumber+1 {
		panic("block number is not equal to the last block number + 1")
	}
	prevPis := b.LastValidityWitness.ToValidityPublicInputs()
	if prevPis.PublicState.AccountTreeRoot != b.AccountTree.GetRoot() {
		panic("account tree root is not equal to the last account tree root")
	}
	if prevPis.PublicState.BlockTreeRoot != b.BlockTree.GetRoot() {
		panic("block tree root is not equal to the last block tree root")
	}

	// TODO

	// let block_merkle_proof = self
	// 	.block_tree
	// 	.prove(block_witness.block.block_number as usize);
	// self.block_tree.push(block_witness.block.hash());

	// let sender_leaves =
	// 	get_sender_leaves(&block_witness.pubkeys, block_witness.signature.sender_flag);
	// let block_pis = block_witness.to_main_validation_pis();

	// let account_registoration_proofs = {
	// 	if block_pis.is_valid && block_pis.is_registoration_block {
	// 		let mut account_registoration_proofs = Vec::new();
	// 		for sender_leaf in &sender_leaves {
	// 			let last_block_number = if sender_leaf.is_valid {
	// 				block_pis.block_number
	// 			} else {
	// 				0
	// 			};
	// 			let is_dummy_pubkey = sender_leaf.sender.is_dummy_pubkey();
	// 			let proof = if is_dummy_pubkey {
	// 				AccountRegistorationProof::dummy(ACCOUNT_TREE_HEIGHT)
	// 			} else {
	// 				self.account_tree
	// 					.prove_and_insert(sender_leaf.sender, last_block_number as u64)
	// 					.unwrap()
	// 			};
	// 			account_registoration_proofs.push(proof);
	// 		}
	// 		Some(account_registoration_proofs)
	// 	} else {
	// 		None
	// 	}
	// };

	// let account_update_proofs = {
	// 	if block_pis.is_valid && (!block_pis.is_registoration_block) {
	// 		let mut account_update_proofs = Vec::new();
	// 		let block_number = block_pis.block_number;
	// 		for sender_leaf in sender_leaves.iter() {
	// 			let account_id = self.account_tree.index(sender_leaf.sender).unwrap();
	// 			let prev_leaf = self.account_tree.get_leaf(account_id);
	// 			let prev_last_block_number = prev_leaf.value as u32;
	// 			let last_block_number = if sender_leaf.is_valid {
	// 				block_number
	// 			} else {
	// 				prev_last_block_number
	// 			};
	// 			let proof = self
	// 				.account_tree
	// 				.prove_and_update(sender_leaf.sender, last_block_number as u64);
	// 			account_update_proofs.push(proof);
	// 		}
	// 		Some(account_update_proofs)
	// 	} else {
	// 		None
	// 	}
	// };
	// let validity_transition_witness = ValidityTransitionWitness {
	// 	sender_leaves,
	// 	block_merkle_proof,
	// 	account_registoration_proofs,
	// 	account_update_proofs,
	// };
	// ValidityWitness {
	// 	validity_transition_witness,
	// 	block_witness: block_witness.clone(),
	// }

	return nil
}

func (b *MockBlockBuilder) postBlock(
	isRegistrationBlock bool,
	txs []MockTxRequest,
) *ValidityWitness {
	blockWitness, txTree := b.generateBlock(isRegistrationBlock, txs)
	validityWitness := b.generateValidityWitness(blockWitness)
	b.AuxInfo[blockWitness.Block.BlockNumber] =
		AuxInfo{
			TxTree:          txTree,
			ValidityWitness: validityWitness, // clone
			AccountTree:     b.AccountTree,   // clone
			BlockTree:       b.BlockTree,     // clone
		}
	b.LastBlockNumber = blockWitness.Block.BlockNumber
	b.LastValidityWitness = validityWitness // clone

	return validityWitness
}

type ValidityProcessor interface {
	Prove(prevValidityProof *intMaxTypes.Plonky2Proof, validityWitness *ValidityWitness) (*intMaxTypes.Plonky2Proof, error)
}

type ExternalValidityProcessor struct {
}

func NewExternalValidityProcessor() *ExternalValidityProcessor {
	return nil
}

func (p *ExternalValidityProcessor) Prove(prevValidityProof *intMaxTypes.Plonky2Proof, validityWitness *ValidityWitness) (*intMaxTypes.Plonky2Proof, error) {

	return nil, nil
}

type SyncValidityProver struct {
	ValidityProcessor ValidityProcessor
	LastBlockNumber   uint32
	ValidityProofs    map[uint32]*intMaxTypes.Plonky2Proof
}

func NewSyncValidityProver() *SyncValidityProver {
	return &SyncValidityProver{
		ValidityProcessor: NewExternalValidityProcessor(),
		LastBlockNumber:   0,
		ValidityProofs:    make(map[uint32]*intMaxTypes.Plonky2Proof),
	}
}

func (b *SyncValidityProver) Sync(blockBuilder *MockBlockBuilder) {
	currentBlockNumber := blockBuilder.LastBlockNumber
	for blockNumber := b.LastBlockNumber + 1; blockNumber <= currentBlockNumber; blockNumber++ {
		prevValidityProof, ok := b.ValidityProofs[blockNumber-1]
		if !ok && blockNumber != 1 {
			panic("prev validity proof not found")
		}
		auxInfo, ok := blockBuilder.AuxInfo[blockNumber]
		if !ok {
			panic("aux info not found")
		}

		validityProof, err := b.ValidityProcessor.Prove(prevValidityProof, auxInfo.ValidityWitness)

		if err != nil {
			panic(err)
		}
		b.ValidityProofs[blockNumber] = validityProof
	}

	b.LastBlockNumber = currentBlockNumber
}

type CircuitData interface{}

// / A dummy implementation of the transition wrapper circuit used for testing balance proof.
type TransitionWrapperCircuit interface {
	Prove(
		prevPis *ValidityPublicInputs,
		validity_witness *ValidityWitness,
	) (*intMaxTypes.Plonky2Proof, error)
	CircuitData() *CircuitData
}
