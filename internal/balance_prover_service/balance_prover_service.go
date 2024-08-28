package balance_prover_service

import (
	"errors"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/hash/goldenposeidon"
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

func (s *MockBlockBuilder) GetBlockMerkleProof(rootBlockNumber, leafBlockNumber uint32) (*intMaxTree.BlockHashMerkleProof, error) {
	// if rootBlockNumber < leafBlockNumber {
	// 	return nil, errors.New("root block number is less than leaf block number")
	// }

	// auxInfo, ok := s.GetAuxInfo(rootBlockNumber)
	// if !ok {
	// 	return nil, errors.New("current block number not found")
	// }
	// blockMerkleProof := auxInfo.BlockTree.Prove(int(leafBlockNumber))

	// return blockMerkleProof, nil

	return nil, errors.New("not implemented")
}

type BalancePublicInputs struct {
	PubKey                  *intMaxAcc.PublicKey
	PrivateCommitment       *intMaxTypes.PoseidonHashOut
	LastTxHash              *intMaxTypes.PoseidonHashOut
	LastTxInsufficientFlags backup_balance.InsufficientFlags
	PublicState             *block_validity_prover.PublicState
}

const balancePublicInputsLen = 47
const (
	int2Key = 2
	int3Key = 3
	int4Key = 4
	int8Key = 8
)

func (s *BalancePublicInputs) FromPublicInputs(publicInputs []ffg.Element) (*BalancePublicInputs, error) {
	if len(publicInputs) < balancePublicInputsLen {
		return nil, errors.New("invalid length")
	}

	const (
		numHashOutElts                = goldenposeidon.NUM_HASH_OUT_ELTS
		publicKeyOffset               = 0
		privateCommitmentOffset       = publicKeyOffset + int8Key
		lastTxHashOffset              = privateCommitmentOffset + numHashOutElts
		lastTxInsufficientFlagsOffset = lastTxHashOffset + numHashOutElts
		publicStateOffset             = lastTxInsufficientFlagsOffset + backup_balance.InsufficientFlagsLen
		end                           = publicStateOffset + block_validity_prover.PublicStateLimbSize
	)

	address := new(intMaxTypes.Uint256).FromFieldElementSlice(publicInputs[0:int8Key])
	publicKey, err := new(intMaxAcc.PublicKey).SetBigInt(address.BigInt())
	if err != nil {
		return nil, err
	}
	privateCommitment := poseidonHashOut{
		Elements: [numHashOutElts]ffg.Element{
			publicInputs[privateCommitmentOffset],
			publicInputs[privateCommitmentOffset+1],
			publicInputs[privateCommitmentOffset+int2Key],
			publicInputs[privateCommitmentOffset+int3Key],
		},
	}
	lastTxHash := poseidonHashOut{
		Elements: [numHashOutElts]ffg.Element{
			publicInputs[lastTxHashOffset],
			publicInputs[lastTxHashOffset+1],
			publicInputs[lastTxHashOffset+int2Key],
			publicInputs[lastTxHashOffset+int3Key],
		},
	}
	lastTxInsufficientFlags := new(backup_balance.InsufficientFlags).FromFieldElementSlice(
		publicInputs[lastTxInsufficientFlagsOffset:publicStateOffset],
	)
	publicState := new(block_validity_prover.PublicState).FromFieldElementSlice(
		publicInputs[publicStateOffset:end],
	)

	return &BalancePublicInputs{
		PubKey:                  publicKey,
		PrivateCommitment:       &privateCommitment,
		LastTxHash:              &lastTxHash,
		LastTxInsufficientFlags: *lastTxInsufficientFlags,
		PublicState:             publicState,
	}, nil
}
