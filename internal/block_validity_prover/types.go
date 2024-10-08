package block_validity_prover

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
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

const base10 = 10

type PublicState struct {
	BlockTreeRoot       *intMaxGP.PoseidonHashOut `json:"blockTreeRoot"`
	PrevAccountTreeRoot *intMaxGP.PoseidonHashOut `json:"prevAccountTreeRoot"`
	AccountTreeRoot     *intMaxGP.PoseidonHashOut `json:"accountTreeRoot"`
	DepositTreeRoot     common.Hash               `json:"depositTreeRoot"`
	BlockHash           common.Hash               `json:"blockHash"`
	BlockNumber         uint32                    `json:"blockNumber"`
}

func (ps *PublicState) Genesis() *PublicState {
	blockTree, err := NewBlockHashTree(intMaxTree.BLOCK_HASH_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	genesisBlockHash := new(block_post_service.PostedBlock).Genesis().Hash()
	blockTreeRoot := blockTree.GetRoot()

	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("genesis accountTreeRoot: %s\n", accountTree.GetRoot().String())
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	prevAccountTreeRoot := accountTree.GetRoot()
	accountTreeRoot := accountTree.GetRoot()
	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()
	return &PublicState{
		BlockTreeRoot:       blockTreeRoot,
		PrevAccountTreeRoot: prevAccountTreeRoot,
		AccountTreeRoot:     accountTreeRoot,
		DepositTreeRoot:     depositTreeRoot,
		BlockHash:           genesisBlockHash,
		BlockNumber:         1,
	}
}

func (ps *PublicState) Set(other *PublicState) *PublicState {
	if other == nil {
		ps = nil
		return nil
	}

	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.BlockTreeRoot)
	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.PrevAccountTreeRoot)
	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(other.AccountTreeRoot)
	ps.DepositTreeRoot = other.DepositTreeRoot
	ps.BlockHash = other.BlockHash
	ps.BlockNumber = other.BlockNumber

	return ps
}

func (ps *PublicState) Equal(other *PublicState) bool {
	if !ps.BlockTreeRoot.Equal(other.BlockTreeRoot) {
		return false
	}
	if !ps.PrevAccountTreeRoot.Equal(other.PrevAccountTreeRoot) {
		return false
	}
	if !ps.AccountTreeRoot.Equal(other.AccountTreeRoot) {
		return false
	}
	if ps.DepositTreeRoot != other.DepositTreeRoot {
		return false
	}
	if ps.BlockHash != other.BlockHash {
		return false
	}
	if ps.BlockNumber != other.BlockNumber {
		return false
	}

	return true
}

func FieldElementSliceToUint32Slice(value []ffg.Element) []uint32 {
	v := make([]uint32, len(value))
	for i, x := range value {
		y := x.ToUint64Regular()
		if y >= uint64(1)<<int32Key {
			panic("overflow")
		}
		v[i] = uint32(y)
	}

	return v
}

const (
	prevAccountTreeRootOffset = intMaxGP.NUM_HASH_OUT_ELTS
	accountTreeRootOffset     = prevAccountTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
	depositTreeRootOffset     = accountTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
	blockHashOffset           = depositTreeRootOffset + int8Key
	blockNumberOffset         = blockHashOffset + int8Key
	PublicStateLimbSize       = blockNumberOffset + 1
)

func (ps *PublicState) FromFieldElementSlice(value []ffg.Element) *PublicState {
	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[:intMaxGP.NUM_HASH_OUT_ELTS])
	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[prevAccountTreeRootOffset:accountTreeRootOffset])
	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(value[accountTreeRootOffset:depositTreeRootOffset])
	depositTreeRoot := intMaxTypes.Bytes32{}
	copy(depositTreeRoot[:], FieldElementSliceToUint32Slice(value[depositTreeRootOffset:blockHashOffset]))
	ps.DepositTreeRoot = common.Hash{}
	copy(ps.DepositTreeRoot[:], depositTreeRoot.Bytes())
	blockHash := intMaxTypes.Bytes32{}
	copy(blockHash[:], FieldElementSliceToUint32Slice(value[blockHashOffset:blockNumberOffset]))
	ps.BlockHash = common.Hash{}
	copy(ps.BlockHash[:], blockHash.Bytes())
	ps.BlockNumber = uint32(value[blockNumberOffset].ToUint64Regular())

	return ps
}

const NumPublicStateBytes = int32Key*int5Key + int4Key

func (ps *PublicState) Marshal() []byte {
	buf := make([]byte, NumPublicStateBytes)
	offset := 0

	copy(buf[offset:offset+int32Key], ps.BlockTreeRoot.Marshal())
	offset += int32Key

	copy(buf[offset:offset+int32Key], ps.PrevAccountTreeRoot.Marshal())
	offset += int32Key

	copy(buf[offset:offset+int32Key], ps.AccountTreeRoot.Marshal())
	offset += int32Key

	copy(buf[offset:offset+int32Key], ps.DepositTreeRoot.Bytes())
	offset += int32Key

	copy(buf[offset:offset+int32Key], ps.BlockHash.Bytes())

	binary.BigEndian.PutUint32(buf, ps.BlockNumber)

	return buf
}

func (ps *PublicState) Unmarshal(data []byte) error {
	if len(data) < NumPublicStateBytes {
		return errors.New("invalid data length")
	}

	offset := 0

	ps.BlockTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.BlockTreeRoot.Unmarshal(data[offset : offset+int32Key])
	offset += int32Key

	ps.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.PrevAccountTreeRoot.Unmarshal(data[offset : offset+int32Key])
	offset += int32Key

	ps.AccountTreeRoot = new(intMaxGP.PoseidonHashOut)
	ps.AccountTreeRoot.Unmarshal(data[offset : offset+int32Key])
	offset += int32Key

	ps.DepositTreeRoot = common.BytesToHash(data[offset : offset+int32Key])
	offset += int32Key

	ps.BlockHash = common.BytesToHash(data[offset : offset+int32Key])
	offset += int32Key

	ps.BlockNumber = binary.BigEndian.Uint32(data[offset : offset+int4Key])

	return nil
}

