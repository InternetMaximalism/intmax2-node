package block_signature

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/worker"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
) block_signature.UseCaseBlockSignature {
	return &uc{
		cfg: cfg,
		w:   w,
	}
}

var ErrDecodeTxHashFail = errors.New("failed to decode tx hash")
var ErrDecodeSignatureFail = errors.New("failed to decode signature")
var ErrUnmarshalSignatureFail = errors.New("failed to unmarshal signature")

func (u *uc) Do(
	ctx context.Context, input *block_signature.UCBlockSignatureInput,
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

	txTreeRootBytes := input.TxTree.TxTreeHash.Marshal()
	signatureBytes, err := hexutil.Decode(input.Signature)
	if err != nil {
		return errors.Join(ErrDecodeSignatureFail, err)
	}

	// Verify signature.
	err = VerifyTxTreeSignature(signatureBytes, input.DecodeSender, txTreeRootBytes)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return err
	}

	prevBalancePublicInputs, err := backup_balance.VerifyEnoughBalanceProof(input.EnoughBalanceProof.PrevBalanceProof)
	transferPublicInputs, err := VerifyTransferStepProof(input.EnoughBalanceProof.TransferStepProof)
	_ = prevBalancePublicInputs.Equal(&transferPublicInputs.PrevBalancePis)
	// TODO: Check public inputs.
	// if !ok {
	// 	open_telemetry.MarkSpanError(spanCtx, ErrInvalidEnoughBalanceProof)
	// 	return ErrInvalidEnoughBalanceProof
	// }

	// input.TxInfo = &worker.TransactionHashesWithSenderAndFile{
	// 	Sender: input.Sender,
	// 	File:   nil,
	// }

	err = u.w.SignTxTreeByAvailableFile(input.Signature, &worker.TransactionHashesWithSenderAndFile{
		Sender: input.TxInfo.Sender,
		File:   input.TxInfo.File,
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrSignTxTreeByAvailableFileFail, err)
	}

	return nil
}

func VerifyTxTreeSignature(signatureBytes []byte, sender *intMaxAcc.PublicKey, txTreeRootBytes []byte) error {
	messagePoint := finite_field.BytesToFieldElementSlice(txTreeRootBytes)

	signature := new(bn254.G2Affine)
	err := signature.Unmarshal(signatureBytes)
	if err != nil {
		return errors.Join(ErrUnmarshalSignatureFail, err)
	}

	err = intMaxAcc.VerifySignature(signature, sender, messagePoint)
	if err != nil {
		return errors.Join(ErrInvalidSignature, err)
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

func VerifyTransferStepProof(transferStepProof *block_signature.Plonky2Proof) (*TransferStepPublicInputs, error) {
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
