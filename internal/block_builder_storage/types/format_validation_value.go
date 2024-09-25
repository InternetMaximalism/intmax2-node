package types

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
)

type FormatValidationValue struct {
	PublicKeys          []intMaxTypes.Uint256
	Signature           *SignatureContent
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	SignatureCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewFormatValidationValue(
	publicKeys []intMaxTypes.Uint256,
	signature *SignatureContent,
) *FormatValidationValue {
	pubkeyCommitment := GetPublicKeyCommitment(publicKeys)
	signatureCommitment := signature.Commitment()
	err := signature.IsValidFormat(publicKeys)

	return &FormatValidationValue{
		PublicKeys:          publicKeys,
		Signature:           signature,
		PublicKeyCommitment: pubkeyCommitment,
		SignatureCommitment: signatureCommitment,
		IsValid:             err == nil,
	}
}
