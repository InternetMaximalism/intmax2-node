package balance_synchronizer

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/finite_field"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const numTransfersInTx = 1 << intMaxTree.TRANSFER_TREE_HEIGHT

type poseidonHashOut = intMaxGP.PoseidonHashOut

type mockWallet struct {
	privateKey    intMaxAcc.PrivateKey
	assetTree     intMaxTree.AssetTree
	nullifierTree intMaxTree.NullifierTree
	nonce         uint32
	salt          balance_prover_service.Salt
	publicState   *block_validity_prover.PublicState

	// cache
	sendWitnesses     map[uint32]*balance_prover_service.SendWitness
	depositCases      map[uint32]*balance_prover_service.DepositCase // depositIndex => DepositCase
	transferWitnesses map[uint32][]*intMaxTypes.TransferWitness
}

type UserState interface {
	AddDepositCase(depositIndex uint32, depositCase *balance_prover_service.DepositCase) error
	UpdateOnSendTx(
		salt balance_prover_service.Salt,
		txWitness *balance_prover_service.TxWitness,
		transferWitnesses []*intMaxTypes.TransferWitness,
	) (*balance_prover_service.SendWitness, error)
	PublicKey() *intMaxAcc.PublicKey
	Nonce() uint32
	Salt() balance_prover_service.Salt
	GenericAddress() (*intMaxTypes.GenericAddress, error)
	PrivateState() *balance_prover_service.PrivateState
	PublicState() *block_validity_prover.PublicState
	Nullifiers() []intMaxTypes.Bytes32
	IsIncludedInNullifierTree(nullifier intMaxTypes.Bytes32) (bool, error)
	AssetLeaves() map[uint32]*intMaxTree.AssetLeaf
	// Returns all block numbers in which the user has made transfers.
	GetAllBlockNumbers() []uint32
	DecryptBalanceData(encryptedBalanceData string) (*block_synchronizer.BalanceData, error)
	UpdatePublicState(publicState *block_validity_prover.PublicState)
	GetLastSendWitness() *balance_prover_service.SendWitness
	GetBalancePublicInputs() (*balance_prover_service.BalancePublicInputs, error)
	GetSendWitness(blockNumber uint32) (*balance_prover_service.SendWitness, error)
	GeneratePrivateWitness(
		newSalt balance_prover_service.Salt,
		tokenIndex uint32,
		amount *big.Int,
		nullifier intMaxTypes.Bytes32,
	) (*balance_prover_service.PrivateWitness, error)
	ReceiveDepositAndUpdate(
		blockValidityService block_validity_prover.BlockValidityService,
		depositIndex uint32,
	) (*balance_prover_service.ReceiveDepositWitness, error)
	ReceiveTransferAndUpdate(
		blockValidityService block_validity_prover.BlockValidityService,
		lastBlockNumber uint32,
		transferWitness *intMaxTypes.TransferWitness,
		senderLastBalanceProof string,
		senderBalanceTransitionProof string,
	) (*balance_prover_service.ReceiveTransferWitness, error)
	PrivateKey() *intMaxAcc.PrivateKey
}

func (w *mockWallet) AddDepositCase(depositIndex uint32, depositCase *balance_prover_service.DepositCase) error {
	w.depositCases[depositIndex] = depositCase
	return nil
}

// TODO: refactor this function
func NewBlockContentFromTxRequests(isRegistrationBlock bool, txs []*block_validity_prover.MockTxRequest) (*intMaxTypes.BlockContent, error) {
	const numOfSenders = 128
	if len(txs) > numOfSenders {
		panic("too many txs")
	}

	// sort and pad txs
	sortedTxs := make([]*block_validity_prover.MockTxRequest, len(txs))
	copy(sortedTxs, txs)
	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[j].Sender.PublicKey.BigInt().Cmp(sortedTxs[i].Sender.PublicKey.BigInt()) == 1
	})

	publicKeys := make([]*intMaxAcc.PublicKey, len(sortedTxs))
	for i, tx := range sortedTxs {
		publicKeys[i] = tx.Sender.Public()
	}

	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, dummyPublicKey)
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	txTree, err := intMaxTree.NewTxTree(uint8(intMaxTree.TX_TREE_HEIGHT), nil, zeroTx.Hash())
	if err != nil {
		panic(err)
	}

	for _, tx := range txs {
		_, index, _ := txTree.GetCurrentRootCountAndSiblings()
		_, err = txTree.AddLeaf(index, tx.Tx)
		if err != nil {
			panic(err)
		}
	}

	txTreeRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()

	flattenTxTreeRoot := finite_field.BytesToFieldElementSlice(txTreeRoot.Marshal())

	addresses := make([]intMaxTypes.Uint256, len(publicKeys))
	for i, publicKey := range publicKeys {
		addresses[i] = *new(intMaxTypes.Uint256).FromBigInt(publicKey.BigInt())
	}
	publicKeysHash := block_validity_prover.GetPublicKeysHash(addresses)

	signatures := make([]*bn254.G2Affine, len(sortedTxs))
	for i, keyPair := range sortedTxs {
		var signature *bn254.G2Affine
		signature, err = keyPair.Sender.WeightByHash(publicKeysHash.Bytes()).Sign(flattenTxTreeRoot)
		if err != nil {
			return nil, err
		}
		signatures[i] = signature
	}

	encodedSignatures := make([]string, len(sortedTxs))
	for i, signature := range signatures {
		encodedSignatures[i] = hexutil.Encode(signature.Marshal())
	}

	var blockContent *intMaxTypes.BlockContent
	blockContent, err = block_post_service.MakeRegistrationBlock(txTreeRoot, publicKeys, encodedSignatures)
	if err != nil {
		return nil, err
	}

	return blockContent, nil
}

