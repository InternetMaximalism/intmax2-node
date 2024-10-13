package balance_synchronizer_test

import (
	"context"
	"encoding/json"
	"testing"

	"intmax2-node/configs"
	"intmax2-node/internal/balance_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMakeTxWitness(t *testing.T) {
	require.NoError(t, configs.LoadDotEnv(2))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, cancel := context.WithCancel(context.Background())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	var err error

	blockValidityService := NewMockBlockValidityService(ctrl)

	validityInputsJson := "{\"PublicState\":{\"blockTreeRoot\":\"0xa2ece33da0b9fdc6b30f5b9c7ef84e42b57ada3b15b98c9bdea14cc2495496cd\",\"prevAccountTreeRoot\":\"0x35268f53bf28368c1b69f12d27c9bf2061328fb5d20d47c1a536cb8fb0d32cdb\",\"accountTreeRoot\":\"0x6af7d9349f55d75cf8a8abdd9586b7636ff4c930fdea8899aad3195abdc31e53\",\"depositTreeRoot\":\"0x9f3ae6197fb367369f0123d6048470424eac47b01f008a63929a25cefb78e090\",\"blockHash\":\"0x52ca16361059b0b483d22a2e247ab41bcf96c40e4f2e9b29fe56b713b9936635\",\"blockNumber\":28},\"TxTreeRoot\":\"0x408f41d61d4a433e077cadb0e0128cdfffbf32d24c2d0e00ed03d089ce8eb946\",\"SenderTreeRoot\":\"0xcf5712def172c0af417a1415d669b595693f55ff2767c3e76e8178e804e7d99e\",\"IsValidBlock\":true}"
	validityInputs := new(block_validity_prover.ValidityPublicInputs)
	err = json.Unmarshal([]byte(validityInputsJson), validityInputs)
	require.NoError(t, err)

	senderLeavesJson := "[{\"sender\":\"18938402705870056724957543220940828166011843165257241232798027402818102547213\",\"isValid\":true},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false},{\"sender\":\"1\",\"isValid\":false}]"
	var senderLeaves []block_validity_prover.SenderLeaf
	err = json.Unmarshal([]byte(senderLeavesJson), &senderLeaves)
	require.NoError(t, err)

	blockValidityService.EXPECT().ValidityPublicInputs(gomock.Any()).Return(validityInputs, senderLeaves, nil).AnyTimes()

	marshaledTxDetailsJson := "{\"TransferTreeRoot\":\"0xbfb767e73a1451dcc761b5ba55baef585f168bf555107c35666e664039d47e98\",\"Nonce\":1,\"Transfers\":[{\"Recipient\":{\"TypeOfAddress\":\"INTMAX\",\"Address\":\"IzOxW4oMM6tGjNxyL6hztkRI9tFHfqE/H5Y9pLO8KFg=\"},\"TokenIndex\":0,\"Amount\":375128738252125,\"Salt\":\"0xab3ff04b77ed712df5c9fb8b88e7cd5964df5e033340d811b7986cdc9ad7a4dd\"},{\"Recipient\":{\"TypeOfAddress\":\"INTMAX\",\"Address\":\"Kd7BjgIR3ua7erfZz074PmUEgJDl9zqTqb8BgAqQ0w0=\"},\"TokenIndex\":0,\"Amount\":1001,\"Salt\":\"0x6eef8134772453b4ee8868e14c03749731b48a152f7e1b2a000e6bb2a150aa92\"}],\"TxTreeRoot\":\"0x408f41d61d4a433e077cadb0e0128cdfffbf32d24c2d0e00ed03d089ce8eb946\",\"TxIndex\":0,\"TxMerkleProof\":[\"0x0000000000000000000000000000000000000000000000000000000000000000\",\"0x3c18a9786cb0b359c4055e3364a246c37953db0ab48808f4c71603f33a1144ca\",\"0xb61a4ad1aaf14fcc8d85581b9901b2cfde8f9ad762a30af22196fc41328ae503\",\"0x61e00af7295ce05a9a247cc59da2de6446fb94bfe956c05f67703a0cc73ca542\",\"0x5154921a064626448cc1b30d2e2c0947167d7cf3bf854d27f522eaa0af88a040\",\"0xe9e6b5c8d15b61ae86ac9b34bd97d439b77e23f0fc590197d0053597686f6672\",\"0x25225f1a5d49614a5a1d2a648eee8f03dda8f741c47dfb1049561260080d30c3\"]}"
	txDetails := new(intMaxTypes.TxDetails)
	err = json.Unmarshal([]byte(marshaledTxDetailsJson), txDetails)
	require.NoError(t, err)

	txWitness, transferWitness, err := balance_synchronizer.MakeTxWitness(blockValidityService, txDetails)
	require.NoError(t, err)

	t.Logf("txWitness: %+v\n", txWitness)
	for i, tw := range transferWitness {
		t.Logf("transferWitness[%d]: %+v\n", i, tw)
	}
}

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