type ValidityPublicInputs struct {
	PublicState    *PublicState
	TxTreeRoot     intMaxTypes.Bytes32
	SenderTreeRoot *intMaxGP.PoseidonHashOut
	IsValidBlock   bool
}

func (vpi *ValidityPublicInputs) Genesis() *ValidityPublicInputs {
	txTreeRoot := intMaxTypes.Bytes32{}
	senderTreeRoot := new(intMaxGP.PoseidonHashOut).SetZero()
	isValidBlock := false

	return &ValidityPublicInputs{
		PublicState:    new(PublicState).Genesis(),
		TxTreeRoot:     txTreeRoot,
		SenderTreeRoot: senderTreeRoot,
		IsValidBlock:   isValidBlock,
	}
}

func (vpi *ValidityPublicInputs) FromPublicInputs(publicInputs []ffg.Element) *ValidityPublicInputs {
	const (
		txTreeRootOffset     = PublicStateLimbSize
		senderTreeRootOffset = txTreeRootOffset + int8Key
		isValidBlockOffset   = senderTreeRootOffset + intMaxGP.NUM_HASH_OUT_ELTS
		end                  = isValidBlockOffset + 1
	)

	vpi.PublicState = new(PublicState).FromFieldElementSlice(publicInputs[:txTreeRootOffset])
	txTreeRoot := intMaxTypes.Bytes32{}
	copy(txTreeRoot[:], FieldElementSliceToUint32Slice(publicInputs[txTreeRootOffset:senderTreeRootOffset]))
	vpi.TxTreeRoot = txTreeRoot
	vpi.SenderTreeRoot = new(intMaxGP.PoseidonHashOut).FromPartial(publicInputs[senderTreeRootOffset:isValidBlockOffset])
	vpi.IsValidBlock = publicInputs[isValidBlockOffset].ToUint64Regular() == 1

	return vpi
}

func (vpi *ValidityPublicInputs) Equal(other *ValidityPublicInputs) bool {
	if !vpi.PublicState.Equal(other.PublicState) {
		return false
	}
	if !vpi.TxTreeRoot.Equal(&other.TxTreeRoot) {
		return false
	}
	if !vpi.SenderTreeRoot.Equal(other.SenderTreeRoot) {
		return false
	}
	if vpi.IsValidBlock != other.IsValidBlock {
		return false
	}

	return true
}

type SenderLeaf struct {
	Sender  *big.Int
	IsValid bool
}

type SerializableSenderLeaf struct {
	Sender  string `json:"sender"`
	IsValid bool   `json:"isValid"`
}

func (leaf *SenderLeaf) ToFieldElementSlice() []ffg.Element {
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))
	sender := intMaxTypes.BigIntToBytes32BeArray(leaf.Sender)
	finite_field.WriteFixedSizeBytes(buf, sender[:], intMaxTypes.NumPublicKeyBytes)
	if leaf.IsValid {
		finite_field.WriteUint32(buf, 1)
	} else {
		finite_field.WriteUint32(buf, 0)
	}

	return buf.Inner()
}

func (leaf *SenderLeaf) Hash() *intMaxGP.PoseidonHashOut {
	return intMaxGP.HashNoPad(leaf.ToFieldElementSlice())
}

func (leaf *SenderLeaf) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableSenderLeaf{
		Sender:  leaf.Sender.String(),
		IsValid: leaf.IsValid,
	})
}

func (leaf *SenderLeaf) UnmarshalJSON(data []byte) error {
	var v SerializableSenderLeaf
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	sender, ok := new(big.Int).SetString(v.Sender, base10)
	if !ok {
		return errors.New("invalid sender")
	}

	leaf.Sender = sender
	leaf.IsValid = v.IsValid

	return nil
}

type AccountRegistrationProofsOption struct {
	Proofs []intMaxTree.IndexedInsertionProof
	IsSome bool
}

