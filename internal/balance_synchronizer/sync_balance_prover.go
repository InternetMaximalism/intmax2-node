package balance_synchronizer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"log"
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
	return &SyncBalanceProver{
		ctx:               ctx,
		cfg:               cfg,
		log:               log,
		storedBalanceData: nil,
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

func (s *SyncBalanceProver) UploadLastBalanceProof(blockNumber uint32, balanceProof string, wallet UserState) error {
	s.setLastBalanceProof(blockNumber, balanceProof)

	s.balanceData.NullifierLeaves = wallet.Nullifiers()
	s.balanceData.AssetLeaves = wallet.AssetLeaves()
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

func (s *SyncBalanceProver) setLastBalanceProof(blockNumber uint32, balanceProof string) {
	compressedBalanceProof, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceProof)
	if err != nil {
		log.Fatalf("failed to set last balance proof: %+v", err.Error())
	}

	s.balanceData.PublicState.BlockNumber = blockNumber
	s.lastBalanceProofBody = compressedBalanceProof.Proof
	s.balanceData.BalanceProofPublicInputs = compressedBalanceProof.PublicInputs
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
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", s.LastUpdatedBlockNumber())
		if s.LastUpdatedBlockNumber() < blockNumber {
			notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
		}
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	for _, blockNumber := range notSyncedBlockNumbers {
		sendWitness, err := wallet.GetSendWitness(blockNumber)
		if err != nil {
			return errors.New("send witness not found")
		}
		blockNumber := sendWitness.GetIncludedBlockNumber()
		prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
		fmt.Printf("FetchUpdateWitness blockNumber: %d\n", blockNumber)
		updateWitness, err := blockValidityService.FetchUpdateWitness(
			wallet.PublicKey(),
			&blockNumber,
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

		// TODO
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
			return err
		}

		// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
		// if err != nil {
		// 	return err
		// }

		fmt.Printf("s.LastUpdatedBlockNumber before SyncSend: %d\n", s.LastUpdatedBlockNumber())
		// s.LastUpdatedBlockNumber = blockNumber
		fmt.Printf("s.LastUpdatedBlockNumber after SyncSend: %d\n", s.LastUpdatedBlockNumber())
		s.UploadLastBalanceProof(blockNumber, balanceProof.Proof, wallet)
		wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)
	}

	return nil
}

// Sync balance proof public state to the latest block
// assuming that there is no un-synced send tx.
func (s *SyncBalanceProver) SyncNoSend(
	log logger.Logger,
	blockValidityService block_validity_prover.BlockValidityService,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
) error {
	fmt.Printf("-----SyncNoSend %s------\n", wallet.PublicKey())

	lastUpdatedBlockNumber := s.LastUpdatedBlockNumber()
	if lastUpdatedBlockNumber == 0 {
		return errors.New("last updated block number is 0")
	}

	allBlockNumbers := wallet.GetAllBlockNumbers()
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", s.LastUpdatedBlockNumber())
		if lastUpdatedBlockNumber < blockNumber {
			return errors.New("sync send tx first")
		}
	}

	fmt.Printf("s.LastUpdatedBlockNumber before FetchUpdateWitness: %d\n", lastUpdatedBlockNumber)
	updateWitness, err := blockValidityService.FetchUpdateWitness(
		wallet.PublicKey(),
		nil, // latest
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

	// let prev_balance_pis = if prev_balance_proof.is_some() {
	//     BalancePublicInputs::from_pis(&prev_balance_proof.as_ref().unwrap().public_inputs)
	// } else {
	//     BalancePublicInputs::new(public_key)
	// };
	// let last_block_number = balance_update_witness.account_membership_proof.get_value();
	// let prev_public_state = &prev_balance_pis.public_state;
	// println!("last_block_number: {}", last_block_number);
	// println!(
	//     "prev_public_state.block_number: {}",
	//     prev_public_state.block_number
	// );
	// if last_block_number > prev_balance_pis.public_state.block_number as u64 {
	// 	return Err("last_block_number is greater than prev_public_state.block_number");
	// }

	var prevBalancePis *balance_prover_service.BalancePublicInputs
	if s.LastBalanceProof() != nil {
		fmt.Println("s.LastBalanceProof != nil")
		lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*s.LastBalanceProof())
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
	prevBalancePisJSON, err := json.Marshal(prevBalancePis)
	if err != nil {
		return err
	}
	fmt.Printf("prevBalancePisJSON: %s", prevBalancePisJSON)

	lastSentTxBlockNumber := updateWitness.AccountMembershipProof.GetLeaf()
	prevPublicState := prevBalancePis.PublicState
	fmt.Printf("sync no send")
	fmt.Printf("lastSentTxBlockNumber: %d\n", lastSentTxBlockNumber)
	fmt.Printf("prevPublicState.BlockNumber: %d\n", prevPublicState.BlockNumber)
	if lastSentTxBlockNumber > uint64(prevPublicState.BlockNumber) {
		return errors.New("last block number is greater than prev public state block number")
	}

	balanceProof, err := balanceProcessor.ProveUpdate(
		wallet.PublicKey(),
		updateWitness,
		s.LastBalanceProof(),
	)
	if err != nil {
		return err
	}

	// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
	// if err != nil {
	// 	return err
	// }

	fmt.Printf("PublicInputs: %+v\n", balanceProof.PublicInputs)
	fmt.Printf("PublicState: %+v\n", balanceProof.PublicInputs.PublicState)
	fmt.Printf("s.LastUpdatedBlockNumber before SyncNoSend: %d\n", s.LastUpdatedBlockNumber())
	// s.LastUpdatedBlockNumber = currentBlockNumber
	fmt.Printf("s.LastUpdatedBlockNumber after SyncNoSend: %d\n", s.LastUpdatedBlockNumber())
	wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)
	s.UploadLastBalanceProof(currentBlockNumber, balanceProof.Proof, wallet)

	return nil
}

