package balance_prover_service

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	SENDER_TREE_HEIGHT     = 7
	balancePublicInputsLen = 47

	int2Key  = 2
	int3Key  = 3
	int4Key  = 4
	int8Key  = 8
	int32Key = 32
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

type GenericAddressInput struct {
	IsPublicKey bool   `json:"isPubkey"`
	Data        string `json:"data"`
}

type TransferInput struct {
	Recipient  GenericAddressInput `json:"recipient"`
	TokenIndex uint32              `json:"tokenIndex"`
	Amount     AmountInput         `json:"amount"`
	Salt       SaltInput           `json:"salt"`
}

func (input *TransferInput) FromTransfer(value *intMaxTypes.Transfer) *TransferInput {
	// input.Recipient = hexutil.Encode(value.Recipient.Address[:])
	data := new(big.Int).SetBytes(value.Recipient.Address)
	input.Recipient = GenericAddressInput{
		IsPublicKey: value.Recipient.TypeOfAddress == intMaxAccTypes.INTMAXAddressType,
		Data:        data.String(),
	}
	input.TokenIndex = value.TokenIndex
	input.Amount = value.Amount.String()
	input.Salt = *value.Salt

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
	input.DepositTreeRoot = value.DepositTreeRoot.Hex()
	input.BlockHash = value.BlockHash.Hex()
	input.BlockNumber = value.BlockNumber

	return input
}

func (input *PublicStateInput) PublicState() *block_validity_prover.PublicState {
	return &block_validity_prover.PublicState{
		BlockTreeRoot:       input.BlockTreeRoot,
		PrevAccountTreeRoot: input.PrevAccountTreeRoot,
		AccountTreeRoot:     input.AccountTreeRoot,
		DepositTreeRoot:     common.HexToHash(input.DepositTreeRoot),
		BlockHash:           common.HexToHash(input.BlockHash),
		BlockNumber:         input.BlockNumber,
	}
}

type ValidityPublicInputsInput struct {
	PublicState    *PublicStateInput `json:"publicState"`
	TxTreeRoot     string            `json:"txTreeRoot"`
	SenderTreeRoot poseidonHashOut   `json:"senderTreeRoot"`
	IsValidBlock   bool              `json:"isValidBlock"`
}

func (input *ValidityPublicInputsInput) FromValidityPublicInputs(value *block_validity_prover.ValidityPublicInputs) *ValidityPublicInputsInput {
	input.PublicState = new(PublicStateInput).FromPublicState(value.PublicState)
	input.TxTreeRoot = hexutil.Encode(value.TxTreeRoot.Bytes())
	input.SenderTreeRoot = *value.SenderTreeRoot
	input.IsValidBlock = value.IsValidBlock

	return input
}

type TxInput struct {
	TransferTreeRoot poseidonHashOut `json:"transferTreeRoot"`
	Nonce            uint32          `json:"nonce"`
}

type TxWitnessInput struct {
	ValidityPis   ValidityPublicInputsInput `json:"validityPis"`
	SenderLeaves  []*SenderLeafInput        `json:"senderLeaves"`
	Tx            TxInput                   `json:"tx"`
	TxIndex       uint32                    `json:"txIndex"`
	TxMerkleProof []*poseidonHashOut        `json:"txMerkleProof"`
}

func (input *TxWitnessInput) FromTxWitness(value *TxWitness) *TxWitnessInput {
	input.ValidityPis = *new(ValidityPublicInputsInput).FromValidityPublicInputs(&value.ValidityPis)
	input.SenderLeaves = make([]*SenderLeafInput, len(value.SenderLeaves))
	for i, leaf := range value.SenderLeaves {
		input.SenderLeaves[i] = &SenderLeafInput{
			Sender:  leaf.Sender.BigInt().String(),
			IsValid: leaf.IsValid,
		}
	}

	input.Tx = TxInput{
		TransferTreeRoot: *value.Tx.TransferTreeRoot,
		Nonce:            value.Tx.Nonce,
	}
	input.TxIndex = value.TxIndex
	input.TxMerkleProof = make([]*poseidonHashOut, len(value.TxMerkleProof))
	copy(input.TxMerkleProof, value.TxMerkleProof)

	return input
}

type InsufficientFlagsInput struct {
	Limbs [backup_balance.InsufficientFlagsLen]uint32 `json:"limbs"`
}

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
	// input.LastTxInsufficientFlags = hexutil.Encode(value.LastTxInsufficientFlags.Bytes())
	input.LastTxInsufficientFlags.Limbs = value.LastTxInsufficientFlags.Limbs
	input.PublicState = new(PublicStateInput).FromPublicState(value.PublicState)

	return input
}

