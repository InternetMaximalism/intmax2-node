//nolint:gocritic
package withdrawal_service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
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
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	base10   = 10
	int32Key = 32
)

type WithdrawalAggregatorService struct {
	ctx                context.Context
	cfg                *configs.Config
	log                logger.Logger
	db                 SQLDriverApp
	scrollClient       *ethclient.Client
	withdrawalContract *bindings.Withdrawal
}

func NewWithdrawalAggregatorService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) (*WithdrawalAggregatorService, error) {
	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Scroll network chain link: %w", err)
	}

	scrollClient, err := utils.NewClient(scrollLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create new scrollClient: %w", err)
	}

	withdrawalContract, err := bindings.NewWithdrawal(common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ScrollMessenger contract: %w", err)
	}

	return &WithdrawalAggregatorService{
		ctx:                ctx,
		cfg:                cfg,
		log:                log,
		db:                 db,
		scrollClient:       scrollClient,
		withdrawalContract: withdrawalContract,
	}, nil
}

func WithdrawalAggregator(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	service, err := NewWithdrawalAggregatorService(ctx, cfg, log, db, sb)
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

	fmt.Println("~~~4")
	withdrawalInfo, err := service.buildMockSubmitWithdrawalProofData(*pendingWithdrawals)
	if err != nil {
		panic(fmt.Sprintf("Failed to build withdrawal proof data: %v", err.Error()))
	}

	fmt.Println("~~~5")
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
	limit := int(w.cfg.Blockchain.WithdrawalAggregatorThreshold)
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

	waitDuration := time.Duration(w.cfg.Blockchain.WithdrawalAggregatorMinutesThreshold) * time.Minute
	if time.Since(minCreatedAt) >= waitDuration {
		w.log.Infof("Pending withdrawals are older than %s, processing", waitDuration)
		return true
	}

	withdrawalCount := len(pendingWithdrawals)
	log.Printf("Number of pending withdrawals: %d (Threshold: %d)", withdrawalCount, w.cfg.Blockchain.WithdrawalAggregatorThreshold)

	return withdrawalCount >= int(w.cfg.Blockchain.WithdrawalAggregatorThreshold)
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

type MerkleProof struct {
	Siblings []*goldenposeidon.PoseidonHashOut `json:"siblings"`
}

type TransferWitness struct {
	Tx                  intMaxTypes.Tx       `json:"tx"`
	Transfer            intMaxTypes.Transfer `json:"transfer"`
	TransferIndex       uint                 `json:"transferIndex"`
	TransferMerkleProof MerkleProof          `json:"transferMerkleProof"`
}

type WithdrawalWitness struct {
	TransferWitness TransferWitness          `json:"transferWitness"`
	BalanceProof    intMaxTypes.Plonky2Proof `json:"balanceProof"`
}

func (w *WithdrawalAggregatorService) buildSubmitWithdrawalProofData(pendingWithdrawals []mDBApp.Withdrawal, withdrawalAggregator common.Address) ([]byte, error) {
	prevWithdrawalProof := new(string)

	for i := range pendingWithdrawals {
		transferTreeRoot := new(goldenposeidon.PoseidonHashOut)
		transferTreeRoot.FromString(pendingWithdrawals[i].Transaction.TransferTreeRoot)

		salt := new(goldenposeidon.PoseidonHashOut)
		salt.FromString(pendingWithdrawals[i].TransferData.Salt)

		transferMerkleSiblings := make([]*goldenposeidon.PoseidonHashOut, 0, len(pendingWithdrawals[i].TransferMerkleProof.Siblings))
		for _, sibling := range pendingWithdrawals[i].TransferMerkleProof.Siblings {
			s := new(goldenposeidon.PoseidonHashOut)
			err := s.FromString(sibling)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal transfer merkle sibling: %w", err)
			}
			transferMerkleSiblings = append(transferMerkleSiblings, s)
		}

		recipientBytes, err := hexutil.Decode(pendingWithdrawals[i].TransferData.Recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to decode recipient address: %w", err)
		}
		recipient, err := intMaxTypes.NewEthereumAddress(recipientBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address: %w", err)
		}

		amount, ok := new(big.Int).SetString(pendingWithdrawals[i].TransferData.Amount, base10)
		if !ok {
			return nil, fmt.Errorf("failed to set amount")
		}

		balanceProof, err := base64.StdEncoding.DecodeString(pendingWithdrawals[i].EnoughBalanceProof.Proof)
		if err != nil {
			return nil, fmt.Errorf("failed to decode balance proof: %w", err)
		}

		decodedBalancePublicInputs, err := base64.StdEncoding.DecodeString(pendingWithdrawals[i].EnoughBalanceProof.PublicInputs)
		if err != nil {
			return nil, fmt.Errorf("failed to decode balance public inputs: %w", err)
		}

		const numGoldilocksFieldBytes = 8
		if len(decodedBalancePublicInputs)%numGoldilocksFieldBytes != 0 {
			return nil, fmt.Errorf("balance public inputs length is not multiple of %d", numGoldilocksFieldBytes)
		}

		balancePublicInputs := make([]ffg.Element, 0, len(decodedBalancePublicInputs))
		for i := 0; i < len(balancePublicInputs); i += numGoldilocksFieldBytes {
			e := binary.BigEndian.Uint64(decodedBalancePublicInputs[i : i+numGoldilocksFieldBytes])
			balancePublicInputs = append(balancePublicInputs, *new(ffg.Element).SetUint64(e))
		}

		withdrawalWitness := WithdrawalWitness{
			TransferWitness: TransferWitness{
				Tx: intMaxTypes.Tx{
					TransferTreeRoot: transferTreeRoot,
					Nonce:            uint64(pendingWithdrawals[i].Transaction.Nonce),
				},
				Transfer: intMaxTypes.Transfer{
					Recipient:  recipient,
					TokenIndex: uint32(pendingWithdrawals[i].TransferData.TokenIndex),
					Amount:     amount,
					Salt:       salt,
				},
				TransferIndex: uint(pendingWithdrawals[i].TransferMerkleProof.Index),
				TransferMerkleProof: MerkleProof{
					Siblings: transferMerkleSiblings,
				},
			},
			BalanceProof: intMaxTypes.Plonky2Proof{
				Proof:        balanceProof,
				PublicInputs: balancePublicInputs,
			},
		}

		withdrawalProof, err := w.RequestWithdrawalProofToProver(&withdrawalWitness, prevWithdrawalProof)
		if err != nil {
			return nil, fmt.Errorf("failed to request withdrawal proof to prover: %w", err)
		}

		prevWithdrawalProof = withdrawalProof
	}

	withdrawalWrapperProof, err := w.RequestWithdrawalWrapperProofToProver(*prevWithdrawalProof, withdrawalAggregator)
	if err != nil {
		return nil, fmt.Errorf("failed to request withdrawal wrapper proof to prover: %w", err)
	}

	return withdrawalWrapperProof, nil
}

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
		amount, ok := new(big.Int).SetString(withdrawal.TransferData.Amount, base10)
		if !ok {
			return nil, fmt.Errorf("failed to set amount")
		}

		var recipientBytes []byte
		recipientBytes, err = hexutil.Decode(withdrawal.TransferData.Recipient)
		if err != nil {
			return nil, fmt.Errorf("failed to decode recipient address: %w", err)
		}
		var recipient *intMaxTypes.GenericAddress
		recipient, err = intMaxTypes.NewEthereumAddress(recipientBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recipient address: %w", err)
		}

		var saltBytes []byte
		saltBytes, err = hexutil.Decode(withdrawal.TransferData.Salt)
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

		nullifier := [int32Key]byte{}
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

	withdrawalWrapperProof, err := w.buildSubmitWithdrawalProofData(pendingWithdrawals, aggregator)
	if err != nil {
		return nil, err
	}

	withdrawalInfo, err := MakeWithdrawalInfo(aggregator, withdrawals, withdrawalWrapperProof)
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

	err = utils.LogTransactionDebugInfo(
		w.log,
		w.cfg.Blockchain.WithdrawalPrivateKeyHex,
		w.cfg.Blockchain.WithdrawalContractAddress,
		withdrawals,
		publicInputs,
		proof,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log transaction debug info: %w", err)
	}

	tx, err := w.withdrawalContract.SubmitWithdrawalProof(transactOpts, withdrawals, publicInputs, proof)
	if err != nil {
		return nil, fmt.Errorf("failed to send submit withdrawal proof transaction: %w", err)
	}

	receipt, err := bind.WaitMined(w.ctx, w.scrollClient, tx)
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