// NOTE: This function is used for testing only
func (w *mockWallet) SendTx(
	blockValidityProver *block_validity_prover.BlockValidityProverMemory,
	transfers []*intMaxTypes.Transfer,
) (*balance_prover_service.TxWitness, []*intMaxTypes.TransferWitness, error) {
	fmt.Printf("-----SendTx-----")
	if len(transfers) >= numTransfersInTx {
		return nil, nil, errors.New("transfers length must be less than numTransfersInTx")
	}
	for len(transfers) < numTransfersInTx {
		transfers = append(transfers, new(intMaxTypes.Transfer).SetZero())
	}
	fmt.Printf("SendTx transfers: %v\n", transfers)

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, nil, zeroTransfer.Hash())
	if err != nil {
		return nil, nil, err
	}

	for _, transfer := range transfers {
		_, index, _ := transferTree.GetCurrentRootCountAndSiblings()
		_, err = transferTree.AddLeaf(index, transfer)
		if err != nil {
			return nil, nil, err
		}
	}

	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	tx := intMaxTypes.Tx{
		TransferTreeRoot: &transferTreeRoot,
		Nonce:            w.nonce,
	}

	txRequest0 := block_validity_prover.MockTxRequest{
		Tx:                  &tx,
		Sender:              &w.privateKey,
		AccountID:           2, // XXX: Use correct account ID
		WillReturnSignature: true,
	}
	txRequests := []*block_validity_prover.MockTxRequest{&txRequest0}
	blockContent, err := NewBlockContentFromTxRequests(true, txRequests)
	if err != nil {
		return nil, nil, err
	}

	lastGeneratedProofBlockNumber, err := blockValidityProver.BlockBuilder().LastGeneratedProofBlockNumber() // XXX: Is this correct block number?
	if err != nil {
		return nil, nil, err
	}
	lastValidityWitness, err := blockValidityProver.BlockBuilder().ValidityWitnessByBlockNumber(lastGeneratedProofBlockNumber)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("IMPORTANT PostBlock")
	validityWitness, err := blockValidityProver.UpdateValidityWitness(
		blockContent,
		lastValidityWitness,
	)
	if err != nil {
		return nil, nil, err
	}

	txLeaves := make([]*intMaxTypes.Tx, len(txRequests))
	for i, tx := range txRequests {
		txLeaves[i] = tx.Tx
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	txTree, err := intMaxTree.NewTxTree(intMaxTree.TX_TREE_HEIGHT, txLeaves, zeroTx.Hash())
	if err != nil {
		return nil, nil, err
	}

	txIndex := uint32(0)
	txMerkleProof, _, err := txTree.ComputeMerkleProof(uint64(txIndex))
	if err != nil {
		return nil, nil, err
	}

	senderWitness := make([]*intMaxTree.SenderLeaf, 0)
	for _, sender := range validityWitness.ValidityTransitionWitness.SenderLeaves {
		senderWitness = append(senderWitness, &intMaxTree.SenderLeaf{
			Sender:  new(intMaxTypes.Uint256).FromBigInt(sender.Sender),
			IsValid: sender.IsValid,
		})
	}

	txWitness := &balance_prover_service.TxWitness{
		ValidityPis:   *validityWitness.ValidityPublicInputs(),
		SenderLeaves:  senderWitness,
		Tx:            tx,
		TxIndex:       txIndex,
		TxMerkleProof: txMerkleProof,
	}

	transferWitnesses := make([]*intMaxTypes.TransferWitness, len(transfers))
	for transferIndex, transfer := range transfers {
		transferMerkleProof, _, _ := transferTree.ComputeMerkleProof(uint64(transferIndex))
		transferWitness := intMaxTypes.TransferWitness{
			Tx:                  tx,
			Transfer:            *transfer,
			TransferIndex:       uint32(transferIndex),
			TransferMerkleProof: transferMerkleProof,
		}
		fmt.Printf("transferWitnesses[%d]: %v\n", transferIndex, transferWitness)
		transferWitnesses[transferIndex] = &transferWitness
	}

	return txWitness, transferWitnesses, nil
}

