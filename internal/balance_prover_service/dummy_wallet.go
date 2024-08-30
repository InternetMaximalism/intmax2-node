package balance_prover_service

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	"math/big"
	"math/rand"
)

type MockWallet struct {
	privateKey        intMaxAcc.PrivateKey
	assetTree         intMaxTree.AssetTree
	nullifierTree     intMaxTree.NullifierTree
	nonce             uint32
	salt              Salt
	publicState       *block_validity_prover.PublicState
	sendWitnesses     map[uint32]*SendWitness
	depositCases      map[uint32]*DepositCase
	transferWitnesses map[uint32][]*ReceiveTransferWitness
}

func (w *MockWallet) AddDepositCase(depositIndex uint32, depositCase *DepositCase) error {
	w.depositCases[depositIndex] = depositCase
	return nil
}

// pub fn new_rand<R: Rng>(rng: &mut R) -> Self {
// 	Self {
// 		key_set: KeySet::rand(rng),
// 		asset_tree: AssetTree::new(ASSET_TREE_HEIGHT),
// 		nullifier_tree: NullifierTree::new(),
// 		nonce: 0,
// 		salt: Salt::default(),
// 		public_state: PublicState::genesis(),
// 		send_witnesses: Vec::new(),
// 		deposit_cases: HashMap::new(),
// 		transfer_witnesses: HashMap::new(),
// 	}
// }

func NewMockWallet(privateKey *intMaxAcc.PrivateKey) (*MockWallet, error) {
	zeroAsset := new(intMaxTree.AssetLeaf).SetDefault()
	assetTree, err := intMaxTree.NewAssetTree(7, nil, zeroAsset.Hash())
	if err != nil {
		return nil, err
	}

	nullifierTree, err := intMaxTree.NewNullifierTree(32)
	if err != nil {
		return nil, err
	}

	return &MockWallet{
		privateKey:        *privateKey,
		assetTree:         *assetTree,
		nullifierTree:     *nullifierTree,
		nonce:             0,
		salt:              Salt{},
		publicState:       new(block_validity_prover.PublicState).Genesis(),
		sendWitnesses:     make(map[uint32]*SendWitness),
		depositCases:      make(map[uint32]*DepositCase), // depositId => DepositCase
		transferWitnesses: make(map[uint32][]*ReceiveTransferWitness),
	}, nil
}

func (s *MockWallet) PublicKey() *intMaxAcc.PublicKey {
	return s.privateKey.Public()
}

func (s *MockWallet) GenericAddress() (*intMaxTypes.GenericAddress, error) {
	return intMaxTypes.NewINTMAXAddress(s.PublicKey().ToAddress().Bytes())
}

func (s *MockWallet) PrivateState() *PrivateState {
	result := new(PrivateState).SetDefault()

	return &PrivateState{
		AssetTreeRoot:     result.AssetTreeRoot,
		NullifierTreeRoot: result.NullifierTreeRoot,
		Nonce:             s.nonce,
		Salt:              s.salt,
	}
}

