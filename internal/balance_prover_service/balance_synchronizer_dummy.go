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
	"intmax2-node/internal/withdrawal_service"
	"math/big"
)

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
		s.log.Fatalf("ETH balance 1")
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
		s.log.Fatalf("ETH balance 2")
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
		s.log.Fatalf("ETH balance 3")
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

	if GetAssetBalance(bobWallet, 0).Cmp(big.NewInt(40)) != 0 {
		s.log.Fatalf("ETH balance 4")
	}

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

	// prove withdrawal
	withdrawalTransferWitness := withdrawalTransferWitnesses[0]
	withdrawalProcessor := NewWithdrawalProcessor(balanceProcessor.BalanceCircuit)
	withdrawalWitness := WithdrawalWitness{
		TransferWitness: withdrawalTransferWitness,
		BalanceProof:    bobBalanceProof,
	}

	if withdrawalWitness.ToWithdrawal().Amount != 10 {
		s.log.Fatalf("withdrawal amount")
	}

	withdrawalProof, err := withdrawalProcessor.Prove(withdrawalWitness, nil)
	if err != nil {
		s.log.Fatalf("failed to prove withdrawal: %+v", err)
	}
}

// pub fn prove(
// 	&self,
// 	withdrawal_witness: &WithdrawalWitness<F, C, D>,
// 	prev_withdrawal_proof: &Option<ProofWithPublicInputs<F, C, D>>,
// ) -> Result<ProofWithPublicInputs<F, C, D>> {

type WithfrawalProcessor struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
}

func NewWithfrawalProcessor(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *WithfrawalProcessor {
	return &WithfrawalProcessor{ctx, cfg, log}
}

// WithdrawalRequestRequest
//

func (p *WithfrawalProcessor) Prove(withdrawalWitness WithdrawalWitness, prevWithdrawalProof *ProofWithPublicInputs) (string, error) {
	WithdrawalRequestRequest
	withdrawalAggregator, err := withdrawal_service.NewWithdrawalAggregatorService(
		p.ctx,
		p.cfg,
		p.log,
		p.db,
		p.sb,
	)
	if err != nil {
		return "", err
	}

	withdrawalAggregator.RequestWithdrawalWrapperProofToProver

	return nil, nil
}

type WithdrawalWitness struct {
	transferWitness TransferWitness
	balanceProof    string
}

func GetAssetBalance(wallet *MockWallet, tokenIndex uint32) *big.Int {
	privateState := wallet.PrivateState()
	if !privateState.AssetTreeRoot.Equal(wallet.assetTree.GetRoot()) {
		fmt.Printf("assetTree (wallet): %v\n", wallet.assetTree.GetRoot())
		fmt.Printf("assetTreeRoot (privateState): %v\n", privateState.AssetTreeRoot.String())
		panic("asset tree root mismatch")
	}
	assetLeaf := wallet.assetTree.GetLeaf(tokenIndex)
	if assetLeaf.IsInsufficient {
		panic("insufficient asset balance")
	}
	return assetLeaf.Amount.BigInt()
}