func MakeTxWitness(
	blockValidityService block_validity_prover.BlockValidityService,
	txDetails *intMaxTypes.TxDetails,
) (*balance_prover_service.TxWitness, []*intMaxTypes.TransferWitness, error) {
	transfers := txDetails.Transfers
	if len(transfers) >= numTransfersInTx {
		return nil, nil, errors.New("transfers length must be less than numTransfersInTx")
	}
	for len(transfers) < numTransfersInTx {
		transfers = append(transfers, new(intMaxTypes.Transfer).SetZero())
	}
	fmt.Printf("MakeTxWitness transfers: %v\n", transfers)

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, nil, zeroTransfer.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create transfer tree: %w", err)
	}

	for _, transfer := range transfers {
		_, index, _ := transferTree.GetCurrentRootCountAndSiblings()
		_, err = transferTree.AddLeaf(index, transfer)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to add leaf to transfer tree: %w", err)
		}
	}

	// transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	// tx := intMaxTypes.Tx{
	// 	TransferTreeRoot: &transferTreeRoot,
	// 	Nonce:            w.nonce,
	// }
	tx := txDetails.Tx

	// txRequest0 := block_validity_prover.MockTxRequest{
	// 	Tx:                  &tx,
	// 	Sender:              &w.privateKey,
	// 	AccountID:           2,
	// 	WillReturnSignature: true,
	// }
	// txRequests := []*block_validity_prover.MockTxRequest{&txRequest0}
	// blockContent, err := NewBlockContentFromTxRequests(true, txRequests)
	// if err != nil {
	// 	return nil, nil, err
	// }

	validityPublicInputs, senderLeaves, err := blockValidityService.ValidityPublicInputs(common.HexToHash(txDetails.TxTreeRoot.String()))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get validity public inputs: %w", err)
	}
	// validityWitness, err := blockValidityService.UpdateValidityWitness(
	// 	blockContent,
	// )

	// txLeaves := make([]*intMaxTypes.Tx, len(txRequests))
	// for i, tx := range txRequests {
	// 	txLeaves[i] = tx.Tx
	// }

	// zeroTx := new(intMaxTypes.Tx).SetZero()
	// txTree, err := intMaxTree.NewTxTree(7, txLeaves, zeroTx.Hash())
	// if err != nil {
	// 	return nil, nil, err
	// }

	// txIndex := uint32(0)
	// txMerkleProof, _, err := txTree.ComputeMerkleProof(uint64(txIndex))
	// if err != nil {
	// 	return nil, nil, err
	// }

	senderWitness := make([]*intMaxTree.SenderLeaf, 0)
	for _, sender := range senderLeaves {
		senderWitness = append(senderWitness, &intMaxTree.SenderLeaf{
			Sender:  new(intMaxTypes.Uint256).FromBigInt(sender.Sender),
			IsValid: sender.IsValid,
		})
	}

	txWitness := &balance_prover_service.TxWitness{
		ValidityPis:   *validityPublicInputs, // *validityWitness.ValidityPublicInputs(),
		SenderLeaves:  senderWitness,
		Tx:            tx,
		TxIndex:       txDetails.TxIndex,
		TxMerkleProof: txDetails.TxMerkleProof,
	}

	transferWitnesses := make([]*intMaxTypes.TransferWitness, len(transfers))
	for transferIndex, transfer := range transfers {
		transferMerkleProof, _, _ := transferTree.ComputeMerkleProof(uint64(transferIndex))
		transferWitness := intMaxTypes.TransferWitness{
			Tx:                  tx,
			Transfer:            *transfer,
			TransferIndex:       uint32(transferIndex),
			TransferMerkleProof: transferMerkleProof,
		}
		fmt.Printf("transferWitnesses[%d]: %v\n", transferIndex, transferWitness)
		transferWitnesses[transferIndex] = &transferWitness
	}

	return txWitness, transferWitnesses, nil
}

func (wallet *mockWallet) CalculateSpentTokenWitness(
	newPrivateStateSalt balance_prover_service.Salt,
	tx *intMaxTypes.Tx,
	transfers []*intMaxTypes.Transfer,
) (*balance_prover_service.SpentTokenWitness, error) {
	prevPrivateState := wallet.PrivateState()

	if tx.Nonce != wallet.nonce {
		fmt.Printf("transaction nonce mismatch: %d != %d", tx.Nonce, wallet.nonce)
		var ErrTransactionNonceMismatch = errors.New("transaction nonce mismatch")
		return nil, ErrTransactionNonceMismatch
	}

	assetTree := new(intMaxTree.AssetTree).Set(&wallet.assetTree)

	assetMerkleProofs := make([]*intMaxTree.AssetMerkleProof, 0, len(transfers))
	prevBalances := make([]*intMaxTree.AssetLeafEntry, 0, len(transfers))
	insufficientFlags := new(backup_balance.InsufficientFlags)
	for i, transfer := range transfers {
		if transfer == nil {
			return nil, fmt.Errorf("transferWitness[%d] is nil", i)
		}

		tokenIndex := transfer.TokenIndex
		prevBalance := assetTree.GetLeaf(tokenIndex)
		assetMerkleProof, _, _ := assetTree.Prove(tokenIndex)
		newBalance := prevBalance.Sub(transfer.Amount)
		_, err := assetTree.UpdateLeaf(tokenIndex, newBalance)
		if err != nil {
			panic(err)
		}
		prevBalanceEntry := intMaxTree.AssetLeafEntry{
			TokenIndex: tokenIndex,
			Leaf:       prevBalance,
		}
		prevBalances = append(prevBalances, &prevBalanceEntry)
		assetMerkleProofs = append(assetMerkleProofs, &assetMerkleProof)
		insufficientFlags.SetBit(i, newBalance.IsInsufficient)
	}

	return &balance_prover_service.SpentTokenWitness{
		PrevPrivateState:    prevPrivateState,
		PrevBalances:        prevBalances,
		AssetMerkleProofs:   assetMerkleProofs,
		InsufficientFlags:   *insufficientFlags,
		Transfers:           transfers,
		NewPrivateStateSalt: newPrivateStateSalt,
	}, nil
}

