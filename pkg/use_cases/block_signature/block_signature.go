package block_signature

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_balance"
	ucBlockSignature "intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/worker"

	"github.com/iden3/go-iden3-crypto/ffg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	w   Worker
}

func New(
	cfg *configs.Config,
	w Worker,
) ucBlockSignature.UseCaseBlockSignature {
	return &uc{
		cfg: cfg,
		w:   w,
	}
}

var ErrDecodeTxHashFail = errors.New("failed to decode tx hash")
var ErrDecodeSignatureFail = errors.New("failed to decode signature")
var ErrUnmarshalSignatureFail = errors.New("failed to unmarshal signature")

func (u *uc) Do(
	ctx context.Context, input *ucBlockSignature.UCBlockSignatureInput,
) (err error) {
	const (
		hName        = "UseCase BlockSignature"
		senderKey    = "sender"
		signatureKey = "signature"
		txHashKey    = "tx_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(senderKey, input.Sender),
			attribute.String(signatureKey, input.Signature),
			attribute.String(txHashKey, input.TxHash),
		))
	defer span.End()

	// NOTICE: Perform signature verification during validation.

	prevBalancePublicInputs, err := backup_balance.VerifyEnoughBalanceProof(input.EnoughBalanceProof.PrevBalanceProof)
	transferPublicInputs, err := VerifyTransferStepProof(input.EnoughBalanceProof.TransferStepProof)
	_ = prevBalancePublicInputs.Equal(&transferPublicInputs.PrevBalancePis)
	/*
		// TODO: Check public inputs.
		if !ok {
			open_telemetry.MarkSpanError(spanCtx, ErrInvalidEnoughBalanceProof)
			return ErrInvalidEnoughBalanceProof
		}

		input.TxInfo = &worker.TransactionHashesWithSenderAndFile{
			Sender: input.Sender,
			File:   nil,
		}
	*/

	err = u.w.SignTxTreeByAvailableFile(
		input.Signature,
		&worker.TransactionHashesWithSenderAndFile{
			Sender: input.TxInfo.Sender,
			TxHash: input.TxInfo.TxHash,
			File:   input.TxInfo.File,
		},
		input.TxTree.TxHash,
		input.TxTree.LeafIndex,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrSignTxTreeByAvailableFileFail, err)
	}

	return nil
}

type TransferStepPublicInputs struct {
	PrevBalancePis backup_balance.BalancePublicInputs
	NextBalancePis backup_balance.BalancePublicInputs
}

func (pis *TransferStepPublicInputs) FromPublicInputs(publicInputs []ffg.Element) *TransferStepPublicInputs {
	return pis
}

func (pis *TransferStepPublicInputs) Verify() error {
	return nil
}

func VerifyTransferStepProof(transferStepProof *ucBlockSignature.Plonky2Proof) (*TransferStepPublicInputs, error) {
	publicInputs := make([]ffg.Element, len(transferStepProof.PublicInputs))
	for i, publicInput := range transferStepProof.PublicInputs {
		publicInputs[i].SetUint64(publicInput)
	}
	decodedPublicInputs := new(TransferStepPublicInputs).FromPublicInputs(publicInputs)
	err := decodedPublicInputs.Verify()
	if err != nil {
		return nil, err
	}

	// TODO: Verify enough balance proof by using Balance Validity Prover.
	return decodedPublicInputs, nil
}
