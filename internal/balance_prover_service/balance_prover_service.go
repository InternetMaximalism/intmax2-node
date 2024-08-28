package balance_prover_service

import (
	"errors"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type BalanceValidityAuxInfo struct {
	ValidityWitness *block_validity_prover.ValidityWitness
}

type ValidityVerifierData struct{}

type ValidityProcessor struct{}

func (s *ValidityProcessor) Prove(
	prevValidityProof *intMaxTypes.Plonky2Proof,
	validityWitness *block_validity_prover.ValidityWitness,
) (*intMaxTypes.Plonky2Proof, error) {
	return nil, errors.New("not implemented")
}

type MockBlockBuilder struct{}

func (s *MockBlockBuilder) GetAuxInfo(blockNumber uint32) (*BalanceValidityAuxInfo, bool) {
	return nil, false
}

func (s *MockBlockBuilder) LastBlockNumber() uint32 {
	return 0
}

func (s *MockBlockBuilder) GetBlockNumber() uint32 {
	return 0
}

func (s *MockBlockBuilder) GetDepositTreeProof(index uint32) *intMaxTree.MerkleProof {
	return nil
}

// pub struct BalancePublicInputs {
//     pub pubkey: U256,
//     pub private_commitment: PoseidonHashOut,
//     pub last_tx_hash: PoseidonHashOut,
//     pub last_tx_insufficient_flags: InsufficientFlags,
//     pub public_state: PublicState,
// }

type BalancePublicInputs struct {
	PubKey                  *intMaxAcc.PublicKey
	PrivateCommitment       *intMaxTypes.PoseidonHashOut
	LastTxHash              *intMaxTypes.PoseidonHashOut
	LastTxInsufficientFlags backup_balance.InsufficientFlags
	PublicState             *block_validity_prover.PublicState
}

func (s *BalancePublicInputs) FromPublicInputs(publicInputs []ffg.Element) (*BalancePublicInputs, error) {
	return nil, errors.New("not implemented")
}
