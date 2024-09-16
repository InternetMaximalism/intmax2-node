package tx_transfer_service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	"intmax2-node/internal/balance_service"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)

const (
	base10Key = 10
	uint64Key = 64
)

func TransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressStr string,
	userEthPrivateKey string,
) error {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return fmt.Errorf("fail to get wallet from private key: %w", err)
	}

	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return fmt.Errorf("fail to parse user private key: %w", err)
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, cfg, sb, *tokenInfo)
	if err != nil {
		return err
	}

	// TODO: Create balance proof locally
	// fmt.Printf("User's INTMAX Address: %s\n", userAccount.ToAddress().String())
	// balanceProver := balance_prover_service.NewSyncBalanceProver()
	// balanceSynchronizer := balance_prover_service.NewSynchronizer(ctx, cfg, log, sb, db)
	// blockSynchronizer, err := block_synchronizer.NewBlockSynchronizer(
	// 	ctx, cfg, log,
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to create block synchronizer: %w", err)
	// }
	// blockValidityProver, err := block_validity_prover.NewBlockValidityProver(ctx, cfg, log, sb, db)
	// if err != nil {
	// 	return fmt.Errorf("failed to create block validity prover: %w", err)
	// }

	// // syncValidityProver, err := balance_prover_service.NewSyncValidityProver(ctx, cfg, log, sb, db)
	// // if err != nil {
	// // 	return fmt.Errorf("failed to create sync validity prover: %w", err)
	// // }
	// balanceProcessor := balance_prover_service.NewBalanceProcessor(ctx, cfg, log)

	// // balanceProcessor *BalanceProcessor,
	// err = SyncLocally(ctx, cfg, log, balanceProver, balanceSynchronizer, blockValidityProver, blockSynchronizer, balanceProcessor, userAccount)
	// if err != nil {
	// 	return fmt.Errorf("failed to sync balance proof: %w", err)
	// }

	fmt.Println("Fetching balances...")
	// log := logger.NewCommandLineLogger()
	// blockValidityService := balance_prover_service.NewExternalBlockValidityProver(ctx, cfg)
	// balanceSynchronizer, err := balance_synchronizer.SyncLocally(
	// 	ctx,
	// 	cfg,
	// 	log,
	// 	sb,
	// 	blockValidityService,
	// 	userAccount,
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to sync balance proof: %w", err)
	// }

	l1Balance, err := balance_service.GetTokenBalance(ctx, cfg, log, sb, *wallet.WalletAddress, *tokenInfo)
	if err != nil {
		return fmt.Errorf(ErrFailedToGetBalanceEth, "Ethereum")
	}

	balance, err := balance_service.GetUserBalance(ctx, cfg, userAccount, tokenIndex)
	if err != nil {
		return fmt.Errorf(ErrFailedToGetBalance.Error()+": %v", err)
	}

	if strings.TrimSpace(amountStr) == "" {
		return fmt.Errorf("amount is required")
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", amountStr)
	}

	if balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient funds for total amount: balance %s, total amount %s", balance, amount)
	}

	var dataBlockInfo *BlockInfoResponseData
	dataBlockInfo, err = GetBlockInfo(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get the block info data: %w", err)
	}

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	var initialLeaves []*intMaxTypes.Transfer

	gasFee, gasOK := dataBlockInfo.TransferFee[new(big.Int).SetUint64(uint64(tokenIndex)).String()]
	if gasOK {
		// Send transfer transaction
		var recipient *intMaxAcc.PublicKey
		recipient, err = intMaxAcc.NewPublicKeyFromAddressHex(dataBlockInfo.IntMaxAddress)
		if err != nil {
			return fmt.Errorf("failed to parse recipient address: %v", err)
		}

		var recipientAddress *intMaxTypes.GenericAddress
		recipientAddress, err = intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
		if err != nil {
			return fmt.Errorf("failed to create recipient address: %v", err)
		}

		var amountGasFee uint256.Int
		err = amountGasFee.Scan(gasFee)
		if err != nil {
			return fmt.Errorf("failed to convert string to uint256.Int: %w", err)
		}

		transfer := intMaxTypes.NewTransferWithRandomSalt(
			recipientAddress,
			tokenIndex,
			amountGasFee.ToBig(),
		)

		initialLeaves = append(initialLeaves, transfer)
	}

	const base10 = 10
	gasFeeInt, ok := new(big.Int).SetString(gasFee, base10)
	if !ok {
		return fmt.Errorf("failed to convert gas fee to int: %w", err)
	}
	totalAmountWithGas := new(big.Int).Add(amount, gasFeeInt)
	if l1Balance.Cmp(totalAmountWithGas) < 0 {
		return fmt.Errorf("insufficient funds for tx cost: l1Balance %s, tx cost %s", l1Balance, totalAmountWithGas)
	}

	// Send transfer transaction
	var recipient *intMaxAcc.PublicKey
	recipient, err = intMaxAcc.NewPublicKeyFromAddressHex(recipientAddressStr)
	if err != nil {
		return fmt.Errorf("failed to parse recipient address: %v", err)
	}

	var recipientAddress *intMaxTypes.GenericAddress
	recipientAddress, err = intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
	if err != nil {
		return fmt.Errorf("failed to create recipient address: %v", err)
	}

	transfer := intMaxTypes.NewTransferWithRandomSalt(
		recipientAddress,
		tokenIndex,
		amount,
	)

	initialLeaves = append(initialLeaves, transfer)

	var transferTree *intMaxTree.TransferTree
	transferTree, err = intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		return fmt.Errorf("failed to create transfer tree: %v", err)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	// lastBalanceProof := ""
	// balanceTransitionProof := ""
	// nonce := balanceSynchronizer.CurrentNonce() + 1
	nonce := uint32(1) // TODO: Get nonce from balance synchronizer

	err = SendTransferTransaction(
		ctx,
		cfg,
		userAccount,
		transfersHash,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	var proposedBlock *BlockProposedResponseData
	proposedBlock, err = GetBlockProposed(
		ctx, cfg, userAccount, transfersHash, nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The proposed block has been successfully received.")

	var tx *intMaxTypes.Tx
	tx, err = intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()

	txDetails := intMaxTypes.TxDetails{
		Tx: intMaxTypes.Tx{
			TransferTreeRoot: &transfersHash,
			Nonce:            nonce,
		},
		Transfers:     initialLeaves,
		TxTreeRoot:    &proposedBlock.TxTreeRoot,
		TxMerkleProof: proposedBlock.TxTreeMerkleProof,
	}
	backupTx, err := transaction.NewBackupTransactionData(
		userAccount.Public(),
		txDetails,
		txHash,
		"0x",
	)
	if err != nil {
		return fmt.Errorf("failed to make backup transaction data: %v", err)
	}

	// lastBalanceProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(lastBalanceProof)
	// if err != nil {
	// 	return fmt.Errorf("failed to create last balance proof: %v", err)
	// }

	// balanceTransitionProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceTransitionProof)
	// if err != nil {
	// 	return fmt.Errorf("failed to create balance transition proof: %v", err)
	// }

	backupTransfers := make([]*transaction.BackupTransferInput, len(initialLeaves))
	for i := range initialLeaves {
		transferMerkleProof, _, err := transferTree.ComputeMerkleProof(uint64(i))
		if err != nil {
			return fmt.Errorf("failed to compute merkle proof: %v", err)
		}
		transferWitness := intMaxTypes.TransferWitness{
			Transfer:            *initialLeaves[i],
			TransferIndex:       uint32(i),
			Tx:                  *tx,
			TransferMerkleProof: transferMerkleProof,
		}
		transferDetails := intMaxTypes.TransferDetails{
			TransferWitness: &transferWitness,
			TxTreeRoot:      &proposedBlock.TxTreeRoot,
			TxMerkleProof:   proposedBlock.TxTreeMerkleProof,
			// SenderLastBalancePublicInputs:       lastBalanceProofWithPis.PublicInputsBytes(),
			// SenderBalanceTransitionPublicInputs: balanceTransitionProofWithPis.PublicInputsBytes(),
		}
		backupTransfers[i], err = MakeTransferBackupData(
			&transferDetails,
			// lastBalanceProofWithPis.ProofBase64String(),
			// balanceTransitionProofWithPis.ProofBase64String(),
		)
		if err != nil {
			return fmt.Errorf("failed to make backup transfer data: %v", err)
		}
		fmt.Printf("SenderLastBalanceProofBody[%d]: %v\n", i, backupTransfers[i].SenderLastBalanceProofBody)
		fmt.Printf("SenderTransitionProofBody[%d]: %v\n", i, backupTransfers[i].SenderTransitionProofBody)
	}

	// Accept proposed block
	err = SendSignedProposedBlock(
		ctx, cfg, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
		backupTx, backupTransfers,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction has been successfully sent.")

	return nil
}

var ErrFailedToCreateRecipientAddress = errors.New("failed to create recipient address")
var ErrFailedToGetRecipientPublicKey = errors.New("failed to get recipient public key")
var ErrFailedToEncryptTransfer = errors.New("failed to encrypt transfer")
var ErrFailedToDecodeFromBase64 = errors.New("failed to decode from base64")
var ErrFailedToDecrypt = errors.New("failed to decrypt")
var ErrFailedToUnmarshal = errors.New("failed to unmarshal")

func MakeTransferBackupData(
	transferDetails *intMaxTypes.TransferDetails,
	// senderLastBalanceProofBody string,
	// senderBalanceTransitionProofBody string,
) (backupTransfer *transaction.BackupTransferInput, _ error) {
	// if len(senderLastBalanceProofBody) == 0 {
	// 	return nil, errors.New("sender last balance proof body is empty")
	// }

	// if len(senderBalanceTransitionProofBody) == 0 {
	// 	return nil, errors.New("sender balance transition proof body is empty")
	// }

	transfer := transferDetails.TransferWitness.Transfer
	if transfer.Recipient.TypeOfAddress != intMaxAccTypes.INTMAXAddressType {
		return nil, errors.New("recipient address should be INTMAX")
	}

	recipientAddress, err := transfer.Recipient.ToINTMAXAddress()
	if err != nil {
		return nil, errors.Join(ErrFailedToCreateRecipientAddress, err)
	}
	recipientPublicKey, err := recipientAddress.Public()
	if err != nil {
		return nil, errors.Join(ErrFailedToGetRecipientPublicKey, err)
	}

	encryptedTransfer, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		recipientPublicKey,
		transferDetails.Marshal(),
	)
	if err != nil {
		return nil, errors.Join(ErrFailedToEncryptTransfer, err)
	}

	return &transaction.BackupTransferInput{
		Recipient:                hexutil.Encode(transfer.Recipient.Marshal()),
		TransferHash:             transfer.Hash().String(),
		EncodedEncryptedTransfer: base64.StdEncoding.EncodeToString(encryptedTransfer),
		// SenderLastBalanceProofBody: senderLastBalanceProofBody,
		// SenderTransitionProofBody:  senderBalanceTransitionProofBody,
		SenderLastBalanceProofBody: "",
		SenderTransitionProofBody:  "",
	}, nil
}

func GetTransactionFromBackupData(
	encryptedTransaction *GetTransactionData,
	senderAccount *intMaxAcc.PrivateKey,
) (*intMaxTypes.TxDetails, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedTransaction.EncryptedTx)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeFromBase64, err)
	}

	var message []byte
	message, err = senderAccount.DecryptECIES(ciphertext)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecrypt, err)
	}

	var txDetails intMaxTypes.TxDetails
	err = txDetails.Unmarshal(message)
	if err != nil {
		return nil, errors.Join(ErrFailedToUnmarshal, err)
	}

	return &txDetails, nil
}

