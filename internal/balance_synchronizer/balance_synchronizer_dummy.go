package balance_synchronizer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/mnemonic_wallet/models"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/withdrawal_service"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

func (s *balanceSynchronizerDummy) TestE2E(
	blockValidityProver *block_validity_prover.BlockValidityProverMemory,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	blockBuilderWallet *models.Wallet,
	withdrawalAggregator *withdrawal_service.WithdrawalAggregatorService,
) {
	withdrawalWitness, err := s.TestE2EWithoutWithdrawal(blockValidityProver, blockSynchronizer, blockBuilderWallet, withdrawalAggregator)
	if err != nil {
		s.log.Fatalf("failed to test e2e: %+v", err)
		return
	}
	withdrawalWitnessInput := new(withdrawal_service.WithdrawalWitnessInput).FromWithdrawalWitness(withdrawalWitness)

	withdrawalWitnessJSON, err := json.Marshal(withdrawalWitnessInput)
	if err != nil {
		s.log.Fatalf("failed to marshal withdrawal witness: %+v", err)
	}
	fmt.Printf("withdrawalWitnessJSON: %s\n", withdrawalWitnessJSON)

	// withdrawalWitness := new(withdrawal_service.WithdrawalWitnessInput)
	// err := json.Unmarshal([]byte(EncodedWithdrawalWitness), &withdrawalWitness)
	// if err != nil {
	// 	s.log.Fatalf("failed to unmarshal withdrawal witness: %+v", err)
	// }

	withdrawalProcessor := NewWithdrawalProcessor(withdrawalAggregator)
	withdrawalProofJSON, err := withdrawalProcessor.Prove(withdrawalWitnessInput, nil)
	if err != nil {
		s.log.Fatalf("failed to prove withdrawal: %+v", err)
	}

	s.log.Debugf("withdrawal proof: %v\n", withdrawalProofJSON)

	s.log.Infof("Done")
}

