package balance_synchronizer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"sort"
)

type SyncBalanceProver struct {
	ctx               context.Context
	cfg               *configs.Config
	log               logger.Logger
	storedBalanceData *block_synchronizer.BackupBalanceData
	balanceData       *block_synchronizer.BalanceData
	// LastUpdatedBlockNumber uint32
	lastBalanceProofBody []byte
	LastSenderProof      *string
}

func NewSyncBalanceProver(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *SyncBalanceProver {
	storedBalanceData := new(block_synchronizer.BackupBalanceData)
	return &SyncBalanceProver{
		ctx:               ctx,
		cfg:               cfg,
		log:               log,
		storedBalanceData: storedBalanceData,
		balanceData:       nil,
		// LastUpdatedBlockNumber: 0,
		lastBalanceProofBody: nil,
		LastSenderProof:      nil,
	}
}

func (s *SyncBalanceProver) UploadLastBalanceProof(blockNumber uint32, balanceProof string, wallet UserState) error {
	// s.setLastBalanceProof(blockNumber, balanceProof)
	fmt.Printf("size of balanceProof: %d\n", len(balanceProof))
	compressedBalanceProof, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceProof)
	if err != nil {
		s.log.Fatalf("failed to set last balance proof: %+v", err.Error())
	}
	fmt.Printf("size of compressedBalanceProof.Proof: %d\n", len(compressedBalanceProof.Proof))

	// s.balanceData.PublicState.BlockNumber = blockNumber
	s.lastBalanceProofBody = compressedBalanceProof.Proof
	if s.balanceData == nil {
		fmt.Printf("s.balanceData is initialized\n")
		s.balanceData = new(block_synchronizer.BalanceData)
	}

	assetLeaves := wallet.AssetLeaves()
	assetLeafEntries := make([]*tree.AssetLeafEntry, 0, len(assetLeaves))
	for tokenIndex, leaf := range assetLeaves {
		assetLeafEntries = append(assetLeafEntries, &tree.AssetLeafEntry{
			TokenIndex: tokenIndex,
			Leaf:       leaf,
		})
	}

	s.balanceData.BalanceProofPublicInputs = compressedBalanceProof.PublicInputs
	s.balanceData.NullifierLeaves = wallet.Nullifiers()
	s.balanceData.AssetLeafEntries = assetLeafEntries
	s.balanceData.Nonce = wallet.Nonce()
	s.balanceData.Salt = wallet.Salt()
	s.balanceData.PublicState = wallet.PublicState()

	newBalanceData := new(block_synchronizer.BalanceData).Set(s.balanceData)

	encryptedNewBalanceData, err := newBalanceData.Encrypt(wallet.PublicKey())
	if err != nil {
		return err
	}
	if s.lastBalanceProofBody == nil {
		return errors.New("last balance proof is nil")
	}

	lastBalanceProofBody := base64.StdEncoding.EncodeToString(s.lastBalanceProofBody)

	signature := "0x" // TODO: authentication
	storedBalanceData, err := block_synchronizer.BackupBalanceProof(s.ctx, s.cfg, s.log,
		wallet.PublicKey().ToAddress(), s.storedBalanceData.ID, lastBalanceProofBody, encryptedNewBalanceData,
		s.storedBalanceData.EncryptedTxs, s.storedBalanceData.EncryptedTransfers, s.storedBalanceData.EncryptedDeposits,
		signature, uint64(s.balanceData.PublicState.BlockNumber))
	if err != nil {
		// Fatal error
		return err
	}

	return s.SetEncryptedBalanceData(wallet, storedBalanceData)
}

func (s *SyncBalanceProver) SetEncryptedBalanceData(wallet UserState, storedBalanceData *block_synchronizer.BackupBalanceData) error {
	if storedBalanceData.EncryptedBalanceData == "" {
		return nil
	}

	balanceData, err := wallet.DecryptBalanceData(storedBalanceData.EncryptedBalanceData)
	if err != nil {
		return err
	}

	s.storedBalanceData = storedBalanceData
	s.balanceData = balanceData

	return nil
}

func (s *SyncBalanceProver) LastBalanceProof() *string {
	if s.lastBalanceProofBody == nil {
		return nil
	}

	proof := intMaxTypes.Plonky2Proof{
		PublicInputs: s.balanceData.BalanceProofPublicInputs,
		Proof:        s.lastBalanceProofBody,
	}

	encodedProof := proof.ProofBase64String()

	return &encodedProof
}

