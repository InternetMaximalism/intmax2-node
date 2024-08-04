//nolint:gocritic
package withdrawal_service

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	WithdrawalThreshold = 8
	WaitDuration        = 15 * time.Minute

	int10Key = 10
	int32Key = 32
)

type WithdrawalAggregatorService struct {
	ctx                context.Context
	cfg                *configs.Config
	log                logger.Logger
	db                 SQLDriverApp
	client             *ethclient.Client
	withdrawalContract *bindings.Withdrawal
}

func newWithdrawalAggregatorService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) (*WithdrawalAggregatorService, error) {
	link, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	}

	client, err := utils.NewClient(link)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	withdrawalContract, err := bindings.NewWithdrawal(common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ScrollMessenger contract: %w", err)
	}

	return &WithdrawalAggregatorService{
		ctx:                ctx,
		cfg:                cfg,
		log:                log,
		db:                 db,
		client:             client,
		withdrawalContract: withdrawalContract,
	}, nil
}

func WithdrawalAggregator(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	service, err := newWithdrawalAggregatorService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize WithdrawalAggregatorService: %v", err.Error()))
	}

	pendingWithdrawals, err := service.fetchPendingWithdrawals()
	if err != nil {
		panic(fmt.Sprintf("Failed to retrieve withdrawals %v", err.Error()))
	}

	if len(*pendingWithdrawals) == 0 {
		log.Infof("No pending withdrawal requests found")
		return
	}

	shouldSubmit := service.shouldProcessWithdrawals(*pendingWithdrawals)
	if !shouldSubmit {
		log.Infof("Not enough pending withdrawal requests to process")
		return
	}

	// proofs, err := service.fetchWithdrawalProofsFromProver(*pendingWithdrawals)
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to fetch withdrawal proofs %v", err.Error()))
	// }

	// withdrawalInfo, err := service.buildSubmitWithdrawalProofData(*pendingWithdrawals, proofs)
	// if err != nil {
	// 	panic("NEED_TO_BE_IMPLEMENTED")
	// }

	withdrawalInfo, err := service.buildMockSubmitWithdrawalProofData(*pendingWithdrawals)
	if err != nil {
		panic(fmt.Sprintf("Failed to build withdrawal proof data: %v", err.Error()))
	}

	receipt, err := service.submitWithdrawalProof(
		withdrawalInfo.Withdrawals,
		withdrawalInfo.WithdrawalProofPublicInputs,
		withdrawalInfo.Proof,
	)
	if err != nil {
		// TODO: NEED_TO_BE_CHANGED change status depends on the result of the proof
		if err.Error() == "WithdrawalProofVerificationFailed" {
			log.Errorf("Failed to submit withdrawal proof: %v", err.Error())
			err = db.UpdateWithdrawalsStatus(extractIds(*pendingWithdrawals), mDBApp.WS_FAILED)
			if err != nil {
				panic(fmt.Sprintf("Failed to update withdrawal status: %v", err.Error()))
			}
			return
		}
		panic(fmt.Sprintf("Failed to submit withdrawal proof: %v", err.Error()))
	}

	if receipt == nil {
		panic("Received nil receipt for transaction")
	}

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		log.Infof("Successfully submit withdrawal proof %d withdrawals. Transaction Hash: %v", len(*pendingWithdrawals), receipt.TxHash.Hex())
	case types.ReceiptStatusFailed:
		panic(fmt.Sprintf("Transaction failed: submit withdrawal proof unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
	default:
		panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
	}

	err = db.UpdateWithdrawalsStatus(extractIds(*pendingWithdrawals), mDBApp.WS_SUCCESS)
	if err != nil {
		panic(fmt.Sprintf("Failed to update withdrawal status: %v", err.Error()))
	}
}

func (w *WithdrawalAggregatorService) fetchPendingWithdrawals() (*[]mDBApp.Withdrawal, error) {
	limit := int(WithdrawalThreshold)
	withdrawals, err := w.db.WithdrawalsByStatus(mDBApp.WS_PENDING, &limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending withdrawals: %w", err)
	}
	if withdrawals == nil {
		return nil, fmt.Errorf("failed to get pending withdrawals because withdrawals is nil")
	}
	return withdrawals, nil
}

