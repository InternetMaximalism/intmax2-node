package balance_prover_service

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	SENDER_TREE_HEIGHT     = 7
	balancePublicInputsLen = 47

	int2Key = 2
	int3Key = 3
	int4Key = 4
	int8Key = 8
)

type poseidonHashOut = intMaxTypes.PoseidonHashOut

type TxWitness struct {
	ValidityPis   block_validity_prover.ValidityPublicInputs `json:"validityPis"`
	SenderLeaves  []*intMaxTree.SenderLeaf                   `json:"senderLeaves"`
	Tx            intMaxTypes.Tx                             `json:"tx"`
	TxIndex       uint32                                     `json:"txIndex"`
	TxMerkleProof []*poseidonHashOut                         `json:"txMerkleProof"`
}

func (w *TxWitness) GetSenderTree() (*intMaxTree.SenderTree, error) {
	senderTree, err := intMaxTree.NewSenderTree(SENDER_TREE_HEIGHT, w.SenderLeaves)
	if err != nil {
		return nil, err
	}

	return senderTree, nil
}

type TransferInput struct {
	Recipient  string      `json:"recipient"`
	TokenIndex uint32      `json:"tokenIndex"`
	Amount     AmountInput `json:"amount"`
	Salt       SaltInput   `json:"salt"`
}

func (input *TransferInput) FromTransfer(value *intMaxTypes.Transfer) *TransferInput {
	input.Recipient = hexutil.Encode(value.Recipient.Address[:])
	input.TokenIndex = value.TokenIndex
	input.Amount = value.Amount.String()
	input.Salt = value.Salt.String()

	return input
}

type SenderLeafInput struct {
	Sender  string `json:"sender"`
	IsValid bool   `json:"isValid"`
}

type PublicStateInput struct {
	BlockTreeRoot       *poseidonHashOut `json:"blockTreeRoot"`
	PrevAccountTreeRoot *poseidonHashOut `json:"prevAccountTreeRoot"`
	AccountTreeRoot     *poseidonHashOut `json:"accountTreeRoot"`
	DepositTreeRoot     string           `json:"depositTreeRoot"`
	BlockHash           string           `json:"blockHash"`
	BlockNumber         uint32           `json:"blockNumber"`
}

func (input *PublicStateInput) FromPublicState(value *block_validity_prover.PublicState) *PublicStateInput {
	input.BlockTreeRoot = value.BlockTreeRoot
	input.PrevAccountTreeRoot = value.PrevAccountTreeRoot
	input.AccountTreeRoot = value.AccountTreeRoot
	input.DepositTreeRoot = value.DepositTreeRoot.String()
	input.BlockHash = value.BlockHash.String()
	input.BlockNumber = value.BlockNumber

	return input
}

type ValidityPublicInputsInput struct {
	PublicState    *PublicStateInput
	TxTreeRoot     string
	SenderTreeRoot string
	IsValidBlock   bool
}

type TxInput struct {
	TransferTreeRoot string
	Nonce            uint64
}

type TxWitnessInput struct {
	ValidityPis   ValidityPublicInputsInput `json:"validityPis"`
	SenderLeaves  []*SenderLeafInput        `json:"senderLeaves"`
	Tx            TxInput                   `json:"tx"`
	TxIndex       uint32                    `json:"txIndex"`
	TxMerkleProof []string                  `json:"txMerkleProof"`
}

type InsufficientFlagsInput = string

type PoseidonHashOutInput = string

type BalancePublicInputsInput struct {
	PublicKey               PoseidonHashOutInput   `json:"pubkey"`
	PrivateCommitment       PoseidonHashOutInput   `json:"privateCommitment"`
	LastTxHash              PoseidonHashOutInput   `json:"lastTxHash"`
	LastTxInsufficientFlags InsufficientFlagsInput `json:"lastTxInsufficientFlags"`
	PublicState             *PublicStateInput      `json:"publicState"`
}