func (w *mockWallet) UpdateOnSendTx(
	newSalt balance_prover_service.Salt,
	txWitness *balance_prover_service.TxWitness,
	transferWitnesses []*intMaxTypes.TransferWitness,
) (*balance_prover_service.SendWitness, error) {
	prevPrivateState := w.PrivateState()
	prevBalancePis, err := w.GetBalancePublicInputs()
	if err != nil {
		return nil, err
	}

	if txWitness.Tx.Nonce != w.nonce {
		panic("nonce mismatch")
	}

	w.nonce += 1
	w.salt = newSalt
	w.publicState = txWitness.ValidityPis.PublicState

	transfers := make([]*intMaxTypes.Transfer, 0, len(transferWitnesses))
	assetMerkleProofs := make([]*intMaxTree.AssetMerkleProof, 0, len(transferWitnesses))
	prevBalances := make([]*intMaxTree.AssetLeafEntry, 0, len(transferWitnesses))
	insufficientFlags := new(backup_balance.InsufficientFlags)
	// insufficientBits := make([]bool, 0, len(transferWitnesses))
	fmt.Printf("transferWitnesses: %v\n", transferWitnesses)
	for i, transferWitness := range transferWitnesses {
		if transferWitness == nil {
			return nil, fmt.Errorf("transferWitness[%d] is nil", i)
		}

		transfer := transferWitness.Transfer
		tokenIndex := transfer.TokenIndex
		prevBalance := w.assetTree.GetLeaf(tokenIndex)
		assetMerkleProof, _, _ := w.assetTree.Prove(tokenIndex)
		newBalance := prevBalance.Sub(transfer.Amount)
		_, err = w.assetTree.UpdateLeaf(tokenIndex, newBalance)
		if err != nil {
			panic(err)
		}
		transfers = append(transfers, &transfer)
		prevBalanceEntry := intMaxTree.AssetLeafEntry{
			TokenIndex: tokenIndex,
			Leaf:       prevBalance,
		}
		prevBalances = append(prevBalances, &prevBalanceEntry)
		assetMerkleProofs = append(assetMerkleProofs, &assetMerkleProof)
		insufficientFlags.SetBit(i, newBalance.IsInsufficient)
	}

	sendWitness := balance_prover_service.SendWitness{
		SpentTokenWitness: &balance_prover_service.SpentTokenWitness{
			PrevPrivateState:    prevPrivateState,
			PrevBalances:        prevBalances,
			AssetMerkleProofs:   assetMerkleProofs,
			InsufficientFlags:   *insufficientFlags,
			Transfers:           transfers,
			NewPrivateStateSalt: newSalt,
		},
		PrevBalancePis: prevBalancePis,
		TxWitness:      *txWitness,
	}

	w.sendWitnesses[sendWitness.GetIncludedBlockNumber()] = &sendWitness
	w.transferWitnesses[sendWitness.GetIncludedBlockNumber()] = transferWitnesses

	return &sendWitness, nil
}

func (w *mockWallet) SendTxAndUpdate(
	blockValidityService *block_validity_prover.BlockValidityProverMemory,
	transfers []*intMaxTypes.Transfer,
) (*balance_prover_service.SendWitness, error) {
	txWitness, transferWitnesses, err := w.SendTx(blockValidityService, transfers)
	if err != nil {
		return nil, err
	}
	newSalt, err := new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		return nil, err
	}
	return w.UpdateOnSendTx(*newSalt, txWitness, transferWitnesses)
}

func NewMockWallet(privateKey *intMaxAcc.PrivateKey) (*mockWallet, error) {
	zeroAsset := new(intMaxTree.AssetLeaf).SetDefault()
	const assetTreeHeight = 32
	const nullifierTreeHeight = 32
	assetTree, err := intMaxTree.NewAssetTree(assetTreeHeight, nil, zeroAsset.Hash())
	if err != nil {
		return nil, err
	}

	nullifierTree, err := intMaxTree.NewNullifierTree(nullifierTreeHeight)
	if err != nil {
		return nil, err
	}

	return &mockWallet{
		privateKey:        *privateKey,
		assetTree:         *assetTree,
		nullifierTree:     *nullifierTree,
		nonce:             0,
		salt:              balance_prover_service.Salt{},
		publicState:       new(block_validity_prover.PublicState).Genesis(),
		sendWitnesses:     make(map[uint32]*balance_prover_service.SendWitness),
		depositCases:      make(map[uint32]*balance_prover_service.DepositCase), // depositId => DepositCase
		transferWitnesses: make(map[uint32][]*intMaxTypes.TransferWitness),
	}, nil
}

func (s *mockWallet) PrivateKey() *intMaxAcc.PrivateKey {
	return &s.privateKey
}

func (s *mockWallet) PublicKey() *intMaxAcc.PublicKey {
	return s.privateKey.Public()
}

func (s *mockWallet) Nonce() uint32 {
	return s.nonce
}

func (s *mockWallet) Salt() balance_prover_service.Salt {
	return s.salt
}

