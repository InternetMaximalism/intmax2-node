package block_signature

import (
	"context"
)

type UCBlockSignature struct {
	Message string
}

type Plonky2Proof struct {
	PublicInputs []uint64 `json:"publicInputs"`
	Proof        []byte   `json:"proof"`
}

func (dst *Plonky2Proof) Set(src *Plonky2Proof) *Plonky2Proof {
	dst.PublicInputs = make([]uint64, len(src.PublicInputs))
	copy(dst.PublicInputs, src.PublicInputs)
	dst.Proof = make([]byte, len(src.Proof))
	copy(dst.Proof, src.Proof)

	return dst
}

type EnoughBalanceProofInput struct {
	PrevBalanceProof  *Plonky2Proof `json:"prevBalanceProof"`
	TransferStepProof *Plonky2Proof `json:"transferStepProof"`
}

func (dst *EnoughBalanceProofInput) Set(src *EnoughBalanceProofInput) *EnoughBalanceProofInput {
	dst.PrevBalanceProof = new(Plonky2Proof).Set(src.PrevBalanceProof)
	dst.TransferStepProof = new(Plonky2Proof).Set(src.TransferStepProof)

	return dst
}

type UCBlockSignatureInput struct {
	Sender             string                   `json:"sender"`
	TxHash             string                   `json:"txHash"`
	Signature          string                   `json:"signature"`
	EnoughBalanceProof *EnoughBalanceProofInput `json:"enoughBalanceProof"`
}

// UseCaseBlockSignature describes BlockSignature contract.
type UseCaseBlockSignature interface {
	Do(ctx context.Context, input *UCBlockSignatureInput) (*UCBlockSignature, error)
}