func (s *MockWallet) GetAllBlockNumbers() []uint32 {
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

func (s *MockWallet) UpdatePublicState(publicState *block_validity_prover.PublicState) {
	s.publicState = publicState
}

func (s *MockWallet) GetLastSendWitness() *SendWitness {
	if len(s.sendWitnesses) == 0 {
		return nil
	}

	lastBlockNumber := uint32(0)
	var lastSendWitness *SendWitness
	for blockNumber, sendWitness := range s.sendWitnesses {
		if blockNumber > lastBlockNumber {
			lastBlockNumber = blockNumber
			lastSendWitness = sendWitness
		}
	}

	return lastSendWitness
}

func (s *MockWallet) GetBalancePublicInputs() (*BalancePublicInputs, error) {
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

	return &BalancePublicInputs{
		PubKey:                  s.privateKey.Public(),
		PrivateCommitment:       s.PrivateState().Commitment(),
		LastTxHash:              lastTxHash,
		LastTxInsufficientFlags: lastTxInsufficientFlags,
		PublicState:             s.publicState,
	}, nil
}

func (s *MockWallet) GetSendWitness(blockNumber uint32) (*SendWitness, error) {
	result, ok := s.sendWitnesses[blockNumber]
	if !ok {
		return nil, errors.New("send witness not found")
	}

	return result, nil
}

func (s *MockWallet) GeneratePrivateWitness(
	newSalt Salt,
	tokenIndex uint32,
	amount *big.Int,
	nullifier intMaxTypes.Bytes32,
) (*PrivateWitness, error) {
	assetTree := s.assetTree
	nullifierTree := s.nullifierTree
	prevPrivateState := s.PrivateState()

	prevAssetLeaf := assetTree.GetLeaf(tokenIndex)
	assetMerkleProof, _, err := assetTree.Prove(tokenIndex)
	if err != nil {
		return nil, err
	}
	newAssetLeaf := prevAssetLeaf.Add(amount)
	_, err = assetTree.UpdateLeaf(tokenIndex, newAssetLeaf)
	if err != nil {
		return nil, err
	}
	nullifierProof, err := nullifierTree.Insert(nullifier)
	if err != nil {
		return nil, errors.New("nullifier already exists")
	}

	return &PrivateWitness{
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

func (s *MockWallet) updateOnReceive(witness *PrivateWitness) error {
	nullifierTree := s.nullifierTree
	assetTree := s.assetTree

	nullifier := new(intMaxTypes.Uint256).FromFieldElementSlice(witness.Nullifier.ToFieldElementSlice())
	oldNullifierTreeRoot := nullifierTree.GetRoot()
	newNullifierTreeRoot, err := witness.NullifierProof.GetNewRoot(nullifier.BigInt(), 0, oldNullifierTreeRoot)
	if err != nil {
		return errors.New("invalid nullifier proof")
	}

	assetMerkleProof := witness.AssetMerkleProof
	err = assetMerkleProof.Verify(
		witness.PrevAssetLeaf,
		witness.TokenIndex,
		assetTree.GetRoot(),
	)
	if err != nil {
		return errors.New("invalid asset merkle proof")
	}

	newAssetLeaf := witness.PrevAssetLeaf.Add(witness.Amount)
	newAssetTreeRoot := witness.AssetMerkleProof.GetRoot(newAssetLeaf, witness.TokenIndex)
	_, err = nullifierTree.Insert(witness.Nullifier)
	if err != nil {
		return errors.New("nullifier already exists")
	}
	_, err = assetTree.UpdateLeaf(witness.TokenIndex, newAssetLeaf)
	if err != nil {
		return err
	}

	nullifierTreeRoot := nullifierTree.GetRoot()
	if !nullifierTreeRoot.Equal(newNullifierTreeRoot) {
		return errors.New("nullifier tree root not equal")
	}

	assetTreeRoot := assetTree.GetRoot()
	if assetTreeRoot.Equal(newAssetTreeRoot) {
		return errors.New("asset tree root not equal")
	}

	s.salt = witness.NewSalt

	return nil
}

type MockBlockBuilder = block_validity_prover.BlockBuilderStorage

func (s *MockWallet) ReceiveDepositAndUpdate(
	blockBuilder MockBlockBuilder,
	depositId uint32,
) (*ReceiveDepositWitness, error) {
	depositCase, ok := s.depositCases[depositId]
	if !ok {
		return nil, errors.New("deposit not found")
	}

	depositMerkleProof, err := blockBuilder.DepositTreeProof(depositCase.DepositIndex)
	if err != nil {
		return nil, err
	}
	depositWitness := DepositWitness{
		DepositMerkleProof: depositMerkleProof,
		DepositSalt:        depositCase.DepositSalt,
		DepositIndex:       uint(depositCase.DepositIndex),
		Deposit:            depositCase.Deposit,
	}
	deposit := depositWitness.Deposit
	nullifier := deposit.Hash()

	newSalt, err := new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}

	nullifierBytes32 := intMaxTypes.Bytes32{}
	nullifierBytes32.FromBytes(nullifier[:])
	privateWitness, err := s.GeneratePrivateWitness(Salt(*newSalt), deposit.TokenIndex, deposit.Amount, nullifierBytes32)
	if err != nil {
		return nil, err
	}

	// delete deposit
	delete(s.depositCases, depositId)

	// update
	s.updateOnReceive(privateWitness)

	return &ReceiveDepositWitness{
		DepositWitness: &depositWitness,
		PrivateWitness: privateWitness,
	}, nil
}

func (s *MockWallet) ReceiveTransferAndUpdate(
	rng *rand.Rand,
	blockBuilder MockBlockBuilder,
	lastBlockNumber uint32,
	transferWitness *TransferWitness,
	senderBalanceProof *intMaxTypes.Plonky2Proof,
) (*ReceiveTransferWitness, error) {
	receiveTransferWitness, err := s.GenerateReceiveTransferWitness(
		rng,
		blockBuilder,
		lastBlockNumber,
		transferWitness,
		senderBalanceProof,
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

func (s *MockWallet) GenerateReceiveTransferWitness(
	rng *rand.Rand,
	blockBuilder MockBlockBuilder,
	receiverBlockNumber uint32,
	transferWitness *TransferWitness,
	senderBalanceProof *intMaxTypes.Plonky2Proof,
	skipInsufficientCheck bool,
) (*ReceiveTransferWitness, error) {
	transfer := transferWitness.Transfer
	recipientAddress, err := s.GenericAddress()
	if err != nil {
		return nil, err
	}
	if !transfer.Recipient.Equal(recipientAddress) {
		return nil, errors.New("invalid recipient address")
	}

	balancePis, err := new(BalancePublicInputs).FromPublicInputs(senderBalanceProof.PublicInputs)
	if err != nil {
		return nil, err
	}

	if balancePis.PublicState.BlockNumber > receiverBlockNumber {
		return nil, errors.New("receiver's balance proof does not include the incomming tx")
	}

	if !balancePis.LastTxHash.Equal(transferWitness.Tx.Hash()) {
		return nil, errors.New("last tx hash mismatch")
	}

	if !skipInsufficientCheck {
		if balancePis.LastTxInsufficientFlags.RandomAccess(int(transfer.TokenIndex)) {
			return nil, errors.New("tx insufficient check failed")
		}
	}

	nullifier := transfer.Hash()
	nullifierBytes32 := intMaxTypes.Bytes32{}
	nullifierBytes32.FromPoseidonHashOut(nullifier)
	privateWitness, err := s.GeneratePrivateWitness(Salt(*new(poseidonHashOut)), transfer.TokenIndex, transfer.Amount, nullifierBytes32)
	if err != nil {
		return nil, err
	}

	// blockMerkleProof, err := blockBuilder.GetBlockMerkleProof(receiverBlockNumber, balancePis.PublicState.BlockNumber)
	blockMerkleProof, err := blockBuilder.BlockTreeProof(balancePis.PublicState.BlockNumber)
	if err != nil {
		return nil, err
	}

	return &ReceiveTransferWitness{
		TransferWitness:  transferWitness,
		BalanceProof:     senderBalanceProof,
		PrivateWitness:   privateWitness,
		BlockMerkleProof: blockMerkleProof,
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

// func (w *MockWallet) Deposit(salt Salt, blockBuilder MockBlockBuilder, tokenIndex uint32, amount *big.Int) uint32 {
// 	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(w.PublicKey().BigInt(), &salt)
// 	deposit := intMaxTree.DepositLeaf{
// 		RecipientSaltHash: recipientSaltHash,
// 		TokenIndex:        tokenIndex,
// 		Amount:            amount,
// 	}
// 	depositIndex := blockBuilder.Deposit(deposit)

// 	depositCase := DepositCase{
// 		DepositSalt:  salt,
// 		DepositIndex: depositIndex,
// 		Deposit:      deposit,
// 	}
// 	w.depositCases[depositId] = &depositCase

// 	return depositIndex
// }