func (s *mockWallet) Balance(tokenIndex uint32) *intMaxTypes.Uint256 {
	assets := s.assetTree.Leaves
	for _, asset := range assets {
		fmt.Printf("asset: %v", asset)
	}

	return s.assetTree.GetLeaf(tokenIndex).Amount
}

func (s *mockWallet) GenericAddress() (*intMaxTypes.GenericAddress, error) {
	return intMaxTypes.NewINTMAXAddress(s.PublicKey().ToAddress().Bytes())
}

func (s *mockWallet) PrivateState() *balance_prover_service.PrivateState {
	return &balance_prover_service.PrivateState{
		AssetTreeRoot:     s.assetTree.GetRoot(),
		NullifierTreeRoot: s.nullifierTree.GetRoot(),
		TransactionCount:  s.nonce,
		Salt:              s.salt,
	}
}

func (s *mockWallet) PublicState() *block_validity_prover.PublicState {
	return s.publicState
}

func (s *mockWallet) Nullifiers() []intMaxTypes.Bytes32 {
	return s.nullifierTree.Nullifiers()
}

func (s *mockWallet) IsIncludedInNullifierTree(nullifier intMaxTypes.Bytes32) (bool, error) {
	membershipProof, _, err := s.nullifierTree.ProveMembership(nullifier)
	if err != nil {
		return false, err
	}

	return membershipProof.IsIncluded, nil
}

func (s *mockWallet) AssetLeaves() map[uint32]*intMaxTree.AssetLeaf {
	return s.assetTree.Leaves
}

func (s *mockWallet) GetAllBlockNumbers() []uint32 {
	existedBlockNumbers := make(map[uint32]bool)
	for _, w := range s.sendWitnesses {
		blockNumber := w.GetIncludedBlockNumber()
		existedBlockNumbers[blockNumber] = true
	}

	result := make([]uint32, 0, len(existedBlockNumbers))
	for blockNumber := range existedBlockNumbers {
		result = append(result, blockNumber)
	}

	return result
}

func (s *mockWallet) UpdatePublicState(publicState *block_validity_prover.PublicState) {
	s.publicState = new(block_validity_prover.PublicState).Set(publicState)
}

func (s *mockWallet) DecryptBalanceData(encryptedBalanceData string) (*block_synchronizer.BalanceData, error) {
	balanceData := new(block_synchronizer.BalanceData)
	err := balanceData.Decrypt(&s.privateKey, encryptedBalanceData)
	if err != nil {
		return nil, err
	}

	return balanceData, nil
}

func (s *mockWallet) GetLastSendWitness() *balance_prover_service.SendWitness {
	if len(s.sendWitnesses) == 0 {
		return nil
	}

	lastBlockNumber := uint32(0)
	var lastSendWitness *balance_prover_service.SendWitness
	for blockNumber, sendWitness := range s.sendWitnesses {
		if blockNumber > lastBlockNumber {
			lastBlockNumber = blockNumber
			lastSendWitness = sendWitness
		}
	}

	return lastSendWitness
}

func (s *mockWallet) GetBalancePublicInputs() (*balance_prover_service.BalancePublicInputs, error) {
	lastSendWitness := s.GetLastSendWitness()
	lastTxHash := new(poseidonHashOut)
	lastTxInsufficientFlags := backup_balance.InsufficientFlags{}
	if lastSendWitness != nil {
		nextLastTx, err := lastSendWitness.GetNextLastTx()
		if err != nil {
			return nil, err
		}
		lastTxHash = nextLastTx.LastTxHash
		lastTxInsufficientFlags = nextLastTx.LastTxInsufficientFlags
	}

	return &balance_prover_service.BalancePublicInputs{
		PubKey:                  s.privateKey.Public(),
		PrivateCommitment:       s.PrivateState().Commitment(),
		LastTxHash:              lastTxHash,
		LastTxInsufficientFlags: lastTxInsufficientFlags,
		PublicState:             s.publicState,
	}, nil
}

func (s *mockWallet) GetSendWitness(blockNumber uint32) (*balance_prover_service.SendWitness, error) {
	result, ok := s.sendWitnesses[blockNumber]
	if !ok {
		return nil, errors.New("send witness not found")
	}

	return result, nil
}

