package block_validity_prover

import (
	"context"
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

type AccountRegistrationProofs struct {
	Proofs  []intMaxTree.IndexedInsertionProof `json:"proofs"`
	IsValid bool                               `json:"isValid"`
}

func (arp *AccountRegistrationProofs) Set(other *AccountRegistrationProofs) *AccountRegistrationProofs {
	arp.IsValid = other.IsValid
	arp.Proofs = make([]intMaxTree.IndexedInsertionProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}

type AccountUpdateProofs struct {
	Proofs  []intMaxTree.IndexedUpdateProof `json:"proofs"`
	IsValid bool                            `json:"isValid"`
}

func (arp *AccountUpdateProofs) Set(other *AccountUpdateProofs) *AccountUpdateProofs {
	arp.IsValid = other.IsValid
	arp.Proofs = make([]intMaxTree.IndexedUpdateProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}

type ValidityTransitionWitness struct {
	SenderLeaves              []SenderLeaf              `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.MerkleProof    `json:"blockMerkleProof"`
	AccountRegistrationProofs AccountRegistrationProofs `json:"accountRegistrationProofs"`
	AccountUpdateProofs       AccountUpdateProofs       `json:"accountUpdateProofs"`
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
	LowLeafProof *intMaxTree.MerkleProof      `json:"lowLeafProof,omitempty"`
	LeafProof    *intMaxTree.MerkleProof      `json:"leafProof,omitempty"`
	Index        uint64                       `json:"index"`
	LowLeafIndex uint64                       `json:"lowLeafIndex"`
	PrevLowLeaf  intMaxTree.IndexedMerkleLeaf `json:"prevLowLeaf"`
}

type CompressedValidityTransitionWitness struct {
	SenderLeaves                         []SenderLeaf                       `json:"senderLeaves"`
	BlockMerkleProof                     intMaxTree.MerkleProof             `json:"blockMerkleProof"`
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

func (vtw *ValidityTransitionWitness) Compress(maxAccountID uint64) (compressed *CompressedValidityTransitionWitness, err error) {
	compressed = &CompressedValidityTransitionWitness{
		SenderLeaves:             vtw.SenderLeaves,
		BlockMerkleProof:         vtw.BlockMerkleProof,
		CommonAccountMerkleProof: make([]*intMaxGP.PoseidonHashOut, 0),
	}

	significantHeight := int(effectiveBits(uint(maxAccountID)))

	if vtw.AccountRegistrationProofs.IsValid {
		accountRegistrationProofs := vtw.AccountRegistrationProofs.Proofs
		compressed.CommonAccountMerkleProof = accountRegistrationProofs[0].LowLeafProof.Siblings[significantHeight:]
		significantAccountRegistrationProofs := make([]AccountRegistrationProofOrDummy, 0)
		for _, proof := range accountRegistrationProofs {
			var lowLeafProof *intMaxTree.MerkleProof = nil
			if !proof.LowLeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LowLeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid low leaf proof")
					}

					lowLeafProof = &intMaxTree.MerkleProof{
						Siblings: proof.LowLeafProof.Siblings[:significantHeight],
					}
				}
			}

			var leafProof *intMaxTree.MerkleProof = nil
			if !proof.LeafProof.IsDummy(intMaxTree.ACCOUNT_TREE_HEIGHT) {
				for i := 0; i < int(intMaxTree.ACCOUNT_TREE_HEIGHT)-significantHeight; i++ {
					if !proof.LeafProof.Siblings[significantHeight+i].Equal(compressed.CommonAccountMerkleProof[i]) {
						panic("invalid leaf proof")
					}

					leafProof = &intMaxTree.MerkleProof{
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

	if vtw.AccountUpdateProofs.IsValid {
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

func (proof *AccountMerkleProof) Verify(accountTreeRoot *intMaxGP.PoseidonHashOut, accountID uint64, publicKey intMaxTypes.Uint256) error {
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
		return nil
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
	fmt.Printf("messagePointExpected: %v\n", messagePointExpected)
	fmt.Printf("messagePoint: %v\n", messagePoint)
	if !messagePointExpected.Equal(messagePoint) {
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

type BlockWitness struct {
	Block                   *block_post_service.PostedBlock      `json:"block"`
	Signature               SignatureContent                     `json:"signature"`
	PublicKeys              []intMaxTypes.Uint256                `json:"pubkeys"`
	PrevAccountTreeRoot     *intMaxTree.PoseidonHashOut          `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot       *intMaxTree.PoseidonHashOut          `json:"prevBlockTreeRoot"`
	AccountIdPacked         *AccountIdPacked                     `json:"accountIdPacked,omitempty"`         // in account id case
	AccountMerkleProofs     *[]AccountMerkleProof                `json:"accountMerkleProofs,omitempty"`     // in account id case
	AccountMembershipProofs *[]intMaxTree.IndexedMembershipProof `json:"accountMembershipProofs,omitempty"` // in pubkey case
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
	if blockWitness.AccountMerkleProofs != nil {
		bw.AccountMerkleProofs = new([]AccountMerkleProof)
		copy(*bw.AccountMerkleProofs, *blockWitness.AccountMerkleProofs)
	}
	if blockWitness.AccountMembershipProofs != nil {
		bw.AccountMembershipProofs = new([]intMaxTree.IndexedMembershipProof)
		copy(*bw.AccountMembershipProofs, *blockWitness.AccountMembershipProofs)
	}

	return bw
}

func (bw *BlockWitness) Genesis() *BlockWitness {
	blockHashTree, err := intMaxTree.NewBlockHashTreeWithInitialLeaves(intMaxTree.BLOCK_HASH_TREE_HEIGHT, nil)
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
		PrevBlockTreeRoot:       &prevBlockTreeRoot,
		AccountIdPacked:         nil,
		AccountMerkleProofs:     nil,
		AccountMembershipProofs: nil,
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

	if bw.AccountMerkleProofs != nil {
		if len(*bw.AccountMerkleProofs) == 0 {
			significantAccountMerkleProofs := make([]AccountMerkleProof, 0)
			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		} else {
			accountMerkleProofs := *bw.AccountMerkleProofs
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
						intMaxTree.MerkleProof{
							Siblings: significantMerkleProof,
						},
					),
					Leaf: proof.Leaf,
				})
			}

			compressed.SignificantAccountMerkleProofs = &significantAccountMerkleProofs
		}
	}

	if bw.AccountMembershipProofs != nil {
		if len(*bw.AccountMerkleProofs) == 0 {
			significantAccountMembershipProofs := make([]intMaxTree.IndexedMembershipProof, 0)
			compressed.SignificantAccountMembershipProofs = &significantAccountMembershipProofs
		} else {
			accountMembershipProofs := *bw.AccountMembershipProofs
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
) (*AccountExclusionValue, error) {
	result := true
	for i, proof := range accountMembershipProofs {
		err := proof.Verify(publicKeys[i].BigInt(), accountTreeRoot)
		if err != nil {
			var ErrAccountMembershipProofInvalid = errors.New("account membership proof is invalid")
			return nil, errors.Join(ErrAccountMembershipProofInvalid, err)
		}

		isDummy := publicKeys[i].IsDummyPublicKey()
		fmt.Printf("proof.IsIncluded: %v\n", proof.IsIncluded)
		fmt.Printf("isDummy: %v\n", isDummy)
		isExcluded := !proof.IsIncluded || isDummy
		fmt.Printf("isExcluded: %v\n", isExcluded)
		result = result && isExcluded
	}
	fmt.Printf("publicKeys[127] = %v\n", publicKeys[127])

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
		err := proof.Verify(accountTreeRoot, accountID, publicKey)
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
) *FormatValidationValue {
	pubkeyCommitment := getPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.IsValidFormat(publicKeys)
	if err != nil {
		fmt.Printf("NewFormatValidationValue err: %v\n", err)
	}

	return &FormatValidationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: pubkeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
	}
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
		fmt.Printf("isPubkeyEq: %v\n", isPubkeyEq)
		result = result && isPubkeyEq
	}
	if isRegistrationBlock {
		if w.AccountMembershipProofs == nil {
			panic("account membership proofs should be given")
		}

		// Account exclusion verification
		accountExclusionValue, err := NewAccountExclusionValue(
			accountTreeRoot,
			*w.AccountMembershipProofs,
			publicKeys,
		)
		if err != nil {
			panic("account exclusion value is invalid: " + err.Error())
		}

		fmt.Printf("accountExclusionValue.IsValid (registration): %v\n", accountExclusionValue.IsValid)
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

		fmt.Printf("accountInclusionValue.IsValid (non registration): %v\n", accountInclusionValue.IsValid)
		result = result && accountInclusionValue.IsValid
	}

	fmt.Printf("result: %v\n", result)

	// Format validation
	formatValidationValue :=
		NewFormatValidationValue(publicKeys, signature)
	fmt.Printf("formatValidationValue.IsValid: %v\n", formatValidationValue.IsValid)
	result = result && formatValidationValue.IsValid

	if result {
		aggregationValue := NewAggregationValue(publicKeys, signature)
		fmt.Printf("aggregationValue.IsValid: %v\n", aggregationValue.IsValid)
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
	prevBlockTreeRoot := vw.BlockWitness.PrevBlockTreeRoot

	// Check transition block tree root
	block := vw.BlockWitness.Block
	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	fmt.Printf("old block root: %s\n", prevBlockTreeRoot.String())
	err := vw.ValidityTransitionWitness.BlockMerkleProof.Verify(
		defaultLeaf.Hash(),
		int(block.BlockNumber),
		prevBlockTreeRoot,
	)

	if err != nil {
		panic("Block merkle proof is invalid")
	}
	blockHashLeaf := intMaxTree.NewBlockHashLeaf(block.Hash())
	blockTreeRoot := vw.ValidityTransitionWitness.BlockMerkleProof.GetRoot(blockHashLeaf.Hash(), int(block.BlockNumber))
	fmt.Printf("new block root: %s\n", blockTreeRoot.String())

	mainValidationPis := vw.BlockWitness.MainValidationPublicInputs()

	// transition account tree root
	prevAccountTreeRoot := vw.BlockWitness.PrevAccountTreeRoot
	accountTreeRoot := new(intMaxGP.PoseidonHashOut).Set(prevAccountTreeRoot)
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
	// ValidityWitness *ValidityWitness
	BlockContent *intMaxTypes.BlockContent
	PostedBlock  *block_post_service.PostedBlock
}

