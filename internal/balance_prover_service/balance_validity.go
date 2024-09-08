package balance_prover_service

import (
	"fmt"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
)

type ValidBalanceTransition interface {
	BlockNumber() uint32
}

type ValidSentTx struct {
	TxHash      *poseidonHashOut
	blockNumber uint32
	Tx          *intMaxTypes.TxDetails
}

func (v ValidSentTx) BlockNumber() uint32 {
	return v.blockNumber
}

type ValidReceivedDeposit struct {
	DepositHash common.Hash
	blockNumber uint32
	Deposit     *DepositDetails
}

func (v ValidReceivedDeposit) BlockNumber() uint32 {
	return v.blockNumber
}

type ValidReceivedTransfer struct {
	TransferHash *poseidonHashOut
	blockNumber  uint32
	Transfer     *intMaxTypes.TransferDetailsWithProofBody
}

func (v ValidReceivedTransfer) BlockNumber() uint32 {
	return v.blockNumber
}

func ExtractValidSentTransactions(userData *DecodedUserData, syncValidityProver *syncValidityProver) ([]ValidSentTx, []*poseidonHashOut, error) {
	sentBlockNumbers := make([]ValidSentTx, 0, len(userData.Deposits))
	invalidTxHashes := make([]*poseidonHashOut, 0, len(userData.Deposits))
	for _, tx := range userData.Transactions {
		txHash := tx.Tx.Hash() // TODO: validate transaction

		txRoot := tx.TxTreeRoot.String()[:2]
		blockContent, err := syncValidityProver.ValidityProver.BlockBuilder().BlockContentByTxRoot(txRoot)
		if err != nil {
			fmt.Printf("failed to get block content by tx root %s: %v\n", txHash.String(), err)
			invalidTxHashes = append(invalidTxHashes, txHash)
			continue
		}

		blockNumber := blockContent.BlockNumber
		sentBlockNumbers = append(sentBlockNumbers, ValidSentTx{
			TxHash:      txHash,
			blockNumber: blockNumber,
		})

		fmt.Printf("valid transaction: %s\n", txHash.String())
	}

	return sentBlockNumbers, invalidTxHashes, nil
}

func ExtractValidReceivedDeposits(userData *DecodedUserData, syncValidityProver *syncValidityProver) ([]ValidReceivedDeposit, []common.Hash, error) {
	sentBlockNumbers := make([]ValidReceivedDeposit, 0, len(userData.Deposits))
	invalidDepositHashes := make([]common.Hash, 0, len(userData.Deposits))
	for _, deposit := range userData.Deposits {
		defaultDepositHash := common.Hash{}
		if deposit.DepositHash == defaultDepositHash {
			fmt.Printf("deposit hash should not be zero\n")
			continue
		}

		depositHash := deposit.DepositHash // TODO: validate deposit

		_, depositIndex, err := syncValidityProver.ValidityProver.BlockBuilder().GetDepositLeafAndIndexByHash(depositHash)
		if err != nil {
			fmt.Printf("failed to get deposit index by hash %s: %v\n", depositHash.String(), err)
			invalidDepositHashes = append(invalidDepositHashes, depositHash)
			continue
		}

		blockNumber, err := syncValidityProver.ValidityProver.BlockBuilder().BlockNumberByDepositIndex(*depositIndex)
		if err != nil {
			fmt.Printf("failed to get block number by deposit index %d: %v\n", *depositIndex, err)
			invalidDepositHashes = append(invalidDepositHashes, depositHash)
			continue
		}

		sentBlockNumbers = append(sentBlockNumbers, ValidReceivedDeposit{
			DepositHash: depositHash,
			blockNumber: blockNumber,
		})

		fmt.Printf("valid deposit: %s\n", depositHash.String())
	}

	return sentBlockNumbers, invalidDepositHashes, nil
}

func ExtractValidReceivedTransfers(userData *DecodedUserData, syncValidityProver *syncValidityProver) ([]ValidReceivedTransfer, []*poseidonHashOut, error) {
	receivedBlockNumbers := make([]ValidReceivedTransfer, 0, len(userData.Transfers))
	invalidTransferHashes := make([]*poseidonHashOut, 0, len(userData.Transfers))
	for _, transfer := range userData.Transfers {
		transferHash := transfer.TransferDetails.TransferWitness.Transfer.Hash() // TODO: validate transfer

		txRoot := transfer.TransferDetails.TxTreeRoot.String()[:2]
		blockContent, err := syncValidityProver.ValidityProver.BlockBuilder().BlockContentByTxRoot(txRoot)
		if err != nil {
			fmt.Printf("failed to get block content by transfer root %s: %v\n", transferHash.String(), err)
			invalidTransferHashes = append(invalidTransferHashes, transferHash)
			continue
		}

		blockNumber := blockContent.BlockNumber
		receivedBlockNumbers = append(receivedBlockNumbers, ValidReceivedTransfer{
			TransferHash: transferHash,
			blockNumber:  blockNumber,
		})

		fmt.Printf("valid transfer: %s\n", transferHash.String())
	}

	return receivedBlockNumbers, invalidTransferHashes, nil
}