func (input *BalancePublicInputsInput) FromBalancePublicInputs(value *BalancePublicInputs) *BalancePublicInputsInput {
	input.PublicKey = value.PubKey.BigInt().String()
	input.PrivateCommitment = value.PrivateCommitment.String()
	input.LastTxHash = value.LastTxHash.String()
	input.LastTxInsufficientFlags = hexutil.Encode(value.LastTxInsufficientFlags.Bytes())
	input.PublicState = new(PublicStateInput).FromPublicState(value.PublicState)

	return input
}

// Information needed to prove that a balance has been sent
type SendWitness struct {
	PrevBalancePis      *BalancePublicInputs             `json:"prevBalancePis"`
	PrevPrivateState    *PrivateState                    `json:"prevPrivateState"`
	PrevBalances        []*intMaxTree.AssetLeaf          `json:"prevBalances"`
	AssetMerkleProofs   []*intMaxTree.AssetMerkleProof   `json:"assetMerkleProofs"`
	InsufficientFlags   backup_balance.InsufficientFlags `json:"insufficientFlags"`
	Transfers           []*intMaxTypes.Transfer          `json:"transfers"`
	TxWitness           TxWitness                        `json:"txWitness"`
	NewPrivateStateSalt Salt                             `json:"newPrivateStateSalt"`
}

type SendWitnessInput struct {
	PrevBalancePis      *BalancePublicInputsInput `json:"prevBalancePis"`
	PrevPrivateState    *PrivateStateInput        `json:"prevPrivateState"`
	PrevBalances        []*AssetLeafInput         `json:"prevBalances"`
	AssetMerkleProofs   []AssetMerkleProofInput   `json:"assetMerkleProofs"`
	InsufficientFlags   InsufficientFlagsInput    `json:"insufficientFlags"`
	Transfers           []*TransferInput          `json:"transfers"`
	TxWitness           TxWitnessInput            `json:"txWitness"`
	NewPrivateStateSalt SaltInput                 `json:"newPrivateStateSalt"`
}

func (input *SendWitnessInput) FromSendWitness(value *SendWitness) *SendWitnessInput {
	input.PrevBalancePis = new(BalancePublicInputsInput).FromBalancePublicInputs(value.PrevBalancePis)
	input.PrevPrivateState = &PrivateStateInput{
		AssetTreeRoot:     value.PrevPrivateState.AssetTreeRoot,
		NullifierTreeRoot: value.PrevPrivateState.NullifierTreeRoot,
		Nonce:             value.PrevPrivateState.Nonce,
		Salt:              value.PrevPrivateState.Salt.String(),
	}
	input.PrevBalances = make([]*AssetLeafInput, len(value.PrevBalances))
	for i, balance := range value.PrevBalances {
		input.PrevBalances[i] = &AssetLeafInput{
			IsInsufficient: balance.IsInsufficient,
			Amount:         balance.Amount.BigInt().String(),
		}
	}
	input.AssetMerkleProofs = make([]AssetMerkleProofInput, len(value.AssetMerkleProofs))
	for i, proof := range value.AssetMerkleProofs {
		input.AssetMerkleProofs[i] = make([]*poseidonHashOut, len(proof.Siblings))
		copy(input.AssetMerkleProofs[i], proof.Siblings)
	}
	input.InsufficientFlags = hexutil.Encode(value.InsufficientFlags.Bytes())

	return input
}

func (w *SendWitness) GetIncludedBlockNumber() uint32 {
	return w.TxWitness.ValidityPis.PublicState.BlockNumber
}

func (w *SendWitness) GetPrevBlockNumber() uint32 {
	return w.PrevBalancePis.PublicState.BlockNumber
}

type Salt = poseidonHashOut

// func (s *Salt) SetRandom() *Salt {
// 	for _, e := range s.Elements {
// 		e.SetRandom()
// 	}

// 	return s
// }