func (w *WithdrawalAggregatorService) shouldProcessWithdrawals(pendingWithdrawals []mDBApp.Withdrawal) bool {
	minCreatedAt := pendingWithdrawals[0].CreatedAt
	for _, withdrawal := range pendingWithdrawals[1:] {
		if withdrawal.CreatedAt.Before(minCreatedAt) {
			minCreatedAt = withdrawal.CreatedAt
		}
	}

	if time.Since(minCreatedAt) >= WaitDuration {
		w.log.Infof("Pending withdrawals are older than %s, processing", WaitDuration)
		return true
	}

	withdrawalCount := len(pendingWithdrawals)
	log.Printf("Number of pending withdrawals: %d (Threshold: %d)", withdrawalCount, WithdrawalThreshold)

	return withdrawalCount >= WithdrawalThreshold
}

func HashWithdrawalPis(pis bindings.WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs) (common.Hash, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, pis.LastWithdrawalHash); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, pis.WithdrawalAggregator); err != nil {
		return common.Hash{}, err
	}
	packed := buf.Bytes()

	hash := crypto.Keccak256Hash(packed)

	return hash, nil
}

func HashWithdrawalWithPrevHash(
	prevHash common.Hash,
	withdrawal bindings.ChainedWithdrawalLibChainedWithdrawal,
) (common.Hash, error) {
	amount := intMaxTypes.BigIntToBytes32BeArray(withdrawal.Amount)

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, prevHash[:]); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, withdrawal.Recipient); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, withdrawal.TokenIndex); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, amount); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, withdrawal.Nullifier); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, withdrawal.BlockHash); err != nil {
		return common.Hash{}, err
	}
	if err := binary.Write(&buf, binary.BigEndian, withdrawal.BlockNumber); err != nil {
		return common.Hash{}, err
	}
	packed := buf.Bytes()

	hash := crypto.Keccak256Hash(packed)

	return hash, nil
}

type WithdrawalInfo struct {
	Withdrawals                 []bindings.ChainedWithdrawalLibChainedWithdrawal
	WithdrawalProofPublicInputs bindings.WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs
	PisHash                     common.Hash
	Proof                       []byte
}

func MakeWithdrawalInfo(
	aggregator common.Address,
	withdrawals []bindings.ChainedWithdrawalLibChainedWithdrawal,
	proof []byte,
) (*WithdrawalInfo, error) {
	hash := common.Hash{}
	var err error
	for _, withdrawal := range withdrawals {
		hash, err = HashWithdrawalWithPrevHash(hash, withdrawal)
		if err != nil {
			var ErrHashWithdrawalWithPrevHash = errors.New("failed to hash withdrawal with previous hash")
			return nil, errors.Join(ErrHashWithdrawalWithPrevHash, err)
		}
	}
	pis := bindings.WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs{
		LastWithdrawalHash:   hash,
		WithdrawalAggregator: aggregator,
	}
	pisHash, err := HashWithdrawalPis(pis)
	if err != nil {
		var ErrHashWithdrawalPis = errors.New("failed to hash withdrawal proof public inputs")
		return nil, errors.Join(ErrHashWithdrawalPis, err)
	}

	return &WithdrawalInfo{
		Withdrawals:                 withdrawals,
		WithdrawalProofPublicInputs: pis,
		PisHash:                     pisHash,
		Proof:                       proof,
	}, nil
}

// func (w *WithdrawalAggregatorService) fetchWithdrawalProofsFromProver(pendingWithdrawals []mDBApp.Withdrawal) ([]ProofValue, error) {
// 	var idsQuery string
// 	for _, pendingWithdrawal := range pendingWithdrawals {
// 		idsQuery += fmt.Sprintf("ids=%s&", pendingWithdrawal.ID)
// 	}
// 	if len(idsQuery) > 0 {
// 		idsQuery = idsQuery[:len(idsQuery)-1]
// 	}
// 	apiUrl := fmt.Sprintf("%s/proofs?%s",
// 		w.cfg.API.WithdrawalProverUrl,
// 		idsQuery,
// 	)