type MerkleTrees struct {
	AccountTree   *intMaxTree.AccountTree
	BlockHashTree *intMaxTree.BlockHashTree
	DepositLeaves []*intMaxTree.DepositLeaf
	// TxTree        *intMaxTree.PoseidonMerkleTree
}

type mockBlockBuilder struct {
	db                            SQLDriverApp
	LastPostedBlockNumber         uint32
	LastGeneratedProofBlockNumber uint32
	AccountTree                   *intMaxTree.AccountTree      // current account tree
	BlockTree                     *intMaxTree.BlockHashTree    // current block hash tree
	DepositTree                   *intMaxTree.KeccakMerkleTree // current deposit tree
	DepositLeaves                 []*intMaxTree.DepositLeaf
	// DepositTreeRoots           []common.Hash
	LastSeenProcessedDepositId uint64
	ValidityProofs             map[uint32]string
	AuxInfo                    map[uint32]*mDBApp.BlockContent
	validityWitnesses          map[uint32]*ValidityWitness
	LatestWitnessBlockNumber   uint32
	MerkleTreeHistory          map[uint32]*MerkleTrees
	DepositTreeHistory         map[string]*intMaxTree.KeccakMerkleTree // deposit hash -> deposit tree
}

type MockBlockBuilderMemory = mockBlockBuilder

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
	validityWitnesses := make(map[uint32]*ValidityWitness)
	validityWitnesses[0] = new(ValidityWitness).Genesis()
	auxInfo := make(map[uint32]*mDBApp.BlockContent)

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
		// txTree: TxTree::new(TX_TREE_HEIGHT),
		// ValidityWitness: validityWitness,
		AccountTree:   new(intMaxTree.AccountTree).Set(accountTree),
		BlockHashTree: new(intMaxTree.BlockHashTree).Set(blockTree),
		DepositLeaves: make([]*intMaxTree.DepositLeaf, len(depositLeaves)),
	}
	copy(merkleTreeHistory[0].DepositLeaves, depositLeaves)

	return &mockBlockBuilder{
		db:                            db,
		LastGeneratedProofBlockNumber: 0,
		validityWitnesses:             validityWitnesses,
		ValidityProofs:                make(map[uint32]string),
		AccountTree:                   accountTree,
		BlockTree:                     blockTree,
		DepositTree:                   depositTree,
		DepositLeaves:                 depositLeaves,
		// DepositLeavesByHash: make(map[common.Hash]*DepositLeafWithId),
		// DepositTreeRoots: []common.Hash{depositTreeRoot},
		// lastSeenProcessDepositsEventBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		// lastSeenBlockPostedEventBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		AuxInfo:           auxInfo,
		MerkleTreeHistory: merkleTreeHistory,
		// DepositTreeHistory: make(map[string]*intMaxTree.KeccakMerkleTree),
	}
}