func (w *WithdrawalAggregatorService) RequestWithdrawalProofToProver(witness *WithdrawalWitness, prevProof *string) (*string, error) {
	requestID := uuid.New().String()
	resGeneration, err := w.requestWithdrawalProofToProver(requestID, witness, prevProof)
	if err != nil {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-w.ctx.Done():
				return nil, err
			case <-ticker.C:
				resFetching, errFetching := w.fetchWithdrawalProofToProver(requestID)
				if errFetching != nil {
					var ErrWrappedWithdrawalProofFetching = errors.New("failed to fetch wrapper withdrawal proof")
					err = errors.Join(ErrWrappedWithdrawalProofFetching, errFetching)

					time.Sleep(10 * time.Second)

					continue
				}

				return resFetching, nil
			}
		}
	}

	return resGeneration, nil
}

func (w *WithdrawalAggregatorService) requestWithdrawalProofToProver(requestID string, witness *WithdrawalWitness, prevProof *string) (*string, error) {
	requestBody := map[string]interface{}{
		"id":                  requestID,
		"prevWithdrawalProof": prevProof,
		"withdrawalWitness":   witness,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	apiUrl := fmt.Sprintf("%s/proof",
		w.cfg.API.WithdrawalProverUrl,
	)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody)) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res ProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return res.Proof, nil
}