// Returns the most recent block number that has been reflected in the balance data.
func (s *SyncBalanceProver) LastUpdatedBlockNumber() uint32 {
	if s.balanceData == nil {
		return 0
	}

	return s.balanceData.PublicState.BlockNumber
}

func (s *SyncBalanceProver) LastBalancePublicInputs() (*balance_prover_service.BalancePublicInputs, error) {
	// if s.LastBalanceProof == nil {
	// 	return nil, errors.New("last balance proof is nil")
	// }

	// balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*s.LastBalanceProof)
	// if err != nil {
	// 	return nil, err
	// }

	// balancePublicInputs, err := new(balance_prover_service.BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
	// if err != nil {
	// 	return nil, err
	// }

	// return balancePublicInputs, nil

	return new(balance_prover_service.BalancePublicInputs).FromPublicInputs(s.balanceData.BalanceProofPublicInputs)
}

func (s *SyncBalanceProver) SyncSend(
	log logger.Logger,
	blockValidityService block_validity_prover.BlockValidityService,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
	latestIntMaxBlockNumber uint32,
) error {
	fmt.Printf("-----SyncSend %s------\n", wallet.PublicKey())

	lastUpdatedBlockNumber := s.LastUpdatedBlockNumber()
	// All block numbers containing transactions sent by the sender,
	// regardless of whether they are valid or not.
	allBlockNumbers := wallet.GetAllBlockNumbers()
	// Things to synchronize from now on:
	// TODO: Transactions sent in invalid blocks do not need to be reflected in the balance proof.
	// Check whether the block is valid. If it has not been posted, remove it.
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", lastUpdatedBlockNumber)
		if blockNumber <= lastUpdatedBlockNumber {
			continue
		}

		if blockNumber > latestIntMaxBlockNumber {
			continue
		}

		notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	for _, blockNumber := range notSyncedBlockNumbers {
		sendWitness, err := wallet.GetSendWitness(blockNumber) // XXX: not need store sendWitness
		if err != nil {
			return errors.New("send witness not found")
		}
		fmt.Printf("sendWitness: %d\n", len(sendWitness.SpentTokenWitness.Transfers))
		for _, transfer := range sendWitness.SpentTokenWitness.Transfers {
			if transfer.Amount.Cmp(big.NewInt(0)) != 0 {
				fmt.Printf("(sendWitness) transfer: %+v\n", transfer)
				fmt.Printf("(sendWitness) transfer amount: %s\n", transfer.Amount)
			}
		}

		// sentBlockNumber := sendWitness.GetIncludedBlockNumber()
		prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
		fmt.Printf("FetchUpdateWitness blockNumber: %d\n", blockNumber)
		currentBlockNumber := blockNumber
		updateWitness, err := blockValidityService.FetchUpdateWitness(
			wallet.PublicKey(),
			currentBlockNumber,
			prevBalancePisBlockNumber,
			true,
		)
		if err != nil {
			return err
		}

		validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(updateWitness.ValidityProof)
		if err != nil {
			return err
		}
		updateWitnessValidityPis := new(block_validity_prover.ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)

		sendWitnessValidityPis := sendWitness.TxWitness.ValidityPis
		if !updateWitnessValidityPis.Equal(&sendWitnessValidityPis) {
			fmt.Printf("update witness validity proof: %v\n", updateWitnessValidityPis)
			fmt.Printf("update witness public state: %v\n", updateWitnessValidityPis.PublicState)
			fmt.Printf("update witness account tree root: %v\n", updateWitnessValidityPis.PublicState.PrevAccountTreeRoot)
			fmt.Printf("update witness account tree root: %v\n", updateWitnessValidityPis.PublicState.AccountTreeRoot)
			fmt.Printf("send witness validity proof: %v\n", sendWitnessValidityPis)
			fmt.Printf("send witness public state: %v\n", sendWitnessValidityPis.PublicState)
			fmt.Printf("send witness account tree root: %v\n", sendWitnessValidityPis.PublicState.PrevAccountTreeRoot)
			fmt.Printf("send witness account tree root: %v\n", sendWitnessValidityPis.PublicState.AccountTreeRoot)
			return errors.New("update witness validity proof is not equal to send witness validity proof")
		}

		if updateWitnessValidityPis.IsValidBlock {
			fmt.Printf("Block %d is valid", updateWitnessValidityPis.PublicState.BlockNumber)
		} else {
			fmt.Printf("Block %d is invalid", updateWitnessValidityPis.PublicState.BlockNumber)
		}

		// TODO: ValidateTxInclusionValue
		// _, err = ValidateTxInclusionValue(
		// 	sendWitness.PrevBalancePis.PubKey,
		// 	sendWitness.PrevBalancePis.PublicState,
		// 	updateWitness.ValidityProof,
		// 	&updateWitness.BlockMerkleProof,
		// 	updateWitness.AccountMembershipProof,
		// 	sendWitness.TxWitness.TxIndex,
		// 	sendWitness.TxWitness.Tx,
		// 	&intMaxTree.MerkleProof{Siblings: sendWitness.TxWitness.TxMerkleProof},
		// 	// senderLeaf,
		// 	// senderMerkleProof,
		// )
		// if err != nil {
		// 	return err
		// }

		balanceProof, err := balanceProcessor.ProveSend(
			wallet.PublicKey(),
			sendWitness,
			updateWitness,
			s.LastBalanceProof(),
		)
		if err != nil {
			return fmt.Errorf("failed to prove send: %w", err)
		}

		// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
		// if err != nil {
		// 	return err
		// }

		fmt.Printf("s.LastUpdatedBlockNumber before SyncSend: %d\n", s.LastUpdatedBlockNumber())
		// s.LastUpdatedBlockNumber = blockNumber
		err = s.UploadLastBalanceProof(blockNumber, balanceProof.Proof, wallet)
		if err != nil {
			return fmt.Errorf("failed to upload last balance proof in SyncSend: %w", err)
		}

		fmt.Printf("s.LastUpdatedBlockNumber after SyncSend: %d\n", s.LastUpdatedBlockNumber())
		wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)
	}

	return nil
}