type SpentValue struct {
	PrevPrivateState      *PrivateState                    `json:"prevPrivateState"`
	NewPrivateStateSalt   Salt                             `json:"newPrivateStateSalt"`
	Transfers             []*intMaxTypes.Transfer          `json:"transfers"`
	PrevBalances          []*intMaxTree.AssetLeaf          `json:"prevBalances"`
	AssetMerkleProofs     []*intMaxTree.AssetMerkleProof   `json:"assetMerkleProofs"`
	PrevPrivateCommitment *poseidonHashOut                 `json:"prevPrivateCommitment"`
	NewPrivateCommitment  *poseidonHashOut                 `json:"newPrivateCommitment"`
	Tx                    intMaxTypes.Tx                   `json:"tx"`
	InsufficientFlags     backup_balance.InsufficientFlags `json:"insufficientFlags"`
	IsValid               bool                             `json:"isValid"`
}

func NewSpentValue(
	prevPrivateState *PrivateState,
	prevBalances []*intMaxTree.AssetLeaf,
	newPrivateStateSalt Salt,
	transfers []*intMaxTypes.Transfer,
	assetMerkleProofs []*intMaxTree.AssetMerkleProof,
	txNonce uint64,
) (*SpentValue, error) {
	const numTransfers = 64
	if len(prevBalances) != numTransfers {
		return nil, errors.New("prevBalances length is not equal to numTransfers")
	}
	if len(transfers) != numTransfers {
		return nil, errors.New("transfers length is not equal to numTransfers")
	}
	if len(assetMerkleProofs) != numTransfers {
		return nil, errors.New("assetMerkleProofs length is not equal to numTransfers")
	}

	insufficientFlags := backup_balance.InsufficientFlags{}
	assetTreeRoot := prevPrivateState.AssetTreeRoot
	for i := 0; i < numTransfers; i++ {
		transfer := transfers[i]
		proof := assetMerkleProofs[i]
		prevBalance := prevBalances[i]

		// if err := proof.Verify(prevBalance, transfer.TokenIndex, assetTreeRoot); err != nil {
		// 	return nil, err
		// }
		newBalance := prevBalance.Sub(transfer.Amount)
		assetTreeRoot = proof.GetRoot(newBalance, transfer.TokenIndex)
		insufficientFlags.SetBit(i, newBalance.IsInsufficient)
	}

	insufficientFlags = backup_balance.InsufficientFlags(insufficientFlags)
	isValid := txNonce == uint64(prevPrivateState.Nonce)
	newPrivateState := PrivateState{
		AssetTreeRoot:     assetTreeRoot,
		Nonce:             prevPrivateState.Nonce + 1,
		Salt:              newPrivateStateSalt,
		NullifierTreeRoot: prevPrivateState.NullifierTreeRoot,
	}

	zeroTransferHash := new(intMaxTypes.Transfer).Hash()
	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, transfers, zeroTransferHash)
	if err != nil {
		return nil, err
	}

	transferRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	tx := intMaxTypes.Tx{
		TransferTreeRoot: &transferRoot,
		Nonce:            txNonce,
	}

	return &SpentValue{
		PrevPrivateState:      prevPrivateState,
		NewPrivateStateSalt:   newPrivateStateSalt,
		Transfers:             transfers,
		PrevBalances:          prevBalances,
		AssetMerkleProofs:     assetMerkleProofs,
		PrevPrivateCommitment: prevPrivateState.Commitment(),
		NewPrivateCommitment:  newPrivateState.Commitment(),
		Tx:                    tx,
		InsufficientFlags:     insufficientFlags,
		IsValid:               isValid,
	}, nil
}

type SendWitnessResult struct {
	IsValid                 bool                             `json:"isValid"`
	LastTxHash              *intMaxTypes.PoseidonHashOut     `json:"lastTxHash"`
	LastTxInsufficientFlags backup_balance.InsufficientFlags `json:"lastTxInsufficientFlags"`
}

// get last_tx_hash and last_tx_insufficient_flags
// assuming that the tx is included in the block
// TODO: consider include validity proof verification
func (w *SendWitness) GetNextLastTx() (*SendWitnessResult, error) {
	spentValue, err := NewSpentValue(
		w.PrevPrivateState,
		w.PrevBalances,
		w.NewPrivateStateSalt,
		w.Transfers,
		w.AssetMerkleProofs,
		w.TxWitness.Tx.Nonce,
	)
	if err != nil {
		return nil, err
	}

	isValid := spentValue.IsValid
	lastTxHash := w.PrevBalancePis.LastTxHash
	lastTxInsufficientFlags := w.PrevBalancePis.LastTxInsufficientFlags
	if isValid {
		lastTxHash = spentValue.Tx.Hash()
		lastTxInsufficientFlags = spentValue.InsufficientFlags
	}

	return &SendWitnessResult{
		IsValid:                 isValid,
		LastTxHash:              lastTxHash,
		LastTxInsufficientFlags: lastTxInsufficientFlags,
	}, nil
}