// Information needed to prove that a balance has been sent
type SendWitness struct {
	PrevBalancePis      *BalancePublicInputs             `json:"prevBalancePis"`
	PrevPrivateState    *PrivateState                    `json:"prevPrivateState"`
	PrevBalances        []*intMaxTree.AssetLeafEntry     `json:"prevBalances"`
	AssetMerkleProofs   []*intMaxTree.AssetMerkleProof   `json:"assetMerkleProofs"`
	InsufficientFlags   backup_balance.InsufficientFlags `json:"insufficientFlags"`
	Transfers           []*intMaxTypes.Transfer          `json:"transfers"`
	TxWitness           TxWitness                        `json:"txWitness"`
	NewPrivateStateSalt Salt                             `json:"newPrivateStateSalt"`
}

type SendWitnessInput struct {
	PrevBalancePis      *BalancePublicInputsInput `json:"prevBalancePis"`
	PrevPrivateState    *PrivateStateInput        `json:"prevPrivateState"`
	PrevBalances        []*AssetLeafEntryInput    `json:"prevBalances"`
	AssetMerkleProofs   []AssetMerkleProofInput   `json:"assetMerkleProofs"`
	InsufficientFlags   InsufficientFlagsInput    `json:"insufficientFlags"`
	Transfers           []*TransferInput          `json:"transfers"`
	TxWitness           *TxWitnessInput           `json:"txWitness"`
	NewPrivateStateSalt SaltInput                 `json:"newPrivateStateSalt"`
}

func (input *SendWitnessInput) FromSendWitness(value *SendWitness) *SendWitnessInput {
	input.PrevBalancePis = new(BalancePublicInputsInput).FromBalancePublicInputs(value.PrevBalancePis)
	input.PrevPrivateState = &PrivateStateInput{
		AssetTreeRoot:     value.PrevPrivateState.AssetTreeRoot,
		NullifierTreeRoot: value.PrevPrivateState.NullifierTreeRoot,
		Nonce:             value.PrevPrivateState.TransactionCount,
		Salt:              value.PrevPrivateState.Salt,
	}
	input.PrevBalances = make([]*AssetLeafEntryInput, len(value.PrevBalances))
	for i, balance := range value.PrevBalances {
		input.PrevBalances[i] = &AssetLeafEntryInput{
			TokenIndex: balance.TokenIndex,
			Leaf: &AssetLeafInput{
				IsInsufficient: balance.Leaf.IsInsufficient,
				Amount:         balance.Leaf.Amount.BigInt().String(),
			},
		}
	}
	input.AssetMerkleProofs = make([]AssetMerkleProofInput, len(value.AssetMerkleProofs))
	for i, proof := range value.AssetMerkleProofs {
		input.AssetMerkleProofs[i] = make([]*poseidonHashOut, len(proof.Siblings))
		copy(input.AssetMerkleProofs[i], proof.Siblings)
	}
	// input.InsufficientFlags = hexutil.Encode(value.InsufficientFlags.Bytes())
	input.InsufficientFlags.Limbs = value.InsufficientFlags.Limbs
	input.Transfers = make([]*TransferInput, len(value.Transfers))
	for i, transfer := range value.Transfers {
		input.Transfers[i] = new(TransferInput).FromTransfer(transfer)
	}
	input.TxWitness = new(TxWitnessInput).FromTxWitness(&value.TxWitness)
	input.NewPrivateStateSalt = value.NewPrivateStateSalt

	return input
}

func (w *SendWitness) GetIncludedBlockNumber() uint32 {
	return w.TxWitness.ValidityPis.PublicState.BlockNumber
}

func (w *SendWitness) GetPrevBalancePisBlockNumber() uint32 {
	return w.PrevBalancePis.PublicState.BlockNumber
}

type Salt = poseidonHashOut

type SpentValue struct {
	PrevPrivateState      *PrivateState                    `json:"prevPrivateState"`
	NewPrivateStateSalt   Salt                             `json:"newPrivateStateSalt"`
	Transfers             []*intMaxTypes.Transfer          `json:"transfers"`
	PrevBalances          []*intMaxTree.AssetLeafEntry     `json:"prevBalances"`
	AssetMerkleProofs     []*intMaxTree.AssetMerkleProof   `json:"assetMerkleProofs"`
	PrevPrivateCommitment *poseidonHashOut                 `json:"prevPrivateCommitment"`
	NewPrivateCommitment  *poseidonHashOut                 `json:"newPrivateCommitment"`
	Tx                    intMaxTypes.Tx                   `json:"tx"`
	InsufficientFlags     backup_balance.InsufficientFlags `json:"insufficientFlags"`
	IsValid               bool                             `json:"isValid"`
}