func (arp *AccountRegistrationProofsOption) Set(other *AccountRegistrationProofsOption) *AccountRegistrationProofsOption {
	arp.IsSome = other.IsSome
	arp.Proofs = make([]intMaxTree.IndexedInsertionProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}

type AccountUpdateProofsOption struct {
	Proofs []intMaxTree.IndexedUpdateProof
	IsSome bool
}

func (arp *AccountUpdateProofsOption) Set(other *AccountUpdateProofsOption) *AccountUpdateProofsOption {
	arp.IsSome = other.IsSome
	arp.Proofs = make([]intMaxTree.IndexedUpdateProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}

type ValidityTransitionWitness struct {
	SenderLeaves              []SenderLeaf                    `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.PoseidonMerkleProof  `json:"blockMerkleProof"`
	AccountRegistrationProofs AccountRegistrationProofsOption `json:"accountRegistrationProofs"`
	AccountUpdateProofs       AccountUpdateProofsOption       `json:"accountUpdateProofs"`
}

type ValidityTransitionWitnessFlatten struct {
	SenderLeaves              []SenderLeaf                       `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.PoseidonMerkleProof     `json:"blockMerkleProof"`
	AccountRegistrationProofs []intMaxTree.IndexedInsertionProof `json:"accountRegistrationProofs,omitempty"`
	AccountUpdateProofs       []intMaxTree.IndexedUpdateProof    `json:"accountUpdateProofs,omitempty"`
}

func (w *ValidityTransitionWitness) MarshalJSON() ([]byte, error) {
	result := ValidityTransitionWitnessFlatten{
		SenderLeaves:              w.SenderLeaves,
		BlockMerkleProof:          w.BlockMerkleProof,
		AccountRegistrationProofs: nil,
		AccountUpdateProofs:       nil,
	}

	if w.AccountRegistrationProofs.IsSome {
		result.AccountRegistrationProofs = w.AccountRegistrationProofs.Proofs
	}

	if w.AccountUpdateProofs.IsSome {
		result.AccountUpdateProofs = w.AccountUpdateProofs.Proofs
	}

	return json.Marshal(&result)
}

func (vtw *ValidityTransitionWitness) Set(other *ValidityTransitionWitness) *ValidityTransitionWitness {
	vtw.SenderLeaves = make([]SenderLeaf, len(other.SenderLeaves))
	copy(vtw.SenderLeaves, other.SenderLeaves)
	vtw.BlockMerkleProof.Set(&other.BlockMerkleProof)
	vtw.AccountRegistrationProofs.Set(&other.AccountRegistrationProofs)
	vtw.AccountUpdateProofs.Set(&other.AccountUpdateProofs)

	return vtw
}

type AccountRegistrationProofOrDummy struct {
	LowLeafProof *intMaxTree.PoseidonMerkleProof `json:"lowLeafProof,omitempty"`
	LeafProof    *intMaxTree.PoseidonMerkleProof `json:"leafProof,omitempty"`
	Index        uint64                          `json:"index"`
	LowLeafIndex uint64                          `json:"lowLeafIndex"`
	PrevLowLeaf  intMaxTree.IndexedMerkleLeaf    `json:"prevLowLeaf"`
}

type CompressedValidityTransitionWitness struct {
	SenderLeaves                         []SenderLeaf                       `json:"senderLeaves"`
	BlockMerkleProof                     intMaxTree.PoseidonMerkleProof     `json:"blockMerkleProof"`
	SignificantAccountRegistrationProofs *[]AccountRegistrationProofOrDummy `json:"significantAccountRegistrationProofs,omitempty"`
	SignificantAccountUpdateProofs       *[]intMaxTree.IndexedUpdateProof   `json:"significantAccountUpdateProofs,omitempty"`
	CommonAccountMerkleProof             []*intMaxGP.PoseidonHashOut        `json:"commonAccountMerkleProof"`
}

func (vtw *ValidityTransitionWitness) Genesis() *ValidityTransitionWitness {
	senderLeaves := make([]SenderLeaf, 0)
	accountRegistrationProofs := make([]intMaxTree.IndexedInsertionProof, 0)
	accountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0)

	// Create a empty block hash tree
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	prevRoot := blockHashTree.GetRoot()
	prevLeafHash := new(intMaxTree.BlockHashLeaf).SetDefault().Hash()
	blockMerkleProof, _, err := blockHashTree.Prove(0)
	if err != nil {
		panic(err)
	}

	// verify
	err = blockMerkleProof.Verify(prevLeafHash, 0, prevRoot)
	if err != nil {
		panic(err)
	}

	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	genesisBlockHash := intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())
	newRoot, err := blockHashTree.AddLeaf(0, genesisBlockHash)
	if err != nil {
		panic(err)
	}
	err = blockMerkleProof.Verify(genesisBlockHash.Hash(), 0, newRoot)
	if err != nil {
		panic(err)
	}

	return &ValidityTransitionWitness{
		SenderLeaves:     senderLeaves,
		BlockMerkleProof: blockMerkleProof,
		AccountRegistrationProofs: AccountRegistrationProofsOption{
			IsSome: false,
			Proofs: accountRegistrationProofs,
		},
		AccountUpdateProofs: AccountUpdateProofsOption{
			IsSome: false,
			Proofs: accountUpdateProofs,
		},
	}
}

func (vtw *ValidityTransitionWitness) Compress(maxAccountID uint64) (compressed *CompressedValidityTransitionWitness, err error) {
	compressed = &CompressedValidityTransitionWitness{
		SenderLeaves:             vtw.SenderLeaves,
		BlockMerkleProof:         vtw.BlockMerkleProof,
		CommonAccountMerkleProof: make([]*intMaxGP.PoseidonHashOut, 0),
	}

	significantHeight := int(effectiveBits(uint(maxAccountID)))

	if vtw.AccountRegistrationProofs.IsSome {
		accountRegistrationProofs := vtw.AccountRegistrationProofs.Proofs
		compressed.CommonAccountMerkleProof = accountRegistrationProofs[0].LowLeafProof.Siblings[significantHeight:]
		significantAccountRegistrationProofs := make([]AccountRegistrationProofOrDummy, 0)
		for _, proof := range accountRegistrationProofs {
			var lowLeafProof *intMaxTree.PoseidonMerkleProof = nil
			if !proof.LowLeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LowLeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid low leaf proof")
					}

					lowLeafProof = &intMaxTree.PoseidonMerkleProof{
						Siblings: proof.LowLeafProof.Siblings[:significantHeight],
					}
				}
			}

			var leafProof *intMaxTree.PoseidonMerkleProof = nil
			if !proof.LeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid leaf proof")
					}

					leafProof = &intMaxTree.PoseidonMerkleProof{
						Siblings: proof.LeafProof.Siblings[:significantHeight],
					}
				}
			}

			significantAccountRegistrationProofs = append(significantAccountRegistrationProofs, AccountRegistrationProofOrDummy{
				LowLeafProof: lowLeafProof,
				LeafProof:    leafProof,
				Index:        uint64(proof.Index),
				LowLeafIndex: uint64(proof.LowLeafIndex),
				PrevLowLeaf:  *proof.PrevLowLeaf,
			})
		}

		compressed.SignificantAccountRegistrationProofs = &significantAccountRegistrationProofs
	}

	if vtw.AccountUpdateProofs.IsSome {
		accountUpdateProofs := vtw.AccountUpdateProofs.Proofs
		compressed.CommonAccountMerkleProof = accountUpdateProofs[0].LeafProof.Siblings[significantHeight:]
		significantAccountUpdateProofs := make([]intMaxTree.IndexedUpdateProof, 0)
		for _, proof := range accountUpdateProofs {
			for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
				if proof.LeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
					panic("invalid leaf proof")
				}

				significantAccountUpdateProofs = append(significantAccountUpdateProofs, intMaxTree.IndexedUpdateProof{
					LeafProof: intMaxTree.IndexedMerkleProof{
						Siblings: proof.LeafProof.Siblings[:significantHeight],
					},
					LeafIndex: proof.LeafIndex,
					PrevLeaf:  proof.PrevLeaf,
				})
			}
		}

		compressed.SignificantAccountUpdateProofs = &significantAccountUpdateProofs
	}

	return compressed, nil
}