// This function returns an error if there is at least one transaction included in a block
// between the last synchronized block number and the block number you want to synchronize to,
// where the transaction was sent by you.
func ShouldSyncNoSend(
	lastUpdatedBlockNumber uint32,
	blockNumber uint32,
	allBlockNumbers []uint32,
) error {
	if lastUpdatedBlockNumber == 0 {
		return nil
	}

	for _, targetBlockNumber := range allBlockNumbers {
		if targetBlockNumber > blockNumber {
			break
		}

		if lastUpdatedBlockNumber < targetBlockNumber {
			return fmt.Errorf("block number is less than last updated block number: %d", targetBlockNumber)
		}
	}

	if blockNumber <= lastUpdatedBlockNumber {
		return fmt.Errorf("block number is less than or equal to last updated block number: %d", blockNumber)
	}

	return nil
}

// SyncNoSend synchronizes the balance prover state without sending any transactions.
// It verifies that the balance prover's last updated block number is consistent with the wallet's block numbers
// and updates the balance prover state if necessary.
func (s *SyncBalanceProver) SyncNoSend(
	log logger.Logger,
	blockValidityService block_validity_prover.BlockValidityService,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
	blockNumber uint32,
) error {
	fmt.Printf("-----SyncNoSend %s------\n", wallet.PublicKey())

	lastUpdatedBlockNumber := s.LastUpdatedBlockNumber()
	allBlockNumbers := wallet.GetAllBlockNumbers()
	fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", lastUpdatedBlockNumber)
	err := ShouldSyncNoSend(lastUpdatedBlockNumber, blockNumber, allBlockNumbers) // XXX
	if err != nil {
		return err
	}

	var prevBalancePis *balance_prover_service.BalancePublicInputs
	lastBalanceProof := s.LastBalanceProof()
	if lastBalanceProof != nil {
		fmt.Println("s.LastBalanceProof != nil")
		var lastBalanceProofWithPis *intMaxTypes.Plonky2Proof
		lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*lastBalanceProof)
		if err != nil {
			return err
		}
		prevBalancePis, err = new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("NewBalancePublicInputsWithPublicKey")
		prevBalancePis = balance_prover_service.NewBalancePublicInputsWithPublicKey(wallet.PublicKey())
	}

	prevPublicState := prevBalancePis.PublicState
	fmt.Printf("blockNumber (FetchUpdateWitness): %d\n", blockNumber)
	fmt.Printf("prevBalancePis.PublicState.BlockNumber (FetchUpdateWitness): %d\n", prevPublicState.BlockNumber)
	fmt.Printf("lastUpdatedBlockNumber (FetchUpdateWitness): %d\n", lastUpdatedBlockNumber)
	if prevPublicState.BlockNumber != lastUpdatedBlockNumber {
		panic("broken balance data")
	}
	if blockNumber == prevPublicState.BlockNumber {
		fmt.Printf("no need to update balance proof: %d\n", blockNumber)
		return nil
	}
	updateWitness, err := blockValidityService.FetchUpdateWitness(
		wallet.PublicKey(),
		blockNumber,
		prevPublicState.BlockNumber, // XXX: lastUpdatedBlockNumber?
		false,
	)
	if err != nil {
		return err
	}

	validityProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(updateWitness.ValidityProof)
	if err != nil {
		return err
	}
	validityPis := new(block_validity_prover.ValidityPublicInputs).FromPublicInputs(validityProofWithPis.PublicInputs)
	currentBlockNumber := validityPis.PublicState.BlockNumber

	// DEBUG
	// prevBalancePisJSON, err := json.Marshal(prevBalancePis)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("prevBalancePisJSON: %s", prevBalancePisJSON)

	lastSentTxBlockNumber := updateWitness.AccountMembershipProof.GetLeaf()
	if lastSentTxBlockNumber > uint64(1)<<32 {
		panic("last sent tx block number is invalid")
	}
	fmt.Printf("lastSentTxBlockNumber: %d\n", lastSentTxBlockNumber)
	fmt.Printf("prevPublicState.BlockNumber: %d\n", prevPublicState.BlockNumber)
	// if uint32(lastSentTxBlockNumber) > prevPublicState.BlockNumber && uint32(lastSentTxBlockNumber) < blockNumber {
	// 	// This indicates that there are unsynchronized transitions that need to be processed in advance.
	// 	return errors.New("last block number is greater than prev public state block number")
	// }
	if uint32(lastSentTxBlockNumber) > prevPublicState.BlockNumber {
		// This indicates that there are unsynchronized transitions that need to be processed in advance.
		return errors.New("last block number is greater than prev public state block number")
	}

	// TODO: blockHashLeaf := blockHistory.BlockHashTree.Leaves[leafBlockNumber]
	blockHashLeaf := tree.NewBlockHashLeaf(prevPublicState.BlockHash)
	fmt.Printf("blockHash (SyncNoSend): %s\n", prevPublicState.BlockHash)
	fmt.Printf("blockHashLeafHash (SyncNoSend): %s\n", blockHashLeaf.Hash())
	fmt.Printf("prevPublicState.BlockNumber (SyncNoSend): %d\n", prevPublicState.BlockNumber)
	fmt.Printf("leafBlockNumber (SyncNoSend): %d\n", lastUpdatedBlockNumber)
	fmt.Printf("blockTreeRoot (SyncNoSend): %s\n", validityPis.PublicState.BlockTreeRoot.String())
	err = updateWitness.BlockMerkleProof.Verify(
		blockHashLeaf.Hash(), // leaf of the pre-updated block tree
		int(lastUpdatedBlockNumber),
		validityPis.PublicState.BlockTreeRoot,
	)
	if err != nil {
		return fmt.Errorf("block merkle proof is invalid: %w", err)
	}

	balanceProof, err := balanceProcessor.ProveUpdate(
		wallet.PublicKey(),
		updateWitness,
		s.LastBalanceProof(),
	)
	if err != nil {
		return fmt.Errorf("failed to prove update: %w", err)
	}

	// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
	// if err != nil {
	// 	return err
	// }

	fmt.Printf("s.LastUpdatedBlockNumber before SyncNoSend: %d\n", s.LastUpdatedBlockNumber())
	// s.LastUpdatedBlockNumber = currentBlockNumber
	err = s.UploadLastBalanceProof(currentBlockNumber, balanceProof.Proof, wallet)
	if err != nil {
		return fmt.Errorf("failed to upload last balance proof in SyncNoSend: %w", err)
	}

	fmt.Printf("s.LastUpdatedBlockNumber after SyncNoSend: %d\n", s.LastUpdatedBlockNumber())
	wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)

	return nil
}

