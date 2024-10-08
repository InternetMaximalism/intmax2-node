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

	"sort"
	"time"
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

// type SyncBalanceProverInterface interface {
// 	BalancePublicInputs() (*BalancePublicInputs, error)
// 	SyncSend(
// 		syncValidityProver *syncValidityProver,
// 		wallet *MockWallet,
// 		balanceProcessor *BalanceProcessor,
// 	) error
// 	SyncNoSend(
// 		syncValidityProver *syncValidityProver,
// 		wallet *MockWallet,
// 		balanceProcessor *BalanceProcessor,
// 	) error
// 	SyncAll(
// 		syncValidityProver *syncValidityProver,
// 		wallet *MockWallet,
// 		balanceProcessor *BalanceProcessor,
// 	) error
// 	ReceiveDeposit(
// 		wallet *MockWallet,
// 		balanceProcessor *BalanceProcessor,
// 		blockBuilder MockBlockBuilder,
// 		depositIndex uint32,
// 	) error
// 	ReceiveTransfer(
// 		wallet *MockWallet,
// 		balanceProcessor *BalanceProcessor,
// 		blockBuilder MockBlockBuilder,
// 		transferWitness *intMaxTypes.TransferWitness,
// 		senderBalanceProof string,
// 	) error
// }

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

// func (s *SyncBalanceProver) Init(intMaxPrivateKey *intMaxAcc.PrivateKey) error {
// 	storedBalanceData, err := block_synchronizer.GetBackupBalance(s.ctx, s.cfg, intMaxPrivateKey.Public())
// 	if err != nil {
// 		const msg = "failed to start Balance Prover Service: %+v"
// 		log.Fatalf(msg, err.Error())
// 	}

// 	balanceData := new(block_synchronizer.BalanceData)
// 	err = balanceData.Decrypt(intMaxPrivateKey, storedBalanceData.EncryptedBalanceData)
// 	if err != nil {
// 		return err
// 	}

// 	s.storedBalanceData = storedBalanceData
// 	s.balanceData = balanceData

// 	newBalanceData := new(block_synchronizer.BalanceData).Set(balanceData)

// 	encryptedNewBalanceData, err := newBalanceData.Encrypt(intMaxPrivateKey.Public())
// 	if err != nil {
// 		return err
// 	}

// 	signature := "0x"
// 	err = block_synchronizer.BackupBalanceProof(s.ctx, s.cfg, s.log,
// 		intMaxPrivateKey.ToAddress(), storedBalanceData.ID, storedBalanceData.BalanceProofBody, encryptedNewBalanceData,
// 		storedBalanceData.EncryptedTxs, storedBalanceData.EncryptedTransfers, storedBalanceData.EncryptedDeposits,
// 		signature, storedBalanceData.BlockNumber)
// 	if err != nil {
// 		const msg = "failed to start Balance Prover Service: %+v"
// 		log.Fatalf(msg, err.Error())
// 	}

// 	return nil
// }

// func (s *SyncBalanceProver) setLastBalanceProof(blockNumber uint32, balanceProof string) {
// 	compressedBalanceProof, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceProof)
// 	if err != nil {
// 		log.Fatalf("failed to set last balance proof: %+v", err.Error())
// 	}

