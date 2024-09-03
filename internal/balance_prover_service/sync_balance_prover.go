package balance_prover_service

import (
	"errors"
	"fmt"
	intMaxTypes "intmax2-node/internal/types"
	"sort"
)

type SyncBalanceProver struct {
	LastBlockNumber  uint32
	LastBalanceProof *string
}

func NewSyncBalanceProver() *SyncBalanceProver {
	return &SyncBalanceProver{
		LastBlockNumber:  0,
		LastBalanceProof: nil,
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
	// blockBuilder MockBlockBuilder,
) error {
	blockBuilder := syncValidityProver.ValidityProcessor.BlockBuilder()
	// syncValidityProver.Sync() // sync validity proofs
	allBlockNumbers := wallet.GetAllBlockNumbers()
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		if s.LastBlockNumber < blockNumber {
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
		prevBlockNumber := sendWitness.GetPrevBlockNumber()
		updateWitness, err := syncValidityProver.FetchUpdateWitness(
			blockBuilder,
			wallet.PublicKey(),
			blockNumber,
			prevBlockNumber,
			true,
		)
		if err != nil {
			return err
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

		s.LastBlockNumber = blockNumber
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
	// blockBuilder MockBlockBuilder,
) error {
	blockBuilder := syncValidityProver.ValidityProcessor.BlockBuilder()
	// err := syncValidityProver.Check() // sync validity proofs
	// if err != nil {
	// 	fmt.Printf("WARNING: not synced: %v\n", err)
	// }
	allBlockNumbers := wallet.GetAllBlockNumbers()
	notSyncedBlockNumbers := []uint32{}
	for _, blockNumber := range allBlockNumbers {
		if s.LastBlockNumber < blockNumber {
			notSyncedBlockNumbers = append(notSyncedBlockNumbers, blockNumber)
		}
	}

	sort.Slice(notSyncedBlockNumbers, func(i, j int) bool {
		return notSyncedBlockNumbers[i] < notSyncedBlockNumbers[j]
	})

	if len(notSyncedBlockNumbers) > 0 {
		return errors.New("sync send tx first")
	}
	currentBlockNumber := blockBuilder.LatestIntMaxBlockNumber()
	updateWitness, err := syncValidityProver.FetchUpdateWitness(
		blockBuilder,
		wallet.PublicKey(),
		currentBlockNumber,
		s.LastBlockNumber,
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
	s.LastBlockNumber = currentBlockNumber
	s.LastBalanceProof = &balanceProof.Proof
	wallet.UpdatePublicState(balanceProof.PublicInputs.PublicState)

	return nil
}

func (s *SyncBalanceProver) SyncAll(
	syncValidityProver *syncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
) (err error) {
	fmt.Printf("len(b.ValidityProofs) before SyncSend: %d\n", len(syncValidityProver.ValidityProcessor.BlockBuilder().ValidityProofs))

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
	receiveTransferWitness, err := wallet.ReceiveTransferAndUpdate(
		blockBuilder,
		s.LastBlockNumber,
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