func (s *SyncBalanceProver) SyncAll(
	log logger.Logger,
	blockValidityService *block_validity_prover.BlockValidityProverMemory,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
) (err error) {
	// latestIntMaxBlockNumber, err := blockValidityService.LastPostedBlockNumber()
	latestIntMaxBlockNumber, err := blockValidityService.LatestSynchronizedBlockNumber()
	if err != nil {
		return err
	}
	fmt.Printf("LatestWitnessNumber before SyncSend: %d\n", latestIntMaxBlockNumber)

	err = s.SyncSend(log, blockValidityService, blockSynchronizer, wallet, balanceProcessor, latestIntMaxBlockNumber)
	if err != nil {
		return err
	}
	err = s.SyncNoSend(log, blockValidityService, blockSynchronizer, wallet, balanceProcessor, latestIntMaxBlockNumber)
	if err != nil {
		return err
	}

	return nil
}

func (s *SyncBalanceProver) ReceiveDeposit(
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
	// blockBuilder MockBlockBuilder,
	blockValidityService block_validity_prover.BlockValidityService,
	depositIndex uint32,
) error {
	receiveDepositWitness, err := wallet.ReceiveDepositAndUpdate(blockValidityService, depositIndex)
	if err != nil {
		if err.Error() == ErrNullifierAlreadyExists.Error() {
			return ErrNullifierAlreadyExists
		}

		return errors.Join(ErrReceiveDepositAndUpdate, err)
	}
	fmt.Println("start ProveReceiveDeposit")
	lastBalanceProof := *s.LastBalanceProof()
	lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(lastBalanceProof)
	if err != nil {
		// fmt.Printf("lastBalanceProof: %s\n", lastBalanceProof)
		fmt.Printf("size of lastBalanceProof: %d\n", len(lastBalanceProof))
		return errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
	}

	lastBalancePublicInputs, err := new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
	if err != nil {
		return errors.Join(ErrBalancePublicInputsFromPublicInputs, err)
	}
	fmt.Printf("lastBalancePublicInputs (ReceiveDeposit) PrivateCommitment commitment: %s\n", lastBalancePublicInputs.PrivateCommitment.String())

	// validation
	{
		depositIndex := receiveDepositWitness.DepositWitness.DepositIndex
		deposit := receiveDepositWitness.DepositWitness.Deposit
		depositMerkleProof := receiveDepositWitness.DepositWitness.DepositMerkleProof
		depositTreeRoot := lastBalancePublicInputs.PublicState.DepositTreeRoot
		userDepositTreeRoot := wallet.PublicState().DepositTreeRoot
		if depositTreeRoot != userDepositTreeRoot {
			s.log.Debugf("depositIndex: %d\n", depositIndex)
			s.log.Debugf("depositTreeRoot in balance proof: %s\n", depositTreeRoot.String())
			s.log.Debugf("DepositTreeRoot in public state: %s\n", userDepositTreeRoot.String())
			panic("deposit tree root is mismatch")
		}

		if depositMerkleProof.Verify(deposit.Hash(), int(depositIndex), depositTreeRoot) != nil {
			panic("deposit merkle proof is invalid") // XXX
		}
	}

	balanceProof, err := balanceProcessor.ProveReceiveDeposit(
		wallet.PublicKey(),
		receiveDepositWitness,
		&lastBalanceProof,
	)
	if err != nil {
		return errors.Join(ErrProveReceiveDeposit, err)
	}

	lastBalanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceProof.Proof)
	if err != nil {
		return errors.Join(ErrNewCompressedPlonky2ProofFromBase64StringFail, err)
	}

	lastBalancePublicInputs, err = new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
	if err != nil {
		return errors.Join(ErrBalancePublicInputsFromPublicInputs, err)
	}

	fmt.Printf("ReceiveDeposit PrivateCommitment commitment (after): %s\n", lastBalancePublicInputs.PrivateCommitment.String())
	fmt.Printf("ReceiveDeposit PrivateCommitment commitment (after, public inputs): %s\n", balanceProof.PublicInputs.PrivateCommitment.String())
	fmt.Printf("wallet private state: %+v\n", wallet.PrivateState())
	fmt.Printf("wallet private state commitment: %s\n", wallet.PrivateState().Commitment().String())

	fmt.Println("finish ProveReceiveDeposit")

	return s.UploadLastBalanceProof(s.LastUpdatedBlockNumber(), balanceProof.Proof, wallet)
}