func (s *mockWallet) GeneratePrivateWitness(
	newSalt balance_prover_service.Salt,
	tokenIndex uint32,
	amount *big.Int,
	nullifier intMaxTypes.Bytes32,
) (*balance_prover_service.PrivateWitness, error) {
	prevPrivateState := s.PrivateState()

	fmt.Printf("s.assetTree: %v\n", s.assetTree)
	assetTree := new(intMaxTree.AssetTree).Set(&s.assetTree)             // clone
	nullifierTree := new(intMaxTree.NullifierTree).Set(&s.nullifierTree) // clone

	prevAssetLeaf := assetTree.GetLeaf(tokenIndex)
	assetMerkleProof, _, err := assetTree.Prove(tokenIndex)
	if err != nil {
		return nil, err
	}

	assetRoot := assetTree.GetRoot()
	fmt.Printf("prev asset leaf isInsufficient: %v\n", prevAssetLeaf.IsInsufficient)
	fmt.Printf("prev asset leaf amount: %v\n", prevAssetLeaf.Amount.BigInt())
	fmt.Printf("prev asset leaf hash: %v\n", prevAssetLeaf.Hash())
	fmt.Printf("prev asset root hash: %s\n", assetRoot.String())
	// fmt.Printf("prev asset root hash: %s\n", assetRoot.String())
	fmt.Printf("prev asset index: %d\n", tokenIndex)
	for i, sibling := range assetMerkleProof.Siblings {
		fmt.Printf("asset Merkle proof: siblings[%d] = %s\n", i, sibling)
	}

	newAssetLeaf := prevAssetLeaf.Add(amount)
	_, err = assetTree.UpdateLeaf(tokenIndex, newAssetLeaf)
	if err != nil {
		return nil, err
	}

	oldNullifierTreeRoot := nullifierTree.GetRoot()
	// fmt.Printf("old nullifier tree root: %v\n", oldNullifierTreeRoot)
	// fmt.Printf("inserting nullifier: %v\n", nullifier)
	fmt.Printf("Adding nullifier: %x\n", nullifier.Bytes())
	for i, nullifierLeaf := range nullifierTree.GetLeaves() {
		fmt.Printf("nullifier leaf[%d]: %x\n", i, nullifierLeaf.Key.Bytes())
	}

	nullifierProof, err := nullifierTree.Insert(nullifier)
	if err != nil {
		fmt.Printf("(GeneratePrivateWitness) insert nullifier error: %v\n", err)
		return nil, errors.New("nullifier already exists")
	}
	// expectedNewNullifierTreeRoot := nullifierTree.GetRoot()
	// fmt.Printf("expected new nullifier tree root: %v\n", expectedNewNullifierTreeRoot)

	// for i, sibling := range nullifierProof.LeafProof.Siblings {
	// 	fmt.Printf("nullifier leaf Merkle proof: siblings[%d] = %s\n", i, sibling.String())
	// }
	// for i, sibling := range nullifierProof.LowLeafProof.Siblings {
	// 	fmt.Printf("nullifier low leaf Merkle proof: siblings[%d] = %s\n", i, sibling.String())
	// }

	nullifierInt := new(intMaxTypes.Uint256).FromFieldElementSlice(nullifier.ToFieldElementSlice())
	newNullifierTreeRoot, err := nullifierProof.GetNewRoot(nullifierInt.BigInt(), 0, oldNullifierTreeRoot)
	if err != nil {
		return nil, errors.Join(errors.New("fail to GetNewRoot"), err)
	}
	fmt.Printf("actual new nullifier tree root: %v\n", newNullifierTreeRoot)

	return &balance_prover_service.PrivateWitness{
		TokenIndex:       tokenIndex,
		Amount:           amount,
		Nullifier:        nullifier,
		NewSalt:          newSalt,
		PrevPrivateState: prevPrivateState,
		NullifierProof:   nullifierProof,
		PrevAssetLeaf:    prevAssetLeaf,
		AssetMerkleProof: &assetMerkleProof,
	}, nil
}

var ErrNullifierTreeProof = errors.New("failed to generate nullifier tree proof")

var ErrNullifierAlreadyExists = errors.New("nullifier already exists")

func (s *mockWallet) updateOnReceive(witness *balance_prover_service.PrivateWitness) error {
	nullifier := new(intMaxTypes.Uint256).FromFieldElementSlice(witness.Nullifier.ToFieldElementSlice())
	oldNullifierTreeRoot := s.nullifierTree.GetRoot()
	// fmt.Printf("old nullifier tree root: %v\n", oldNullifierTreeRoot)
	// fmt.Printf("nullifier: %v\n", nullifier)
	// for i, sibling := range witness.NullifierProof.LeafProof.Siblings {
	// 	fmt.Printf("nullifier leaf Merkle proof: siblings[%d] = %s\n", i, sibling.String())
	// }
	// for i, sibling := range witness.NullifierProof.LowLeafProof.Siblings {
	// 	fmt.Printf("nullifier low leaf Merkle proof: siblings[%d] = %s\n", i, sibling.String())
	// }
	newNullifierTreeRoot, err := witness.NullifierProof.GetNewRoot(nullifier.BigInt(), 0, oldNullifierTreeRoot)
	if err != nil {
		return errors.Join(errors.New("invalid nullifier proof"), err)
	}

	fmt.Printf("nullifier tree root before Insert: %v\n", s.nullifierTree.GetRoot())
	fmt.Printf("inserting nullifier: %s\n", witness.Nullifier.Hex())
	membershipProof, _, err := s.nullifierTree.ProveMembership(witness.Nullifier)
	if err != nil {
		fmt.Printf("nullifier tree proof error: %v\n", err)
		return errors.Join(ErrNullifierTreeProof, err)
	}
	if membershipProof.IsIncluded {
		fmt.Printf("nullifier already exists: %s\n", witness.Nullifier.Hex())
		return ErrNullifierAlreadyExists
	}

	assetMerkleProof := witness.AssetMerkleProof
	err = assetMerkleProof.Verify(
		witness.PrevAssetLeaf.Hash(),
		int(witness.TokenIndex),
		s.assetTree.GetRoot(),
	)
	if err != nil {
		return errors.New("invalid asset merkle proof")
	}

	newAssetLeaf := witness.PrevAssetLeaf.Add(witness.Amount)
	newAssetTreeRoot := witness.AssetMerkleProof.GetRoot(newAssetLeaf.Hash(), int(witness.TokenIndex))

	_, err = s.nullifierTree.Insert(witness.Nullifier)
	if err != nil {
		fmt.Printf("Fatal: nullifier already exists: %s\n", witness.Nullifier.Hex())
		panic(errors.New("nullifier already exists"))
	}
	fmt.Printf("nullifier tree root after Insert: %v\n", s.nullifierTree.GetRoot())
	_, err = s.assetTree.UpdateLeaf(witness.TokenIndex, newAssetLeaf)
	if err != nil {
		return err
	}

	nullifierTreeRoot := s.nullifierTree.GetRoot()
	if !nullifierTreeRoot.Equal(newNullifierTreeRoot) {
		return fmt.Errorf("nullifier tree root not equal: %s != %s", nullifierTreeRoot, newNullifierTreeRoot)
	}

	assetTreeRoot := s.assetTree.GetRoot()
	fmt.Printf("expected asset tree root: %s\n", assetTreeRoot.String())
	fmt.Printf("actual asset tree root: %s\n", newAssetTreeRoot.String())
	if !assetTreeRoot.Equal(newAssetTreeRoot) {
		return fmt.Errorf("asset tree root not equal: %s != %s", assetTreeRoot.String(), newAssetTreeRoot)
	}

	s.salt = witness.NewSalt

	return nil
}