type UpdateWitness struct {
	ValidityProof          string                             `json:"validityProof"`
	BlockMerkleProof       intMaxTree.BlockHashMerkleProof    `json:"blockMerkleProof"`
	AccountMembershipProof *intMaxTree.IndexedMembershipProof `json:"accountMembershipProof"`
}

type DepositWitness struct {
	DepositSalt        Salt                          `json:"depositSalt"`
	DepositIndex       uint                          `json:"depositIndex"`
	Deposit            intMaxTree.DepositLeaf        `json:"deposit"`
	DepositMerkleProof *intMaxTree.KeccakMerkleProof `json:"depositMerkleProof"`
}

type PrivateWitness struct {
	TokenIndex       uint32                            `json:"tokenIndex"`
	Amount           *big.Int                          `json:"amount"`
	Nullifier        intMaxTypes.Bytes32               `json:"nullifier"`
	NewSalt          Salt                              `json:"newSalt"`
	PrevPrivateState *PrivateState                     `json:"prevPrivateState"`
	NullifierProof   *intMaxTree.IndexedInsertionProof `json:"nullifierProof"`
	PrevAssetLeaf    *intMaxTree.AssetLeaf             `json:"prevAssetLeaf"`
	AssetMerkleProof *intMaxTree.AssetMerkleProof      `json:"assetMerkleProof"`
}

type ReceiveDepositWitness struct {
	DepositWitness *DepositWitness `json:"depositWitness"`
	PrivateWitness *PrivateWitness `json:"privateWitness"`
}

type KeccakMerkleProofInput = []string

type DepositLeafInput struct {
	RecipientSaltHash string `json:"pubkeySaltHash"`
	TokenIndex        uint32 `json:"tokenIndex"`
	Amount            string `json:"amount"`
}

type DepositWitnessInput struct {
	DepositSalt        string                  `json:"depositSalt"`
	DepositIndex       uint                    `json:"depositIndex"`
	Deposit            *DepositLeafInput       `json:"deposit"`
	DepositMerkleProof *KeccakMerkleProofInput `json:"depositMerkleProof"`
}

type IndexedMerkleProofInput = []*poseidonHashOut

type IndexedMerkleLeafInput struct {
	Key       string `json:"key"`
	Value     uint64 `json:"value"`
	NextIndex int    `json:"nextIndex"`
	NextKey   string `json:"nextKey"`
}

func (input *IndexedMerkleLeafInput) FromIndexedMerkleLeaf(value *intMaxTree.IndexedMerkleLeaf) *IndexedMerkleLeafInput {
	fmt.Printf("input: %v\n", input)
	fmt.Printf("value.Key: %v\n", value)
	input.Key = value.Key.String()
	input.Value = value.Value
	input.NextIndex = value.NextIndex
	input.NextKey = value.NextKey.String()

	return input
}

type LeafIndexInput = int

type IndexedInsertionProofInput struct {
	Index        LeafIndexInput           `json:"index"`
	LowLeafProof *IndexedMerkleProofInput `json:"lowLeafProof"`
	LeafProof    *IndexedMerkleProofInput `json:"leafProof"`
	LowLeafIndex LeafIndexInput           `json:"lowLeafIndex"`
	PrevLowLeaf  *IndexedMerkleLeafInput  `json:"prevLowLeaf"`
}

type AmountInput = string

type AssetLeafInput struct {
	IsInsufficient bool        `json:"isInsufficient"`
	Amount         AmountInput `json:"amount"`
}

type AssetMerkleProofInput = []*poseidonHashOut

type SaltInput = string