func (s *SyncBalanceProver) ReceiveTransfer(
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
	// blockBuilder MockBlockBuilder,
	blockValidityService block_validity_prover.BlockValidityService,
	transferWitness *intMaxTypes.TransferWitness,
	senderLastBalanceProof string,
	senderBalanceTransitionProof string,
) error {
	fmt.Printf("ReceiveTransfer s.LastUpdatedBlockNumber: %d\n", s.LastUpdatedBlockNumber())
	receiveTransferWitness, err := wallet.ReceiveTransferAndUpdate(
		blockValidityService,
		s.LastUpdatedBlockNumber(),
		transferWitness,
		senderLastBalanceProof,
		senderBalanceTransitionProof,
	)
	if err != nil {
		if err.Error() == ErrNullifierAlreadyExists.Error() {
			return ErrNullifierAlreadyExists
		}

		return err
	}
	balanceProof, err := balanceProcessor.ProveReceiveTransfer(
		wallet.PublicKey(),
		receiveTransferWitness,
		s.LastBalanceProof(),
	)
	if err != nil {
		return err
	}

	// s.LastBalanceProof = &balanceProof.Proof
	return s.UploadLastBalanceProof(s.LastUpdatedBlockNumber(), balanceProof.Proof, wallet)
}

// func (s *SyncBalanceProver) SyncBalanceProof(
// 	ctx context.Context,
// 	cfg *configs.Config,
// 	publicKey *intMaxAcc.PublicKey,
// ) error {
// 	userAllData, err := balance_service.GetUserBalancesRawRequest(ctx, cfg, publicKey.ToAddress().String())
// 	if err != nil {
// 		return fmt.Errorf("failed to get user balances: %w", err)
// 	}
// balanceProverService := NewBalanceProverService(s.ctx, s.cfg, s.log, blockBuilderWallet)

