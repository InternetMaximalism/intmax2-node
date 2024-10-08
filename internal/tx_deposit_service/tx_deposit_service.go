package tx_deposit_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/mnemonic_wallet"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	intMaxTypes "intmax2-node/internal/types"
	postBackupDeposit "intmax2-node/internal/use_cases/post_backup_deposit"
	"intmax2-node/pkg/utils"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
)

const int10Key = 10

const DepositEventSignatureID = "0x1e88950eef3c1bd8dd83d765aec1f21f34ca153104f0acd7a6218bf8f48e8410"

var ErrRecipientPublicKey = errors.New("failed to get recipient public key")
var ErrCreateTransactor = errors.New("failed to create transactor")
var ErrWaitForTransaction = errors.New("failed to wait for transaction to be mined")
var ErrInsufficientERC20Balance = errors.New("the specified ERC20 is not owned by the account or insufficient balance")
var ErrInsufficientERC721Balance = errors.New("the specified ERC721 is not owned by the account")
var ErrFailedToApproveERC20Transaction = errors.New("failed to send ERC20.Approve transaction")
var ErrFailedToApproveERC721Transaction = errors.New("failed to send ERC721.Approve transaction")
var ErrFailedToDecodeFromBase64 = errors.New("failed to decode from base64")
var ErrFailedToDecrypt = errors.New("failed to decrypt")
var ErrFailedToUnmarshal = errors.New("failed to unmarshal")

type TxDepositService struct {
	ctx       context.Context
	cfg       *configs.Config
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

func NewTxDepositService(ctx context.Context, cfg *configs.Config, _ ServiceBlockchain) (*TxDepositService, error) {
	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &TxDepositService{
		ctx:       ctx,
		cfg:       cfg,
		client:    client,
		liquidity: liquidity,
	}, nil
}

func (d *TxDepositService) DepositETHWithRandomSalt(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	amountStr string,
) (uint64, *goldenposeidon.PoseidonHashOut, error) {
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return 0, nil, fmt.Errorf("failed to convert amount to int: %s", amountStr)
	}

	salt, err := new(goldenposeidon.PoseidonHashOut).SetRandom()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to set random salt: %w", err)
	}

	receipt, err := d.depositEth(userEthPrivateKeyHex, recipient, amount, salt)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to deposit ETH: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return 0, nil, fmt.Errorf("failed to deposit ETH: receipt status is %d", receipt.Status)
	}

	var depositID uint64
	ok = false
	for i := 0; i < len(receipt.Logs); i++ {
		if receipt.Logs[i].Topics[0].Hex() == DepositEventSignatureID {
			depositID = receipt.Logs[i].Topics[1].Big().Uint64()
			ok = true
			break
		}
	}
	if !ok {
		return 0, nil, fmt.Errorf("failed to get deposit ID")
	}

	return depositID, salt, nil
}

func (d *TxDepositService) DepositERC20WithRandomSalt(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	tokenAddress common.Address,
	amountStr string,
) (uint64, *goldenposeidon.PoseidonHashOut, error) {
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return 0, nil, fmt.Errorf("failed to convert amount to int: %s", amountStr)
	}

	salt, err := new(goldenposeidon.PoseidonHashOut).SetRandom()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to set random salt: %w", err)
	}

	receipt, err := d.depositErc20(userEthPrivateKeyHex, tokenAddress, recipient, amount, salt)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to deposit ETH: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return 0, nil, fmt.Errorf("failed to deposit ETH: receipt status is %d", receipt.Status)
	}

	var depositID uint64
	ok = false
	for i := 0; i < len(receipt.Logs); i++ {
		if receipt.Logs[i].Topics[0].Hex() == DepositEventSignatureID {
			depositID = receipt.Logs[i].Topics[1].Big().Uint64()
			ok = true
			break
		}
	}
	if !ok {
		return 0, nil, fmt.Errorf("failed to get deposit ID")
	}

	return depositID, salt, nil
}

func (d *TxDepositService) DepositERC721WithRandomSalt(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	tokenAddress common.Address,
	tokenId *big.Int,
) (uint64, *goldenposeidon.PoseidonHashOut, error) {
	salt, err := new(goldenposeidon.PoseidonHashOut).SetRandom()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to set random salt: %w", err)
	}

	receipt, err := d.depositErc721(userEthPrivateKeyHex, tokenAddress, recipient, tokenId, salt)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to deposit ETH: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return 0, nil, fmt.Errorf("failed to deposit ETH: receipt status is %d", receipt.Status)
	}

	var depositID uint64
	ok := false
	for i := 0; i < len(receipt.Logs); i++ {
		if receipt.Logs[i].Topics[0].Hex() == DepositEventSignatureID {
			depositID = receipt.Logs[i].Topics[1].Big().Uint64()
			ok = true
			break
		}
	}
	if !ok {
		return 0, nil, fmt.Errorf("failed to get deposit ID")
	}

	return depositID, salt, nil
}

