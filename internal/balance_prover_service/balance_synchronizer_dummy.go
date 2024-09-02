package balance_prover_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet/models"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
)

// fn e2e_test() {
//     let mut rng = rand::thread_rng();
//     let mut block_builder = MockBlockBuilder::new();
//     let mut sync_validity_prover = SyncValidityProver::<F, C, D>::new();
//     let balance_processor = BalanceProcessor::new(sync_validity_prover.validity_circuit());

//     let mut alice_wallet = MockWallet::new_rand(&mut rng);
//     let mut alice_prover = SyncBalanceProver::<F, C, D>::new();

//     // depost 100wei ETH to alice wallet
//     let deposit_index = alice_wallet.deposit(&mut rng, &mut block_builder, 0, 100.into());

//     // post dummy block to reflect the deposit tree
//     block_builder.post_block(true, vec![]);

//     // sync alice wallet to the latest block, which includes the deposit
//     alice_prover.sync_all(
//         &mut sync_validity_prover,
//         &mut alice_wallet,
//         &balance_processor,
//         &block_builder,
//     );
//     let balance_pis = alice_prover.get_balance_pis();
//     assert_eq!(balance_pis.public_state.block_number, 1); // balance proof synced to block 1

//     // receive deposit and update alice balance proof
//     alice_prover.receive_deposit(
//         &mut rng,
//         &mut alice_wallet,
//         &balance_processor,
//         &block_builder,
//         deposit_index,
//     );
//     assert_eq!(get_asset_balance(&alice_wallet, 0), 100.into()); // check ETH balance

//     let mut bob_wallet = MockWallet::new_rand(&mut rng);
//     let mut bob_prover = SyncBalanceProver::<F, C, D>::new();

//     // transfer 50wei ETH to bob
//     let transfer_to_bob = Transfer {
//         recipient: GenericAddress::from_pubkey(bob_wallet.get_pubkey()),
//         token_index: 0,
//         amount: 50.into(),
//         salt: Salt::rand(&mut rng),
//     };
//     let send_witness =
//         alice_wallet.send_tx_and_update(&mut rng, &mut block_builder, &[transfer_to_bob]);
//     let transfer_witness = alice_wallet
//         .get_transfer_witnesses(send_witness.get_included_block_number())
//         .unwrap()[0] // first transfer in the tx
//         .clone();

//     // update alice balance proof
//     alice_prover.sync_all(
//         &mut sync_validity_prover,
//         &mut alice_wallet,
//         &balance_processor,
//         &block_builder,
//     );
//     assert_eq!(get_asset_balance(&alice_wallet, 0), 50.into()); // check ETH balance
//     let alice_balance_proof = alice_prover.get_balance_proof();

//     // sync bob wallet to the latest block
//     bob_prover.sync_all(
//         &mut sync_validity_prover,
//         &mut bob_wallet,
//         &balance_processor,
//         &block_builder,
//     );

//     // receive transfer and update bob balance proof
//     bob_prover.receive_transfer(
//         &mut rng,
//         &mut bob_wallet,
//         &balance_processor,
//         &block_builder,
//         &transfer_witness,
//         &alice_balance_proof,
//     );
//     assert_eq!(get_asset_balance(&bob_wallet, 0), 50.into()); // check ETH balance

//     // bob withdraw 10wei ETH
//     let bob_eth_address = Address::rand(&mut rng);
//     let withdrawal = Transfer {
//         recipient: GenericAddress::from_address(bob_eth_address),
//         token_index: 0,
//         amount: 10.into(),
//         salt: Salt::rand(&mut rng),
//     };
//     let withdrawal_send_witness =
//         bob_wallet.send_tx_and_update(&mut rng, &mut block_builder, &[withdrawal]);
//     let withdrawal_transfer_witness = bob_wallet
//         .get_transfer_witnesses(withdrawal_send_witness.get_included_block_number())
//         .unwrap()[0] // first transfer in the tx
//         .clone();

//     // update bob balance proof
//     bob_prover.sync_all(
//         &mut sync_validity_prover,
//         &mut bob_wallet,
//         &balance_processor,
//         &block_builder,
//     );
//     assert_eq!(get_asset_balance(&bob_wallet, 0), 40.into());
//     let bob_balance_proof = bob_prover.get_balance_proof();