type AccountMerkleProof struct {
	MerkleProof intMaxTree.IndexedMerkleProof `json:"merkleProof"`
	Leaf        intMaxTree.IndexedMerkleLeaf  `json:"leaf"`
}

func (proof *AccountMerkleProof) Verify(publicKey intMaxTypes.Uint256, accountID uint64, accountTreeRoot *intMaxGP.PoseidonHashOut) error {
	if publicKey.IsDummyPublicKey() {
		return errors.New("public key is zero")
	}

	err := proof.MerkleProof.Verify(&proof.Leaf, int(accountID), accountTreeRoot)
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

func (b *AccountIdPacked) Set(other *AccountIdPacked) *AccountIdPacked {
	if other == nil {
		b = nil
		return b
	}

	for i := 0; i < numAccountIDPackedBytes; i++ {
		b[i] = other[i]
	}

	return b
}

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
	accountIDsBytes := make([]byte, numAccountIDBytes*numOfSenders)
	for i, accountID := range accountIDs {
		chunkBytes := make([]byte, int8Key)
		binary.BigEndian.PutUint64(chunkBytes, accountID)
		copy(accountIDsBytes[i*numAccountIDBytes:(i+1)*numAccountIDBytes], chunkBytes[int8Key-numAccountIDBytes:])
	}
	const defaultAccountID = uint64(1)
	for i := len(accountIDs); i < numOfSenders; i++ {
		chunkBytes := make([]byte, int8Key)
		binary.BigEndian.PutUint64(chunkBytes, defaultAccountID)
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
	commitment := intMaxGP.HashNoPad(flattenSignatureContent)

	return commitment
}

func (s *SignatureContent) Hash() common.Hash {
	commitment := s.Commitment()
	result := new(intMaxTypes.Bytes32).FromPoseidonHashOut(commitment)

	return common.Hash(result.Bytes())
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
	if !messagePointExpected.Equal(messagePoint) {
		// fmt.Printf("messagePointExpected: %v\n", messagePointExpected)
		// fmt.Printf("messagePoint: %v\n", messagePoint)
		return errors.New("message point check failed")
	}

	return nil
}

// Verify that the calculation of agg_pubkey matches.
// It is assumed that the format validation has already passed.
func (s *SignatureContent) VerifyAggregation(publicKey []intMaxTypes.Uint256) error {
	if len(publicKey) != numOfSenders {
		return errors.New("public keys length is invalid")
	}

	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, pubKey := range publicKey {
		senderFlagBit := getBitFromUint32Slice(s.SenderFlag[:], i)
		publicKey, err := intMaxAcc.NewPublicKeyFromAddressInt(pubKey.BigInt())
		if err != nil {
			return errors.New("public key recovery check failed")
		}

		publicKeysHash := s.PublicKeyHash.Bytes()
		if senderFlagBit {
			weightedPublicKey := publicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggPublicKey := intMaxTypes.NewG1AffineFromFlatG1(&s.AggPublicKey)
	if !aggregatedPublicKey.Pk.Equal(aggPublicKey) {
		return errors.New("aggregated public key does not match")
	}

	return nil
}

type AccountMerkleProofsOption struct {
	Proofs []AccountMerkleProof
	IsSome bool
}

func (dst *AccountMerkleProofsOption) Set(src *AccountMerkleProofsOption) *AccountMerkleProofsOption {
	dst.IsSome = src.IsSome
	dst.Proofs = make([]AccountMerkleProof, len(src.Proofs))
	copy(dst.Proofs, src.Proofs)

	return dst
}

type AccountMembershipProofsOption struct {
	Proofs []intMaxTree.IndexedMembershipProof
	IsSome bool
}

func (dst *AccountMembershipProofsOption) Set(src *AccountMembershipProofsOption) *AccountMembershipProofsOption {
	dst.IsSome = src.IsSome
	dst.Proofs = make([]intMaxTree.IndexedMembershipProof, len(src.Proofs))
	copy(dst.Proofs, src.Proofs)

	return dst
}

type BlockWitness struct {
	Block                   *block_post_service.PostedBlock
	Signature               SignatureContent
	PublicKeys              []intMaxTypes.Uint256
	PrevAccountTreeRoot     *intMaxTree.PoseidonHashOut
	PrevBlockTreeRoot       *intMaxTree.PoseidonHashOut
	AccountIdPacked         *AccountIdPacked              // in account id case
	AccountMerkleProofs     AccountMerkleProofsOption     // in account id case
	AccountMembershipProofs AccountMembershipProofsOption // in pubkey case
}

type BlockWitnessFlatten struct {
	Block                   *block_post_service.PostedBlock     `json:"block"`
	Signature               SignatureContent                    `json:"signature"`
	PublicKeys              []intMaxTypes.Uint256               `json:"pubkeys"`
	PrevAccountTreeRoot     *intMaxTree.PoseidonHashOut         `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot       *intMaxTree.PoseidonHashOut         `json:"prevBlockTreeRoot"`
	AccountIdPacked         *AccountIdPacked                    `json:"accountIdPacked,omitempty"`
	AccountMerkleProofs     []AccountMerkleProof                `json:"accountMerkleProofs"`
	AccountMembershipProofs []intMaxTree.IndexedMembershipProof `json:"accountMembershipProofs"`
}

func (bw *BlockWitness) MarshalJSON() ([]byte, error) {
	result := BlockWitnessFlatten{
		Block:                   bw.Block,
		Signature:               bw.Signature,
		PublicKeys:              bw.PublicKeys,
		PrevAccountTreeRoot:     bw.PrevAccountTreeRoot,
		PrevBlockTreeRoot:       bw.PrevBlockTreeRoot,
		AccountIdPacked:         bw.AccountIdPacked,
		AccountMerkleProofs:     nil,
		AccountMembershipProofs: nil,
	}

	if bw.AccountMembershipProofs.IsSome {
		result.AccountMembershipProofs = bw.AccountMembershipProofs.Proofs
	}
	if bw.AccountMerkleProofs.IsSome {
		result.AccountMerkleProofs = bw.AccountMerkleProofs.Proofs
	}

	return json.Marshal(&result)
}

type CompressedBlockWitness struct {
	Block                              *block_post_service.PostedBlock      `json:"block"`
	Signature                          SignatureContent                     `json:"signature"`
	PublicKeys                         []intMaxTypes.Uint256                `json:"pubkeys"`
	PrevAccountTreeRoot                intMaxTree.PoseidonHashOut           `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot                  intMaxTree.PoseidonHashOut           `json:"prevBlockTreeRoot"`
	AccountIdPacked                    *AccountIdPacked                     `json:"accountIdPacked,omitempty"`                    // in account id case
	SignificantAccountMerkleProofs     *[]AccountMerkleProof                `json:"significantAccountMerkleProofs,omitempty"`     // in account id case
	SignificantAccountMembershipProofs *[]intMaxTree.IndexedMembershipProof `json:"significantAccountMembershipProofs,omitempty"` // in pubkey case
	CommonAccountMerkleProof           []*intMaxGP.PoseidonHashOut          `json:"commonAccountMerkleProof"`
}

func (bw *BlockWitness) Set(blockWitness *BlockWitness) *BlockWitness {
	bw.Block = new(block_post_service.PostedBlock).Set(blockWitness.Block)
	bw.Signature.Set(&blockWitness.Signature)
	bw.PublicKeys = make([]intMaxTypes.Uint256, len(blockWitness.PublicKeys))
	copy(bw.PublicKeys, blockWitness.PublicKeys)

	bw.PrevAccountTreeRoot = new(intMaxGP.PoseidonHashOut).Set(blockWitness.PrevAccountTreeRoot)
	bw.PrevBlockTreeRoot = new(intMaxGP.PoseidonHashOut).Set(blockWitness.PrevBlockTreeRoot)
	bw.AccountIdPacked = new(AccountIdPacked).Set(blockWitness.AccountIdPacked)
	if blockWitness.AccountMerkleProofs.IsSome {
		bw.AccountMerkleProofs.Proofs = make([]AccountMerkleProof, len(blockWitness.AccountMerkleProofs.Proofs))
		copy(bw.AccountMerkleProofs.Proofs, blockWitness.AccountMerkleProofs.Proofs)
	}
	if blockWitness.AccountMembershipProofs.IsSome {
		bw.AccountMembershipProofs.Proofs = make([]intMaxTree.IndexedMembershipProof, len(blockWitness.AccountMembershipProofs.Proofs))
		copy(bw.AccountMembershipProofs.Proofs, blockWitness.AccountMembershipProofs.Proofs)
	}

	return bw
}

func (bw *BlockWitness) Genesis() *BlockWitness {
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Genesis blockHashTree leaves: %v\n", blockHashTree.Leaves)
	prevBlockTreeRoot, _, _ := blockHashTree.GetCurrentRootCountAndSiblings()
	accountTree, err := intMaxTree.NewAccountTree(intMaxTree.ACCOUNT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}
	prevAccountTreeRoot := accountTree.GetRoot()

	return &BlockWitness{
		Block:               new(block_post_service.PostedBlock).Genesis(),
		Signature:           SignatureContent{},
		PublicKeys:          make([]intMaxTypes.Uint256, 0),
		PrevAccountTreeRoot: prevAccountTreeRoot,
		PrevBlockTreeRoot:   &prevBlockTreeRoot,
		AccountIdPacked:     nil,
		AccountMerkleProofs: AccountMerkleProofsOption{
			Proofs: nil,
			IsSome: false,
		},
		AccountMembershipProofs: AccountMembershipProofsOption{
			Proofs: nil,
			IsSome: false,
		},
	}
}

func (bw *BlockWitness) Compress(maxAccountID uint64) (compressed *CompressedBlockWitness, err error) {
	compressed = &CompressedBlockWitness{
		Block:                    bw.Block,
		Signature:                bw.Signature,
		PublicKeys:               bw.PublicKeys,
		PrevAccountTreeRoot:      *bw.PrevAccountTreeRoot,
		PrevBlockTreeRoot:        *bw.PrevBlockTreeRoot,
		AccountIdPacked:          bw.AccountIdPacked,
		CommonAccountMerkleProof: make([]*intMaxGP.PoseidonHashOut, 0),
	}

	significantHeight := effectiveBits(uint(maxAccountID))

	if bw.AccountMerkleProofs.IsSome {
		if len(bw.AccountMerkleProofs.Proofs) == 0 {
			significantAccountMerkleProofs := make([]AccountMerkleProof, 0)
			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		} else {
			accountMerkleProofs := bw.AccountMerkleProofs.Proofs
			compressed.CommonAccountMerkleProof = accountMerkleProofs[0].MerkleProof.Siblings[significantHeight:]
			significantAccountMerkleProofs := make([]AccountMerkleProof, 0)
			for _, proof := range accountMerkleProofs {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-int(significantHeight); i++ {
					if !proof.MerkleProof.Siblings[int(significantHeight)+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid common account merkle proof")
					}
				}

				significantMerkleProof := proof.MerkleProof.Siblings[:significantHeight]
				significantAccountMerkleProofs = append(significantAccountMerkleProofs, AccountMerkleProof{
					MerkleProof: intMaxTree.IndexedMerkleProof(
						intMaxTree.PoseidonMerkleProof{
							Siblings: significantMerkleProof,
						},
					),
					Leaf: proof.Leaf,
				})
			}

			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		}
	}

	if bw.AccountMembershipProofs.IsSome {
		if len(bw.AccountMembershipProofs.Proofs) == 0 {
			significantAccountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
			compressed.SignificantAccountMembershipProofs = &significantAccountMembershipProofs
		} else {
			accountMembershipProofs := bw.AccountMembershipProofs.Proofs
			compressed.CommonAccountMerkleProof = accountMembershipProofs[0].LeafProof.Siblings[significantHeight:]
			significantAccountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
			for _, proof := range accountMembershipProofs {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-int(significantHeight); i++ {
					if !proof.LeafProof.Siblings[int(significantHeight)+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid common account merkle proof")
					}
				}

				significantMerkleProof := proof.LeafProof.Siblings[:significantHeight]
				significantAccountMembershipProofs = append(significantAccountMembershipProofs, intMaxTree.IndexedMembershipProof{
					LeafProof: intMaxTree.IndexedMerkleProof{
						Siblings: significantMerkleProof,
					},
					LeafIndex:  proof.LeafIndex,
					Leaf:       proof.Leaf,
					IsIncluded: proof.IsIncluded,
				})
			}

			compressed.SignificantAccountMembershipProofs = &significantAccountMembershipProofs
		}
	}

	return compressed, nil
}

func effectiveBits(n uint) uint32 {
	if n == 0 {
		return 0
	}

	bits := uint32(0)
	for n > 0 {
		n >>= 1
		bits++
	}

	return bits
}

type MainValidationPublicInputs struct {
	PrevBlockHash       common.Hash
	BlockHash           common.Hash
	DepositTreeRoot     common.Hash
	AccountTreeRoot     *intMaxGP.PoseidonHashOut
	TxTreeRoot          intMaxTypes.Bytes32
	SenderTreeRoot      *intMaxGP.PoseidonHashOut
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
) (*AccountExclusionValue, string) {
	isValid := true
	invalidReason := ""
	for i, proof := range accountMembershipProofs {
		err := proof.Verify(publicKeys[i].BigInt(), accountTreeRoot)
		if err != nil {
			for i, sibling := range proof.LeafProof.Siblings {
				fmt.Printf("sibling[%d]: %v\n", i, sibling)
			}
			fmt.Printf("leaf index: %v\n", proof.LeafIndex)
			fmt.Printf("leaf: %v\n", proof.Leaf)
			fmt.Printf("accountTreeRoot: %s\n", accountTreeRoot.String())

			var ErrAccountMembershipProofInvalid = errors.New("account membership proof is invalid")
			// return nil, errors.Join(ErrAccountMembershipProofInvalid, err)
			panic(errors.Join(ErrAccountMembershipProofInvalid, err))
		}

		isDummy := publicKeys[i].IsDummyPublicKey()
		if !isDummy {
			fmt.Printf("isDummy: %v, ", isDummy)
			fmt.Printf("proof.IsIncluded: %v\n", proof.IsIncluded)
		}
		isExcluded := !proof.IsIncluded || isDummy
		if isValid && !isExcluded {
			fmt.Printf("proof: %+v\n", proof)
			fmt.Printf("proof.IsIncluded: %v\n", proof.IsIncluded)
			fmt.Printf("isDummy: %v\n", isDummy)
			invalidReason = fmt.Sprintf("account %d is not excluded", i)
		}
		isValid = isValid && isExcluded
	}

	publicKeyCommitment := getPublicKeyCommitment(publicKeys)

	fmt.Printf("NewAccountExclusionValue isValid: %v\n", isValid)
	return &AccountExclusionValue{
		AccountTreeRoot:         accountTreeRoot,
		AccountMembershipProofs: accountMembershipProofs,
		PublicKeys:              publicKeys,
		PublicKeyCommitment:     publicKeyCommitment,
		IsValid:                 isValid,
	}, invalidReason
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
	accountTreeRoot *intMaxTree.PoseidonHashOut,
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
		err := proof.Verify(publicKey, accountID, accountTreeRoot)
		result = result && err == nil
	}

	publicKeyCommitment := getPublicKeyCommitment(publicKeys)

	return &AccountInclusionValue{
		AccountIDPacked:     *accountIDPacked,
		AccountIDHash:       accountIDHash,
		AccountTreeRoot:     accountTreeRoot,
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
) (*FormatValidationValue, string) {
	pubkeyCommitment := getPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.IsValidFormat(publicKeys)
	var invalidReason string
	if err != nil {
		invalidReason = err.Error()
	}

	return &FormatValidationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: pubkeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
	}, invalidReason
}

type AggregationValue struct {
	PublicKeys          []intMaxTypes.Uint256
	Signature           *SignatureContent
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	SignatureCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewAggregationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *AggregationValue {
	publicKeyCommitment := getPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.VerifyAggregation(publicKeys)

	return &AggregationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: publicKeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
	}
}

func GetSenderTreeRoot(publicKeys []intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) *intMaxGP.PoseidonHashOut {
	if len(publicKeys) != numOfSenders {
		panic("public keys length should be equal to number of senders")
	}

	senderLeafHashes := make([]*intMaxGP.PoseidonHashOut, len(publicKeys))
	for i, publicKey := range publicKeys {
		isValid := getBitFromUint32Slice(senderFlag[:], i)
		senderLeaf := SenderLeaf{Sender: publicKey.BigInt(), IsValid: isValid}
		senderLeafHashes[i] = senderLeaf.Hash()
	}

	zeroHash := new(intMaxGP.PoseidonHashOut).SetZero()
	senderTree, err := intMaxTree.NewPoseidonIncrementalMerkleTree(intMaxTree.TX_TREE_HEIGHT, senderLeafHashes, zeroHash)
	if err != nil {
		panic(err)
	}

	root, _, _ := senderTree.GetCurrentRootCountAndSiblings()

	return &root
}

// NOTICE: If the content can be posted to the blockchain, the value of MainValidationPublicInputs is output
// even if it is invalid, and return the reason as a string.
func (w *BlockWitness) MainValidationPublicInputs() (*MainValidationPublicInputs, string) {
	var invalidReason string
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
		}, "genesis block is invalid"
	}

	result := true
	block := new(block_post_service.PostedBlock).Set(w.Block)
	signature := new(SignatureContent).Set(&w.Signature)
	publicKeys := make([]intMaxTypes.Uint256, len(w.PublicKeys))
	copy(publicKeys, w.PublicKeys)

	prevAccountTreeRoot := w.PrevAccountTreeRoot

	publicKeysHash := GetPublicKeysHash(publicKeys)
	isRegistrationBlock := signature.IsRegistrationBlock
	isPublicKeyEq := signature.PublicKeyHash == publicKeysHash
	if isRegistrationBlock {
		if !isPublicKeyEq {
			panic("pubkey hash mismatch")
		}
		fmt.Printf("blockWitness blockNumber: %v\n", w.Block.BlockNumber)
		fmt.Printf("blockWitness accountMembershipProof2: %v\n", w.AccountMembershipProofs.IsSome)
		fmt.Printf("blockWitness accountMerkleProof2: %v\n", w.AccountMerkleProofs.IsSome)

		if !w.AccountMembershipProofs.IsSome {
			panic("account membership proofs should be given")
		}

		// Account exclusion verification
		accountExclusionValue, invalidAccountExclusionValueReason := NewAccountExclusionValue(
			prevAccountTreeRoot,
			w.AccountMembershipProofs.Proofs,
			publicKeys,
		)
		if invalidAccountExclusionValueReason != "" {
			fmt.Printf("WARNING: invalid reason (MainValidationPublicInputs): %s\n", invalidAccountExclusionValueReason)
			// panic("account exclusion value is invalid: " + invalidReason)
		}

		if result && !accountExclusionValue.IsValid {
			fmt.Printf("prevAccountTreeRoot: %s\n", prevAccountTreeRoot.String())
			invalidReason = fmt.Sprintf("account exclusion value is invalid: %s", invalidAccountExclusionValueReason)
		}

		fmt.Printf("accountExclusionValue.IsValid: %v\n", accountExclusionValue.IsValid)
		result = result && accountExclusionValue.IsValid
	} else {

		if result && !isPublicKeyEq {
			invalidReason = "public key is invalid"
		}
		result = result && isPublicKeyEq

		if w.AccountIdPacked != nil {
			panic("account id packed should be given")
		}

		if !w.AccountMerkleProofs.IsSome {
			panic("account merkle proofs should be given")
		}

		// Account inclusion verification
		accountInclusionValue, err := NewAccountInclusionValue(
			prevAccountTreeRoot,
			w.AccountIdPacked,
			w.AccountMerkleProofs.Proofs,
			publicKeys,
		)
		if err != nil {
			panic(fmt.Errorf("account inclusion value is invalid: %v", err))
		}

		if result && !accountInclusionValue.IsValid {
			invalidReason = "account inclusion value is invalid"
		}
		result = result && accountInclusionValue.IsValid
	}

	// Format validation
	formatValidationValue, invalidFormatValidationReason := NewFormatValidationValue(publicKeys, signature)
	if result && !formatValidationValue.IsValid {
		invalidReason = fmt.Sprintf("formatValidationValue is invalid: %s", invalidFormatValidationReason)
	}
	result = result && formatValidationValue.IsValid

	if result {
		aggregationValue := NewAggregationValue(publicKeys, signature)
		if !aggregationValue.IsValid {
			invalidReason = "aggregationValue is invalid"
		}
		result = result && aggregationValue.IsValid
	}

	prev_block_hash := block.PrevBlockHash
	blockHash := block.Hash()
	senderTreeRoot := GetSenderTreeRoot(publicKeys, signature.SenderFlag)

	txTreeRoot := signature.TxTreeRoot

	return &MainValidationPublicInputs{
		PrevBlockHash:       prev_block_hash,
		BlockHash:           blockHash,
		DepositTreeRoot:     block.DepositRoot,
		AccountTreeRoot:     prevAccountTreeRoot,
		TxTreeRoot:          txTreeRoot,
		SenderTreeRoot:      senderTreeRoot,
		BlockNumber:         block.BlockNumber,
		IsRegistrationBlock: isRegistrationBlock,
		IsValid:             result,
	}, invalidReason
}