func (s *SyncBalanceProver) SyncAll(
	log logger.Logger,
	blockValidityService *block_validity_prover.BlockValidityProverMemory,
	blockSynchronizer block_validity_prover.BlockSynchronizer,
	wallet UserState,
	balanceProcessor balance_prover_service.BalanceProcessor,
) (err error) {
	latestIntMaxBlockNumber, err := blockValidityService.LatestIntMaxBlockNumber()
	if err != nil {
		return err
	}
	fmt.Printf("LatestWitnessNumber before SyncSend: %d\n", latestIntMaxBlockNumber)

	err = s.SyncSend(log, blockValidityService, blockSynchronizer, wallet, balanceProcessor)
	if err != nil {
		return err
	}
	err = s.SyncNoSend(log, blockValidityService, blockSynchronizer, wallet, balanceProcessor)
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
		return err
	}
	fmt.Println("start ProveReceiveDeposit")
	lastBalanceProof := *s.LastBalanceProof()
	lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(lastBalanceProof)
	if err != nil {
		return err
	}

	lastBalancePublicInputs, err := new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
	if err != nil {
		return err
	}
	fmt.Printf("lastBalancePublicInputs (ReceiveDeposit) PrivateCommitment commitment: %s\n", lastBalancePublicInputs.PrivateCommitment.String())

	balanceProof, err := balanceProcessor.ProveReceiveDeposit(
		wallet.PublicKey(),
		receiveDepositWitness,
		&lastBalanceProof,
	)
	if err != nil {
		return err
	}

	lastBalanceProofWithPis, err = intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceProof.Proof)
	if err != nil {
		return err
	}

	lastBalancePublicInputs, err = new(balance_prover_service.BalancePublicInputs).FromPublicInputs(lastBalanceProofWithPis.PublicInputs)
	if err != nil {
		return err
	}

	fmt.Printf("ReceiveDeposit PrivateCommitment commitment (after): %s\n", lastBalancePublicInputs.PrivateCommitment.String())
	fmt.Printf("ReceiveDeposit PrivateCommitment commitment (after, public inputs): %s\n", balanceProof.PublicInputs.PrivateCommitment.String())
	fmt.Printf("wallet private state: %+v\n", wallet.PrivateState())
	fmt.Printf("wallet private state commitment: %s\n", wallet.PrivateState().Commitment().String())

	fmt.Println("finish ProveReceiveDeposit")

	s.UploadLastBalanceProof(s.LastUpdatedBlockNumber(), balanceProof.Proof, wallet)

	return nil
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
	s.UploadLastBalanceProof(s.LastUpdatedBlockNumber(), balanceProof.Proof, wallet)

	return nil
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