func (b *mockBlockBuilder) Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error) {
	return b.db.Exec(ctx, input, executor)
}

type DepositLeafWithId struct {
	DepositLeaf *intMaxTree.DepositLeaf
	DepositId   uint32
}

func (b *mockBlockBuilder) GenerateBlock(
	blockContent *intMaxTypes.BlockContent,
	postedBlock *block_post_service.PostedBlock,
) (*BlockWitness, error) {
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
				return nil, errors.New("account membership proof error")
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

func (db *mockBlockBuilder) SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error {
	fmt.Printf("---------------------- SetValidityWitness ----------------------\n")
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
			if depositTreeRoot == witness.BlockWitness.Block.DepositRoot {
				break
			}
		}
	}

	db.LatestWitnessBlockNumber = blockNumber
	db.validityWitnesses[blockNumber] = new(ValidityWitness).Set(witness)
	db.MerkleTreeHistory[blockNumber] = &MerkleTrees{
		AccountTree:   db.AccountTree,
		BlockHashTree: db.BlockTree,
		DepositLeaves: depositTree.Leaves,
		// TxTree:        txTree,
	}

	return nil
}

func (db *mockBlockBuilder) LastValidityWitness() (*ValidityWitness, error) {
	// return db.db.LastValidityWitness()
	// fmt.Printf("LastValidityWitness: %v\n", db.lastValidityWitness.BlockWitness.PrevBlockTreeRoot)

	return db.validityWitnesses[uint32(len(db.validityWitnesses)-1)], nil
}