type ValidityWitness struct {
	BlockWitness              *BlockWitness              `json:"blockWitness"`
	ValidityTransitionWitness *ValidityTransitionWitness `json:"validityTransitionWitness"`
}

func (vw *ValidityWitness) Set(validityWitness *ValidityWitness) *ValidityWitness {
	vw.BlockWitness = new(BlockWitness).Set(validityWitness.BlockWitness)
	vw.ValidityTransitionWitness = new(ValidityTransitionWitness).Set(validityWitness.ValidityTransitionWitness)

	return vw
}

func (vw *ValidityWitness) Genesis() *ValidityWitness {
	return &ValidityWitness{
		BlockWitness:              new(BlockWitness).Genesis(),
		ValidityTransitionWitness: new(ValidityTransitionWitness).Genesis(),
	}
}

type CompressedValidityWitness struct {
	BlockWitness              *CompressedBlockWitness              `json:"blockWitness"`
	ValidityTransitionWitness *CompressedValidityTransitionWitness `json:"validityTransitionWitness"`
}

func (w *ValidityWitness) Compress(maxAccountID uint64) (*CompressedValidityWitness, error) {
	blockWitness, err := w.BlockWitness.Compress(maxAccountID)
	if err != nil {
		return nil, err
	}

	validityTransitionWitness, err := w.ValidityTransitionWitness.Compress(maxAccountID)
	if err != nil {
		return nil, err
	}

	return &CompressedValidityWitness{
		BlockWitness:              blockWitness,
		ValidityTransitionWitness: validityTransitionWitness,
	}, nil
}

