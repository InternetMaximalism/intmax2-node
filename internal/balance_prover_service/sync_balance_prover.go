package balance_prover_service

import (
	"errors"
	"fmt"
	"intmax2-node/internal/block_validity_prover"
	intMaxTypes "intmax2-node/internal/types"
	"sort"
)

type SyncBalanceProver struct {
	LastUpdatedBlockNumber uint32
	LastBalanceProof       *string
}

func NewSyncBalanceProver() *SyncBalanceProver {
	return &SyncBalanceProver{
		LastUpdatedBlockNumber: 0,
		LastBalanceProof:       nil,
	}
}

func (s *SyncBalanceProver) BalancePublicInputs() (*BalancePublicInputs, error) {
	if s.LastBalanceProof == nil {
		return nil, errors.New("last balance proof is nil")
	}

	balanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(*s.LastBalanceProof)
	if err != nil {
		return nil, err
	}

	balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProofWithPis.PublicInputs)
	if err != nil {
		return nil, err
	}

	return balancePublicInputs, nil
}

func (s *SyncBalanceProver) SyncSend(
	syncValidityProver *syncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
) error {
	fmt.Println("-----SyncSend------")
	err := syncValidityProver.ValidityProver.SyncBlockProver() // sync validity proofs
	if err != nil {
		return err
	}
	allBlockNumbers := wallet.GetAllBlockNumbers()
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", s.LastUpdatedBlockNumber)
		if s.LastUpdatedBlockNumber < blockNumber {
			notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
		}
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	blockBuilder := syncValidityProver.ValidityProver.BlockBuilder()
	for _, blockNumber := range notSyncedBlockNumbers {
		sendWitness, err := wallet.GetSendWitness(blockNumber)
		if err != nil {
			return errors.New("send witness not found")
		}
		blockNumber := sendWitness.GetIncludedBlockNumber()
		prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
		fmt.Printf("FetchUpdateWitness blockNumber: %d\n", blockNumber)
		updateWitness, err := syncValidityProver.FetchUpdateWitness(
			blockBuilder,
			wallet.PublicKey(),
			blockNumber,
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
		fmt.Printf("--------CHECK-------------")
		if !updateWitnessValidityPis.Equal(&sendWitnessValidityPis) {
			fmt.Printf("update witness validity proof: %v\n", updateWitnessValidityPis)
			fmt.Printf("update witness public state: %v\n", updateWitnessValidityPis.PublicState)
			fmt.Printf("update witness account tree root: %v\n", updateWitnessValidityPis.PublicState.PrevAccountTreeRoot) // 0x32
			fmt.Printf("update witness account tree root: %v\n", updateWitnessValidityPis.PublicState.AccountTreeRoot)     // XXX: 0x32adcd085f3953448a528b510df7a30499a6fe6ba144865da4cd23c94f3a4e15
			fmt.Printf("send witness validity proof: %v\n", sendWitnessValidityPis)
			fmt.Printf("send witness public state: %v\n", sendWitnessValidityPis.PublicState)
			fmt.Printf("send witness account tree root: %v\n", sendWitnessValidityPis.PublicState.PrevAccountTreeRoot) // 0x32
			fmt.Printf("send witness account tree root: %v\n", sendWitnessValidityPis.PublicState.AccountTreeRoot)     // 0x06350dc0d3599f02b51e6ca00c055df6661a0c6e783a6599fca46fe0a88a5257
			return errors.New("update witness validity proof is not equal to send witness validity proof")
		}

		balanceProof, err := balanceProcessor.ProveSend(
			wallet.PublicKey(),
			sendWitness,
			updateWitness,
			s.LastBalanceProof,
		)
		if err != nil {
			return err
		}

		// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
		// if err != nil {
		// 	return err
		// }

		fmt.Printf("s.LastUpdatedBlockNumber before SyncSend: %d\n", s.LastUpdatedBlockNumber)
		s.LastUpdatedBlockNumber = blockNumber
		fmt.Printf("s.LastUpdatedBlockNumber after SyncSend: %d\n", s.LastUpdatedBlockNumber)
		s.LastBalanceProof = &balanceProof.Proof
		wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)
	}

	return nil
}