type PrivateStateInput struct {
	AssetTreeRoot     *poseidonHashOut `json:"assetTreeRoot"`
	NullifierTreeRoot *poseidonHashOut `json:"nullifierTreeRoot"`
	Nonce             uint32           `json:"nonce"`
	Salt              SaltInput        `json:"salt"`
}

type PrivateWitnessInput struct {
	TokenIndex       uint32                      `json:"tokenIndex"`
	Amount           AmountInput                 `json:"amount"`
	Nullifier        intMaxTypes.Bytes32         `json:"nullifier"`
	NewSalt          string                      `json:"newSalt"`
	PrevPrivateState *PrivateStateInput          `json:"prevPrivateState"`
	NullifierProof   *IndexedInsertionProofInput `json:"nullifierProof"`
	PrevAssetLeaf    *AssetLeafInput             `json:"prevAssetLeaf"`
	AssetMerkleProof *AssetMerkleProofInput      `json:"assetMerkleProof"`
}

func (input *PrivateWitnessInput) FromPrivateWitness(value *PrivateWitness) *PrivateWitnessInput {
	input = &PrivateWitnessInput{
		TokenIndex: value.TokenIndex,
		Amount:     value.Amount.String(),
		Nullifier:  value.Nullifier,
		NewSalt:    value.NewSalt.String(),
		// PrevPrivateState: value.PrevPrivateState,
		PrevPrivateState: &PrivateStateInput{
			AssetTreeRoot:     value.PrevPrivateState.AssetTreeRoot,
			NullifierTreeRoot: value.PrevPrivateState.NullifierTreeRoot,
			Nonce:             value.PrevPrivateState.Nonce,
			Salt:              value.PrevPrivateState.Salt.String(),
		},
		NullifierProof: &IndexedInsertionProofInput{
			Index:        value.NullifierProof.Index,
			LowLeafProof: &value.NullifierProof.LowLeafProof.Siblings,
			LeafProof:    &value.NullifierProof.LeafProof.Siblings,
			LowLeafIndex: value.NullifierProof.LowLeafIndex,
			PrevLowLeaf: &IndexedMerkleLeafInput{
				Key:       value.NullifierProof.PrevLowLeaf.Key.String(),
				Value:     value.NullifierProof.PrevLowLeaf.Value,
				NextIndex: value.NullifierProof.PrevLowLeaf.NextIndex,
				NextKey:   value.NullifierProof.PrevLowLeaf.NextKey.String(),
			},
		},
		PrevAssetLeaf: &AssetLeafInput{
			IsInsufficient: value.PrevAssetLeaf.IsInsufficient,
			Amount:         value.PrevAssetLeaf.Amount.BigInt().String(),
		},
		AssetMerkleProof: &value.AssetMerkleProof.Siblings,
	}

	return input
}

type ReceiveDepositWitnessInput struct {
	DepositWitness *DepositWitnessInput `json:"depositWitness"`
	PrivateWitness *PrivateWitnessInput `json:"privateWitness"`
}

func (input *ReceiveDepositWitnessInput) FromReceiveDepositWitness(value *ReceiveDepositWitness) *ReceiveDepositWitnessInput {
	depositMerkleProofSiblings := make([]string, 0, len(value.DepositWitness.DepositMerkleProof.Siblings))
	for _, sibling := range value.DepositWitness.DepositMerkleProof.Siblings {
		depositMerkleProofSiblings = append(depositMerkleProofSiblings, hexutil.Encode(sibling[:]))
	}
	input.DepositWitness = &DepositWitnessInput{
		DepositSalt:  value.DepositWitness.DepositSalt.String(),
		DepositIndex: value.DepositWitness.DepositIndex,
		Deposit: &DepositLeafInput{
			RecipientSaltHash: hexutil.Encode(value.DepositWitness.Deposit.RecipientSaltHash[:]),
			TokenIndex:        value.DepositWitness.Deposit.TokenIndex,
			Amount:            value.DepositWitness.Deposit.Amount.String(),
		},
		DepositMerkleProof: &depositMerkleProofSiblings,
	}

	input.PrivateWitness = new(PrivateWitnessInput).FromPrivateWitness(value.PrivateWitness)

	return input
}

