package balance_prover_service

import (
	"errors"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"
)

const SENDER_TREE_HEIGHT = 7

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

func (w *SendWitness) GetIncludedBlockNumber() uint32 {
	return w.TxWitness.ValidityPis.PublicState.BlockNumber
}

func (w *SendWitness) GetPrevBlockNumber() uint32 {
	return w.PrevBalancePis.PublicState.BlockNumber
}

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
	ValidityProof          intMaxTypes.Plonky2Proof          `json:"validityProof"`
	BlockMerkleProof       intMaxTree.BlockHashMerkleProof   `json:"blockMerkleProof"`
	AccountMembershipProof intMaxTree.IndexedMembershipProof `json:"accountMembershipProof"`
}

type DepositWitness struct {
	DepositSalt        Salt                   `json:"depositSalt"`
	DepositIndex       uint                   `json:"depositIndex"`
	Deposit            intMaxTree.DepositLeaf `json:"deposit"`
	DepositMerkleProof intMaxTree.MerkleProof `json:"depositMerkleProof"`
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

type TransferWitness struct {
	Tx                  intMaxTypes.Tx          `json:"tx"`
	Transfer            intMaxTypes.Transfer    `json:"transfer"`
	TransferIndex       uint32                  `json:"transferIndex"`
	TransferMerkleProof *intMaxTree.MerkleProof `json:"transferMerkleProof"`
}

type ReceiveTransferWitness struct {
	TransferWitness  *TransferWitness                 `json:"transferWitness"`
	PrivateWitness   *PrivateWitness                  `json:"privateWitness"`
	BalanceProof     *intMaxTypes.Plonky2Proof        `json:"balanceProof"`
	BlockMerkleProof *intMaxTree.BlockHashMerkleProof `json:"blockMerkleProof"`
}
