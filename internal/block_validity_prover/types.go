package block_validity_prover

import (
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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/ffg"
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

func (proof *AccountMerkleProof) Verify(accountTreeRoot intMaxGP.PoseidonHashOut, accountID uint64, publicKey intMaxTypes.Uint256) error {
	if publicKey.IsDummyPublicKey() {
		return errors.New("public key is zero")
	}

	err := proof.MerkleProof.Verify(&proof.Leaf, int(accountID), &accountTreeRoot)
	if err != nil {
		var ErrMerkleProofInvalid = errors.New("given Merkle proof is invalid")
		return errors.Join(ErrMerkleProofInvalid, err)
	}

	if publicKey.BigInt().Cmp(proof.Leaf.Key) != 0 {
		return errors.New("public key does not match leaf key")
	}

	return nil
}

const (
	numAccountIDBytes       = 5
	numUint32Bytes          = 4
	numAccountIDPackedBytes = numOfSenders * numAccountIDBytes / numUint32Bytes
)

type AccountIdPacked [numAccountIDPackedBytes]uint32

func (b *AccountIdPacked) FromBytes(bytes []byte) {
	if len(bytes) > numAccountIDPackedBytes*numUint32Bytes {
		panic("invalid bytes length")
	}

	if len(bytes) < numAccountIDPackedBytes*numUint32Bytes {
		panic("invalid bytes length")
	}

	for i := 0; i < numAccountIDPackedBytes; i++ {
		b[i] = binary.BigEndian.Uint32(bytes[i*numUint32Bytes : (i+1)*numUint32Bytes])
	}
}

func (b *AccountIdPacked) Bytes() []byte {
	bytes := make([]byte, numAccountIDPackedBytes*numUint32Bytes)
	for i := 0; i < numOfSenders; i++ {
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

func (accountIDsPacked *AccountIdPacked) Pack(accountIDs []uint64) *AccountIdPacked {
	accountIDsBytes := make([]byte, numAccountIDBytes*len(accountIDs))
	for i, accountID := range accountIDs {
		chunkBytes := make([]byte, int8Key)
		binary.BigEndian.PutUint64(chunkBytes, accountID)
		copy(accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes], chunkBytes[int8Key-numAccountIDBytes:])
	}

	accountIDsPacked.FromBytes(accountIDsBytes)

	return accountIDsPacked
}

func (accountIDsPacked *AccountIdPacked) Unpack() []uint64 {
	accountIDsBytes := accountIDsPacked.Bytes()
	accountIDs := make([]uint64, 0)
	for i := 0; i < numOfSenders; i++ {
		chunkBytes := make([]byte, int8Key)
		copy(chunkBytes[int8Key-numAccountIDBytes:], accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes])

		accountID := binary.BigEndian.Uint64(chunkBytes)
		accountIDs = append(accountIDs, accountID)
	}

	return accountIDs
}

func (accountIDPacked *AccountIdPacked) Hash() intMaxTypes.Bytes32 {
	h := crypto.Keccak256(accountIDPacked.Bytes()) // TODO: Is this correct hash?
	var b intMaxTypes.Bytes32
	b.FromBytes(h)

	return b
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

func (s *SignatureContent) ToFieldElementSlice() []ffg.Element {
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))
	var isRegistrationBlock uint32 = 0
	if s.IsRegistrationBlock {
		isRegistrationBlock = 1
	}
	finite_field.WriteUint32(buf, isRegistrationBlock)
	for _, d := range s.TxTreeRoot {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range s.SenderFlag {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range s.PublicKeyHash {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range s.AccountIDHash {
		finite_field.WriteUint32(buf, d)
	}
	for i := range s.AggPublicKey {
		coord := s.AggPublicKey[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}
	for i := range s.AggSignature {
		coord := s.AggSignature[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}
	for i := range s.MessagePoint {
		coord := s.MessagePoint[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}

	return buf.Inner()
}

func (s *SignatureContent) Commitment() *intMaxGP.PoseidonHashOut {
	flattenSignatureContent := s.ToFieldElementSlice()
	fmt.Printf("len(flattenSignatureContent): %d\n", len(flattenSignatureContent))
	commitment := intMaxGP.HashNoPad(flattenSignatureContent)

	return commitment
}

func (s *SignatureContent) IsValidFormat(publicKeys []intMaxTypes.Uint256) error {
	if len(publicKeys) != numOfSenders {
		return errors.New("public keys length is invalid")
	}

	// sender flag check
	zeroSenderFlag := intMaxTypes.Bytes16{}
	if s.SenderFlag == zeroSenderFlag {
		return errors.New("sender flag is zero")
	}

	// public key order check
	curPublicKey := publicKeys[0]
	for i := 1; i < len(publicKeys); i++ {
		publicKey := publicKeys[i]
		if curPublicKey.BigInt().Cmp(publicKey.BigInt()) != 1 && !publicKey.IsDummyPublicKey() {
			return errors.New("public key order is invalid")
		}

		curPublicKey = publicKey
	}

	// public keys order and recovery check
	for _, publicKey := range publicKeys {
		_, err := intMaxAcc.NewPublicKeyFromAddressInt(publicKey.BigInt())
		if err != nil {
			return errors.New("public key recovery check failed")
		}
	}

	// message point check
	txTreeRoot := s.TxTreeRoot.ToFieldElementSlice()
	messagePointExpected := intMaxGP.HashToG2(txTreeRoot)
	messagePoint := intMaxTypes.NewG2AffineFromFlatG2(&s.MessagePoint)
	if messagePointExpected.Equal(messagePoint) {
		return errors.New("message point check failed")
	}

	return nil
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

func GetPublicKeysHash(publicKeys []intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	publicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range publicKeys {
		publicKeyBytes := sender.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes)
	}
	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < intMaxTypes.NumOfSenders; i++ {
		publicKeyBytes := dummyPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes[:])
	}

	publicKeysHash := crypto.Keccak256(publicKeysBytes) // TODO: Is this correct hash?

	var result intMaxTypes.Bytes32
	result.FromBytes(publicKeysHash)

	return result
}

func GetAccountIDsHash(accountIDs []uint64) intMaxTypes.Bytes32 {
	accountIDsPacked := new(AccountIdPacked).Pack(accountIDs)

	return accountIDsPacked.Hash()
}

type AccountExclusionValue struct {
	AccountTreeRoot         *intMaxGP.PoseidonHashOut
	AccountMembershipProofs []intMaxTree.IndexedMembershipProof
	PublicKeys              []intMaxTypes.Uint256
	PublicKeyCommitment     *intMaxGP.PoseidonHashOut
	IsValid                 bool
}

func getPublicKeyCommitment(publicKeys []intMaxTypes.Uint256) *intMaxGP.PoseidonHashOut {
	publicKeyFlattened := make([]ffg.Element, 0)
	for _, publicKey := range publicKeys {
		publicKeyFlattened = append(publicKeyFlattened, publicKey.ToFieldElementSlice()...)
	}

	return intMaxGP.HashNoPad(publicKeyFlattened)
}

func NewAccountExclusionValue(
	accountTreeRoot *intMaxGP.PoseidonHashOut,
	accountMembershipProofs []intMaxTree.IndexedMembershipProof,
	publicKeys []intMaxTypes.Uint256,
) (*AccountExclusionValue, error) {
	result := true
	for i, proof := range accountMembershipProofs {
		err := proof.Verify(publicKeys[i].BigInt(), accountTreeRoot)
		if err != nil {
			var ErrAccountMembershipProofInvalid = errors.New("account membership proof is invalid")
			return nil, errors.Join(ErrAccountMembershipProofInvalid, err)
		}

		isDummy := publicKeys[i].IsDummyPublicKey()
		isExcluded := !proof.IsIncluded || isDummy
		result = result && isExcluded
	}

	publicKeyCommitment := getPublicKeyCommitment(publicKeys)

	return &AccountExclusionValue{
		AccountTreeRoot:         accountTreeRoot,
		AccountMembershipProofs: accountMembershipProofs,
		PublicKeys:              publicKeys,
		PublicKeyCommitment:     publicKeyCommitment,
		IsValid:                 result,
	}, nil
}

type AccountInclusionValue struct {
	AccountIDPacked     AccountIdPacked
	AccountIDHash       intMaxTypes.Bytes32
	AccountTreeRoot     *intMaxGP.PoseidonHashOut
	AccountMerkleProofs []AccountMerkleProof
	PublicKeys          []intMaxTypes.Uint256
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewAccountInclusionValue(
	accountTreeRoot intMaxTree.PoseidonHashOut,
	accountIDPacked *AccountIdPacked,
	accountMerkleProofs []AccountMerkleProof,
	publicKeys []intMaxTypes.Uint256,
) (*AccountInclusionValue, error) {
	if len(accountMerkleProofs) != numOfSenders {
		return nil, errors.New("account merkle proofs length should be equal to number of senders")
	}

	if len(publicKeys) != numOfSenders {
		return nil, errors.New("public keys length should be equal to number of senders")
	}

	result := true
	accountIDHash := accountIDPacked.Hash()
	accountIDs := accountIDPacked.Unpack()
	for i := range accountIDs {
		accountID := accountIDs[i]
		proof := accountMerkleProofs[i]
		publicKey := publicKeys[i]
		err := proof.Verify(accountTreeRoot, accountID, publicKey)
		result = result && err == nil
	}

	publicKeyCommitment := getPublicKeyCommitment(publicKeys)

	return &AccountInclusionValue{
		AccountIDPacked:     *accountIDPacked,
		AccountIDHash:       accountIDHash,
		AccountTreeRoot:     &accountTreeRoot,
		AccountMerkleProofs: accountMerkleProofs,
		PublicKeys:          publicKeys,
		PublicKeyCommitment: publicKeyCommitment,
		IsValid:             true,
	}, nil
}

type FormatValidationValue struct {
	PublicKeys          []intMaxTypes.Uint256
	Signature           *SignatureContent
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	SignatureCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewFormatValidationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *FormatValidationValue {
	pubkeyCommitment := getPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.IsValidFormat(publicKeys)

	return &FormatValidationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: pubkeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
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

func (w *BlockWitness) MainValidationPublicInputs() *MainValidationPublicInputs {
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
		accountExclusionValue, err := NewAccountExclusionValue(
			&accountTreeRoot,
			*w.AccountMembershipProofs,
			publicKeys,
		)
		if err != nil {
			panic("account exclusion value is invalid: " + err.Error())
		}

		result = result && accountExclusionValue.IsValid
	} else {
		if w.AccountIdPacked != nil {
			panic("account id packed should be given")
		}

		if w.AccountMerkleProofs == nil {
			panic("account merkle proofs should be given")
		}

		// Account inclusion verification
		accountInclusionValue, err := NewAccountInclusionValue(
			accountTreeRoot,
			w.AccountIdPacked,
			*w.AccountMerkleProofs,
			publicKeys,
		)
		if err != nil {
			panic("account inclusion value is invalid: " + err.Error())
		}

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

func (vw *ValidityWitness) ValidityPublicInputs() *ValidityPublicInputs {
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

	mainValidationPis := vw.BlockWitness.MainValidationPublicInputs()

	// transition account tree root
	prevAccountTreeRoot := vw.BlockWitness.PrevAccountTreeRoot
	accountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(&prevAccountTreeRoot)
	if mainValidationPis.IsValid && mainValidationPis.IsRegistrationBlock {
		accountRegistrationProofs := vw.ValidityTransitionWitness.AccountRegistrationProofs
		if !accountRegistrationProofs.IsValid {
			panic("account registration proofs should be given")
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
				fmt.Printf("senderLeaf.Sender: %s\n", senderLeaf.Sender.String())
				panic("Invalid account registration proof: " + err.Error())
			}
		}
	}
	if mainValidationPis.IsValid && !mainValidationPis.IsRegistrationBlock {
		accountUpdateProofs := vw.ValidityTransitionWitness.AccountUpdateProofs
		if !accountUpdateProofs.IsValid {
			panic("account update proofs should be given")
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
	LastBlockNumber                         uint32
	AccountTree                             *intMaxTree.AccountTree      // current account tree
	BlockTree                               *intMaxTree.BlockHashTree    // current block hash tree
	DepositTree                             *intMaxTree.KeccakMerkleTree // current deposit tree
	DepositLeaves                           map[common.Hash]*DepositLeafWithId
	DepositTreeRoots                        []common.Hash
	LastSeenProcessDepositsEventBlockNumber uint64
	LastSeenBlockPostedEventBlockNumber     uint64
	LastSeenProcessedDepositId              uint64
	LastValidityWitness                     *ValidityWitness
	AuxInfo                                 map[uint32]AuxInfo
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
		LastBlockNumber:                         0,
		LastValidityWitness:                     validityWitness,
		AccountTree:                             accountTree,
		BlockTree:                               blockTree,
		DepositTree:                             depositTree,
		DepositLeaves:                           make(map[common.Hash]*DepositLeafWithId),
		DepositTreeRoots:                        []common.Hash{depositTreeRoot},
		LastSeenProcessDepositsEventBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		LastSeenBlockPostedEventBlockNumber:     cfg.Blockchain.RollupContractDeployedBlockNumber,
		AuxInfo:                                 auxInfo,
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
		publicKeys[i] = *publicKey
		accountIDs[i] = sender.AccountID
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
		AggPublicKey:        intMaxTypes.FlattenG1Affine(blockContent.AggregatedPublicKey.Pk),
		AggSignature:        intMaxTypes.FlattenG2Affine(blockContent.AggregatedSignature),
		MessagePoint:        intMaxTypes.FlattenG2Affine(blockContent.MessagePoint),
	}
	copy(signature.TxTreeRoot[:], intMaxTypes.CommonHashToUint32Slice(blockContent.TxTreeRoot))
	signature.SenderFlag.FromBytes(senderFlagBytes[:])

	prevAccountTreeRoot := b.AccountTree.GetRoot()
	prevBlockTreeRoot := b.BlockTree.GetRoot()

	if isRegistrationBlock {
		accountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, len(blockContent.Senders))
		for i, sender := range blockContent.Senders {
			accountMembershipProof, _, err := b.AccountTree.ProveMembership(sender.PublicKey.BigInt())
			if err != nil {
				panic(err)
			}

			accountMembershipProofs[i] = *accountMembershipProof
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

	accountMerkleProofs := make([]AccountMerkleProof, len(blockContent.Senders))
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

		accountMerkleProofs[i] = AccountMerkleProof{
			MerkleProof: accountMerkleProof.LeafProof,
			Leaf:        accountMerkleProof.Leaf,
		}
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

func (b *MockBlockBuilder) generateValidityWitness(blockWitness *BlockWitness) (*ValidityWitness, error) {
	if blockWitness.Block.BlockNumber != b.LastBlockNumber+1 {
		return nil, errors.New("block number is not equal to the last block number + 1")
	}
	prevPis := b.LastValidityWitness.ValidityPublicInputs()
	if prevPis.PublicState.AccountTreeRoot != b.AccountTree.GetRoot() {
		return nil, errors.New("account tree root is not equal to the last account tree root")
	}
	if prevPis.PublicState.BlockTreeRoot != b.BlockTree.GetRoot() {
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}

	blockMerkleProof, _, err := b.BlockTree.Prove(blockWitness.Block.BlockNumber)
	if err != nil {
		var ErrBlockTreeProve = errors.New("block tree prove error")
		return nil, errors.Join(ErrBlockTreeProve, err)
	}

	blockHashLeaf := intMaxTree.NewBlockHashLeaf(blockWitness.Block.Hash())
	_, count, _ := b.BlockTree.GetCurrentRootCountAndSiblings()
	if count != blockWitness.Block.BlockNumber {
		return nil, errors.New("block number is not equal to the current block number")
	}

	_, err = b.BlockTree.AddLeaf(count, blockHashLeaf)
	if err != nil {
		var ErrBlockTreeAddLeaf = errors.New("block tree add leaf error")
		return nil, errors.Join(ErrBlockTreeAddLeaf, err)
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
				proof, err = b.AccountTree.Insert(senderLeaf.Sender, uint64(lastBlockNumber))
				if err != nil {
					// invalid block
					var ErrAccountTreeInsert = errors.New("account tree insert error")
					return nil, errors.Join(ErrAccountTreeInsert, err)
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
			accountID, ok := b.AccountTree.GetAccountID(senderLeaf.Sender)
			if !ok {
				return nil, errors.New("account id not found")
			}
			prevLeaf := b.AccountTree.GetLeaf(accountID)
			prevLastBlockNumber := uint32(prevLeaf.Value)
			lastBlockNumber := prevLastBlockNumber
			if senderLeaf.IsValid {
				lastBlockNumber = blockPis.BlockNumber
			}
			var proof *intMaxTree.IndexedUpdateProof
			proof, err = b.AccountTree.Update(senderLeaf.Sender, uint64(lastBlockNumber))
			if err != nil {
				var ErrAccountTreeUpdate = errors.New("account tree update error")
				return nil, errors.Join(ErrAccountTreeUpdate, err)
			}
			accountUpdateProofs = append(accountUpdateProofs, *proof)
		}

		accountUpdateProofsWitness = AccountUpdateProofs{
			IsValid: true,
			Proofs:  accountUpdateProofs,
		}
	}

	return &ValidityWitness{
		BlockWitness: blockWitness,
		ValidityTransitionWitness: &ValidityTransitionWitness{
			SenderLeaves:              senderLeaves,
			BlockMerkleProof:          blockMerkleProof,
			AccountRegistrationProofs: accountRegistrationProofsWitness,
			AccountUpdateProofs:       accountUpdateProofsWitness,
		},
	}, nil
}

func (b *MockBlockBuilder) postBlock(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) *ValidityWitness {
	blockWitness, txTree := b.generateBlock(blockContent, postedBlock)
	validityWitness, err := b.generateValidityWitness(blockWitness)
	if err != nil {
		panic(err)
	}

	b.AuxInfo[blockWitness.Block.BlockNumber] =
		AuxInfo{
			TxTree:          txTree,
			ValidityWitness: validityWitness,
			AccountTree:     b.AccountTree,
			BlockTree:       b.BlockTree,
		}
	b.LastBlockNumber = blockWitness.Block.BlockNumber
	b.LastValidityWitness = validityWitness

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
