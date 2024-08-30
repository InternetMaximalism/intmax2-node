package balance_prover_service

import (
	"errors"
	intMaxTypes "intmax2-node/internal/types"
	"math/rand"
	"sort"
)

type SyncBalanceProver struct {
	LastBlockNumber  uint32
	LastBalanceProof *intMaxTypes.Plonky2Proof
}

func NewSyncBalanceProver() *SyncBalanceProver {
	return &SyncBalanceProver{
		LastBlockNumber:  0,
		LastBalanceProof: nil,
	}
}

func (s *SyncBalanceProver) SyncSend(
	syncValidityProver *SyncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
) error {
	syncValidityProver.Sync() // sync validity proofs
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

		balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
		if err != nil {
			return err
		}
		s.LastBlockNumber = blockNumber
		s.LastBalanceProof = balanceProof
		wallet.UpdatePublicState(balancePublicInputs.PublicState)
	}

	return nil
}

// Sync balance proof public state to the latest block
// assuming that there is no un-synced send tx.
func (s *SyncBalanceProver) SyncNoSend(
	syncValidityProver *SyncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
) error {
	syncValidityProver.Sync() // sync validity proofs
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
	balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(balanceProof.PublicInputs)
	if err != nil {
		return err
	}

	s.LastBlockNumber = currentBlockNumber
	s.LastBalanceProof = balanceProof
	wallet.UpdatePublicState(balancePublicInputs.PublicState)

	return nil
}

func (s *SyncBalanceProver) SyncAll(
	syncValidityProver *SyncValidityProver,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
) (err error) {
	err = s.SyncSend(syncValidityProver, wallet, balanceProcessor, blockBuilder)
	if err != nil {
		return err
	}
	err = s.SyncNoSend(syncValidityProver, wallet, balanceProcessor, blockBuilder)
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
	balanceProof, err := balanceProcessor.ProveReceiveDeposit(
		wallet.PublicKey(),
		receiveDepositWitness,
		s.LastBalanceProof,
	)
	if err != nil {
		return err
	}
	s.LastBalanceProof = balanceProof

	return nil
}

func (s *SyncBalanceProver) ReceiveTransfer(
	rng *rand.Rand,
	wallet *MockWallet,
	balanceProcessor *BalanceProcessor,
	blockBuilder MockBlockBuilder,
	transferWitness *TransferWitness,
	senderBalanceProof *intMaxTypes.Plonky2Proof,
) error {
	receiveTransferWitness, err := wallet.ReceiveTransferAndUpdate(
		rng,
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
	s.LastBalanceProof = balanceProof

	return nil
}