// Sync balance proof public state to the latest block
// assuming that there is no un-synced send tx.
func (s *SyncBalanceProver) SyncNoSend(
	syncValidityProver *syncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
) error {
	fmt.Println("-----SyncNoSend------")
	blockBuilder := syncValidityProver.ValidityProver.BlockBuilder()
	err := syncValidityProver.ValidityProver.SyncBlockProver()
	if err != nil {
		return err
	}
	allBlockNumbers := wallet.GetAllBlockNumbers()
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		fmt.Printf("s.LastUpdatedBlockNumber after GetAllBlockNumbers: %d\n", s.LastUpdatedBlockNumber)
		if s.LastUpdatedBlockNumber < blockNumber {
			notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
		}
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	if len(notSyncedBlockNumbers) > 0 {
		return errors.New("sync send tx first")
	}
	currentBlockNumber := blockBuilder.LatestWitnessBlockNumber
	fmt.Printf("currentBlockNumber before FetchUpdateWitness: %d\n", currentBlockNumber)
	fmt.Printf("s.LastUpdatedBlockNumber before FetchUpdateWitness: %d\n", s.LastUpdatedBlockNumber)
	updateWitness, err := syncValidityProver.FetchUpdateWitness(
		blockBuilder,
		wallet.PublicKey(),
		currentBlockNumber,
		s.LastUpdatedBlockNumber,
		false,
	)
	if err != nil {
		return err
	}
	balanceProof, err := balanceProcessor.ProveUpdate(
		wallet.PublicKey(),
		updateWitness,
		s.LastBalanceProof,
	)
	if err != nil {
		return err
	}

	// balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
	// if err != nil {
	// 	return err
	// }

	fmt.Printf("PublicInputs: %v\n", balanceProof.PublicInputs)
	fmt.Printf("PublicState: %v\n", balanceProof.PublicInputs.PublicState)
	fmt.Printf("s.LastUpdatedBlockNumber before SyncNoSend: %d\n", s.LastUpdatedBlockNumber)
	s.LastUpdatedBlockNumber = currentBlockNumber
	fmt.Printf("s.LastUpdatedBlockNumber after SyncNoSend: %d\n", s.LastUpdatedBlockNumber)
	s.LastBalanceProof = &balanceProof.Proof
	wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)

	return nil
}

func (s *SyncBalanceProver) SyncAll(
	syncValidityProver *syncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
) (err error) {
	fmt.Printf("LatestWitnessNumber before SyncSend: %d\n", syncValidityProver.ValidityProver.BlockBuilder().LatestIntMaxBlockNumber())

	err = s.SyncSend(syncValidityProver, wallet, balanceProcessor)
	if err != nil {
		return err
	}
	err = s.SyncNoSend(syncValidityProver, wallet, balanceProcessor)
	if err != nil {
		return err
	}

	return nil
}

func (s *SyncBalanceProver) ReceiveDeposit(
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
	depositId uint32,
) error {
	receiveDepositWitness, err := wallet.ReceiveDepositAndUpdate(blockBuilder, depositId)
	if err != nil {
		return err
	}
	fmt.Println("start ProveReceiveDeposit")
	balanceProof, err := balanceProcessor.ProveReceiveDeposit(
		wallet.PublicKey(),
		receiveDepositWitness,
		s.LastBalanceProof,
	)
	if err != nil {
		return err
	}
	fmt.Println("finish ProveReceiveDeposit")

	s.LastBalanceProof = &balanceProof.Proof

	return nil
}

func (s *SyncBalanceProver) ReceiveTransfer(
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
	transferWitness *TransferWitness,
	senderBalanceProof string,
) error {
	fmt.Printf("ReceiveTransfer s.LastUpdatedBlockNumber: %d\n", s.LastUpdatedBlockNumber)
	receiveTransferWitness, err := wallet.ReceiveTransferAndUpdate(
		blockBuilder,
		s.LastUpdatedBlockNumber,
		transferWitness,
		senderBalanceProof,
	)
	if err != nil {
		return err
	}
	balanceProof, err := balanceProcessor.ProveReceiveTransfer(
		wallet.PublicKey(),
		receiveTransferWitness,
		s.LastBalanceProof,
	)
	if err != nil {
		return err
	}

	s.LastBalanceProof = &balanceProof.Proof

	return nil
}
