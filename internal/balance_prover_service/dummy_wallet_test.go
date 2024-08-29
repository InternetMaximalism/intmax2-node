package balance_prover_service_test

// func TestDummyWallet(t *testing.T) {
// 	privateKey, err := intMaxAcc.NewPrivateKeyFromString("7397927abf5b7665c4667e8cb8b92e929e287625f79264564bb66c1fa2232b2c")
// 	require.NoError(t, err)

// 	wallet, err := balance_prover_service.NewMockWallet(privateKey)
// 	if err != nil {
// 		t.Errorf("NewMockWallet got err: %v, want: nil", err)
// 	}

// 	got := wallet.PublicKey()
// 	if got == nil {
// 		t.Errorf("PublicKey got: nil, want: not nil")
// 	}

// 	// let rng = &mut rand::thread_rng();
// 	// // shared state
// 	// let mut block_builder = MockBlockBuilder::new();
// 	// let mut sync_validity_prover = SyncValidityProver::<F, C, D>::new();
// 	// let balance_processor = BalanceProcessor::new(sync_validity_prover.validity_circuit());

// 	// // alice deposit
// 	// let mut alice = MockWallet::new_rand(rng);
// 	// let mut alice_balance_prover = SyncBalanceProver::<F, C, D>::new();
// 	// let deposit_amount = U256::rand_small(rng);
// 	// let first_deposit_index = alice.deposit(rng, &mut block_builder, 0, deposit_amount);
// 	// alice.deposit(rng, &mut block_builder, 1, deposit_amount); // dummy deposit

// 	// // post dummy block
// 	// let transfer = Transfer::rand(rng);
// 	// alice.send_tx_and_update(rng, &mut block_builder, &[transfer]);
// 	// alice_balance_prover.sync_send(
// 	// 	&mut sync_validity_prover,
// 	// 	&mut alice,
// 	// 	&balance_processor,
// 	// 	&block_builder,
// 	// );
// 	// let alice_balance_proof = alice_balance_prover.last_balance_proof.clone().unwrap();

// 	// let receive_deposit_witness =
// 	// 	alice.receive_deposit_and_update(rng, &block_builder, first_deposit_index);
// 	// let _new_alice_balance_proof = balance_processor.prove_receive_deposit(
// 	// 	alice.get_pubkey(),
// 	// 	&receive_deposit_witness,
// 	// 	&Some(alice_balance_proof),
// 	// );

// 	blockBuilder := balance_prover_service.MockBlockBuilder{}
// 	// syncValidityProver := balance_prover_service.SyncValidityProver{}

// 	balanceProcessor := balance_prover_service.BalanceProcessor{}
// 	balanceValidityProver := balance_prover_service.NewSyncBalanceProver()

// 	depositAmount := big.NewInt(10000)
// 	salt, err := new(balance_prover_service.Salt).SetRandom()
// 	if err != nil {
// 		t.Errorf("SetRandom got err: %v, want: nil", err)
// 	}
// 	// depositIndex := wallet.Deposit(*salt, blockBuilder, 0, depositAmount)
// 	depositIndex := 3
// 	balanceValidityProver.ReceiveDeposit(wallet, &balanceProcessor, &blockBuilder, depositIndex)

// }