// 	s.balanceData.PublicState.BlockNumber = blockNumber
// 	s.lastBalanceProofBody = compressedBalanceProof.Proof
// 	s.balanceData.BalanceProofPublicInputs = compressedBalanceProof.PublicInputs
// }

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
) error {
	fmt.Printf("-----SyncSend %s------\n", wallet.PublicKey())

	allBlockNumbers := wallet.GetAllBlockNumbers()
	lastUpdatedBlockNumber := s.LastUpdatedBlockNumber()
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", lastUpdatedBlockNumber)
		if lastUpdatedBlockNumber < blockNumber {
			notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
		}
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	for _, blockNumber := range notSyncedBlockNumbers {
		sendWitness, err := wallet.GetSendWitness(blockNumber) // XXX: not need store sendWitness
		if err != nil {
			return errors.New("send witness not found")
		}
		// sentBlockNumber := sendWitness.GetIncludedBlockNumber()
		prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
		fmt.Printf("FetchUpdateWitness blockNumber: %d\n", blockNumber)
		currentBlockNumber := blockNumber
		updateWitness, err := blockValidityService.FetchUpdateWitness(
			wallet.PublicKey(),
			&currentBlockNumber,
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
	// if lastUpdatedBlockNumber == 0 {
	// 	return errors.New("last updated block number is 0")
	// }

	allBlockNumbers := wallet.GetAllBlockNumbers()
	fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", lastUpdatedBlockNumber)
	for _, targetBlockNumber := range allBlockNumbers {
		if lastUpdatedBlockNumber < targetBlockNumber {
			return errors.New("sync send tx first")
		}
	}

	if blockNumber <= lastUpdatedBlockNumber {
		// var ErrBlockNumberLessThanLastUpdatedBlockNumber = errors.New("block number is less than or equal to last updated block number")
		// return ErrBlockNumberLessThanLastUpdatedBlockNumber
		return nil
	}

	updateWitness, err := blockValidityService.FetchUpdateWitness(
		wallet.PublicKey(),
		&blockNumber,
		lastUpdatedBlockNumber,
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

	var prevBalancePis *balance_prover_service.BalancePublicInputs
	if s.LastBalanceProof() != nil {
		fmt.Println("s.LastBalanceProof != nil")
		var lastBalanceProofWithPis *intMaxTypes.Plonky2Proof
		lastBalanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*s.LastBalanceProof())
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

	// DEBUG
	// prevBalancePisJSON, err := json.Marshal(prevBalancePis)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("prevBalancePisJSON: %s", prevBalancePisJSON)

	lastSentTxBlockNumber := updateWitness.AccountMembershipProof.GetLeaf()
	prevPublicState := prevBalancePis.PublicState
	fmt.Printf("lastSentTxBlockNumber: %d\n", lastSentTxBlockNumber)
	fmt.Printf("prevPublicState.BlockNumber: %d\n", prevPublicState.BlockNumber)
	if lastSentTxBlockNumber > uint64(prevPublicState.BlockNumber) {
		// This indicates that there are unsynchronized transitions that need to be processed in advance.
		return errors.New("last block number is greater than prev public state block number")
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

	err = s.SyncSend(log, blockValidityService, blockSynchronizer, wallet, balanceProcessor)
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
	// err = balanceSynchronizer.Sync(userWalletState.PrivateKey())
	// if err != nil {
	// 	const msg = "failed to sync: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }

	// balanceTransitionData, err := balance_prover_service.NewBalanceTransitionData(ctx, cfg, log, userWalletState.PrivateKey())
	// if err != nil {
	// 	const msg = "failed to start Balance Prover Service: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }
	// fmt.Println("end NewBalanceTransitionData")
	// sortedValidUserData, err := balanceTransitionData.SortValidUserData(log, blockValidityService)
	// if err != nil {
	// 	const msg = "failed to sort valid user data: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }
	// fmt.Printf("size of sortedValidUserData: %v\n", len(sortedValidUserData))
	// for _, transition := range sortedValidUserData {
	// 	fmt.Printf("transition block number: %d\n", transition.BlockNumber())
	// }

	// storedBalanceData, err := block_synchronizer.GetBackupBalance(ctx, cfg, userWalletState.PublicKey())
	// if err != nil {
	// 	const msg = "failed to start Balance Prover Service: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }

	// err = syncBalanceProver.SetEncryptedBalanceData(userWalletState, storedBalanceData)
	// if err != nil {
	// 	const msg = "failed to start Balance Prover Service: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }

	timeout := 5 * time.Second
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Warnf("Received cancel signal from context, stopping...")
			return nil, errors.New("received cancel signal from context")
		case <-ticker.C:
			err = balanceSynchronizer.syncProcessing(userWalletState.PrivateKey())
			if err != nil {
				if errors.Is(err, ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber) ||
					errors.Is(err, ErrNoValidUserData) {
					return balanceSynchronizer, nil
				}

				if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
					continue
				}

				const msg = "failed to start sync processing: %+v"
				log.Fatalf(msg, err.Error())
			}

			// var validityProverInfo *block_validity_prover.ValidityProverInfo
			// validityProverInfo, err = blockValidityService.FetchValidityProverInfo()
			// if err != nil {
			// 	const msg = "failed to fetch validity prover info: %+v"
			// 	panic(fmt.Sprintf(msg, err.Error()))
			// }

			// // When the sync is done, we should stop the loop.
			// latestSynchronizedBlockNumber := validityProverInfo.BlockNumber
			// log.Debugf("latestSynchronizedBlockNumber: %d\n", latestSynchronizedBlockNumber)
			// log.Debugf("syncBalanceProver.LastUpdatedBlockNumber(): %d\n", syncBalanceProver.LastUpdatedBlockNumber())
			// if latestSynchronizedBlockNumber == 0 {
			// 	log.Debugf("latestSynchronizedBlockNumber is 0\n")
			// 	continue
			// }

			// if latestSynchronizedBlockNumber <= syncBalanceProver.LastUpdatedBlockNumber() && syncBalanceProver.LastUpdatedBlockNumber() != 0 {
			// 	return balanceSynchronizer, nil
			// }

			// if len(sortedValidUserData) == 0 {
			// 	return balanceSynchronizer, nil
			// }

			// for _, transition := range sortedValidUserData {
			// 	log.Debugf("valid transition: %v\n", transition)

			// 	switch transition := transition.(type) {
			// 	case balance_prover_service.ValidSentTx:
			// 		err = balanceSynchronizer.validSentTx(&transition)
			// 		if err != nil {
			// 			const msg = "failed to send transaction: %+v"
			// 			log.Warnf(msg, err.Error())
			// 			continue
			// 		}
			// 	case balance_prover_service.ValidReceivedDeposit:
			// 		err = balanceSynchronizer.validReceivedDeposit(&transition)
			// 		if err != nil {
			// 			if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
			// 				errors.Is(err, ErrApplyReceivedDepositTransitionFail) || errors.Is(err, ErrNullifierAlreadyExists) {
			// 				const msg = "failed to receive deposit: %+v"
			// 				log.Warnf(msg, err.Error())
			// 				continue
			// 			} else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
			// 				return nil, err
			// 			}

			// 			const msg = "failed to sync balance prover: %+v"
			// 			log.Fatalf(msg, err.Error())
			// 		}
			// 	case balance_prover_service.ValidReceivedTransfer:
			// 		err = balanceSynchronizer.validReceivedTransfer(&transition)
			// 		if err != nil {
			// 			if errors.Is(err, block_validity_prover.ErrNoValidityProofByBlockNumber) ||
			// 				errors.Is(err, ErrApplyReceivedTransferTransitionFail) || errors.Is(err, ErrNullifierAlreadyExists) {
			// 				const msg = "failed to receive transfer: %+v"
			// 				log.Warnf(msg, err.Error())
			// 				continue
			// 			} else if errors.Is(err, block_validity_prover.ErrBlockUnSynchronization) {
			// 				return nil, err
			// 			}

			// 			const msg = "failed to sync balance prover: %+v"
			// 			log.Fatalf(msg, err.Error())
			// 		}
			// 	default:
			// 		log.Warnf("unknown transition: %v\n", transition)
			// 	}
			// }
		}
	}
}