// 	return nil
// }

// type balanceSynchronizer struct {
//     ctx context.Context
//     cfg *configs.Config
//     log logger.Logger
//     sb  block_validity_prover.ServiceBlockchain
//     db  block_validity_prover.SQLDriverApp
// }

// type syncValidityProver struct {
//     log               logger.Logger
//     ValidityProver    block_validity_prover.BlockValidityProver
//     blockSynchronizer block_validity_prover.BlockSynchronizer
// }

type BalanceSynchronizer interface {
	CurrentNonce() uint32
	LastBalanceProof() *intMaxTypes.Plonky2Proof
	ProveSendTransition(spentTokenWitness *balance_prover_service.SpentTokenWitness) (string, error)
}

func SyncUserBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	blockValidityService block_validity_prover.BlockValidityService,
	userWalletState UserState,
) (BalanceSynchronizer, error) {
	blockSynchronizer, err := block_synchronizer.NewBlockSynchronizer(
		ctx, cfg, log,
	)
	if err != nil {
		const msg = "failed to get Block Synchronizer: %+v"
		return nil, fmt.Errorf(msg, err.Error())
	}

	syncBalanceProver := NewSyncBalanceProver(ctx, cfg, log)

	balanceProcessor := balance_prover_service.NewBalanceProcessor(
		ctx, cfg, log,
	)
	balanceSynchronizer := NewSynchronizer(ctx, cfg, log, sb, blockSynchronizer, blockValidityService, balanceProcessor, syncBalanceProver, userWalletState)

	// timeout := 5 * time.Second
	// ticker := time.NewTicker(timeout)
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		ticker.Stop()
	// 		log.Warnf("Received cancel signal from context, stopping...")
	// 		return nil, errors.New("received cancel signal from context")
	// 	case <-ticker.C:
	err = balanceSynchronizer.syncProcessing(userWalletState.PrivateKey())
	if err != nil {
		if errors.Is(err, ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber) ||
			errors.Is(err, ErrNoValidUserData) {
			return balanceSynchronizer, nil
		}

		// if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
		// 	return nil, err
		// }

		const msg = "failed to start sync processing: %+v"
		log.Fatalf(msg, err.Error())
	}

	return balanceSynchronizer, nil
}
