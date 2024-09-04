package balance_prover_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
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

func (s *balanceSynchronizerDummy) TestE2E(syncValidityProver *syncValidityProver, blockBuilderWallet *models.Wallet) {
	// blockBuilder := block_validity_prover.NewMockBlockBuilder(s.cfg, s.db)
	balanceProcessor := NewBalanceProcessor(s.ctx, s.cfg, s.log)
	blockBuilder := syncValidityProver.ValidityProver.BlockBuilder()

	alicePrivateKey, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	if err != nil {
		s.log.Fatalf("failed to create private key: %+v", err)
	}
	fmt.Printf("alice public key: %v\n", alicePrivateKey.Public().ToAddress().String())

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

	// fmt.Printf("SyncBlockProver")
	// err = syncValidityProver.ValidityProver.SyncBlockProver()
	// if err != nil {
	// 	s.log.Fatalf("failed to sync block prover: %+v", err)
	// }

	// fmt.Printf("len(b.ValidityProofs) after SetValidityProof: %d\n", len(syncValidityProver.ValidityProver.BlockBuilder().ValidityProofs))

	// sync alice wallet to the latest block, which includes the deposit
	err = aliceProver.SyncAll(syncValidityProver, aliceWallet, balanceProcessor)
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
	fmt.Printf("-----------------ReceiveDeposit----------------------")
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
	fmt.Printf("bob public key: %v\n", bobPrivateKey.Public().ToAddress().String())

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

	fmt.Printf("-----------------SendTxAndUpdate----------------------")
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
	fmt.Printf("-----------------SyncAll alice----------------------")
	err = aliceProver.SyncAll(syncValidityProver, aliceWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	if GetAssetBalance(aliceWallet, 0).Cmp(big.NewInt(50)) != 0 {
		s.log.Fatalf("ETH balance")
	}

	aliceBalanceProof := *aliceProver.LastBalanceProof

	// sync bob wallet to the latest block

	fmt.Printf("-----------------SyncAll bob----------------------")
	err = bobProver.SyncAll(syncValidityProver, bobWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	// receive transfer and update bob balance proof
	fmt.Printf("-----------------ReceiveTransfer----------------------")
	err = bobProver.ReceiveTransfer(bobWallet, balanceProcessor, blockBuilder, transferWitness, aliceBalanceProof)
	if err != nil {
		s.log.Fatalf("failed to receive transfer: %+v", err)
	}

	if GetAssetBalance(bobWallet, 0).Cmp(big.NewInt(50)) != 0 {
		s.log.Fatalf("ETH balance")
	}

	// bob withdraw 10wei ETH
	const (
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"
	)

	bobEthPrivateKey, err := mnemonic_wallet.New().WalletGenerator(
		derivation, mnPassword,
	)
	if err != nil {
		s.log.Fatalf("failed to generate wallet: %+v", err)
	}
	bobGenericEthAddress, err := intMaxTypes.NewEthereumAddress(bobEthPrivateKey.WalletAddress[:])
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
	err = bobProver.SyncAll(syncValidityProver, bobWallet, balanceProcessor)
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
		fmt.Printf("assetTree (wallet): %v\n", wallet.assetTree.GetRoot()) // XXX
		fmt.Printf("assetTreeRoot (privateState): %v\n", privateState.AssetTreeRoot.String())
		panic("asset tree root mismatch")
	}
	assetLeaf := wallet.assetTree.GetLeaf(tokenIndex)
	if assetLeaf.IsInsufficient {
		panic("insufficient asset balance")
	}
	return assetLeaf.Amount.BigInt()
}
