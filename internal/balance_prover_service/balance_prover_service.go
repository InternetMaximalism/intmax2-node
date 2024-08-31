package balance_prover_service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet/models"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

type balanceProverService struct {
	ctx               context.Context
	cfg               *configs.Config
	log               logger.Logger
	wallet            *models.Wallet
	BalanceProcessor  *BalanceProcessor
	SyncBalanceProver *SyncBalanceProver
}

func NewBalanceProverService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	wallet *models.Wallet,
) *balanceProverService {
	balanceProcessor := NewBalanceProcessor(
		ctx,
		cfg,
		log,
	)
	syncBalanceProver := NewSyncBalanceProver()
	return &balanceProverService{
		ctx,
		cfg,
		log,
		wallet,
		balanceProcessor,
		syncBalanceProver,
	}
}

func (s *balanceProverService) DecodeUserData() (*DecodedUserData, error) {
	fmt.Printf("Starting balance prover service: %s\n", s.wallet.IntMaxWalletAddress)

	userAllData, err := balance_service.GetUserBalancesRawRequest(s.ctx, s.cfg, s.wallet.IntMaxWalletAddress)
	if err != nil {
		const msg = "failed to get user all data: %+v"
		s.log.Fatalf(msg, err.Error())
	}

	intMaxPrivateKey, err := intMaxAcc.NewPrivateKeyFromString(s.wallet.IntMaxPrivateKey)
	if err != nil {
		return nil, err
	}

	decodedUserAllData, err := DecodeBackupData(s.ctx, s.cfg, userAllData, intMaxPrivateKey)
	if err != nil {
		return nil, err
	}

	fmt.Printf("user deposits: %d\n", len(decodedUserAllData.Deposits))
	fmt.Printf("user transfers: %d\n", len(decodedUserAllData.Transfers))
	fmt.Printf("user transactions: %d\n", len(decodedUserAllData.Transactions))
	fmt.Println("Finished balance prover service")
	return decodedUserAllData, nil
}

type DepositDetails struct {
	Recipient         *intMaxAcc.PublicKey
	TokenIndex        uint32
	Amount            *big.Int
	Salt              *intMaxGP.PoseidonHashOut
	RecipientSaltHash common.Hash
	DepositID         uint32
	DepositHash       common.Hash
}

type DecodedUserData struct {
	Transactions []*intMaxTypes.TxDetails
	Deposits     []*DepositDetails
	Transfers    []*intMaxTypes.Transfer
}

