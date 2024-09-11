package balance_prover_service

import (
	"fmt"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
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

func ExtractValidSentTransactions(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityProver block_validity_prover.BlockValidityProver,
) ([]ValidSentTx, []*poseidonHashOut, error) {
	sentBlockNumbers := make([]ValidSentTx, 0, len(userData.Deposits))
	invalidTxHashes := make([]*poseidonHashOut, 0, len(userData.Deposits))
	for _, tx := range userData.Transactions {
		txHash := tx.Tx.Hash() // TODO: validate transaction

		fmt.Printf("transaction hash: %s\n", txHash.String())
		if tx.TxTreeRoot == nil {
			// TODO: If TxTreeRoot is nil, the account is no longer valid.
			log.Warnf("transaction tx tree root is nil\n")
			invalidTxHashes = append(invalidTxHashes, txHash)
			continue
		}

		txRoot := tx.TxTreeRoot.String()[:2]
		blockContent, err := blockValidityProver.BlockContentByTxRoot(txRoot)
		if err != nil {
			log.Warnf("failed to get block content by tx root %s: %v\n", txHash.String(), err)
			continue
		}

		blockNumber := blockContent.BlockNumber
		sentBlockNumbers = append(sentBlockNumbers, ValidSentTx{
			TxHash:      txHash,
			blockNumber: blockNumber,
			Tx:          tx,
		})

		log.Debugf("valid transaction: %s\n", txHash.String())
	}

	return sentBlockNumbers, invalidTxHashes, nil
}

func ExtractValidReceivedDeposits(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityProver block_validity_prover.BlockValidityProver,
) ([]ValidReceivedDeposit, []common.Hash, error) {
	sentBlockNumbers := make([]ValidReceivedDeposit, 0, len(userData.Deposits))
	invalidDepositHashes := make([]common.Hash, 0, len(userData.Deposits))
	for _, deposit := range userData.Deposits {
		defaultDepositHash := common.Hash{}
		if deposit.DepositHash == defaultDepositHash {
			log.Warnf("deposit hash should not be zero\n")
			continue
		}

		depositHash := deposit.DepositHash // TODO: validate deposit
		log.Debugf("deposit hash: %s\n", depositHash.String())

		_, depositIndex, err := blockValidityProver.GetDepositLeafAndIndexByHash(depositHash)
		if err != nil {
			log.Warnf("failed to get deposit index by hash %s: %v\n", depositHash.String(), err)
			continue
		}

		blockNumber, err := blockValidityProver.BlockNumberByDepositIndex(*depositIndex)
		if err != nil {
			log.Warnf("failed to get block number by deposit index %d: %v\n", *depositIndex, err)
			continue
		}

		sentBlockNumbers = append(sentBlockNumbers, ValidReceivedDeposit{
			DepositHash: depositHash,
			blockNumber: blockNumber,
			Deposit:     deposit,
		})

		log.Debugf("valid deposit: %s\n", depositHash.String())
	}

	return sentBlockNumbers, invalidDepositHashes, nil
}

func ExtractValidReceivedTransfers(
	log logger.Logger,
	userData *BalanceTransitionData,
	blockValidityProver block_validity_prover.BlockValidityProver,
) ([]ValidReceivedTransfer, []*poseidonHashOut, error) {
	receivedBlockNumbers := make([]ValidReceivedTransfer, 0, len(userData.Transfers))
	invalidTransferHashes := make([]*poseidonHashOut, 0, len(userData.Transfers))
	for _, transfer := range userData.Transfers {
		transferHash := transfer.TransferDetails.TransferWitness.Transfer.Hash() // TODO: validate transfer

		log.Debugf("transfer hash: %s\n", transferHash.String())
		if transfer.TransferDetails.TxTreeRoot == nil {
			log.Warnf("transfer tx tree root is nil\n")
			invalidTransferHashes = append(invalidTransferHashes, transferHash)
			continue
		}

		txRoot := transfer.TransferDetails.TxTreeRoot.String()[:2]
		blockContent, err := blockValidityProver.BlockContentByTxRoot(txRoot)
		if err != nil {
			log.Warnf("failed to get block content by transfer root %s: %v\n", transferHash.String(), err)
			continue
		}

		blockNumber := blockContent.BlockNumber
		receivedBlockNumbers = append(receivedBlockNumbers, ValidReceivedTransfer{
			TransferHash: transferHash,
			blockNumber:  blockNumber,
			Transfer:     transfer,
		})

		log.Debugf("valid transfer: %s\n", transferHash.String())
	}

	return receivedBlockNumbers, invalidTransferHashes, nil
}
