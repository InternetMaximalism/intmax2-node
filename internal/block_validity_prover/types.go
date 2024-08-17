package block_validity_prover

import (
	"encoding/binary"
	"encoding/json"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type PublicState struct {
	BlockTreeRoot       intMaxGP.PoseidonHashOut `json:"blockTreeRoot"`
	PrevAccountTreeRoot intMaxGP.PoseidonHashOut `json:"prevAccountTreeRoot"`
	AccountTreeRoot     intMaxGP.PoseidonHashOut `json:"accountTreeRoot"`
	DepositTreeRoot     common.Hash              `json:"depositTreeRoot"`
	BlockHash           common.Hash              `json:"blockHash"`
	BlockNumber         uint32                   `json:"blockNumber"`
}

func (ps *PublicState) Genesis() *PublicState {
	blockTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	genesisBlockHash := new(block_post_service.PostedBlock).Genesis().Hash()
	blockTreeRoot, err := blockTree.AddLeaf(0, intMaxTree.NewBlockHashLeaf(genesisBlockHash))
	if err != nil {
		panic(err)
	}

	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT)
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

func (pis *PublicState) Equal(other *PublicState) bool {
	if !pis.BlockTreeRoot.Equal(&other.BlockTreeRoot) {
		return false
	}
	if !pis.PrevAccountTreeRoot.Equal(&other.PrevAccountTreeRoot) {
		return false
	}
	if !pis.AccountTreeRoot.Equal(&other.AccountTreeRoot) {
		return false
	}
	if pis.DepositTreeRoot != other.DepositTreeRoot {
		return false
	}
	if pis.BlockHash != other.BlockHash {
		return false
	}
	if pis.BlockNumber != other.BlockNumber {
		return false
	}
	return true
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
	Sender  *big.Int `json:"sender"`
	IsValid bool     `json:"isValid"`
}

type AccountRegistrationProofs struct {
	Proofs  []intMaxTree.IndexedInsertionProof `json:"proofs"`
	IsValid bool                               `json:"isValid"`
}

type AccountUpdateProofs struct {
	Proofs  []intMaxTree.IndexedUpdateProof `json:"proofs"`
	IsValid bool                            `json:"isValid"`
}

type ValidityTransitionWitness struct {
	SenderLeaves              []SenderLeaf              `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.MerkleProof    `json:"blockMerkleProof"`
	AccountRegistrationProofs AccountRegistrationProofs `json:"accountRegistrationProofs"`
	AccountUpdateProofs       AccountUpdateProofs       `json:"accountUpdateProofs"`
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
	MerkleProof intMaxTree.IndexedMerkleProof `json:"merkleProof"`
	Leaf        intMaxTree.IndexedMerkleLeaf  `json:"leaf"`
}

const (
	numAccountIDBytes       = 5
	numUint32Bytes          = 4
	numAccountIDPackedBytes = numOfSenders * numAccountIDBytes / numUint32Bytes
)

type AccountIdPacked [numAccountIDPackedBytes]uint32

func (b *AccountIdPacked) FromBytes(bytes []byte) {
	for i := 0; i < numAccountIDPackedBytes/numUint32Bytes; i++ {
		b[i] = binary.BigEndian.Uint32(bytes[i*numUint32Bytes : (i+1)*numUint32Bytes])
	}
}

func (b *AccountIdPacked) Bytes() []byte {
	bytes := make([]byte, int16Key)
	for i := 0; i < numAccountIDPackedBytes/numUint32Bytes; i++ {
		binary.BigEndian.PutUint32(bytes[i*numUint32Bytes:(i+1)*numUint32Bytes], b[i])
	}

	return bytes
}

func (b *AccountIdPacked) Hex() string {
	return hexutil.Encode(b.Bytes())
}

func (b *AccountIdPacked) FromHex(s string) error {
	bytes, err := hexutil.Decode(s)
	if err != nil {
		return err
	}

	b.FromBytes(bytes)
	return nil
}

func (b *AccountIdPacked) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Hex())
}

func (b *AccountIdPacked) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	return b.FromHex(s)
}

type SignatureContent struct {
	IsRegistrationBlock bool                `json:"isRegistrationBlock"`
	TxTreeRoot          intMaxTypes.Bytes32 `json:"txTreeRoot"`
	SenderFlag          intMaxTypes.Bytes16 `json:"senderFlag"`
	PublicKeyHash       intMaxTypes.Bytes32 `json:"pubkeyHash"`
	AccountIDHash       intMaxTypes.Bytes32 `json:"accountIdHash"`
	AggPublicKey        intMaxTypes.FlatG1  `json:"aggPubkey"`
	AggSignature        intMaxTypes.FlatG2  `json:"aggSignature"`
	MessagePoint        intMaxTypes.FlatG2  `json:"messagePoint"`
}

func (sc *SignatureContent) Set(other *SignatureContent) *SignatureContent {
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
	Block                   *block_post_service.PostedBlock      `json:"block"`
	Signature               SignatureContent                     `json:"signature"`
	PublicKeys              []intMaxTypes.Uint256                `json:"pubkeys"`
	PrevAccountTreeRoot     intMaxTree.PoseidonHashOut           `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot       intMaxTree.PoseidonHashOut           `json:"prevBlockTreeRoot"`
	AccountIdPacked         *AccountIdPacked                     `json:"accountIdPacked"`         // in account id case
	AccountMerkleProofs     *[]AccountMerkleProof                `json:"accountMerkleProofs"`     // in account id case
	AccountMembershipProofs *[]intMaxTree.IndexedMembershipProof `json:"accountMembershipProofs"` // in pubkey case
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
		Block:                   new(block_post_service.PostedBlock).Genesis(),
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

// TODO: Implement this
func GetPublicKeysHash(publicKeys []intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	return intMaxTypes.Bytes32{}
}

func GetAccountIDsHash(accountIDs []uint64) intMaxTypes.Bytes32 {
	return intMaxTypes.Bytes32{}
}

type AccountExclusionValue struct {
	IsValid bool
}

// TODO: Implement this
func NewAccountExclusionValue(
	accountTreeRoot intMaxTree.PoseidonHashOut,
	accountMembershipProofs []intMaxTree.IndexedMembershipProof,
	publicKeys []intMaxTypes.Uint256,
) *AccountExclusionValue {
	return &AccountExclusionValue{
		IsValid: true,
	}
}

type AccountInclusionValue struct {
	IsValid bool
}

// TODO: Implement this
func NewAccountInclusionValue(
	accountTreeRoot intMaxTree.PoseidonHashOut,
	accountIdPacked *AccountIdPacked,
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

// TODO: Implement this
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

// TODO: Implement this
func NewAggregationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *AggregationValue {
	return &AggregationValue{
		IsValid: true,
	}
}

// TODO: Implement this
func GetSenderTreeRoot(publicKeys *[]intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) intMaxGP.PoseidonHashOut {
	return intMaxGP.PoseidonHashOut{}
}

func (w *BlockWitness) ToMainValidationPublicInputs() *MainValidationPublicInputs {
	if new(block_post_service.PostedBlock).Genesis().Equals(w.Block) {
		validityPis := new(ValidityPublicInputs).Genesis()
		return &MainValidationPublicInputs{
			PrevBlockHash:       new(block_post_service.PostedBlock).Genesis().PrevBlockHash,
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
	block := new(block_post_service.PostedBlock).Set(w.Block)
	signature := new(SignatureContent).Set(&w.Signature)
	publicKeys := make([]intMaxTypes.Uint256, len(w.PublicKeys))
	copy(publicKeys, w.PublicKeys)

	accountTreeRoot := w.PrevAccountTreeRoot

	publicKeysHash := GetPublicKeysHash(publicKeys)
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
			w.AccountIdPacked,
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
	BlockWitness              *BlockWitness              `json:"blockWitness"`
	ValidityTransitionWitness *ValidityTransitionWitness `json:"validityTransitionWitness"`
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
	LastBlockNumber            uint32
	AccountTree                *intMaxTree.AccountTree      // current account tree
	BlockTree                  *intMaxTree.BlockHashTree    // current block hash tree
	DepositTree                *intMaxTree.KeccakMerkleTree // current deposit tree
	DepositLeaves              map[common.Hash]*DepositLeafWithId
	DepositTreeRoots           []common.Hash
	LastSeenEventBlockNumber   uint64
	LastSeenProcessedDepositId uint64
	LastValidityWitness        *ValidityWitness
	AuxInfo                    map[uint32]AuxInfo
}

func NewMockBlockBuilder(cfg *configs.Config) *MockBlockBuilder {
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	blockHashes := make([][32]byte, 1)
	blockHashes[0] = new(block_post_service.PostedBlock).Genesis().Hash()
	blockTree, err := intMaxTree.NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT, blockHashes)
	if err != nil {
		panic(err)
	}

	zeroDepositHash := new(intMaxTree.DepositLeaf).SetZero().Hash()
	depositTree, err := intMaxTree.NewKeccakMerkleTree(intMaxTree.DEPOSIT_TREE_HEIGHT, nil, zeroDepositHash)
	if err != nil {
		panic(err)
	}
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()

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
		LastBlockNumber:          0,
		LastValidityWitness:      validityWitness,
		AccountTree:              accountTree,
		BlockTree:                blockTree,
		DepositTree:              depositTree,
		DepositLeaves:            make(map[common.Hash]*DepositLeafWithId),
		DepositTreeRoots:         []common.Hash{depositTreeRoot},
		LastSeenEventBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		AuxInfo:                  auxInfo,
	}
}

type DepositLeafWithId struct {
	DepositLeaf *intMaxTree.DepositLeaf
	DepositId   uint64
}

// TODO: Implement *intMaxTree.TxTree
func (b *MockBlockBuilder) generateBlock(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*BlockWitness, *intMaxTree.TxTree) {
	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	accountIDs := make([]uint64, len(blockContent.Senders))
	senderFlagBytes := [int16Key]byte{}
	for i, sender := range blockContent.Senders {
		publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
		publicKeys = append(publicKeys, *publicKey)
		accountIDs = append(accountIDs, sender.AccountID)
		var flag uint8 = 0
		if sender.IsSigned {
			flag = 1
		}
		senderFlagBytes[i/int8Key] |= flag << (int8Key - 1 - i%int8Key)
	}

	signature := SignatureContent{
		IsRegistrationBlock: isRegistrationBlock,
		TxTreeRoot:          intMaxTypes.Bytes32{},
		SenderFlag:          intMaxTypes.Bytes16{},
		PublicKeyHash:       GetPublicKeysHash(publicKeys),
		AccountIDHash:       GetAccountIDsHash(accountIDs),
		AggPublicKey:        G1ToSolidityType(blockContent.AggregatedPublicKey.Pk),
		AggSignature:        G2ToSolidityType(blockContent.AggregatedSignature),
		MessagePoint:        G2ToSolidityType(blockContent.MessagePoint),
	}
	copy(signature.TxTreeRoot[:], intMaxTypes.CommonHashToUint32Slice(blockContent.TxTreeRoot))
	signature.SenderFlag.FromBytes(senderFlagBytes[:])

	prevAccountTreeRoot := b.AccountTree.GetRoot()
	prevBlockTreeRoot := b.BlockTree.GetRoot()

	if isRegistrationBlock {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
		for _, sender := range blockContent.Senders {
			accountMembershipProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
			if err != nil {
				panic(err)
			}

			accountMembershipProofs = append(accountMembershipProofs, *accountMembershipProof)
		}

		blockWitness := &BlockWitness{
			Block:                   postedBlock,
			Signature:               signature,
			PublicKeys:              publicKeys,
			PrevAccountTreeRoot:     prevAccountTreeRoot,
			PrevBlockTreeRoot:       prevBlockTreeRoot,
			AccountIdPacked:         nil,
			AccountMerkleProofs:     nil,
			AccountMembershipProofs: &accountMembershipProofs,
		}

		return blockWitness, nil
	}

	accountMerkleProofs := make([]AccountMerkleProof, 0)
	accountIDPackedBytes := make([]byte, numAccountIDPackedBytes)
	for i, sender := range blockContent.Senders {
		accountIDByte := make([]byte, int8Key)
		binary.BigEndian.PutUint64(accountIDByte, sender.AccountID)
		copy(accountIDPackedBytes[i/int8Key:i/int8Key+int5Key], accountIDByte[int8Key-int5Key:])
		accountMerkleProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
		if err != nil {
			panic(err)
		}
		if !accountMerkleProof.IsIncluded {
			panic("account is not included")
		}

		accountMerkleProofs = append(accountMerkleProofs, AccountMerkleProof{
			MerkleProof: accountMerkleProof.LeafProof,
			Leaf:        accountMerkleProof.Leaf,
		})
	}

	accountIDPacked := new(AccountIdPacked)
	accountIDPacked.FromBytes(accountIDPackedBytes)
	blockWitness := &BlockWitness{
		Block:                   postedBlock,
		Signature:               signature,
		PublicKeys:              publicKeys,
		PrevAccountTreeRoot:     prevAccountTreeRoot,
		PrevBlockTreeRoot:       prevBlockTreeRoot,
		AccountIdPacked:         accountIDPacked,
		AccountMerkleProofs:     &accountMerkleProofs,
		AccountMembershipProofs: nil,
	}

	return blockWitness, nil
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

	// TODO: Implement this

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
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) *ValidityWitness {
	blockWitness, txTree := b.generateBlock(blockContent, postedBlock)
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

func G1ToSolidityType(pk *bn254.G1Affine) [2]intMaxTypes.Uint256 {
	x := intMaxTypes.Uint256{}
	y := intMaxTypes.Uint256{}
	x.FromBigInt(pk.X.BigInt(new(big.Int)))
	y.FromBigInt(pk.Y.BigInt(new(big.Int)))

	return [2]intMaxTypes.Uint256{x, y}
}

func G2ToSolidityType(sig *bn254.G2Affine) [int4Key]intMaxTypes.Uint256 {
	x_a0 := intMaxTypes.Uint256{}
	x_a1 := intMaxTypes.Uint256{}
	y_a0 := intMaxTypes.Uint256{}
	y_a1 := intMaxTypes.Uint256{}
	x_a0.FromBigInt(sig.X.A0.BigInt(new(big.Int)))
	x_a1.FromBigInt(sig.X.A1.BigInt(new(big.Int)))
	y_a0.FromBigInt(sig.Y.A0.BigInt(new(big.Int)))
	y_a1.FromBigInt(sig.Y.A1.BigInt(new(big.Int)))

	return [int4Key]intMaxTypes.Uint256{x_a1, y_a1, y_a1, y_a0}
}