type TransferWitness struct {
	Tx                  intMaxTypes.Tx          `json:"tx"`
	Transfer            intMaxTypes.Transfer    `json:"transfer"`
	TransferIndex       uint32                  `json:"transferIndex"`
	TransferMerkleProof *intMaxTree.MerkleProof `json:"transferMerkleProof"`
}

type TransferWitnessInput struct {
	Tx                  TxInput          `json:"tx"`
	Transfer            TransferInput    `json:"transfer"`
	TransferIndex       uint32           `json:"transferIndex"`
	TransferMerkleProof MerkleProofInput `json:"transferMerkleProof"`
}

func (input *TransferWitnessInput) FromTransferWitness(value *TransferWitness) *TransferWitnessInput {
	input.Tx = TxInput{
		TransferTreeRoot: value.Tx.TransferTreeRoot.String(),
		Nonce:            value.Tx.Nonce,
	}
	input.Transfer = TransferInput{
		Recipient:  hexutil.Encode(value.Transfer.Recipient.Address[:]),
		TokenIndex: value.Transfer.TokenIndex,
		Amount:     value.Transfer.Amount.String(),
		Salt:       value.Transfer.Salt.String(),
	}
	input.TransferIndex = value.TransferIndex
	input.TransferMerkleProof = make([]string, len(value.TransferMerkleProof.Siblings))
	for i := 0; i < len(value.TransferMerkleProof.Siblings); i++ {
		input.TransferMerkleProof[i] = value.TransferMerkleProof.Siblings[i].String()
	}

	return input
}

type ReceiveTransferWitness struct {
	TransferWitness  *TransferWitness                 `json:"transferWitness"`
	PrivateWitness   *PrivateWitness                  `json:"privateWitness"`
	BalanceProof     string                           `json:"balanceProof"`
	BlockMerkleProof *intMaxTree.BlockHashMerkleProof `json:"blockMerkleProof"`
}

type ReceiveTransferWitnessInput struct {
	TransferWitness  *TransferWitnessInput `json:"transferWitness"`
	PrivateWitness   *PrivateWitnessInput  `json:"privateWitness"`
	BalanceProof     string                `json:"balanceProof"`
	BlockMerkleProof MerkleProofInput      `json:"blockMerkleProof"`
}

func (input *ReceiveTransferWitnessInput) FromReceiveTransferWitness(value *ReceiveTransferWitness) *ReceiveTransferWitnessInput {
	transferMerkleProof := make([]string, len(value.TransferWitness.TransferMerkleProof.Siblings))
	for i, sibling := range value.TransferWitness.TransferMerkleProof.Siblings {
		transferMerkleProof[i] = sibling.String()
	}
	input.TransferWitness = new(TransferWitnessInput).FromTransferWitness(value.TransferWitness)
	input.PrivateWitness = new(PrivateWitnessInput).FromPrivateWitness(value.PrivateWitness)
	input.BalanceProof = value.BalanceProof
	input.BlockMerkleProof = make([]string, len(value.BlockMerkleProof.Siblings))
	for i, sibling := range value.BlockMerkleProof.Siblings {
		input.BlockMerkleProof[i] = sibling.String()
	}

	return input
}

type DepositCase struct {
	DepositSalt  Salt                   `json:"depositSalt"`
	DepositIndex uint32                 `json:"depositIndex"`
	Deposit      intMaxTree.DepositLeaf `json:"deposit"`
}

type PrivateState struct {
	AssetTreeRoot     *poseidonHashOut `json:"assetTreeRoot"`
	NullifierTreeRoot *poseidonHashOut `json:"nullifierTreeRoot"`
	Nonce             uint32           `json:"nonce"`
	Salt              Salt             `json:"salt"`
}