func (db *mockBlockBuilder) AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error) {
	return db.AccountTree.GetRoot(), nil
}

func (db *mockBlockBuilder) GetAccountMembershipProof(blockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error) {
	blockHistory, ok := db.MerkleTreeHistory[blockNumber]
	if !ok {
		return nil, fmt.Errorf("current block number %d not found", blockNumber)
	}
	proof, _, err := blockHistory.AccountTree.ProveMembership(publicKey)
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

func (db *mockBlockBuilder) DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, error) {
	depositLeaves := db.MerkleTreeHistory[blockNumber].DepositLeaves

	if depositIndex >= uint32(len(depositLeaves)) {
		return nil, errors.New("block number is out of range")
	}

	leaves := make([][32]byte, 0)
	for i, depositLeaf := range depositLeaves {
		fmt.Printf("leaves[%d]: %v\n", i, depositLeaf.Hash())
		leaves = append(leaves, [32]byte(depositLeaf.Hash()))
	}
	proof, root, err := db.DepositTree.ComputeMerkleProof(depositIndex, leaves)
	if err != nil {
		var ErrDepositTreeProof = errors.New("deposit tree proof error")
		return nil, errors.Join(ErrDepositTreeProof, err)
	}
	fmt.Printf("DepositTreeProof deposit tree root: %s\n", root.String())

	return proof, nil
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
		return nil, errors.New("account id not found")
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

	fmt.Printf("sortedTxs: %v\n", sortedTxs)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, keyPair := range sortedTxs {
		weightedPublicKey := keyPair.Sender.Public().WeightByHash(publicKeysHash.Bytes())
		aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
	}

	if aggregatedPublicKey.Pk == nil {
		aggregatedPublicKey.Pk = new(bn254.G1Affine)
		aggregatedPublicKey.Pk.X.SetZero()
		aggregatedPublicKey.Pk.Y.SetZero()
	}

	fmt.Printf("aggregatedPublicKey: %v\n", aggregatedPublicKey)
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
		fmt.Printf("tx: %v\n", tx)
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
	blockNumber := lastValidityWitness.BlockWitness.Block.BlockNumber

	var accountIDPacked *AccountIdPacked
	var accountMerkleProofs []AccountMerkleProof
	var accountMembershipProofs []intMaxTree.IndexedMembershipProof
	accountIDs := make([]uint64, 0)
	if isRegistrationBlock {
		accountMembershipProofs = make([]intMaxTree.IndexedMembershipProof, 0)
		for _, publicKey := range publicKeys {
			isDummy := publicKey.BigInt().Cmp(intMaxAcc.NewDummyPublicKey().BigInt()) == 0
			if !isDummy {
				proof, err := db.GetAccountMembershipProof(blockNumber, publicKey.BigInt())
				if err != nil {
					panic("account membership proof error")
				}

				accountMembershipProofs = append(accountMembershipProofs, *proof)
			}
		}
	} else {
		accountMerkleProofs = make([]AccountMerkleProof, 0)
		for _, publicKey := range publicKeys {
			accountID, ok := db.AccountTree.GetAccountID(publicKey.BigInt())
			if !ok {
				panic("account id not found")
			}
			proof, err := db.ProveInclusion(accountID)
			if err != nil {
				panic("account inclusion proof error")
			}

			accountIDs = append(accountIDs, accountID)
			accountMerkleProofs = append(accountMerkleProofs, AccountMerkleProof{
				MerkleProof: proof.MerkleProof,
				Leaf:        proof.Leaf,
			})
		}

		accountIDPacked = new(AccountIdPacked).Pack(accountIDs)
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
		GetAccountIDsHash(accountIDs),
		isRegistrationBlock,
		sortedTxs,
	)
	if err != nil {
		panic(err)
	}

	depositRoot, _, _ := db.DepositTree.GetCurrentRootCountAndSiblings()
	lastValidityWitness, err = db.LastValidityWitness()
	if err != nil {
		panic(err)
	}
	block := &block_post_service.PostedBlock{
		PrevBlockHash: lastValidityWitness.BlockWitness.Block.Hash(),
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

	fmt.Printf("blockWitness.PublicKeys[127] = %v\n", blockWitness.PublicKeys[127].BigInt())
	if !blockWitness.MainValidationPublicInputs().IsValid && len(txs) > 0 {
		panic("should be valid block") // XXX
	}

	return blockWitness, txTree, nil
}

type MockTxRequest struct {
	Sender              *intMaxAcc.PrivateKey
	Tx                  *intMaxTypes.Tx
	WillReturnSignature bool
}

func (db *mockBlockBuilder) PostBlock(
	isRegistrationBlock bool,
	txs []*MockTxRequest,
) (*ValidityWitness, error) {
	blockWitness, _, err := db.GenerateBlockWithTxTree(
		isRegistrationBlock,
		txs,
	)
	if err != nil {
		panic(err)
	}

	prevValidityWitness, err := db.LastValidityWitness()
	if err != nil {
		panic(err)
	}
	fmt.Printf("prev block number is %d (PostBlock)\n", prevValidityWitness.BlockWitness.Block.BlockNumber)
	validityWitness, err := generateValidityWitness(
		db,
		blockWitness,
		prevValidityWitness,
	)
	if err != nil {
		panic(err)
	}

	db.SetValidityWitness(blockWitness.Block.BlockNumber, validityWitness)

	fmt.Printf("post block #%d\n", validityWitness.BlockWitness.Block.BlockNumber)
	return validityWitness, nil
}

func generateValidityWitness(db BlockBuilderStorage, blockWitness *BlockWitness, prevValidityWitness *ValidityWitness) (*ValidityWitness, error) {
	if blockWitness.Block.BlockNumber != db.LatestIntMaxBlockNumber()+1 {
		fmt.Printf("blockWitness.Block.BlockNumber: %d\n", blockWitness.Block.BlockNumber)
		fmt.Printf("db.LatestIntMaxBlockNumber(): %d\n", db.LatestIntMaxBlockNumber())
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

	blockTreeRoot, err := db.BlockTreeRoot()
	if err != nil {
		return nil, errors.New("block tree root error")
	}
	if prevPis.IsValidBlock {
		fmt.Printf("block number %d is valid\n", prevPis.PublicState.BlockNumber+1)
	} else {
		fmt.Printf("block number %d is invalid\n", prevPis.PublicState.BlockNumber+1)
	}
	fmt.Printf("blockTreeRoot: %s\n", blockTreeRoot.String())
	if !prevPis.PublicState.BlockTreeRoot.Equal(blockTreeRoot) {
		fmt.Printf("prevPis.PublicState.BlockTreeRoot is not the same with blockTreeRoot, %s != %s", prevPis.PublicState.BlockTreeRoot.String(), blockTreeRoot.String())
		return nil, errors.New("block tree root is not equal to the last block tree root")
	}

	defaultLeaf := new(intMaxTree.BlockHashLeaf).SetDefault()
	prevBlockTreeRoot := prevPis.PublicState.BlockTreeRoot
	blockMerkleProof, err := db.CurrentBlockTreeProof(blockWitness.Block.BlockNumber)
	if err != nil {
		var ErrBlockTreeProve = errors.New("block tree prove error")
		return nil, errors.Join(ErrBlockTreeProve, err)
	}

	// debug
	err = blockMerkleProof.Verify(
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
	return b.LatestWitnessBlockNumber
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

	// b.lastSeenBlockPostedEventBlockNumber = blockNumber

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

	validityProofs, ok := b.ValidityProofs[blockNumber]
	if !ok {
		fmt.Printf("blockNumber (GetValidityProof): %d\n", blockNumber)
		return nil, ErrNoValidityProofByBlockNumber
	}

	return &validityProofs, nil
}

func (b *mockBlockBuilder) SetValidityProof(blockNumber uint32, proof string) error {
	fmt.Printf("blockNumber (SetValidityProof): %d\n", blockNumber)
	// if blockNumber != uint32(len(b.ValidityProofs)) {
	// 	return errors.New("block number should be equal to the last block number + 1")
	// }

	if _, ok := b.ValidityProofs[blockNumber]; ok {
		fmt.Printf("already exists block number %d\n", blockNumber)
		return nil
	}

	fmt.Printf("s.LastBlockNumber before SetValidityProof = %d\n", b.LastGeneratedProofBlockNumber)
	fmt.Printf("s.LastValidityProofBlockNumber before SetValidityProof = %d\n", b.LastPostedBlockNumber)
	b.ValidityProofs[blockNumber] = proof
	b.LastGeneratedProofBlockNumber = blockNumber
	// b.LastPostedBlockNumber = blockNumber
	fmt.Printf("s.LastBlockNumber after SetValidityProof= %d\n", b.LastGeneratedProofBlockNumber)

	return nil
}

func (b *mockBlockBuilder) BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContent, error) {
	return b.db.BlockContentByBlockNumber(blockNumber)

	// auxInfo, ok := b.AuxInfo[blockNumber]
	// if !ok {
	// 	return nil, false
	// }

	// return auxInfo, true
}

func BlockAuxInfo(db BlockBuilderStorage, blockNumber uint32) (*AuxInfo, error) {
	auxInfo, err := db.BlockContentByBlockNumber(blockNumber)
	if err != nil {
		return nil, errors.New("block content by block number error")
	}

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

func setAuxInfo(
	db BlockBuilderStorage,
	postedBlock *block_post_service.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
) error {
	storedBlockContent, err := db.CreateBlockContent(
		postedBlock,
		blockContent,
	)
	if err != nil {
		return err
	}

	if storedBlockContent.BlockNumber != postedBlock.BlockNumber {
		// Fatal error
		panic(fmt.Sprintf("block %d is ErrBlockNumberMismatch", postedBlock.BlockNumber))
	}

	return nil
}

func (b *mockBlockBuilder) CreateBlockContent(
	postedBlock *block_post_service.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
) (*mDBApp.BlockContent, error) {
	return b.db.CreateBlockContent(
		postedBlock,
		blockContent,
	)
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

type CircuitData interface{}

// / A dummy implementation of the transition wrapper circuit used for testing balance proof.
type TransitionWrapperCircuit interface {
	Prove(
		prevPis *ValidityPublicInputs,
		validity_witness *ValidityWitness,
	) (*intMaxTypes.Plonky2Proof, error)
	CircuitData() *CircuitData
}