func GetSignatureFromBackupData(
	encryptedSignature string,
	senderAccount *intMaxAcc.PrivateKey,
) (*bn254.G2Affine, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedSignature)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeFromBase64, err)
	}

	var message []byte
	message, err = senderAccount.DecryptECIES(ciphertext)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecrypt, err)
	}

	var sign bn254.G2Affine
	err = sign.Unmarshal(message)
	if err != nil {
		return nil, errors.Join(ErrFailedToUnmarshal, err)
	}

	return &sign, nil
}

type BackupWithdrawal struct {
	SenderAddress       intMaxAcc.Address                 `json:"senderAddress"`
	Transfer            *intMaxTypes.Transfer             `json:"transfer"`
	TransferMerkleProof []*intMaxTypes.PoseidonHashOut    `json:"transferMerkleProof"`
	TransferIndex       int32                             `json:"transferIndex"`
	TransferTreeRoot    intMaxTypes.PoseidonHashOut       `json:"transferTreeRoot"`
	Nonce               uint64                            `json:"nonce"`
	TxTreeMerkleProof   []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`
	TxIndex             int32                             `json:"txIndex"`
	TxTreeRoot          goldenposeidon.PoseidonHashOut    `json:"txTreeRoot"`
}

type withdrawalTransfer struct {
	Recipient common.Address `json:"recipient"`

	TokenIndex uint32 `json:"tokenIndex"`

	// Amount is a decimal string
	Amount string `json:"amount"`

	Salt *goldenposeidon.PoseidonHashOut `json:"salt"`
}

type backupWithdrawal struct {
	SenderAddress string `json:"senderAddress"`

	Transfer withdrawalTransfer `json:"transfer"`

	TransferMerkleProof []*goldenposeidon.PoseidonHashOut `json:"transferMerkleProof"`

	TransferIndex int32 `json:"transferIndex"`

	TransferTreeRoot goldenposeidon.PoseidonHashOut `json:"transferTreeRoot"`

	// Nonce is a decimal string
	Nonce string `json:"nonce"`

	TxTreeMerkleProof []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`

	TxIndex int32 `json:"txIndex"`

	TxTreeRoot goldenposeidon.PoseidonHashOut `json:"txTreeRoot"`
}

