package balance_synchronizer

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const numTransfersInTx = 1 << intMaxTree.TRANSFER_TREE_HEIGHT

type poseidonHashOut = intMaxGP.PoseidonHashOut

type mockWallet struct {
	privateKey        intMaxAcc.PrivateKey
	assetTree         intMaxTree.AssetTree
	nullifierTree     intMaxTree.NullifierTree
	nonce             uint32
	salt              balance_prover_service.Salt
	publicState       *block_validity_prover.PublicState
	sendWitnesses     map[uint32]*balance_prover_service.SendWitness
	depositCases      map[uint32]*balance_prover_service.DepositCase // depositIndex => DepositCase
	transferWitnesses map[uint32][]*intMaxTypes.TransferWitness
}

type UserState interface {
	AddDepositCase(depositIndex uint32, depositCase *balance_prover_service.DepositCase) error
	SendTx(
		// blockBuilder *block_validity_prover.MockBlockBuilderMemory,
		blockValidityService block_validity_prover.BlockValidityService,
		transfers []*intMaxTypes.Transfer,
	) (*balance_prover_service.TxWitness, []*intMaxTypes.TransferWitness, error)
	UpdateOnSendTx(
		salt balance_prover_service.Salt,
		txWitness *balance_prover_service.TxWitness,
		transferWitnesses []*intMaxTypes.TransferWitness,
	) (*balance_prover_service.SendWitness, error)
	SendTxAndUpdate(
		// blockBuilder *block_validity_prover.MockBlockBuilderMemory,
		blockValidityService block_validity_prover.BlockValidityService,
		transfers []*intMaxTypes.Transfer,
	) (*balance_prover_service.SendWitness, error)
	// PrivateKey() *intMaxAcc.PrivateKey
	PublicKey() *intMaxAcc.PublicKey
	Nonce() uint32
	GenericAddress() (*intMaxTypes.GenericAddress, error)
	PrivateState() *balance_prover_service.PrivateState
	GetAllBlockNumbers() []uint32
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
		// blockBuilder MockBlockBuilder,
		blockValidityService block_validity_prover.BlockValidityService,
		depositIndex uint32,
	) (*balance_prover_service.ReceiveDepositWitness, error)
	ReceiveTransferAndUpdate(
		// blockBuilder MockBlockBuilder,
		blockValidityService block_validity_prover.BlockValidityService,
		lastBlockNumber uint32,
		transferWitness *intMaxTypes.TransferWitness,
		senderLastBalanceProof string,
		senderBalanceTransitionProof string,
	) (*balance_prover_service.ReceiveTransferWitness, error)
}

func (w *mockWallet) AddDepositCase(depositIndex uint32, depositCase *balance_prover_service.DepositCase) error {
	w.depositCases[depositIndex] = depositCase
	return nil
}

func (w *mockWallet) SendTx(
	// blockBuilder *block_validity_prover.MockBlockBuilderMemory,
	blockValidityService block_validity_prover.BlockValidityService,
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
		_, err := transferTree.AddLeaf(index, transfer)
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
		WillReturnSignature: true,
	}
	txRequests := []*block_validity_prover.MockTxRequest{&txRequest0}

	fmt.Printf("IMPORTANT PostBlock")
	validityWitness, err := blockValidityService.PostBlock(
		w.nonce == 0,
		txRequests,
	)
	if err != nil {
		return nil, nil, err
	}

	txLeaves := make([]*intMaxTypes.Tx, len(txRequests))
	for i, tx := range txRequests {
		txLeaves[i] = tx.Tx
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	txTree, err := intMaxTree.NewTxTree(7, txLeaves, zeroTx.Hash())
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

func (w *mockWallet) UpdateOnSendTx(
	salt balance_prover_service.Salt,
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
	w.salt = salt
	w.publicState = txWitness.ValidityPis.PublicState

	transfers := make([]*intMaxTypes.Transfer, 0, len(transferWitnesses))
	assetMerkleProofs := make([]*intMaxTree.AssetMerkleProof, 0, len(transferWitnesses))
	prevBalances := make([]*intMaxTree.AssetLeaf, 0, len(transferWitnesses))
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
		// _, err = w.assetTree.UpdateLeaf(tokenIndex, newBalance)
		if tokenIndex < uint32(len(w.assetTree.Leaves)) {
			_, err = w.assetTree.UpdateLeaf(tokenIndex, newBalance)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = w.assetTree.AddLeaf(tokenIndex, newBalance)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			panic(err)
		}
		transfers = append(transfers, &transfer)
		prevBalances = append(prevBalances, prevBalance)
		assetMerkleProofs = append(assetMerkleProofs, assetMerkleProof)
		insufficientFlags.SetBit(i, newBalance.IsInsufficient)
	}

	sendWitness := balance_prover_service.SendWitness{
		PrevBalancePis:      prevBalancePis,
		PrevPrivateState:    prevPrivateState,
		PrevBalances:        prevBalances,
		AssetMerkleProofs:   assetMerkleProofs,
		InsufficientFlags:   *insufficientFlags,
		Transfers:           transfers,
		TxWitness:           *txWitness,
		NewPrivateStateSalt: salt,
	}

	w.sendWitnesses[sendWitness.GetIncludedBlockNumber()] = &sendWitness
	w.transferWitnesses[sendWitness.GetIncludedBlockNumber()] = transferWitnesses

	return &sendWitness, nil
}