func NewSpentValue(
	prevPrivateState *PrivateState,
	prevBalances []*intMaxTree.AssetLeafEntry,
	newPrivateStateSalt Salt,
	transfers []*intMaxTypes.Transfer,
	assetMerkleProofs []*intMaxTree.AssetMerkleProof,
	txNonce uint64,
) (*SpentValue, error) {
	const numTransfers = 1 << intMaxTree.TRANSFER_TREE_HEIGHT
	if len(prevBalances) != numTransfers {
		return nil, errors.New("prevBalances length is not equal to numTransfers")
	}
	if len(transfers) != numTransfers {
		return nil, errors.New("transfers length is not equal to numTransfers")
	}
	if len(assetMerkleProofs) != numTransfers {
		return nil, errors.New("assetMerkleProofs length is not equal to numTransfers")
	}

	prevBalancesMap := make(map[uint32]*intMaxTree.AssetLeaf)
	for _, balance := range prevBalances {
		prevBalancesMap[balance.TokenIndex] = balance.Leaf
	}

	insufficientFlags := backup_balance.InsufficientFlags{}
	assetTreeRoot := prevPrivateState.AssetTreeRoot
	for i := 0; i < numTransfers; i++ {
		transfer := transfers[i]
		proof := assetMerkleProofs[i]
		prevBalance, ok := prevBalancesMap[transfer.TokenIndex]
		if !ok {
			prevBalance = &intMaxTree.AssetLeaf{
				IsInsufficient: false,
				Amount:         new(intMaxTypes.Uint256),
			}
		}

		// if err := proof.Verify(prevBalance, transfer.TokenIndex, assetTreeRoot); err != nil {
		// 	return nil, err
		// }
		newBalance := prevBalance.Sub(transfer.Amount)
		assetTreeRoot = proof.GetRoot(newBalance.Hash(), int(transfer.TokenIndex))
		insufficientFlags.SetBit(i, newBalance.IsInsufficient)
	}

	insufficientFlags = backup_balance.InsufficientFlags(insufficientFlags)
	newPrivateState := PrivateState{
		AssetTreeRoot:     assetTreeRoot,
		TransactionCount:  prevPrivateState.TransactionCount + 1,
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
		Nonce:            uint32(txNonce),
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
		IsValid:               uint32(txNonce) == prevPrivateState.TransactionCount,
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
		uint64(w.TxWitness.Tx.Nonce),
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

type DepositWitness struct {
	DepositSalt        Salt                          `json:"depositSalt"`
	DepositIndex       uint                          `json:"depositIndex"`
	Deposit            intMaxTree.DepositLeaf        `json:"deposit"`
	DepositRoot        common.Hash                   `json:"-"`
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

type AssetLeafEntryInput struct {
	TokenIndex uint32          `json:"tokenIndex"`
	Leaf       *AssetLeafInput `json:"assetLeaf"`
}

type AssetMerkleProofInput = []*poseidonHashOut

type SaltInput = poseidonHashOut

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
			Nonce:             value.PrevPrivateState.TransactionCount,
			Salt:              value.PrevPrivateState.Salt,
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

type TransferWitnessInput struct {
	Tx                  TxInput          `json:"tx"`
	Transfer            TransferInput    `json:"transfer"`
	TransferIndex       uint32           `json:"transferIndex"`
	TransferMerkleProof MerkleProofInput `json:"transferMerkleProof"`
}

func (input *TransferWitnessInput) FromTransferWitness(value *intMaxTypes.TransferWitness) *TransferWitnessInput {
	input.Tx = TxInput{
		TransferTreeRoot: *value.Tx.TransferTreeRoot,
		Nonce:            value.Tx.Nonce,
	}
	input.Transfer = *new(TransferInput).FromTransfer(&value.Transfer)
	input.TransferIndex = value.TransferIndex
	input.TransferMerkleProof = make([]string, len(value.TransferMerkleProof))
	for i := 0; i < len(value.TransferMerkleProof); i++ {
		input.TransferMerkleProof[i] = value.TransferMerkleProof[i].String()
	}

	return input
}

type ReceiveTransferWitness struct {
	TransferWitness        *intMaxTypes.TransferWitness     `json:"transferWitness"`
	PrivateWitness         *PrivateWitness                  `json:"privateWitness"`
	LastBalanceProof       string                           `json:"lastBalanceProof"`
	BalanceTransitionProof string                           `json:"balanceTransitionProof"`
	BlockMerkleProof       *intMaxTree.BlockHashMerkleProof `json:"blockMerkleProof"`
}

type ReceiveTransferWitnessInput struct {
	TransferWitness        *TransferWitnessInput `json:"transferWitness"`
	PrivateWitness         *PrivateWitnessInput  `json:"privateWitness"`
	LastBalanceProof       string                `json:"lastBalanceProof"`
	BalanceTransitionProof string                `json:"balanceTransitionProof"`
	BlockMerkleProof       MerkleProofInput      `json:"blockMerkleProof"`
}

func (input *ReceiveTransferWitnessInput) FromReceiveTransferWitness(value *ReceiveTransferWitness) *ReceiveTransferWitnessInput {
	transferMerkleProof := make([]string, len(value.TransferWitness.TransferMerkleProof))
	for i, sibling := range value.TransferWitness.TransferMerkleProof {
		transferMerkleProof[i] = sibling.String()
	}
	input.TransferWitness = new(TransferWitnessInput).FromTransferWitness(value.TransferWitness)
	input.PrivateWitness = new(PrivateWitnessInput).FromPrivateWitness(value.PrivateWitness)
	input.LastBalanceProof = value.LastBalanceProof
	input.BalanceTransitionProof = value.BalanceTransitionProof
	input.BlockMerkleProof = make([]string, len(value.BlockMerkleProof.Siblings))
	for i, sibling := range value.BlockMerkleProof.Siblings {
		input.BlockMerkleProof[i] = sibling.String()
	}

	return input
}

type DepositCase struct {
	DepositSalt  Salt                   `json:"depositSalt"`
	DepositID    uint32                 `json:"depositId"`
	DepositIndex uint32                 `json:"depositIndex"`
	Deposit      intMaxTree.DepositLeaf `json:"deposit"`
}

type PrivateState struct {
	AssetTreeRoot     *poseidonHashOut `json:"assetTreeRoot"`
	NullifierTreeRoot *poseidonHashOut `json:"nullifierTreeRoot"`
	TransactionCount  uint32           `json:"nonce"`
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

	assetTreeRoot := assetTree.GetRoot()
	nullifierTreeRoot := nullifierTree.GetRoot()
	return &PrivateState{
		AssetTreeRoot:     assetTreeRoot,
		NullifierTreeRoot: nullifierTreeRoot,
		TransactionCount:  0,
		Salt:              Salt{},
	}
}

func (s *PrivateState) ToFieldElementSlice() []ffg.Element {
	const numPrivateStateElements = int32Key + int32Key + 1 + int32Key
	buf := make([]ffg.Element, 0, numPrivateStateElements)
	buf = append(buf, s.AssetTreeRoot.Elements[:]...)
	buf = append(buf, s.NullifierTreeRoot.Elements[:]...)
	buf = append(buf, *new(ffg.Element).SetUint64(uint64(s.TransactionCount)))
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

func NewBalancePublicInputsWithPublicKey(publicKey *intMaxAcc.PublicKey) *BalancePublicInputs {
	// privateCommitment := new(intMaxTypes.PoseidonHashOut).SetZero()
	privateCommitment := new(PrivateState).SetDefault().Commitment()
	lastTxHash := new(intMaxTypes.PoseidonHashOut).SetZero()
	lastTxInsufficientFlags := backup_balance.InsufficientFlags{}
	publicState := new(block_validity_prover.PublicState).Genesis()

	pis := new(BalancePublicInputs)
	pis.PubKey = publicKey
	pis.PrivateCommitment = privateCommitment
	pis.LastTxHash = lastTxHash
	pis.LastTxInsufficientFlags = lastTxInsufficientFlags
	pis.PublicState = publicState

	return pis
}

const (
	numHashOutElts                = intMaxGP.NUM_HASH_OUT_ELTS
	publicKeyOffset               = 0
	privateCommitmentOffset       = publicKeyOffset + int8Key
	lastTxHashOffset              = privateCommitmentOffset + numHashOutElts
	lastTxInsufficientFlagsOffset = lastTxHashOffset + numHashOutElts
	publicStateOffset             = lastTxInsufficientFlagsOffset + backup_balance.InsufficientFlagsLen
	sizeOfBalancePublicInputs     = publicStateOffset + block_validity_prover.PublicStateLimbSize
)

func (s *BalancePublicInputs) FromPublicInputs(publicInputs []ffg.Element) (*BalancePublicInputs, error) {
	if len(publicInputs) < balancePublicInputsLen {
		return nil, errors.New("invalid length")
	}

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
		publicInputs[publicStateOffset:sizeOfBalancePublicInputs],
	)

	return &BalancePublicInputs{
		PubKey:                  publicKey,
		PrivateCommitment:       &privateCommitment,
		LastTxHash:              &lastTxHash,
		LastTxInsufficientFlags: *lastTxInsufficientFlags,
		PublicState:             publicState,
	}, nil
}

func ValidateTxInclusionValue(
	publicKey *intMaxAcc.PublicKey,
	prevPublicState *block_validity_prover.PublicState,
	validityProof string,
	blockMerkleProof *intMaxTree.BlockHashMerkleProof,
	prevAccountMembershipProof *intMaxTree.IndexedMembershipProof,
	senderIndex uint32,
	tx intMaxTypes.Tx,
	txMerkleProof *intMaxTree.PoseidonMerkleProof,
	// senderLeaf *intMaxTree.SenderLeaf,
	// senderMerkleProof *intMaxTree.MerkleProof,
	// newPublicState             *block_validity_prover.PublicState,
	// isValid                    bool,
) (bool, error) {
	// let validity_pis = ValidityPublicInputs::from_u64_slice(
	// 	&validity_proof.public_inputs[0..VALIDITY_PUBLIC_INPUTS_LEN].to_u64_vec(),
	// );
	// block_merkle_proof
	// 	.verify(
	// 		&prev_public_state.block_hash,
	// 		prev_public_state.block_number as usize,
	// 		validity_pis.public_state.block_tree_root,
	// 	)
	// 	.expect("block merkle proof is invalid");
	// prev_account_membership_proof
	// 	.verify(pubkey, validity_pis.public_state.prev_account_tree_root)
	// 	.expect("prev account membership proof is invalid");
	// let last_block_number = prev_account_membership_proof.get_value() as u32;
	// assert!(last_block_number <= prev_public_state.block_number); // no send tx till one before the last block

	// let tx_tree_root: PoseidonHashOut = validity_pis
	// 	.tx_tree_root
	// 	.try_into()
	// 	.expect("tx tree root is invalid");
	// tx_merkle_proof
	// 	.verify(tx, sender_index, tx_tree_root)
	// 	.expect("tx merkle proof is invalid");
	// sender_merkle_proof
	// 	.verify(sender_leaf, sender_index, validity_pis.sender_tree_root)
	// 	.expect("sender merkle proof is invalid");

	// assert_eq!(sender_leaf.sender, pubkey);
	// let is_valid = sender_leaf.is_valid && validity_pis.is_valid_block;

	validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(validityProof)
	if err != nil {
		return false, err
	}
	validityPis := new(block_validity_prover.ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)
	err = blockMerkleProof.Verify(
		intMaxTree.NewBlockHashLeaf(prevPublicState.BlockHash).Hash(),
		int(prevPublicState.BlockNumber),
		validityPis.PublicState.BlockTreeRoot,
	)
	if err != nil {
		// block merkle proof is invalid
		return false, err
	}

	err = prevAccountMembershipProof.Verify(publicKey.BigInt(), prevPublicState.PrevAccountTreeRoot)
	if err != nil {
		// prev account membership proof is invalid
		fmt.Printf("publicKey: %v\n", publicKey.BigInt())
		fmt.Printf("prevPublicState.PrevAccountTreeRoot: %s\n", prevPublicState.PrevAccountTreeRoot.String())
		for i, sibling := range prevAccountMembershipProof.LeafProof.Siblings {
			fmt.Printf("sibling[%d]: %v\n", i, sibling)
		}
		var ErrInvalidMembershipProof = errors.New("prev account membership proof is invalid")
		return false, errors.Join(ErrInvalidMembershipProof, err)
	}

	lastBlockNumber := prevAccountMembershipProof.GetLeaf()
	if lastBlockNumber > uint64(prevPublicState.BlockNumber) {
		return false, errors.New("no send tx till one before the last block")
	}

	txTreeRoot := validityPis.TxTreeRoot
	err = txMerkleProof.Verify(tx.Hash(), int(senderIndex), txTreeRoot.PoseidonHashOut())
	if err != nil {
		// tx merkle proof is invalid
		return false, err
	}

	// err = senderMerkleProof.Verify(senderLeaf.Hash(), int(senderIndex), validityPis.SenderTreeRoot)
	// if err != nil {
	// 	// sender merkle proof is invalid
	// 	return false, err
	// }

	// if senderLeaf.Sender.BigInt().Cmp(publicKey.BigInt()) != 0 {
	// 	return false, errors.New("sender leaf sender is not equal to pubkey")
	// }

	// isValid := senderLeaf.IsValid && validityPis.IsValidBlock

	// return isValid, nil

	return true, nil
}
