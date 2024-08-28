package balance_prover_service

import (
	"errors"

	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
)

type SyncValidityProver struct {
	ValidityProcessor ValidityProcessor
	LastBlockNumber   uint32
	ValidityProofs    map[uint32]*intMaxTypes.Plonky2Proof
}

func (s *SyncValidityProver) Sync(blockBuilder *MockBlockBuilder) error {
	currentBlockNumber := blockBuilder.LastBlockNumber()
	for blockNumber := s.LastBlockNumber + 1; blockNumber <= currentBlockNumber; blockNumber++ {
		prevValidityProof := s.ValidityProofs[blockNumber-1]
		if prevValidityProof == nil && blockNumber != 1 {
			return errors.New("prev validity proof is nil")
		}
		auxInfo, ok := blockBuilder.GetAuxInfo(blockNumber)
		if !ok {
			return errors.New("aux info not found")
		}
		validityProof, err := s.ValidityProcessor.Prove(prevValidityProof, auxInfo.ValidityWitness)
		if err != nil {
			return errors.New("validity proof is nil")
		}
		s.ValidityProofs[blockNumber] = validityProof
	}
	s.LastBlockNumber = currentBlockNumber

	return nil
}

func (s *SyncValidityProver) FetchUpdateWitness(
	blockBuilder *MockBlockBuilder,
	publicKey *intMaxAcc.PublicKey,
	blockNumber uint32,
	prevBlockNumber uint32,
	shouldProve bool,
) (*UpdateWitness, error) {
	// request validity prover
	return nil, errors.New("not implemented")
}

func (s *SyncValidityProver) ValidityVerifierData() *ValidityVerifierData {
	return nil
}