func (w *mockWallet) SendTxAndUpdate(
	// blockBuilder *block_validity_prover.MockBlockBuilderMemory,
	blockValidityService block_validity_prover.BlockValidityService,
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

func (s *mockWallet) PublicKey() *intMaxAcc.PublicKey {
	return s.privateKey.Public()
}

func (s *mockWallet) Nonce() uint32 {
	return s.nonce
}

func (s *mockWallet) GenericAddress() (*intMaxTypes.GenericAddress, error) {
	return intMaxTypes.NewINTMAXAddress(s.PublicKey().ToAddress().Bytes())
}

func (s *mockWallet) PrivateState() *balance_prover_service.PrivateState {
	return &balance_prover_service.PrivateState{
		AssetTreeRoot:     s.assetTree.GetRoot(),
		NullifierTreeRoot: s.nullifierTree.GetRoot(),
		Nonce:             s.nonce,
		Salt:              s.salt,
	}
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
	fmt.Printf("prevPrivateState: %v\n", prevPrivateState)

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
	fmt.Printf("prev asset leaf amount: %v\n", prevAssetLeaf.Amount)
	fmt.Printf("prev asset leaf hash: %v\n", prevAssetLeaf.Hash())
	fmt.Printf("prev asset root hash: %s\n", assetRoot.String())
	// fmt.Printf("prev asset root hash: %s\n", assetRoot.String())
	fmt.Printf("prev asset index: %d\n", tokenIndex)
	for i, sibling := range assetMerkleProof.Siblings {
		fmt.Printf("asset Merkle proof: siblings[%d] = %s\n", i, sibling)
	}

	newAssetLeaf := prevAssetLeaf.Add(amount)
	if tokenIndex < uint32(len(assetTree.Leaves)) {
		_, err = assetTree.UpdateLeaf(tokenIndex, newAssetLeaf)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = assetTree.AddLeaf(tokenIndex, newAssetLeaf)
		if err != nil {
			return nil, err
		}
	}

	oldNullifierTreeRoot := nullifierTree.GetRoot()
	// fmt.Printf("old nullifier tree root: %v\n", oldNullifierTreeRoot)
	// fmt.Printf("inserting nullifier: %v\n", nullifier)
	nullifierProof, err := nullifierTree.Insert(nullifier)
	if err != nil {
		fmt.Printf("insert nullifier error: %v\n", err)
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
		AssetMerkleProof: assetMerkleProof,
	}, nil
}

func (s *mockWallet) updateOnReceive(witness *balance_prover_service.PrivateWitness) error {
	fmt.Printf("s.assetTree: %v\n", s.assetTree)

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

	assetMerkleProof := witness.AssetMerkleProof
	err = assetMerkleProof.Verify(
		witness.PrevAssetLeaf,
		witness.TokenIndex,
		s.assetTree.GetRoot(),
	)
	if err != nil {
		return errors.New("invalid asset merkle proof")
	}

	newAssetLeaf := witness.PrevAssetLeaf.Add(witness.Amount)
	newAssetTreeRoot := witness.AssetMerkleProof.GetRoot(newAssetLeaf, witness.TokenIndex)

	fmt.Printf("nullifier tree root before Insert: %v\n", s.nullifierTree.GetRoot())
	fmt.Printf("inserting nullifier: %s\n", witness.Nullifier.Hex())
	_, err = s.nullifierTree.Insert(witness.Nullifier)
	if err != nil {
		fmt.Printf("insert nullifier error: %v\n", err)
		return errors.New("nullifier already exists")
	}
	fmt.Printf("nullifier tree root after Insert: %v\n", s.nullifierTree.GetRoot())

	if witness.TokenIndex < uint32(len(s.assetTree.Leaves)) {
		_, err = s.assetTree.UpdateLeaf(witness.TokenIndex, newAssetLeaf)
		if err != nil {
			return err
		}
	} else {
		_, err = s.assetTree.AddLeaf(witness.TokenIndex, newAssetLeaf)
		if err != nil {
			return err
		}
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

type MockBlockBuilder = block_validity_prover.BlockBuilderStorage

func (s *mockWallet) ReceiveDepositAndUpdate(
	// blockBuilder MockBlockBuilder,
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

	depositMerkleProof, depositTreeRoot, err := blockValidityService.DepositTreeProof(depositIndex)
	if err != nil {
		return nil, err
	}

	fmt.Printf("expected deposit tree root: %s\n", depositTreeRoot.String())
	fmt.Printf("deposit index: %d\n", depositIndex)
	fmt.Printf("ReceiveDepositAndUpdate deposit tree root: %s\n", depositTreeRoot.String())
	fmt.Printf("depositCase.Deposit: %v\n", depositCase.Deposit)
	fmt.Printf("depositCase.Deposit hash: %v\n", depositCase.Deposit.Hash())
	for i, sibling := range depositMerkleProof.Siblings {
		fmt.Printf("depositCase.Deposit Merkle proof: siblings[%d] = %s\n", i, common.Hash(sibling))
	}
	err = depositMerkleProof.Verify(depositTreeRoot, int(depositCase.DepositIndex), depositCase.Deposit.Hash())
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
	fmt.Printf("deposit: %v\n", deposit)
	fmt.Printf("deposit (nullifier) dummy: %v\n", nullifier)

	newSalt, err := new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}

	for i, leaf := range s.assetTree.Leaves {
		fmt.Printf("asset tree leaves[%d]: %v, %v\n", i, leaf.Amount, leaf.IsInsufficient)
	}
	for i, leaf := range s.nullifierTree.GetLeaves() {
		fmt.Printf("nullifier tree leaves[%d]: %v\n", i, leaf)
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
	// blockBuilder MockBlockBuilder,
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
		return nil, err
	}

	return receiveTransferWitness, nil
}

func (s *mockWallet) GenerateReceiveTransferWitness(
	// blockBuilder MockBlockBuilder,
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
	salt := balance_prover_service.Salt(*new(poseidonHashOut))
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

// /// Deposit tokens on the layer 1.
// pub fn deposit<R: Rng>(
//     &mut self,
//     rng: &mut R,
//     block_builder: &mut MockBlockBuilder,
//     token_index: u32,
//     amount: U256,
// ) -> usize {
//     let pubkey = self.get_pubkey();
//     let salt = Salt::rand(rng);
//     let pubkey_salt_hash = get_pubkey_salt_hash(pubkey, salt);

//     let deposit = Deposit {
//         pubkey_salt_hash,
//         token_index,
//         amount,
//     };
//     let deposit_index = block_builder.deposit(&deposit);

//     let deposit_case = DepositCase {
//         deposit_salt: salt,
//         deposit_index,
//         deposit,
//     };
//     self.deposit_cases.insert(deposit_index, deposit_case);
//     deposit_index
// }

func (w *mockWallet) Deposit(b *block_validity_prover.MockBlockBuilderMemory, salt balance_prover_service.Salt, tokenIndex uint32, amount *big.Int) uint32 {
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(w.PublicKey().BigInt(), &salt)
	depositLeaf := intMaxTree.DepositLeaf{
		RecipientSaltHash: recipientSaltHash,
		TokenIndex:        tokenIndex,
		Amount:            amount,
	}
	// depositIndex, err := blockBuilder.Deposit(depositLeaf)
	// if err != nil {
	// 	panic(err)
	// }
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
	w.AddDepositCase(depositIndex, &depositCase)

	return depositIndex
}