func (vw *ValidityWitness) ValidityPublicInputs() *ValidityPublicInputs {
	fmt.Printf("ValidityWitness.ValidityPublicInputs\n")
	blockWitness := vw.BlockWitness
	validityTransitionWitness := vw.ValidityTransitionWitness

	prevBlockTreeRoot := blockWitness.PrevBlockTreeRoot

	// Check transition block tree root
	block := blockWitness.Block
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	fmt.Printf("old block root: %s\n", prevBlockTreeRoot.String())
	err := validityTransitionWitness.BlockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(block.BlockNumber),
		prevBlockTreeRoot,
	)
	if err != nil {
		panic("Block merkle proof is invalid")
	}

	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	blockTreeRoot := validityTransitionWitness.BlockMerkleProof.GetRoot(blockHashLeaf.Hash(), int(block.BlockNumber))
	fmt.Printf("new block root: %s\n", blockTreeRoot.String())

	mainValidationPis, invalidReason := blockWitness.MainValidationPublicInputs()

	// transition account tree root
	prevAccountTreeRoot := blockWitness.PrevAccountTreeRoot
	accountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(prevAccountTreeRoot) // mutable
	fmt.Printf("mainValidationPis.IsValid: %v\n", mainValidationPis.IsValid)
	if !mainValidationPis.IsValid {
		fmt.Printf("WARNING: invalid reason (ValidityPublicInputs): %s\n", invalidReason)
	}
	fmt.Printf("mainValidationPis.IsRegistrationBlock: %v\n", mainValidationPis.IsRegistrationBlock)
	if mainValidationPis.IsValid && mainValidationPis.IsRegistrationBlock {
		accountRegistrationProofs := validityTransitionWitness.AccountRegistrationProofs
		if !accountRegistrationProofs.IsSome {
			panic("account registration proofs should be given")
		}
		for i, senderLeaf := range validityTransitionWitness.SenderLeaves {
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
		accountUpdateProofs := validityTransitionWitness.AccountUpdateProofs
		if !accountUpdateProofs.IsSome {
			panic("account update proofs should be given")
		}
		for i, senderLeaf := range validityTransitionWitness.SenderLeaves {
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

	fmt.Printf("blockNumber (ValidityPublicInputs): %d\n", block.BlockNumber)
	fmt.Printf("prevAccountTreeRoot (ValidityPublicInputs): %s\n", prevAccountTreeRoot.String())
	fmt.Printf("accountTreeRoot (ValidityPublicInputs): %s\n", accountTreeRoot.String())
	return &ValidityPublicInputs{
		PublicState: &PublicState{
			BlockTreeRoot:       blockTreeRoot,
			PrevAccountTreeRoot: prevAccountTreeRoot,
			AccountTreeRoot:     accountTreeRoot,
			DepositTreeRoot:     block.DepositRoot,
			BlockHash:           mainValidationPis.BlockHash,
			BlockNumber:         block.BlockNumber,
		},
		TxTreeRoot:     mainValidationPis.TxTreeRoot,
		SenderTreeRoot: mainValidationPis.SenderTreeRoot,
		IsValidBlock:   mainValidationPis.IsValid,
	}
}

// AuxInfo is a structure for recording past tree states.
type AuxInfo struct {
	BlockContent *intMaxTypes.BlockContent
	PostedBlock  *block_post_service.PostedBlock
}

type MerkleTrees struct {
	AccountTree   *intMaxTree.AccountTree
	BlockHashTree *intMaxTree.BlockHashTree
	DepositLeaves []*intMaxTree.DepositLeaf
}

type MerkleTreeHistory struct {
	MerkleTrees     map[uint32]*MerkleTrees
	lastBlockNumber uint32
}

func NewMerkleTreeHistory(lastBlockNumber uint32, merkleTrees map[uint32]*MerkleTrees) *MerkleTreeHistory {
	return &MerkleTreeHistory{
		MerkleTrees:     merkleTrees,
		lastBlockNumber: lastBlockNumber,
	}
}

func (history *MerkleTreeHistory) LastBlockNumber() uint32 {
	return history.lastBlockNumber
}

func (history *MerkleTreeHistory) PushHistory(merkleTrees *MerkleTrees) {
	_, nextBlockNumber, _ := merkleTrees.BlockHashTree.GetCurrentRootCountAndSiblings()
	newBlockNumber := history.lastBlockNumber + 1
	fmt.Printf("nextBlockNumber (PushHistory): %d\n", nextBlockNumber)
	fmt.Printf("newBlockNumber (PushHistory): %d\n", newBlockNumber)
	if nextBlockNumber == newBlockNumber {
		panic("Block number mismatch")
	}
	history.MerkleTrees[newBlockNumber] = merkleTrees
	history.lastBlockNumber = newBlockNumber
}
