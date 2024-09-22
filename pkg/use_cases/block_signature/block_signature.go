package block_signature

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_balance"
	ucBlockSignature "intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/worker"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	w   Worker
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	w Worker,
) ucBlockSignature.UseCaseBlockSignature {
	return &uc{
		cfg: cfg,
		log: log,
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

	/*
		prevBalancePublicInputs, err := backup_balance.VerifyEnoughBalanceProof(input.EnoughBalanceProof.PrevBalanceProof)
		if err != nil {
			return err
		}
		transferPublicInputs, err := VerifyTransferStepProof(input.EnoughBalanceProof.TransferStepProof)
		if err != nil {
			return err
		}
		_ = prevBalancePublicInputs.Equal(&transferPublicInputs.PrevBalancePis)

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
	b, err := block_synchronizer.NewBlockSynchronizer(ctx, u.cfg, u.log)
	if err != nil {
		var ErrNewBlockPostServiceFail = errors.New("new block post service fail")
		return errors.Join(ErrNewBlockPostServiceFail, err)
	}

	// Backup transaction and transfer
	blockNumber := uint64(1) // dummy
	sender, err := intMaxAcc.NewPublicKeyFromAddressHex(input.Sender)
	if innerErr := b.BackupTransaction(
		sender.ToAddress(),
		input.BackupTx.TxHash,
		input.BackupTx.EncodedEncryptedTx,
		input.BackupTx.Signature,
		blockNumber,
	); innerErr != nil {
		open_telemetry.MarkSpanError(spanCtx, innerErr)
		return innerErr
	}

	for i := 0; i < len(input.BackupTransfers); i++ {
		encodedEncryptedTransfer := input.BackupTransfers[i]
		var addressBytes []byte
		addressBytes, err = hexutil.Decode(encodedEncryptedTransfer.Recipient)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return err
		}
		recipient := new(intMaxTypes.GenericAddress)
		err = recipient.Unmarshal(addressBytes)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return err
		}

		// TODO: Write the process when the recipient is Ethereum.
		if recipient.TypeOfAddress == "INTMAX" {
			var intMaxAddress intMaxAcc.Address
			intMaxAddress, err = recipient.ToINTMAXAddress()
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return err
			}
			u.log.Printf("INTMAX Address: %s\n", intMaxAddress.String())

			if encodedEncryptedTransfer.SenderLastBalanceProofBody != "" {
				var senderLastBalanceProofBody []byte
				senderLastBalanceProofBody, err = hexutil.Decode(encodedEncryptedTransfer.SenderLastBalanceProofBody)
				if err != nil {
					open_telemetry.MarkSpanError(spanCtx, err)
					return errors.Join(ErrDecodeSenderLastBalanceProofBodyFail, err)
				}

				var senderBalanceTransitionProofBody []byte
				senderBalanceTransitionProofBody, err = hexutil.Decode(encodedEncryptedTransfer.SenderTransitionProofBody)
				if err != nil {
					open_telemetry.MarkSpanError(spanCtx, err)
					return errors.Join(ErrDecodeSenderTransitionProofBodyFail, err)
				}

				if innerErr := b.BackupTransfer(
					intMaxAddress, encodedEncryptedTransfer.TransferHash, encodedEncryptedTransfer.EncodedEncryptedTransfer,
					senderLastBalanceProofBody, senderBalanceTransitionProofBody, blockNumber,
				); innerErr != nil {
					open_telemetry.MarkSpanError(spanCtx, innerErr)
					return errors.Join(ErrBackupTransferFail, innerErr)
				}
			} else {
				if innerErr := b.BackupTransfer(
					intMaxAddress, encodedEncryptedTransfer.TransferHash, encodedEncryptedTransfer.EncodedEncryptedTransfer, nil, nil, blockNumber,
				); innerErr != nil {
					open_telemetry.MarkSpanError(spanCtx, innerErr)
					return errors.Join(ErrBackupTransferFail, innerErr)
				}
			}
		} else {
			var ethAddress common.Address
			ethAddress, err = recipient.ToEthereumAddress()
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return err
			}

			u.log.Printf("ETH Address: %s\n", ethAddress.String())
			if innerErr := b.BackupWithdrawal(
				ethAddress, encodedEncryptedTransfer.TransferHash, encodedEncryptedTransfer.EncodedEncryptedTransfer, blockNumber,
			); innerErr != nil {
				open_telemetry.MarkSpanError(spanCtx, innerErr)
				return innerErr
			}
		}
	}

	err = u.w.SignTxTreeByAvailableFile(
		input.Signature,
		&worker.TransactionHashesWithSenderAndFile{
			Sender: input.TxInfo.Sender,
			TxHash: input.TxInfo.TxHash,
			File:   input.TxInfo.File,
		},
		input.TxTree.LeafIndex,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrSignTxTreeByAvailableFileFail, err)
	}

	return nil
}

type TransferStepPublicInputs struct {
	PrevBalancePis backup_balance.BalancePublicInputs `json:"prevBalancePis"`
	NextBalancePis backup_balance.BalancePublicInputs `json:"nextBalancePis"`
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