func (w *WithdrawalAggregatorService) fetchWithdrawalProofToProver(requestID string) (*string, error) {
	requestBody := map[string]interface{}{
		"id": requestID,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	apiUrl := fmt.Sprintf("%s/proof/%s",
		w.cfg.API.WithdrawalProverUrl,
		requestID,
	)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody)) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res ProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		if res.ErrorMessage != nil {
			return nil, fmt.Errorf("prover request failed %s", *res.ErrorMessage)
		}

		return nil, fmt.Errorf("prover request failed")
	}

	return res.Proof, nil
}

func (w *WithdrawalAggregatorService) RequestWithdrawalWrapperProofToProver(withdrawalProof string, withdrawalAggregator common.Address) (wrappedProof []byte, err error) {
	requestID := uuid.New().String()
	resGeneration, err := w.requestWithdrawalWrapperProofToProver(requestID, withdrawalProof, withdrawalAggregator)
	if err != nil {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-w.ctx.Done():
				return nil, err
			case <-ticker.C:
				resFetching, errFetching := w.fetchWithdrawalWrapperProofToProver(requestID)
				if errFetching != nil {
					var ErrWrappedWithdrawalProofFetching = errors.New("failed to fetch wrapper withdrawal proof")
					err = errors.Join(ErrWrappedWithdrawalProofFetching, errFetching)
					continue
				}

				return []byte(*resFetching), nil
			}
		}
	}

	return []byte(*resGeneration), nil
}

func (w *WithdrawalAggregatorService) requestWithdrawalWrapperProofToProver(requestID string, withdrawalProof string, withdrawalAggregator common.Address) (wrappedProof *string, err error) {
	requestBody := map[string]interface{}{
		"id":                   requestID,
		"withdrawalProof":      withdrawalProof,
		"withdrawalAggregator": withdrawalAggregator.Hex(), // with 0x prefix
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	apiUrl := fmt.Sprintf("%s/proof/wrapper",
		w.cfg.API.WithdrawalProverUrl,
	)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonBody)) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res ProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		if res.ErrorMessage != nil {
			return nil, fmt.Errorf("prover request failed %s", *res.ErrorMessage)
		}

		return nil, fmt.Errorf("prover request failed")
	}

	return res.Proof, nil
}

func (w *WithdrawalAggregatorService) fetchWithdrawalWrapperProofToProver(requestID string) (wrappedProof *string, err error) {
	apiUrl := fmt.Sprintf("%s/proof/wrapper?id=%s",
		w.cfg.API.WithdrawalProverUrl,
		requestID,
	)
	resp, err := http.Get(apiUrl) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res ProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if !res.Success {
		return nil, fmt.Errorf("prover request failed %s", *res.ErrorMessage)
	}

	if res.Proof == nil {
		return nil, fmt.Errorf("proof is nil")
	}

	return res.Proof, nil
}

func (w *WithdrawalAggregatorService) RequestWithdrawalGnarkProofToProver(wrappedProofJSON []byte) (gnarkProof *GnarkGetProofResponseResult, err error) {
	JobID, err := w.requestWithdrawalGnarkProofToProver(wrappedProofJSON)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-w.ctx.Done():
			return nil, err
		case <-ticker.C:
			resFetching, errFetching := w.fetchWithdrawalGnarkProofToProver(JobID)
			if errFetching != nil {
				var ErrWrappedWithdrawalProofFetching = errors.New("failed to fetch wrapper withdrawal proof")
				err = errors.Join(ErrWrappedWithdrawalProofFetching, errFetching)
				continue
			}

			return resFetching, nil
		}
	}
}

func (w *WithdrawalAggregatorService) requestWithdrawalGnarkProofToProver(wrappedProofJSON []byte) (string, error) {
	apiUrl := fmt.Sprintf("%s/start-proof",
		w.cfg.API.WithdrawalGnarkProverUrl,
	)
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(wrappedProofJSON)) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res GnarkStartProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return res.JobID, nil
}

func (w *WithdrawalAggregatorService) fetchWithdrawalGnarkProofToProver(jobID string) (wrappedProof *GnarkGetProofResponseResult, err error) {
	apiUrl := fmt.Sprintf("%s/get-proof?jobId=%s",
		w.cfg.API.WithdrawalGnarkProverUrl,
		jobID,
	)
	resp, err := http.Get(apiUrl) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to request API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res GnarkGetProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if res.Status != "done" {
		return nil, fmt.Errorf("prover request failed")
	}

	return &res.Result, nil
}