func (s *balanceSynchronizerDummy) TestE2EWithoutWithdrawal(
	blockValidityProver *block_validity_prover.BlockValidityProverMemory,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	blockBuilderWallet *models.Wallet,
	withdrawalAggregator *withdrawal_service.WithdrawalAggregatorService,
) (*withdrawal_service.WithdrawalWitness, error) {
	balanceProcessor := balance_prover_service.NewBalanceProcessor(s.ctx, s.cfg, s.log)
	blockBuilder := blockValidityProver.BlockBuilder()

	alicePrivateKey, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	if err != nil {
		s.log.Fatalf("failed to create private key: %+v", err)
	}
	fmt.Printf("alice public key: %v\n", alicePrivateKey.Public().ToAddress().String())

	aliceWallet, err := NewMockWallet(alicePrivateKey)
	if err != nil {
		s.log.Fatalf("failed to create mock wallet: %+v", err)
	}
	aliceProver := NewSyncBalanceProver(s.ctx, s.cfg, s.log)

	salt, err := new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		s.log.Fatalf("failed to set random salt: %+v", err)
	}

	// depost 100wei ETH to alice wallet
	depositIndex := aliceWallet.Deposit(blockBuilder, *salt, 0, big.NewInt(100))

	// post dummy block to reflect the deposit tree
	emptyBlockContent, err := NewBlockContentFromTxRequests(true, []*block_validity_prover.MockTxRequest{})
	if err != nil {
		s.log.Fatalf("failed to create block content: %+v", err)
	}

	lastValidityWitness, err := blockBuilder.LastValidityWitness()
	if err != nil {
		var ErrLastValidityWitnessNotFound = errors.New("last validity witness not found")
		return nil, errors.Join(ErrLastValidityWitnessNotFound, err)
	}
	_, err = blockBuilder.UpdateValidityWitness(emptyBlockContent, lastValidityWitness)
	if err != nil {
		s.log.Fatalf("failed to post block: %+v", err)
	}

	// sync alice wallet to the latest block, which includes the deposit
	err = aliceProver.SyncAll(s.log, blockValidityProver, blockSynchronizer, aliceWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	balancePis, err := aliceProver.LastBalancePublicInputs()
	if err != nil {
		s.log.Fatalf("failed to get balance public inputs: %+v", err)
	}

	if balancePis.PublicState.BlockNumber != 1 {
		s.log.Fatalf("balance proof synced to block 1")
	}

	// receive deposit and update alice balance proof
	fmt.Printf("-----------------ReceiveDeposit----------------------")
	err = aliceProver.ReceiveDeposit(aliceWallet, balanceProcessor, blockValidityProver, depositIndex)
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
	bobProver := NewSyncBalanceProver(s.ctx, s.cfg, s.log)

	// transfer 50wei ETH to bob
	recipientAddress, err := intMaxTypes.NewINTMAXAddress(bobWallet.PublicKey().ToAddress().Bytes())
	if err != nil {
		s.log.Fatalf("failed to create recipient address: %+v", err)
	}
	salt, err = new(balance_prover_service.Salt).SetRandom()
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
	sendWitness, err := aliceWallet.SendTxAndUpdate(blockValidityProver, []*intMaxTypes.Transfer{&transferToBob})
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
	err = aliceProver.SyncAll(s.log, blockValidityProver, blockSynchronizer, aliceWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	if GetAssetBalance(aliceWallet, 0).Cmp(big.NewInt(50)) != 0 {
		s.log.Fatalf("ETH balance 2")
	}

	aliceLastBalanceProof := *aliceProver.LastBalanceProof()
	aliceSenderTransitionProof := *aliceProver.LastSenderProof

	// sync bob wallet to the latest block

	fmt.Printf("-----------------SyncAll bob----------------------")
	err = bobProver.SyncAll(s.log, blockValidityProver, blockSynchronizer, bobWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	// receive transfer and update bob balance proof
	fmt.Printf("-----------------ReceiveTransfer----------------------")
	err = bobProver.ReceiveTransfer(bobWallet, balanceProcessor, blockValidityProver, transferWitness, aliceLastBalanceProof, aliceSenderTransitionProof)
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

	salt, err = new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		s.log.Fatalf("failed to set random salt: %+v", err)
	}
	withdrawal := intMaxTypes.Transfer{
		Recipient:  bobGenericEthAddress,
		TokenIndex: 0,
		Amount:     big.NewInt(10),
		Salt:       salt,
	}

	withdrawalSendWitness, err := bobWallet.SendTxAndUpdate(blockValidityProver, []*intMaxTypes.Transfer{&withdrawal})
	if err != nil {
		s.log.Fatalf("failed to send tx and update: %+v", err)
	}
	withdrawalTransferWitnesses := bobWallet.transferWitnesses[withdrawalSendWitness.GetIncludedBlockNumber()]
	fmt.Printf("size of withdrawalTransferWitnesses: %v\n", len(withdrawalTransferWitnesses))

	// update bob balance proof
	err = bobProver.SyncAll(s.log, blockValidityProver, blockSynchronizer, bobWallet, balanceProcessor)
	if err != nil {
		s.log.Fatalf("failed to sync all: %+v", err)
	}

	if GetAssetBalance(bobWallet, 0).Cmp(big.NewInt(40)) != 0 {
		s.log.Fatalf("ETH balance 4")
	}

	// prove withdrawal
	withdrawalTransferWitness := withdrawalTransferWitnesses[0]
	// bobBalanceProof, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*bobProver.LastBalanceProof)
	// if err != nil {
	// 	s.log.Fatalf("failed to create balance proof: %+v", err)
	// }

	withdrawalWitness := withdrawal_service.WithdrawalWitness{
		TransferWitness: &intMaxTypes.TransferWitness{
			Tx:                  withdrawalTransferWitness.Tx,
			Transfer:            withdrawalTransferWitness.Transfer,
			TransferIndex:       withdrawalTransferWitness.TransferIndex,
			TransferMerkleProof: withdrawalTransferWitness.TransferMerkleProof,
		},
		BalanceProof: bobProver.LastBalanceProof(),
	}

	// if withdrawalWitness.ToWithdrawal().Amount != 10 {
	// 	s.log.Fatalf("withdrawal amount")
	// }

	return &withdrawalWitness, nil
}

type WithdrawalProcessor struct {
	withdrawalAggregator *withdrawal_service.WithdrawalAggregatorService
}

func NewWithdrawalProcessor(
	withdrawalAggregator *withdrawal_service.WithdrawalAggregatorService,
) *WithdrawalProcessor {
	return &WithdrawalProcessor{withdrawalAggregator}
}

func BuildSubmitWithdrawalProofData(
	w *withdrawal_service.WithdrawalAggregatorService,
	pendingWithdrawals []withdrawal_service.WithdrawalWitnessInput,
	withdrawalAggregator common.Address,
) (*withdrawal_service.GnarkGetProofResponseResult, error) {
	prevWithdrawalProof := new(string)

	if len(pendingWithdrawals) == 0 {
		return nil, fmt.Errorf("no pending withdrawals")
	}

	for i := range pendingWithdrawals {
		// withdrawalWitness := new(withdrawal_service.WithdrawalWitnessInput).FromWithdrawalWitness(&pendingWithdrawals[i])
		withdrawalProof, err := w.RequestWithdrawalProofToProver(&pendingWithdrawals[i], prevWithdrawalProof)
		if err != nil {
			return nil, fmt.Errorf("failed to request withdrawal proof to prover: %w", err)
		}

		prevWithdrawalProof = withdrawalProof
	}

	withdrawalWrapperProof, err := w.RequestWithdrawalWrapperProofToProver(*prevWithdrawalProof, withdrawalAggregator)
	if err != nil {
		return nil, fmt.Errorf("failed to request withdrawal wrapper proof to prover: %w", err)
	}

	gnarkProof, err := w.RequestWithdrawalGnarkProofToProver(withdrawalWrapperProof)
	if err != nil {
		return nil, fmt.Errorf("failed to request withdrawal gnark proof to prover: %w", err)
	}

	return gnarkProof, nil
}

func (p *WithdrawalProcessor) Prove(withdrawalWitness *withdrawal_service.WithdrawalWitnessInput, prevWithdrawalProof *string) (*withdrawal_service.GnarkGetProofResponseResult, error) {
	// encodedWithdrawalWitness, err := json.Marshal(withdrawalWitness)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal withdrawal witness: %w", err)
	// }
	// fmt.Printf("encodedWithdrawalWitness: %s\n", encodedWithdrawalWitness)

	aggregator := common.Address{}
	pendingWithdrawals := []withdrawal_service.WithdrawalWitnessInput{*withdrawalWitness}
	withdrawalWrapperProof, err := BuildSubmitWithdrawalProofData(p.withdrawalAggregator, pendingWithdrawals, aggregator)
	if err != nil {
		return nil, fmt.Errorf("failed to build submit withdrawal proof data: %w", err)
	}
	fmt.Printf("withdrawalWrapperProof: %v\n", withdrawalWrapperProof)

	// withdrawalWrapperProof :=
	// p.withdrawalAggregator.RequestWithdrawalProofToProver(
	// 	withdrawalWitness.TransferWitness, withdrawalWitness.BalanceProof,
	// )

	// wrappedProof, err := p.withdrawalAggregator.RequestWithdrawalWrapperProofToProver(
	// 	withdrawalProof, withdrawalAggregator,
	// )

	return withdrawalWrapperProof, nil
}

func GetAssetBalance(wallet *mockWallet, tokenIndex uint32) *big.Int {
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