func SyncLocally(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb block_validity_prover.ServiceBlockchain,
	blockValidityService block_validity_prover.BlockValidityService,
	userWalletState UserState,
) (*balanceSynchronizer, error) {
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

	balanceTransitionData, err := balance_prover_service.NewBalanceTransitionData(ctx, cfg, userWalletState.PrivateKey())
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		log.Fatalf(msg, err.Error())
	}
	fmt.Println("end NewBalanceTransitionData")
	sortedValidUserData, err := balanceTransitionData.SortValidUserData(log, blockValidityService)
	if err != nil {
		const msg = "failed to sort valid user data: %+v"
		log.Fatalf(msg, err.Error())
	}
	fmt.Printf("size of sortedValidUserData: %v\n", len(sortedValidUserData))
	for _, transition := range sortedValidUserData {
		fmt.Printf("transition block number: %d\n", transition.BlockNumber())
	}

	storedBalanceData, err := block_synchronizer.GetBackupBalance(ctx, cfg, userWalletState.PublicKey())
	if err != nil {
		if err.Error() != "failed to start Balance Prover Service: no assets found" {
			// default value
			storedBalanceData = &block_synchronizer.BackupBalanceData{
				ID:                   "",
				BalanceProofBody:     "",
				EncryptedBalanceData: "",
				BlockNumber:          0,
			}
		} else {
			const msg = "failed to start Balance Prover Service: %+v"
			log.Fatalf(msg, err.Error())
		}
	}
	log.Debugf("end GetBackupBalance\n")

	err = syncBalanceProver.SetEncryptedBalanceData(userWalletState, storedBalanceData)
	if err != nil {
		const msg = "failed to start Balance Prover Service: %+v"
		log.Fatalf(msg, err.Error())
	}

	timeout := 1 * time.Second
	ticker := time.NewTicker(timeout)
	for {
		log.Debugf("start SyncLocally loop\n")
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Warnf("Received cancel signal from context, stopping...")
			return nil, errors.New("received cancel signal from context")
		case <-ticker.C:
			validityProverInfo, err := blockValidityService.FetchValidityProverInfo()
			if err != nil {
				const msg = "failed to fetch validity prover info: %+v"
				panic(fmt.Sprintf(msg, err.Error()))
			}

			// When the sync is done, we should stop the loop.
			latestSynchronizedBlockNumber := validityProverInfo.BlockNumber
			log.Debugf("latestSynchronizedBlockNumber: %d\n", latestSynchronizedBlockNumber)
			log.Debugf("syncBalanceProver.LastUpdatedBlockNumber(): %d\n", syncBalanceProver.LastUpdatedBlockNumber())
			if latestSynchronizedBlockNumber == 0 {
				log.Debugf("latestSynchronizedBlockNumber is 0\n")
				continue
			}

			if latestSynchronizedBlockNumber <= syncBalanceProver.LastUpdatedBlockNumber() && syncBalanceProver.LastUpdatedBlockNumber() != 0 {
				return balanceSynchronizer, nil
			}

			for _, transition := range sortedValidUserData {
				log.Debugf("valid transition: %v\n", transition)

				switch transition := transition.(type) {
				case balance_prover_service.ValidSentTx:
					log.Debugf("valid sent transaction: %v\n", transition.TxHash)
					err := applySentTransactionTransition(
						log,
						transition.Tx,
						blockValidityService,
						blockSynchronizer,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)
					if err != nil {
						const msg = "failed to send transaction: %+v"
						log.Warnf(msg, err.Error())
						continue
					}
				case balance_prover_service.ValidReceivedDeposit:
					log.Debugf("valid received deposit: %v\n", transition.DepositHash)
					transitionBlockNumber := transition.BlockNumber()
					log.Debugf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						log,
						blockValidityService,
						blockSynchronizer,
						userWalletState,
						balanceProcessor,
					)
					if err != nil {
						const msg = "failed to sync balance prover: %+v"
						panic(fmt.Sprintf(msg, err.Error()))
					}

					err := applyReceivedDepositTransition(
						transition.Deposit,
						blockValidityService,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)
					if err != nil {
						const msg = "failed to receive deposit: %+v"
						log.Warnf(msg, err.Error())
						continue
					}
				case balance_prover_service.ValidReceivedTransfer:
					log.Debugf("valid received transfer: %v\n", transition.TransferHash)
					transitionBlockNumber := transition.BlockNumber()
					log.Debugf("transitionBlockNumber: %d", transitionBlockNumber)
					err = syncBalanceProver.SyncNoSend(
						log,
						blockValidityService,
						blockSynchronizer,
						userWalletState,
						balanceProcessor,
					)
					if err != nil {
						const msg = "failed to sync balance prover: %+v"
						panic(fmt.Sprintf(msg, err.Error()))
					}

					err := applyReceivedTransferTransition(
						transition.Transfer,
						blockValidityService,
						balanceProcessor,
						syncBalanceProver,
						userWalletState,
					)
					if err != nil {
						const msg = "failed to receive transfer: %+v"
						log.Warnf(msg, err.Error())
						continue
					}
				default:
					log.Warnf("unknown transition: %v\n", transition)
				}
			}
		}
	}
}
