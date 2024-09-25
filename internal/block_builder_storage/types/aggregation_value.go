package types

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
)

type AggregationValue struct {
	PublicKeys          []intMaxTypes.Uint256
	Signature           *SignatureContent
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	SignatureCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewAggregationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *AggregationValue {
	publicKeyCommitment := GetPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.VerifyAggregation(publicKeys)

	return &AggregationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: publicKeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
	}
}