//     // prove withdrawal
//     let withdrawal_processor = WithdrawalProcessor::new(&balance_processor.balance_circuit);
//     let withdrawal_witness = WithdrawalWitness {
//         transfer_witness: withdrawal_transfer_witness,
//         balance_proof: bob_balance_proof,
//     };
//     let withdrawal = withdrawal_witness.to_withdrawal();
//     assert_eq!(withdrawal.amount, 10.into()); // check withdrawal amount
//     let _withdrawal_proof = withdrawal_processor
//         .prove(&withdrawal_witness, &None)
//         .unwrap();
// }

type balanceSynchronizerDummy struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
	sb  block_validity_prover.ServiceBlockchain
	db  block_validity_prover.SQLDriverApp
}

func NewSynchronizerDummy(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	db block_validity_prover.SQLDriverApp,
) *balanceSynchronizerDummy {
	return &balanceSynchronizerDummy{
		ctx: ctx,
		cfg: cfg,
		log: log,
		sb:  sb,
		db:  db,
	}
}

func (s *balanceSynchronizerDummy) TestE2E(blockValidityProver block_validity_prover.BlockValidityProver, blockBuilderWallet *models.Wallet) {
	blockBuilder := block_validity_prover.NewMockBlockBuilder(s.cfg, s.db)
	syncValidityProver, err := NewSyncValidityProver(s.ctx, s.cfg, s.log, s.sb, s.db)
	if err != nil {
		s.log.Fatalf("failed to create sync validity prover: %+v", err)
	}
	balanceProcessor := NewBalanceProcessor(s.ctx, s.cfg, s.log)

	alicePrivateKey, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	if err != nil {
		s.log.Fatalf("failed to create private key: %+v", err)
	}

	aliceWallet, err := NewMockWallet(alicePrivateKey)
	if err != nil {
		s.log.Fatalf("failed to create mock wallet: %+v", err)
	}
	aliceProver := NewSyncBalanceProver()

	salt, err := new(Salt).SetRandom()
	if err != nil {
		s.log.Fatalf("failed to set random salt: %+v", err)
	}

	// depost 100wei ETH to alice wallet
	depositIndex := aliceWallet.Deposit(blockBuilder, *salt, 0, big.NewInt(100))

	// post dummy block to reflect the deposit tree
	_, err = blockBuilder.PostBlock(true, []*block_validity_prover.MockTxRequest{})
	if err != nil {
		s.log.Fatalf("failed to post block: %+v", err)
	}

	// sync alice wallet to the latest block, which includes the deposit
	err = aliceProver.SyncAll(syncValidityProver, aliceWallet, balanceProcessor, blockBuilder)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	balancePis, err := aliceProver.BalancePublicInputs()
	if err != nil {
		s.log.Fatalf("failed to get balance public inputs: %+v", err)
	}

	if balancePis.PublicState.BlockNumber != 1 {
		s.log.Fatalf("balance proof synced to block 1")
	}

	// receive deposit and update alice balance proof
	err = aliceProver.ReceiveDeposit(aliceWallet, balanceProcessor, blockBuilder, depositIndex)
	if err != nil {
		s.log.Fatalf("failed to receive deposit: %+v", err)
	}

	if GetAssetBalance(aliceWallet, 0).Cmp(big.NewInt(100)) != 0 {
		s.log.Fatalf("ETH balance")
	}

	bobPrivateKey, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	if err != nil {
		s.log.Fatalf("failed to create private key: %+v", err)
	}
	bobWallet, err := NewMockWallet(bobPrivateKey)
	if err != nil {
		s.log.Fatalf("failed to create mock wallet: %+v", err)
	}
	bobProver := NewSyncBalanceProver()

	// transfer 50wei ETH to bob
	recipientAddress, err := intMaxTypes.NewINTMAXAddress(bobWallet.PublicKey().ToAddress().Bytes())
	if err != nil {
		s.log.Fatalf("failed to create recipient address: %+v", err)
	}
	salt, err = new(Salt).SetRandom()
	if err != nil {
		s.log.Fatalf("failed to set random salt: %+v", err)
	}
	transferToBob := intMaxTypes.Transfer{
		Recipient:  recipientAddress,
		TokenIndex: 0,
		Amount:     big.NewInt(50),
		Salt:       salt,
	}

	sendWitness, err := aliceWallet.SendTxAndUpdate(blockBuilder, []*intMaxTypes.Transfer{&transferToBob})
	if err != nil {
		s.log.Fatalf("failed to send tx and update: %+v", err)
	}
	transferWitnesses, ok := aliceWallet.transferWitnesses[sendWitness.GetIncludedBlockNumber()]
	if !ok {
		s.log.Fatalf("failed to get transfer witnesses")
	}
	transferWitness := transferWitnesses[0]

	// update alice balance proof
	err = aliceProver.SyncAll(syncValidityProver, aliceWallet, balanceProcessor, blockBuilder)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	if GetAssetBalance(aliceWallet, 0).Cmp(big.NewInt(50)) != 0 {
		s.log.Fatalf("ETH balance")
	}

	aliceBalanceProof := *aliceProver.LastBalanceProof

	// sync bob wallet to the latest block

	err = bobProver.SyncAll(syncValidityProver, bobWallet, balanceProcessor, blockBuilder)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	// receive transfer and update bob balance proof
	err = bobProver.ReceiveTransfer(bobWallet, balanceProcessor, blockBuilder, transferWitness, aliceBalanceProof)
	if err != nil {
		s.log.Fatalf("failed to receive transfer: %+v", err)
	}

	if GetAssetBalance(bobWallet, 0).Cmp(big.NewInt(50)) != 0 {
		s.log.Fatalf("ETH balance")
	}

	// bob withdraw 10wei ETH
	bobGenericEthAddress, err := intMaxTypes.NewEthereumAddress([]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21})
	if err != nil {
		s.log.Fatalf("failed to create generic eth address: %+v", err)
	}

	salt, err = new(Salt).SetRandom()
	if err != nil {
		s.log.Fatalf("failed to set random salt: %+v", err)
	}
	withdrawal := intMaxTypes.Transfer{
		Recipient:  bobGenericEthAddress,
		TokenIndex: 0,
		Amount:     big.NewInt(10),
		Salt:       salt,
	}

	withdrawalSendWitness, err := bobWallet.SendTxAndUpdate(blockBuilder, []*intMaxTypes.Transfer{&withdrawal})
	if err != nil {
		s.log.Fatalf("failed to send tx and update: %+v", err)
	}
	withdrawalTransferWitnesses := bobWallet.transferWitnesses[withdrawalSendWitness.GetIncludedBlockNumber()]
	fmt.Printf("size of withdrawalTransferWitnesses: %v\n", len(withdrawalTransferWitnesses))

	// update bob balance proof
	err = bobProver.SyncAll(syncValidityProver, bobWallet, balanceProcessor, blockBuilder)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	if GetAssetBalance(bobWallet, 0) != big.NewInt(40) {
		s.log.Fatalf("ETH balance")
	}

	// bobBalanceProof := bobProver.GetBalanceProof()

	// // prove withdrawal
	// withdrawalTransferWitness := withdrawalTransferWitnesses[0]
	// withdrawalProcessor := NewWithdrawalProcessor(balanceProcessor.BalanceCircuit)
	// withdrawalWitness := WithdrawalWitness{
	// 	TransferWitness: withdrawalTransferWitness,
	// 	BalanceProof:    bobBalanceProof,
	// }

	// if withdrawalWitness.ToWithdrawal().Amount != 10 {
	// 	s.log.Fatalf("withdrawal amount")
	// }

	// withdrawalProof, err := withdrawalProcessor.Prove(withdrawalWitness, nil)
	// if err != nil {
	// 	s.log.Fatalf("failed to prove withdrawal: %+v", err)
	// }
}

// type WithdrawalWitness struct {
// 	transferWitness TransferWitness
// 	balanceProof    string
// }

// fn get_asset_balance(wallet: &MockWallet, token_index: u32) -> U256 {
//     let private_state = wallet.get_private_state();
//     assert_eq!(
//         private_state.asset_tree_root,
//         wallet.asset_tree.get_root(),
//         "asset tree root mismatch"
//     );
//     let asset_leaf = wallet.asset_tree.get_leaf(token_index as usize);
//     assert!(!asset_leaf.is_insufficient, "insufficient asset balance");
//     asset_leaf.amount
// }

func GetAssetBalance(wallet *MockWallet, tokenIndex uint32) *big.Int {
	privateState := wallet.PrivateState()
	if !privateState.AssetTreeRoot.Equal(wallet.assetTree.GetRoot()) {
		panic("asset tree root mismatch")
	}
	assetLeaf := wallet.assetTree.GetLeaf(tokenIndex)
	if assetLeaf.IsInsufficient {
		panic("insufficient asset balance")
	}
	return assetLeaf.Amount.BigInt()
}