// type MockBlockBuilder = block_validity_prover.BlockBuilderStorage

func (s *mockWallet) ReceiveDepositAndUpdate(
	blockValidityService block_validity_prover.BlockValidityService,
	depositIndex uint32,
) (*balance_prover_service.ReceiveDepositWitness, error) {
	fmt.Printf("-----ReceiveDepositAndUpdate %d-----\n", depositIndex)
	for index, depositCase := range s.depositCases {
		fmt.Printf("depositCase[%d]: %v\n", index, depositCase)
		fmt.Printf("depositHash[%d]: %v\n", index, depositCase.Deposit.Hash())
	}

	depositCase, ok := s.depositCases[depositIndex]
	if !ok {
		return nil, errors.New("deposit not found")
	}

	userDepositTreeRoot := s.publicState.DepositTreeRoot
	blockNumber := s.publicState.BlockNumber
	fmt.Printf("user deposit tree root: %s\n", userDepositTreeRoot.String())
	depositMerkleProof, depositTreeRoot, err := blockValidityService.DepositTreeProof(blockNumber, depositIndex)
	if err != nil {
		return nil, err
	}

	fmt.Printf("depositCase.Deposit hash: %v\n", depositCase.Deposit.Hash())
	fmt.Printf("depositCase.Deposit: %v\n", depositCase.Deposit)
	fmt.Printf("depositCase.Deposit RecipientSaltHash: %v\n", common.Hash(depositCase.Deposit.RecipientSaltHash).String())
	fmt.Printf("depositCase.DepositIndex: %d\n", depositCase.DepositIndex)
	for i, sibling := range depositMerkleProof.Siblings {
		fmt.Printf("depositCase.Deposit Merkle proof: siblings[%d] = %s\n", i, common.Hash(sibling))
	}
	fmt.Printf("deposit index: %d\n", depositIndex)
	fmt.Printf("ReceiveDepositAndUpdate deposit tree root: %s\n", depositTreeRoot.String())
	if depositTreeRoot != userDepositTreeRoot {
		return nil, errors.New("deposit tree root is mismatch")
	}

	err = depositMerkleProof.Verify(depositCase.Deposit.Hash(), int(depositCase.DepositIndex), depositTreeRoot)
	if err != nil {
		fmt.Printf("deposit Merkle proof verify error: %v\n", err)
		return nil, err
	}

	depositWitness := balance_prover_service.DepositWitness{
		DepositMerkleProof: depositMerkleProof,
		DepositSalt:        depositCase.DepositSalt,
		DepositIndex:       uint(depositCase.DepositIndex),
		DepositRoot:        depositTreeRoot,
		Deposit:            depositCase.Deposit,
	}
	deposit := depositWitness.Deposit
	nullifier := deposit.Nullifier()
	fmt.Printf("deposit: %+v\n", deposit)
	fmt.Printf("deposit (nullifier) dummy: %v\n", nullifier)

	newSalt, err := new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}

	nullifierBytes32 := intMaxTypes.Bytes32{}
	nullifierBytes32.FromPoseidonHashOut(nullifier)
	salt := balance_prover_service.Salt(*newSalt)
	privateWitness, err := s.GeneratePrivateWitness(salt, deposit.TokenIndex, deposit.Amount, nullifierBytes32)
	if err != nil {
		return nil, err
	}

	// delete deposit
	delete(s.depositCases, depositIndex)

	// update
	err = s.updateOnReceive(privateWitness)
	if err != nil {
		if err.Error() == ErrNullifierAlreadyExists.Error() {
			return nil, ErrNullifierAlreadyExists
		}

		fmt.Printf("updateOnReceive error: %v\n", err)
		return nil, err
	}
	fmt.Println("finish updateOnReceive")

	return &balance_prover_service.ReceiveDepositWitness{
		DepositWitness: &depositWitness,
		PrivateWitness: privateWitness,
	}, nil
}