// 	resp, err := http.Get(apiUrl) // nolint:gosec
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to request API: %w", err)
// 	}
// 	defer func() {
// 		_ = resp.Body.Close()
// 	}()

// 	var res ProofsResponse
// 	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
// 		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
// 	}

// 	if !res.Success {
// 		return nil, fmt.Errorf("prover request failed %s", res.ErrorMessage)
// 	}

// 	return res.Values, nil
// }

// func (w *WithdrawalAggregatorService) buildSubmitWithdrawalProofData(pendingWithdrawals []mDBApp.Withdrawal, proofs []ProofValue) error {
// 	return nil
// }

func (w *WithdrawalAggregatorService) buildMockSubmitWithdrawalProofData(pendingWithdrawals []mDBApp.Withdrawal) (*WithdrawalInfo, error) {
	// private key to address
	decKey, err := hex.DecodeString(w.cfg.Blockchain.WithdrawalPrivateKeyHex)
	if err != nil {
		return nil, err
	}

	pk, err := crypto.ToECDSA(decKey)
	if err != nil {
		return nil, err
	}

	aggregator := crypto.PubkeyToAddress(pk.PublicKey)
	fmt.Printf("Aggregator address: %s\n", aggregator.String())

	withdrawals := make([]bindings.ChainedWithdrawalLibChainedWithdrawal, 0, len(pendingWithdrawals))
	for _, withdrawal := range pendingWithdrawals {
		amount, ok := new(big.Int).SetString(withdrawal.TransferData.Amount, int10Key)
		if !ok {
			return nil, fmt.Errorf("failed to set amount")
		}

		recipientBytes, err := hexutil.Decode(withdrawal.TransferData.Recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to decode recipient address: %w", err)
		}
		recipient, err := intMaxTypes.NewEthereumAddress(recipientBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address: %w", err)
		}

		saltBytes, err := hexutil.Decode(withdrawal.TransferData.Salt)
		if err != nil {
			return nil, fmt.Errorf("failed to decode salt: %w", err)
		}
		salt := new(goldenposeidon.PoseidonHashOut)
		err = salt.Unmarshal(saltBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal salt: %w", err)
		}

		withdrawalTransfer := intMaxTypes.Transfer{
			Recipient:  recipient,
			TokenIndex: uint32(withdrawal.TransferData.TokenIndex),
			Amount:     amount,
			Salt:       salt,
		}

		nullifier := [32]byte{}
		copy(nullifier[:], withdrawalTransfer.GetWithdrawalNullifier().Marshal())
		singleWithdrawal := bindings.ChainedWithdrawalLibChainedWithdrawal{
			Recipient:   common.HexToAddress(withdrawal.TransferData.Recipient),
			TokenIndex:  uint32(withdrawal.TransferData.TokenIndex),
			Amount:      amount,
			Nullifier:   nullifier,
			BlockHash:   common.HexToHash(withdrawal.BlockHash),
			BlockNumber: uint32(withdrawal.BlockNumber),
		}
		withdrawals = append(withdrawals, singleWithdrawal)
	}

	withdrawalInfo, err := MakeWithdrawalInfo(aggregator, withdrawals, []byte{})
	if err != nil {
		return nil, err
	}

	return withdrawalInfo, nil
}

func (w *WithdrawalAggregatorService) submitWithdrawalProof(
	withdrawals []bindings.ChainedWithdrawalLibChainedWithdrawal,
	publicInputs bindings.WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs,
	proof []byte,
) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(w.cfg.Blockchain.WithdrawalPrivateKeyHex, w.cfg.Blockchain.ScrollNetworkChainID)
	if err != nil {
		return nil, err
	}

	tx, err := w.withdrawalContract.SubmitWithdrawalProof(transactOpts, withdrawals, publicInputs, proof)
	if err != nil {
		return nil, fmt.Errorf("failed to send submit withdrawal proof transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func extractIds(withdrawas []mDBApp.Withdrawal) []string {
	ids := make([]string, len(withdrawas))
	for i, w := range withdrawas {
		ids[i] = w.ID
	}
	return ids
}
