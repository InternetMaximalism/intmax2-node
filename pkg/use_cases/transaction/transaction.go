package transaction

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	w   Worker
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	w Worker,
) transaction.UseCaseTransaction {
	return &uc{
		cfg: cfg,
		log: log,
		w:   w,
	}
}

func (u *uc) Do(ctx context.Context, input *transaction.UCTransactionInput) (err error) {
	const (
		hName           = "UseCase Transaction"
		senderKey       = "sender"
		transferHashKey = "transfer_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCInputEmpty)
		return ErrUCInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.String(transferHashKey, input.TransfersHash),
	)

	// TODO: check 0.1 ETH with Rollup contract

	b, err := block_post_service.NewBlockPostService(ctx, u.cfg, u.log)
	if err != nil {
		var ErrNewBlockPostServiceFail = errors.New("new block post service fail")
		return errors.Join(ErrNewBlockPostServiceFail, err)
	}

	// Backup transaction and transfer
	blockNumber := uint64(1) // dummy
	sender, err := intMaxAcc.NewPublicKeyFromAddressHex(input.Sender)
	fmt.Printf("input.EncodedEncryptedTx: %v", input.BackupTx)
	if innerErr := b.BackupTransaction(
		sender.ToAddress(),
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
			fmt.Printf("INTMAX Address: %s\n", intMaxAddress.String())
			if innerErr := b.BackupTransfer(
				intMaxAddress, encodedEncryptedTransfer.EncodedEncryptedTransfer, blockNumber,
			); innerErr != nil {
				open_telemetry.MarkSpanError(spanCtx, innerErr)
				return innerErr
			}
		}

		/*
			else {
				ethAddress, err := recipient.ToEthereumAddress()
				if err != nil {
					open_telemetry.MarkSpanError(spanCtx, err)
					return err
				}

				fmt.Printf("ETH Address: %s\n", ethAddress.String())
				if err := b.BackupWithdrawal(
					ethAddress, encodedEncryptedTransfer.EncodedEncryptedTransfer, blockNumber,
				); err != nil {
					open_telemetry.MarkSpanError(spanCtx, err)
					return err
				}
			}
		*/
	}

	transferData := make([]*intMaxTypes.Transfer, len(input.TransferData))
	for key := range input.TransferData {
		transferData[key] = &intMaxTypes.Transfer{
			Recipient:  input.TransferData[key].DecodeRecipient,
			TokenIndex: uint32(input.TransferData[key].DecodeTokenIndex.Uint64()),
			Amount:     input.TransferData[key].DecodeAmount,
			Salt:       input.TransferData[key].DecodeSalt,
		}
	}

	err = u.w.Receiver(&worker.ReceiverWorker{
		Sender:        input.DecodeSender.ToAddress().String(),
		Nonce:         input.Nonce,
		TransfersHash: input.TransfersHash,
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrTransferWorkerReceiverFail, err)
	}

	return nil
}