func (d *TxDepositService) DepositERC1155WithRandomSalt(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	tokenAddress common.Address,
	tokenId *big.Int,
	amountStr string,
) (uint64, *goldenposeidon.PoseidonHashOut, error) {
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return 0, nil, fmt.Errorf("failed to convert amount to int: %s", amountStr)
	}

	salt, err := new(goldenposeidon.PoseidonHashOut).SetRandom()
	if err != nil {
		var ErrSetRandomSalt = errors.New("failed to set random salt")
		return 0, nil, errors.Join(ErrSetRandomSalt, err)
	}

	receipt, err := d.depositErc1155(userEthPrivateKeyHex, tokenAddress, recipient, tokenId, amount, salt)
	if err != nil {
		var ErrDepositErc1155 = errors.New("failed to deposit ERC1155")
		return 0, nil, errors.Join(ErrDepositErc1155, err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		var ErrDepositErc1155ReceiptStatus = errors.New("failed to deposit ERC1155")
		return 0, nil, errors.Join(ErrDepositErc1155ReceiptStatus, fmt.Errorf("receipt status is %d", receipt.Status))
	}

	var depositID uint64
	ok = false
	for i := 0; i < len(receipt.Logs); i++ {
		if receipt.Logs[i].Topics[0].Hex() == DepositEventSignatureID {
			depositID = receipt.Logs[i].Topics[1].Big().Uint64()
			ok = true
			break
		}
	}
	if !ok {
		var ErrGetDepositID = errors.New("failed to get deposit ID")
		return 0, nil, ErrGetDepositID
	}

	return depositID, salt, nil
}

