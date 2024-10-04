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
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/balance_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_signature"
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
	db block_validity_prover.SQLDriverApp, // TODO: Remove this
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

	tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, cfg, log, sb, *tokenInfo)
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
	// blockValidityProver, err := block_validity_prover.NewBlockValidityProver(ctx, cfg, log, sb, db)
	// if err != nil {
	// 	const msg = "failed to start Block Validity Prover: %+v"
	// 	log.Fatalf(msg, err.Error())
	// }
	blockValidityService, err := block_validity_prover.NewBlockValidityService(ctx, cfg, log, sb, db)
	if err != nil {
		const msg = "failed to start Block Validity Service: %+v"
		log.Fatalf(msg, err.Error())
	}

	// wg := sync.WaitGroup{}

	// wg.Add(1)
	// go func() {
	// 	defer func() {
	// 		wg.Done()
	// 	}()

	// 	var blockSynchronizer block_synchronizer.BlockSynchronizer
	// 	blockSynchronizer, err = block_synchronizer.NewBlockSynchronizer(ctx, cfg, log)
	// 	if err != nil {
	// 		const msg = "failed to start Block Synchronizer: %+v"
	// 		log.Fatalf(msg, err.Error())
	// 	}

	// 	latestSynchronizedDepositIndex, err := blockValidityService.FetchLastDepositIndex()
	// 	if err != nil {
	// 		const msg = "failed to fetch last deposit index: %+v"
	// 		log.Fatalf(msg, err.Error())
	// 	}

	// 	timeout := 5 * time.Second
	// 	ticker := time.NewTicker(timeout)
	// 	for {
	// 		fmt.Printf("block validity ticker\n")
	// 		select {
	// 		case <-ctx.Done():
	// 			ticker.Stop()
	// 			return
	// 		case <-ticker.C:
	// 			fmt.Println("block validity ticker.C")
	// 			err = blockValidityProver.SyncDepositedEvents()
	// 			if err != nil {
	// 				const msg = "failed to sync deposited events: %+v"
	// 				log.Fatalf(msg, err.Error())
	// 			}

	// 			err = blockValidityProver.SyncDepositTree(nil, latestSynchronizedDepositIndex)
	// 			if err != nil {
	// 				const msg = "failed to sync deposit tree: %+v"
	// 				log.Fatalf(msg, err.Error())
	// 			}

	// 			// sync block content
	// 			startBlock, err := blockValidityService.LastSeenBlockPostedEventBlockNumber()
	// 			if err != nil {
	// 				startBlock = cfg.Blockchain.RollupContractDeployedBlockNumber
	// 				// var ErrLastSeenBlockPostedEventBlockNumberFail = errors.New("last seen block posted event block number fail")
	// 				// panic(errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err))
	// 			}

	// 			endBlock, err := blockValidityProver.SyncBlockTree(blockSynchronizer, startBlock)
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			err = blockValidityService.SetLastSeenBlockPostedEventBlockNumber(endBlock)
	// 			if err != nil {
	// 				var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
	// 				panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
	// 			}

	// 			fmt.Printf("Block %d is searched\n", endBlock)
	// 		}
	// 	}
	// }()

	userWalletState, err := balance_synchronizer.NewMockWallet(userAccount)
	if err != nil {
		const msg = "failed to get Mock Wallet: %+v"
		return fmt.Errorf(msg, err.Error())
	}

	fmt.Println("start SyncLocally")
	balanceSynchronizer, err := balance_synchronizer.SyncLocally(
		ctx,
		cfg,
		log,
		sb,
		blockValidityService,
		userWalletState,
	)
	fmt.Println("end SyncLocally")
	if err != nil {
		return fmt.Errorf("failed to sync balance proof: %w", err)
	}

	l2Balance := userWalletState.Balance(tokenIndex).BigInt()
	// balance, err := balance_service.GetUserBalance(ctx, cfg, userAccount, tokenIndex)
	// if err != nil {
	// 	return fmt.Errorf(ErrFailedToGetBalance.Error()+": %v", err)
	// }

	if strings.TrimSpace(amountStr) == "" {
		return fmt.Errorf("amount is required")
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", amountStr)
	}

	if l2Balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient funds for total amount: balance %s, total amount %s", l2Balance, amount)
	}

	var dataBlockInfo *BlockInfoResponseData
	dataBlockInfo, err = GetBlockInfo(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to get the block info data: %w", err)
	}

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	var transfers []*intMaxTypes.Transfer

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

		transfers = append(transfers, transfer)
	}

	const base10 = 10
	gasFeeInt, ok := new(big.Int).SetString(gasFee, base10)
	if !ok {
		return fmt.Errorf("failed to convert gas fee to int: %w", err)
	}
	totalAmountWithGas := new(big.Int).Add(amount, gasFeeInt)
	if l2Balance.Cmp(totalAmountWithGas) < 0 {
		return fmt.Errorf("insufficient funds for tx cost: balance %s, tx cost %s", l2Balance, totalAmountWithGas)
	}

	var transferTree *intMaxTree.TransferTree
	transferTree, err = intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, transfers, zeroTransfer.Hash())
	if err != nil {
		return fmt.Errorf("failed to create transfer tree: %v", err)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	// lastBalanceProof := ""
	// balanceTransitionProof := ""
	nonce := balanceSynchronizer.CurrentNonce()
	// nonce := uint32(1) // TODO: Get nonce from balance synchronizer

	err = SendTransferTransaction(
		ctx,
		cfg,
		log,
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
		ctx, cfg, log, userAccount, transfersHash, nonce,
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
		Transfers:     transfers,
		TxTreeRoot:    &proposedBlock.TxTreeRoot,
		TxMerkleProof: proposedBlock.TxTreeMerkleProof,
	}

	lastBalanceProofWithPis := balanceSynchronizer.LastBalanceProof()

	// txWitness, transferWitnesses, err := balance_synchronizer.MakeTxWitness(blockValidityService, &txDetails)
	// if err != nil {
	// 	return fmt.Errorf("failed to make tx witness: %w", err)
	// }
	newSalt, err := new(balance_prover_service.Salt).SetRandom()
	if err != nil {
		const msg = "failed to set random: %+v"
		return fmt.Errorf(msg, err.Error())
	}
	spentTokenWitness, err := userWalletState.CalculateSpentTokenWitness(
		*newSalt, tx, transfers,
	)
	if err != nil {
		return fmt.Errorf("failed to calculate spent witness: %w", err)
	}

	// prevBalancePisBlockNumber := sendWitness.GetPrevBalancePisBlockNumber()
	// currentBlockNumber := sendWitness.GetIncludedBlockNumber()
	// updateWitness, err := blockValidityService.FetchUpdateWitness(
	// 	userWalletState.PublicKey(),
	// 	&currentBlockNumber,
	// 	prevBalancePisBlockNumber,
	// 	true,
	// )
	// if err != nil {
	// 	return err
	// }

	balanceTransitionProof, err := balanceSynchronizer.ProveSendTransition(spentTokenWitness)
	if err != nil {
		return fmt.Errorf("failed to create balance transition proof: %w", err)
	}
	balanceTransitionProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(balanceTransitionProof.Proof)
	if err != nil {
		return fmt.Errorf("failed to create balance transition proof with pis: %w", err)
	}

	backupTx, err := transaction.NewBackupTransactionData(
		userAccount.Public(),
		txDetails,
		txHash,
		"0x",
	)
	if err != nil {
		return fmt.Errorf("failed to make backup transaction data: %w", err)
	}

	enoughBalanceProofBody := block_signature.EnoughBalanceProofBody{
		PrevBalanceProofBody:  lastBalanceProofWithPis.Proof,
		TransferStepProofBody: balanceTransitionProofWithPis.Proof,
	}
	enoughBalanceProof := new(block_signature.EnoughBalanceProofBodyInput).FromEnoughBalanceProofBody(&enoughBalanceProofBody)
	enoughBalanceProofHash := enoughBalanceProofBody.Hash()

	backupTransfers := make([]*transaction.BackupTransferInput, len(transfers))
	for i := range transfers {
		var transferMerkleProof []*intMaxTypes.PoseidonHashOut
		transferMerkleProof, _, err = transferTree.ComputeMerkleProof(uint64(i))
		if err != nil {
			return fmt.Errorf("failed to compute merkle proof: %v", err)
		}
		transferWitness := intMaxTypes.TransferWitness{
			Transfer:            *transfers[i],
			TransferIndex:       uint32(i),
			Tx:                  *tx,
			TransferMerkleProof: transferMerkleProof,
		}
		transferDetails := intMaxTypes.TransferDetails{
			TransferWitness:                     &transferWitness,
			TxTreeRoot:                          &proposedBlock.TxTreeRoot,
			TxMerkleProof:                       proposedBlock.TxTreeMerkleProof,
			SenderLastBalancePublicInputs:       lastBalanceProofWithPis.PublicInputsBytes(),
			SenderBalanceTransitionPublicInputs: balanceTransitionProofWithPis.PublicInputsBytes(),
			SenderEnoughBalanceProofBodyHash:    enoughBalanceProofHash,
		}
		backupTransfers[i], err = MakeTransferBackupData(
			&transferDetails,
		)
		if err != nil {
			return fmt.Errorf("failed to make backup transfer data: %v", err)
		}
	}

	// Accept proposed block
	err = SendSignedProposedBlock(
		ctx, cfg, log, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
		backupTx, backupTransfers, enoughBalanceProof,
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
) (backupTransfer *transaction.BackupTransferInput, _ error) {
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
		TransferHash:             transfer.Hash().String(),
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