func (bw *BackupWithdrawal) MarshalJSON() ([]byte, error) {
	if bw.Transfer.Recipient.TypeOfAddress != intMaxAccTypes.EthereumAddressType {
		return nil, errors.New("recipient address should be ETHEREUM")
	}

	recipient, err := bw.Transfer.Recipient.ToEthereumAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient address: %w", err)
	}

	return json.Marshal(&backupWithdrawal{
		SenderAddress: bw.SenderAddress.String(),
		Transfer: withdrawalTransfer{
			Recipient:  recipient,
			TokenIndex: bw.Transfer.TokenIndex,
			Amount:     bw.Transfer.Amount.String(),
			Salt:       bw.Transfer.Salt,
		},
		TransferMerkleProof: bw.TransferMerkleProof,
		TransferIndex:       bw.TransferIndex,
		TransferTreeRoot:    bw.TransferTreeRoot,
		Nonce:               strconv.FormatUint(bw.Nonce, base10Key),
		TxTreeMerkleProof:   bw.TxTreeMerkleProof,
		TxIndex:             bw.TxIndex,
		TxTreeRoot:          bw.TxTreeRoot,
	})
}

func (bw *BackupWithdrawal) UnmarshalJSON(data []byte) error {
	var withdrawal backupWithdrawal
	err := json.Unmarshal(data, &withdrawal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	recipientBytes := withdrawal.Transfer.Recipient.Bytes()
	recipient, err := intMaxTypes.NewEthereumAddress(recipientBytes)
	if err != nil {
		return fmt.Errorf("failed to create recipient address: %w", err)
	}
	amount, ok := new(big.Int).SetString(withdrawal.Transfer.Amount, base10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", withdrawal.Transfer.Amount)
	}
	transfer := intMaxTypes.NewTransfer(
		recipient,
		withdrawal.Transfer.TokenIndex,
		amount,
		withdrawal.Transfer.Salt,
	)

	nonce, err := strconv.ParseUint(withdrawal.Nonce, base10Key, uint64Key)
	if err != nil {
		return fmt.Errorf("failed to parse nonce: %w", err)
	}

	senderAddress, err := intMaxAcc.NewAddressFromHex(withdrawal.SenderAddress)
	if err != nil {
		return fmt.Errorf("failed to parse sender address: %w", err)
	}

	bw.SenderAddress = senderAddress
	bw.Transfer = transfer
	bw.TransferMerkleProof = withdrawal.TransferMerkleProof
	bw.TransferIndex = withdrawal.TransferIndex
	bw.TransferTreeRoot = withdrawal.TransferTreeRoot
	bw.Nonce = nonce
	bw.TxTreeMerkleProof = withdrawal.TxTreeMerkleProof
	bw.TxIndex = withdrawal.TxIndex
	bw.TxTreeRoot = withdrawal.TxTreeRoot

	return nil
}

func MakeWithdrawalBackupData(
	transfer *intMaxTypes.Transfer,
	senderAddress intMaxAcc.Address,
	transfersHash goldenposeidon.PoseidonHashOut,
	nonce uint32,
	txTreeRoot goldenposeidon.PoseidonHashOut,
	txTreeMerkleProof []*goldenposeidon.PoseidonHashOut,
	transferMerkleProof []*goldenposeidon.PoseidonHashOut,
	txIndex int32,
	transferIndex int32,
) (backupTransfer *transaction.BackupTransferInput, _ error) {
	if transfer.Recipient.TypeOfAddress != "ETHEREUM" {
		return nil, errors.New("recipient address should be ETHEREUM")
	}

	recipient, err := transfer.Recipient.ToEthereumAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient address: %w", err)
	}

	withdrawal := backupWithdrawal{
		SenderAddress: senderAddress.String(),
		Transfer: withdrawalTransfer{
			Recipient:  recipient,
			TokenIndex: transfer.TokenIndex,
			Amount:     transfer.Amount.String(),
			Salt:       transfer.Salt,
		},
		TransferMerkleProof: transferMerkleProof,
		TransferTreeRoot:    transfersHash,
		TransferIndex:       transferIndex,
		Nonce:               strconv.FormatUint(uint64(nonce), base10Key),
		TxTreeMerkleProof:   txTreeMerkleProof,
		TxTreeRoot:          txTreeRoot,
		TxIndex:             txIndex,
	}

	// No encryption
	encryptedTransfer, err := json.Marshal(&withdrawal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// TODO: Only use one of the implementations.
	encryptedTransfer2, err := json.Marshal(&BackupWithdrawal{
		SenderAddress:       senderAddress,
		Transfer:            transfer,
		TransferMerkleProof: transferMerkleProof,
		TransferIndex:       transferIndex,
		TransferTreeRoot:    transfersHash,
		Nonce:               uint64(nonce),
		TxTreeMerkleProof:   txTreeMerkleProof,
		TxIndex:             txIndex,
		TxTreeRoot:          txTreeRoot,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if !bytes.Equal(encryptedTransfer, encryptedTransfer2) {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return &transaction.BackupTransferInput{
		Recipient:                hexutil.Encode(transfer.Recipient.Marshal()),
		EncodedEncryptedTransfer: base64.StdEncoding.EncodeToString(encryptedTransfer),
	}, nil
}

func TransactionsList(
	ctx context.Context,
	cfg *configs.Config,
	input *GetTransactionsListInput,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return nil, errors.Join(ErrRecoverWalletFromPrivateKey, err)
	}

	fmt.Printf("User's INTMAX Address: %s\n", userAccount.ToAddress().String())

	return GetTransactionsListWithRawRequest(ctx, cfg, input, userAccount)
}

func TransactionByHash(
	ctx context.Context,
	cfg *configs.Config,
	txHash string,
	userEthPrivateKey string,
) (json.RawMessage, error) {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return nil, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return nil, errors.Join(ErrRecoverWalletFromPrivateKey, err)
	}

	return GetTransactionByHashWithRawRequest(ctx, cfg, txHash, userAccount)
}