func (d *TxDepositService) depositEth(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
) (*types.Receipt, error) {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return nil, errors.Join(ErrRecipientPublicKey, err)
	}
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	transactOpts, err := utils.CreateTransactor(userEthPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	transactOpts.Value = amount

	fmt.Printf("Deposit ETH to %s. Processing...\n", recipient.String())
	tx, err := d.liquidity.DepositNativeToken(transactOpts, recipientSaltHash)
	if err != nil {
		return nil, fmt.Errorf("failed to send DepositETH transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	return receipt, nil
}

func (d *TxDepositService) depositErc20(
	userEthPrivateKeyHex string,
	tokenAddress common.Address,
	recipient intMaxAcc.Address,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
) (*types.Receipt, error) {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return nil, errors.Join(ErrRecipientPublicKey, err)
	}
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	transactOpts, err := utils.CreateTransactor(userEthPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	client, err := utils.NewClient(d.cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	erc20, err := bindings.NewErc20(tokenAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a ERC20 contract: %w", err)
	}

	balance, err := erc20.BalanceOf(&bind.CallOpts{
		Pending: false,
		Context: d.ctx,
	}, transactOpts.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get ERC20 token balance: %w", err)
	}
	if balance.Cmp(amount) < 0 {
		return nil, ErrInsufficientERC20Balance
	}

	fmt.Println("Approve ERC20 token. Processing...")
	tx, err := erc20.Approve(transactOpts, common.HexToAddress(d.cfg.Blockchain.LiquidityContractAddress), amount)
	if err != nil {
		return nil, errors.Join(ErrFailedToApproveERC20Transaction, err)
	}

	_, err = bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	fmt.Printf("Deposit ERC20 token to %s. Processing...\n", recipient.String())
	tx2, err := d.liquidity.DepositERC20(transactOpts, tokenAddress, recipientSaltHash, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to send DepositERC20 transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx2)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	return receipt, nil
}

func (d *TxDepositService) depositErc721(
	userEthPrivateKeyHex string,
	tokenAddress common.Address,
	recipient intMaxAcc.Address,
	tokenId *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
) (*types.Receipt, error) {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return nil, errors.Join(ErrRecipientPublicKey, err)
	}
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	transactOpts, err := utils.CreateTransactor(userEthPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	client, err := utils.NewClient(d.cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	erc721, err := bindings.NewErc721(tokenAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a ERC721 contract: %w", err)
	}

	owner, err := erc721.OwnerOf(&bind.CallOpts{
		Pending: false,
		Context: d.ctx,
	}, tokenId)
	if err != nil {
		return nil, fmt.Errorf("failed to get ERC721 token balance: %w", err)
	}
	if owner != transactOpts.From {
		return nil, ErrInsufficientERC721Balance
	}

	fmt.Println("Approve ERC721 token. Processing...")
	tx, err := erc721.Approve(transactOpts, common.HexToAddress(d.cfg.Blockchain.LiquidityContractAddress), tokenId)
	if err != nil {
		return nil, errors.Join(ErrFailedToApproveERC721Transaction, err)
	}

	_, err = bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	fmt.Printf("Deposit ERC721 token to %s. Processing...\n", recipient.String())
	tx2, err := d.liquidity.DepositERC721(transactOpts, tokenAddress, recipientSaltHash, tokenId)
	if err != nil {
		return nil, fmt.Errorf("failed to send DepositERC721 transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx2)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	return receipt, nil
}

func (d *TxDepositService) depositErc1155(
	userEthPrivateKeyHex string,
	tokenAddress common.Address,
	recipient intMaxAcc.Address,
	tokenId *big.Int,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
) (*types.Receipt, error) {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return nil, errors.Join(ErrRecipientPublicKey, err)
	}
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	transactOpts, err := utils.CreateTransactor(userEthPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, errors.Join(ErrCreateTransactor, err)
	}

	// TODO: Implement ERC1155.Approve

	tx, err := d.liquidity.DepositERC1155(transactOpts, tokenAddress, recipientSaltHash, tokenId, amount)
	if err != nil {
		var ErrDepositERC1155Transaction = errors.New("failed to send DepositERC1155 transaction")
		return nil, errors.Join(ErrDepositERC1155Transaction, err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, errors.Join(ErrWaitForTransaction, err)
	}

	return receipt, nil
}

func (d *TxDepositService) BackupDeposit(
	recipient intMaxAcc.Address,
	tokenIndex uint32,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
	depositID uint64,
) error {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return errors.Join(ErrRecipientPublicKey, err)
	}

	deposit := intMaxTypes.Deposit{
		Recipient:  recipientPublicKey,
		TokenIndex: tokenIndex,
		Amount:     amount,
		Salt:       salt,
	}

	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	depositHash := new(intMaxTypes.DepositLeaf).Set(&intMaxTypes.DepositLeaf{
		RecipientSaltHash: recipientSaltHash,
		TokenIndex:        tokenIndex,
		Amount:            amount,
	}).Hash()

	encodedDeposit := deposit.Marshal()
	encryptedDeposit, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		recipientPublicKey,
		encodedDeposit,
	)
	if err != nil {
		return fmt.Errorf("failed to encrypt deposit: %w", err)
	}

	encodedEncryptedText := base64.StdEncoding.EncodeToString(encryptedDeposit)

	err = backupDepositRawRequest(
		d.ctx,
		d.cfg,
		depositHash.String(),
		encodedEncryptedText,
		recipientPublicKey.ToAddress().String(),
		int64(depositID),
	)

	if err != nil {
		return fmt.Errorf("failed to backup deposit: %w", err)
	}

	return nil
}

func backupDepositRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	depositHash string,
	encodedEncryptedText string,
	recipient string,
	depositID int64,
) error {
	ucInput := postBackupDeposit.UCPostBackupDepositInput{
		DepositHash:      depositHash,
		EncryptedDeposit: encodedEncryptedText,
		Recipient:        recipient,
		BlockNumber:      depositID,
	}

	bd, err := json.Marshal(ucInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/deposit", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send of the transaction request: %w"
		return fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "send request error occurred"
		return fmt.Errorf(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to get response")
	}

	response := new(node.BackupDepositResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		if response.Data != nil {
			return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
		}

		return fmt.Errorf("failed to send transaction with unknown error")
	}

	return nil
}

func DepositsListIncoming(
	ctx context.Context,
	cfg *configs.Config,
	input *GetDepositsListInput,
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

	return GetDepositsListWithRawRequest(ctx, cfg, input, userAccount)
}

func GetDepositFromBackupData(
	encryptedDeposit *GetDepositData,
	senderAccount *intMaxAcc.PrivateKey,
) (*intMaxTypes.Deposit, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedDeposit.EncryptedDeposit)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecodeFromBase64, err)
	}

	var message []byte
	message, err = senderAccount.DecryptECIES(ciphertext)
	if err != nil {
		return nil, errors.Join(ErrFailedToDecrypt, err)
	}

	var deposit intMaxTypes.Deposit
	err = deposit.Unmarshal(message)
	if err != nil {
		return nil, errors.Join(ErrFailedToUnmarshal, err)
	}

	return &deposit, nil
}

func DepositByHashIncoming(
	ctx context.Context,
	cfg *configs.Config,
	depositHash string,
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

	return GetDepositByHashIncomingWithRawRequest(ctx, cfg, depositHash, userAccount)
}
