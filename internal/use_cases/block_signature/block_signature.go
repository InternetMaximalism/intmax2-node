package block_signature

import (
	"context"
	"encoding/base64"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

//go:generate mockgen -destination=../mocks/mock_block_signature.go -package=mocks -source=block_signature.go

const (
	SuccessMsg = "Signature accepted."
)

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

type EnoughBalanceProofBodyInput struct {
	PrevBalanceProofBody  string `json:"prevBalanceProof"`
	TransferStepProofBody string `json:"transferStepProof"`
}

func (dst *EnoughBalanceProofBodyInput) Set(src *EnoughBalanceProofBodyInput) *EnoughBalanceProofBodyInput {
	dst.PrevBalanceProofBody = src.PrevBalanceProofBody
	dst.TransferStepProofBody = src.TransferStepProofBody

	return dst
}

func (dst *EnoughBalanceProofBodyInput) FromEnoughBalanceProofInput(src *EnoughBalanceProofInput) *EnoughBalanceProofBodyInput {
	dst.PrevBalanceProofBody = base64.StdEncoding.EncodeToString(src.PrevBalanceProof.Proof)
	dst.TransferStepProofBody = base64.StdEncoding.EncodeToString(src.TransferStepProof.Proof)

	return dst
}

func (proof *EnoughBalanceProofBodyInput) FromEnoughBalanceProofBody(input *EnoughBalanceProofBody) *EnoughBalanceProofBodyInput {
	proof.PrevBalanceProofBody = base64.StdEncoding.EncodeToString(input.PrevBalanceProofBody)
	proof.TransferStepProofBody = base64.StdEncoding.EncodeToString(input.TransferStepProofBody)

	return proof
}

type EnoughBalanceProofBody struct {
	PrevBalanceProofBody  []byte
	TransferStepProofBody []byte
}

func (proof *EnoughBalanceProofBodyInput) EnoughBalanceProofBody() (*EnoughBalanceProofBody, error) {
	prevBalanceProofBodyBytes, err := base64.StdEncoding.DecodeString(proof.PrevBalanceProofBody)
	if err != nil {
		return nil, err
	}

	transferStepProofBodyBytes, err := base64.StdEncoding.DecodeString(proof.TransferStepProofBody)
	if err != nil {
		return nil, err
	}

	return &EnoughBalanceProofBody{
		PrevBalanceProofBody:  prevBalanceProofBodyBytes,
		TransferStepProofBody: transferStepProofBodyBytes,
	}, nil
}

func (proof *EnoughBalanceProofBody) Hash() string {
	buf := []byte{}
	buf = append(buf, proof.PrevBalanceProofBody...)
	buf = append(buf, proof.TransferStepProofBody...)
	output := crypto.Keccak256(buf)

	return hexutil.Encode(output)
}

type UCBlockSignatureInput struct {
	Sender             string                                     `json:"sender"`
	DecodeSender       *intMaxAcc.PublicKey                       `json:"-"`
	TxHash             string                                     `json:"txHash"` // NOTICE: This is TxTreeRoot, not TxHash
	TxTree             *worker.TxTree                             `json:"-"`
	TxInfo             *worker.TransactionHashesWithSenderAndFile `json:"-"`
	Signature          string                                     `json:"signature"`
	EnoughBalanceProof *EnoughBalanceProofBodyInput               `json:"enoughBalanceProof"`
	BackupTx           *transaction.BackupTransactionData         `json:"backupTx"`
	BackupTransfers    []*transaction.BackupTransferInput         `json:"backupTransfers"`
}

// UseCaseBlockSignature describes BlockSignature contract.
type UseCaseBlockSignature interface {
	Do(ctx context.Context, input *UCBlockSignatureInput) error
}