func DecodeBackupData(
	ctx context.Context,
	cfg *configs.Config,
	userAllData *balance_service.GetBalancesResponse,
	userPrivateKey *intMaxAcc.PrivateKey,
) (*DecodedUserData, error) {
	receivedDeposits := make([]*DepositDetails, 0)
	receivedTransfers := make([]*intMaxTypes.Transfer, 0)
	sentTransactions := make([]*intMaxTypes.TxDetails, 0)

	address := userPrivateKey.Public().ToAddress()
	fmt.Printf("Decoding backup data for address: %s\n", address.String())

	for _, deposit := range userAllData.Deposits {
		encryptedDepositBytes, err := base64.StdEncoding.DecodeString(deposit.EncryptedDeposit)
		if err != nil {
			log.Printf("failed to decode deposit: %v", err)
			continue
		}

		encodedDeposit, err := userPrivateKey.DecryptECIES(encryptedDepositBytes)
		if err != nil {
			log.Printf("failed to decrypt deposit: %v", err)
			continue
		}

		var decodedDeposit intMaxTypes.Deposit
		err = decodedDeposit.Unmarshal(encodedDeposit)
		if err != nil {
			log.Printf("failed to unmarshal deposit: %v", err)
			continue
		}

		// Request data store vault if deposit is valid
		depositIDStr := deposit.BlockNumber
		depositID, err := strconv.ParseUint(depositIDStr, 10, 32)
		for err != nil {
			log.Printf("failed to parse deposit ID: %v", err)
		}

		ok, err := balance_service.GetDepositValidityRawRequest(
			ctx,
			cfg,
			depositIDStr,
		)
		if err != nil {
			return nil, errors.Join(balance_service.ErrDepositValidity, err)
		}
		if !ok {
			continue
		}

		recipient, err := intMaxAcc.NewPublicKeyFromAddressHex(deposit.Recipient)
		if err != nil {
			log.Printf("failed to create recipient public key: %v", err)
			continue
		}

		recipientSaltHash := recipient.HashWithSalt(decodedDeposit.Salt)
		depositLeaf := intMaxTree.DepositLeaf{
			RecipientSaltHash: recipientSaltHash,
			TokenIndex:        decodedDeposit.TokenIndex,
			Amount:            decodedDeposit.Amount,
		}
		depositHash := depositLeaf.Hash()
		fmt.Printf("deposit ID: %v\n", depositID)
		fmt.Printf("deposit leaf: %v\n", depositLeaf)
		fmt.Printf("deposit (nullifier): %s\n", depositHash.String())
		deposit := DepositDetails{
			Recipient:         recipient,
			TokenIndex:        decodedDeposit.TokenIndex,
			Amount:            decodedDeposit.Amount,
			Salt:              decodedDeposit.Salt,
			RecipientSaltHash: recipientSaltHash,
			DepositID:         uint32(depositID),
			DepositHash:       depositHash,
		}

		receivedDeposits = append(receivedDeposits, &deposit)

		// if _, ok := balances[tokenIndex]; !ok {
		// 	balances[tokenIndex] = big.NewInt(0)
		// }

		// balances[tokenIndex] = new(big.Int).Add(balances[tokenIndex], decodedDeposit.Amount)
	}

	for _, transfer := range userAllData.Transfers {
		encryptedTransferBytes, err := base64.StdEncoding.DecodeString(transfer.EncryptedTransfer)
		if err != nil {
			log.Printf("failed to decode transfer: %v", err)
			continue
		}
		encodedTransfer, err := userPrivateKey.DecryptECIES(encryptedTransferBytes)
		if err != nil {
			log.Printf("failed to decrypt transfer: %v", err)
			continue
		}
		var decodedTransfer intMaxTypes.Transfer
		err = decodedTransfer.Unmarshal(encodedTransfer)
		if err != nil {
			log.Printf("failed to unmarshal transfer: %v", err)
			continue
		}

		receivedTransfers = append(receivedTransfers, &decodedTransfer)

		// if _, ok := balances[tokenIndex]; !ok {
		// 	balances[tokenIndex] = big.NewInt(0)
		// }

		// balances[tokenIndex] = new(big.Int).Add(balances[tokenIndex], decodedTransfer.Amount)
	}

	for _, transaction := range userAllData.Transactions {
		encryptedTxBytes, err := base64.StdEncoding.DecodeString(transaction.EncryptedTx)
		if err != nil {
			log.Printf("failed to decode transaction: %v", err)
			continue
		}
		encodedTx, err := userPrivateKey.DecryptECIES(encryptedTxBytes)
		if err != nil {
			log.Printf("failed to decrypt transaction: %v", err)
			continue
		}
		var decodedTx intMaxTypes.TxDetails
		err = decodedTx.Unmarshal(encodedTx)
		if err != nil {
			log.Printf("failed to unmarshal transaction: %v", err)
			continue
		}

		sentTransactions = append(sentTransactions, &decodedTx)
		// for _, transfer := range decodedTx.Transfers {
		// 	if _, ok := balances[transfer.TokenIndex]; !ok {
		// 		balances[transfer.TokenIndex] = big.NewInt(0)
		// 	}
		// 	balances[transfer.TokenIndex] = new(big.Int).Sub(balances[transfer.TokenIndex], transfer.Amount)
		// }
	}

	return &DecodedUserData{
		Transactions: sentTransactions,
		Deposits:     receivedDeposits,
		Transfers:    receivedTransfers,
	}, nil
}