func (s *PrivateState) SetDefault() *PrivateState {
	zeroAsset := intMaxTree.AssetLeaf{
		IsInsufficient: false,
		Amount:         new(intMaxTypes.Uint256).FromBigInt(big.NewInt(0)),
	}
	const (
		assetTreeHeight     = 32
		nullifierTreeHeight = 32
	)

	assetTree, err := intMaxTree.NewAssetTree(assetTreeHeight, nil, zeroAsset.Hash())
	if err != nil {
		panic(err)
	}

	nullifierTree, err := intMaxTree.NewNullifierTree(nullifierTreeHeight)
	if err != nil {
		panic(err)
	}

	assetTreeRoot, _, _ := assetTree.GetCurrentRootCountAndSiblings()
	nullifierTreeRoot := nullifierTree.GetRoot()
	return &PrivateState{
		AssetTreeRoot:     &assetTreeRoot,
		NullifierTreeRoot: nullifierTreeRoot,
		Nonce:             0,
		Salt:              Salt{},
	}
}

func (s *PrivateState) ToFieldElementSlice() []ffg.Element {
	buf := make([]ffg.Element, 0, 32+32+1+32)
	buf = append(buf, s.AssetTreeRoot.Elements[:]...)
	buf = append(buf, s.NullifierTreeRoot.Elements[:]...)
	buf = append(buf, *new(ffg.Element).SetUint64(uint64(s.Nonce)))
	buf = append(buf, s.Salt.Elements[:]...)

	return buf
}

func (s *PrivateState) Commitment() *poseidonHashOut {
	return intMaxGP.HashNoPad(s.ToFieldElementSlice())
}

type BalanceValidityAuxInfo struct {
	ValidityWitness *block_validity_prover.ValidityWitness
}

type BalancePublicInputs struct {
	PubKey                  *intMaxAcc.PublicKey
	PrivateCommitment       *intMaxTypes.PoseidonHashOut
	LastTxHash              *intMaxTypes.PoseidonHashOut
	LastTxInsufficientFlags backup_balance.InsufficientFlags
	PublicState             *block_validity_prover.PublicState
}

func (s *BalancePublicInputs) FromPublicInputs(publicInputs []ffg.Element) (*BalancePublicInputs, error) {
	if len(publicInputs) < balancePublicInputsLen {
		return nil, errors.New("invalid length")
	}

	const (
		numHashOutElts                = intMaxGP.NUM_HASH_OUT_ELTS
		publicKeyOffset               = 0
		privateCommitmentOffset       = publicKeyOffset + int8Key
		lastTxHashOffset              = privateCommitmentOffset + numHashOutElts
		lastTxInsufficientFlagsOffset = lastTxHashOffset + numHashOutElts
		publicStateOffset             = lastTxInsufficientFlagsOffset + backup_balance.InsufficientFlagsLen
		end                           = publicStateOffset + block_validity_prover.PublicStateLimbSize
	)

	address := new(intMaxTypes.Uint256).FromFieldElementSlice(publicInputs[0:int8Key])
	publicKey, err := new(intMaxAcc.PublicKey).SetBigInt(address.BigInt())
	if err != nil {
		return nil, err
	}
	privateCommitment := poseidonHashOut{
		Elements: [numHashOutElts]ffg.Element{
			publicInputs[privateCommitmentOffset],
			publicInputs[privateCommitmentOffset+1],
			publicInputs[privateCommitmentOffset+int2Key],
			publicInputs[privateCommitmentOffset+int3Key],
		},
	}
	lastTxHash := poseidonHashOut{
		Elements: [numHashOutElts]ffg.Element{
			publicInputs[lastTxHashOffset],
			publicInputs[lastTxHashOffset+1],
			publicInputs[lastTxHashOffset+int2Key],
			publicInputs[lastTxHashOffset+int3Key],
		},
	}
	lastTxInsufficientFlags := new(backup_balance.InsufficientFlags).FromFieldElementSlice(
		publicInputs[lastTxInsufficientFlagsOffset:publicStateOffset],
	)
	publicState := new(block_validity_prover.PublicState).FromFieldElementSlice(
		publicInputs[publicStateOffset:end],
	)

	return &BalancePublicInputs{
		PubKey:                  publicKey,
		PrivateCommitment:       &privateCommitment,
		LastTxHash:              &lastTxHash,
		LastTxInsufficientFlags: *lastTxInsufficientFlags,
		PublicState:             publicState,
	}, nil
}
