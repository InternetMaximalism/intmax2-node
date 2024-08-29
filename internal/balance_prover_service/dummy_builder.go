package balance_prover_service

// type MockBlockBuilder struct{}

// func (s *MockBlockBuilder) GetAuxInfo(blockNumber uint32) (*BalanceValidityAuxInfo, bool) {
// 	return nil, false
// }

// func (s *MockBlockBuilder) LastBlockNumber() uint32 {
// 	return 0
// }

// func (s *MockBlockBuilder) GetBlockNumber() uint32 {
// 	return 0
// }

// func (s *MockBlockBuilder) GetDepositTreeProof(index uint32) *intMaxTree.MerkleProof {
// 	return nil
// }

// func (s *MockBlockBuilder) GetBlockMerkleProof(rootBlockNumber, leafBlockNumber uint32) (*intMaxTree.BlockHashMerkleProof, error) {
// 	// if rootBlockNumber < leafBlockNumber {
// 	// 	return nil, errors.New("root block number is less than leaf block number")
// 	// }

// 	// auxInfo, ok := s.GetAuxInfo(rootBlockNumber)
// 	// if !ok {
// 	// 	return nil, errors.New("current block number not found")
// 	// }
// 	// blockMerkleProof := auxInfo.BlockTree.Prove(int(leafBlockNumber))

// 	// return blockMerkleProof, nil

// 	return nil, errors.New("not implemented")
// }

// // pub fn deposit(&mut self, deposit: &Deposit) -> usize {
// // 	self.deposit_tree.push(deposit.clone());
// // 	let deposit_index = self.deposit_tree.len() - 1;
// // 	deposit_index
// // }
