package tx_deposit_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gen/service/node"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_deposit"
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

type DepositRequestService struct {
	ctx       context.Context
	cfg       *configs.Config
	log       logger.Logger
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

func NewDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, sc ServiceBlockchain) (*DepositRequestService, error) {
	return newDepositAnalyzerService(ctx, cfg, log, sc)
}

func newDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, _ ServiceBlockchain) (*DepositRequestService, error) {
	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &DepositRequestService{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		client:    client,
		liquidity: liquidity,
	}, nil
}

func (d *DepositRequestService) DepositETHWithRandomSalt(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	tokenIndex uint32,
	amountStr string,
) error {
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %s", amountStr)
	}

	salt, err := new(goldenposeidon.PoseidonHashOut).SetRandom()
	if err != nil {
		return fmt.Errorf("failed to set random salt: %w", err)
	}

	receipt, err := d.depositEth(userEthPrivateKeyHex, recipient, amount, salt)
	if err != nil {
		return fmt.Errorf("failed to deposit ETH: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("failed to deposit ETH: receipt status is %d", receipt.Status)
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
		return fmt.Errorf("failed to get deposit ID")
	}

	err = d.backupDeposit(recipient, 0, amount, salt, depositID)
	if err != nil {
		return fmt.Errorf("failed to backup deposit: %w", err)
	}

	return nil
}

func (d *DepositRequestService) depositEth(
	userEthPrivateKeyHex string,
	recipient intMaxAcc.Address,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
) (*types.Receipt, error) {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient public key: %w", err)
	}
	recipientSaltHash := intMaxAcc.GetPublicKeySaltHash(recipientPublicKey.Pk.X.BigInt(new(big.Int)), salt)

	transactOpts, err := utils.CreateTransactor(userEthPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	transactOpts.Value = amount

	tx, err := d.liquidity.DepositETH(transactOpts, recipientSaltHash)
	if err != nil {
		return nil, fmt.Errorf("failed to send AnalyzeDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func (d *DepositRequestService) backupDeposit(
	recipient intMaxAcc.Address,
	tokenIndex uint32,
	amount *big.Int,
	salt *goldenposeidon.PoseidonHashOut,
	depositID uint64,
) error {
	recipientPublicKey, err := recipient.Public()
	if err != nil {
		return fmt.Errorf("failed to get recipient public key: %w", err)
	}

	deposit := intMaxTypes.Deposit{
		Recipient:  recipientPublicKey,
		TokenIndex: tokenIndex,
		Amount:     amount,
		Salt:       salt,
	}

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
		d.log,
		encodedEncryptedText,
		recipientPublicKey.ToAddress().String(),
		uint32(depositID),
	)

	if err != nil {
		return fmt.Errorf("failed to backup deposit: %w", err)
	}

	return nil
}

func backupDepositRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	encodedEncryptedText string,
	recipient string,
	depositID uint32,
) error {
	ucInput := backup_deposit.UCPostBackupDepositInput{
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

	apiUrl := fmt.Sprintf("%s/v1/backups/deposit", cfg.HTTP.DataStoreVaultUrl)

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
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf("Unexpected status code")
		return err
	}

	response := new(node.BackupDepositResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to send transaction: %s", response.Data.Message)
	}

	return nil
}
