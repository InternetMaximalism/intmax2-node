package types

type CompressedValidityWitness struct {
	BlockWitness              *CompressedBlockWitness              `json:"blockWitness"`
	ValidityTransitionWitness *CompressedValidityTransitionWitness `json:"validityTransitionWitness"`
}