func (s *mockWallet) ReceiveTransferAndUpdate(
	blockValidityService block_validity_prover.BlockValidityService,
	lastBlockNumber uint32,
	transferWitness *intMaxTypes.TransferWitness,
	senderLastBalanceProof string,
	senderBalanceTransitionProof string,
) (*balance_prover_service.ReceiveTransferWitness, error) {
	receiveTransferWitness, err := s.GenerateReceiveTransferWitness(
		blockValidityService,
		lastBlockNumber,
		transferWitness,
		senderLastBalanceProof,
		senderBalanceTransitionProof,
		false, // skipInsufficientCheck
	)
	if err != nil {
		return nil, err
	}

	err = s.updateOnReceive(receiveTransferWitness.PrivateWitness)
	if err != nil {
		if err.Error() == ErrNullifierAlreadyExists.Error() {
			return nil, ErrNullifierAlreadyExists
		}

		return nil, err
	}

	return receiveTransferWitness, nil
}

func (s *mockWallet) GenerateReceiveTransferWitness(
	blockValidityService block_validity_prover.BlockValidityService,
	receiverBlockNumber uint32,
	transferWitness *intMaxTypes.TransferWitness,
	senderLastBalanceProof string,
	senderBalanceTransitionProof string,
	skipInsufficientCheck bool,
) (*balance_prover_service.ReceiveTransferWitness, error) {
	transfer := transferWitness.Transfer
	recipientAddress, err := s.GenericAddress()
	if err != nil {
		return nil, err
	}
	if !transfer.Recipient.Equal(recipientAddress) {
		return nil, errors.New("invalid recipient address")
	}

	senderLastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(senderLastBalanceProof)
	if err != nil {
		return nil, err
	}

	lastBalancePis, err := new(balance_prover_service.BalancePublicInputs).FromPublicInputs(senderLastBalanceProofWithPis.PublicInputs)
	if err != nil {
		return nil, err
	}

	senderBalanceTransitionProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(senderBalanceTransitionProof)
	if err != nil {
		return nil, err
	}

	balanceTransitionPis, err := new(balance_prover_service.SenderPublicInputs).FromPublicInputs(senderBalanceTransitionProofWithPis.PublicInputs)
	if err != nil {
		return nil, err
	}

	nextBalancePublicInputs, err := lastBalancePis.UpdateWithSendTransition(
		balanceTransitionPis,
	)
	if err != nil {
		return nil, err
	}

	// TODO: check sender's balance proof
	if nextBalancePublicInputs.PublicState.BlockNumber > receiverBlockNumber {
		return nil, errors.New("receiver's balance proof does not include the incomming tx")
	}

	if !nextBalancePublicInputs.LastTxHash.Equal(transferWitness.Tx.Hash()) {
		return nil, errors.New("last tx hash mismatch")
	}

	if !skipInsufficientCheck {
		if nextBalancePublicInputs.LastTxInsufficientFlags.RandomAccess(int(transfer.TokenIndex)) {
			return nil, errors.New("tx insufficient check failed")
		}
	}

	nullifier := transfer.Nullifier()
	nullifierBytes32 := intMaxTypes.Bytes32{}
	nullifierBytes32.FromPoseidonHashOut(nullifier)
	salt := balance_prover_service.Salt(poseidonHashOut{})
	privateWitness, err := s.GeneratePrivateWitness(salt, transfer.TokenIndex, transfer.Amount, nullifierBytes32)
	if err != nil {
		return nil, err
	}

	// blockMerkleProof, err := blockBuilder.GetBlockMerkleProof(receiverBlockNumber, balancePis.PublicState.BlockNumber)
	blockMerkleProof, err := blockValidityService.BlockTreeProof(receiverBlockNumber, nextBalancePublicInputs.PublicState.BlockNumber)
	if err != nil {
		return nil, err
	}

	return &balance_prover_service.ReceiveTransferWitness{
		TransferWitness:        transferWitness,
		LastBalanceProof:       senderLastBalanceProof,
		BalanceTransitionProof: senderBalanceTransitionProof,
		PrivateWitness:         privateWitness,
		BlockMerkleProof:       blockMerkleProof,
	}, nil
}

func (w *mockWallet) Deposit(b *block_validity_prover.MockBlockBuilderMemory, salt balance_prover_service.Salt, tokenIndex uint32, amount *big.Int) uint32 {
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(w.PublicKey().BigInt(), &salt)
	depositLeaf := intMaxTree.DepositLeaf{
		RecipientSaltHash: recipientSaltHash,
		TokenIndex:        tokenIndex,
		Amount:            amount,
	}
	b.DepositLeaves = append(b.DepositLeaves, &depositLeaf)
	_, depositIndex, _ := b.DepositTree.GetCurrentRootCountAndSiblings()
	_, err := b.DepositTree.AddLeaf(depositIndex, depositLeaf.Hash())
	if err != nil {
		panic(err)
	}

	depositCase := balance_prover_service.DepositCase{
		DepositSalt:  salt,
		DepositIndex: depositIndex,
		Deposit:      depositLeaf,
	}
	fmt.Printf("(Deposit.AddDepositCase): %+v\n", depositCase)
	err = w.AddDepositCase(depositIndex, &depositCase)
	if err != nil {
		panic(err)
	}

	return depositIndex
}
