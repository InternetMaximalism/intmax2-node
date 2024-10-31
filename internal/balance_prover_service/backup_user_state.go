package balance_prover_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	intMaxTypes "intmax2-node/internal/types"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type BackupUserStateData struct {
	Id                     string
	UserAddress            intMaxAcc.Address
	BlockNumber            uint32
	PrivateStateCommitment goldenposeidon.PoseidonHashOut
	BalanceProof           *BalanceProofWithPublicInputs
	EncryptedUserState     []byte
	AuthSignature          string
	CreatedAt              time.Time
}

func BackupUserStateRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	user intMaxAcc.Address,
	blockNumber uint32,
	balanceProof string,
	encryptedUserState string,
	signature string,
) (*BackupUserStateData, error) {
	userState, err := backupUserStateRawRequest(
		ctx, cfg, log,
		user.String(),
		blockNumber,
		balanceProof,
		encryptedUserState,
		signature,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user address: %w", err)
	}

	return convertBackupUserStateData(userState)
}

func backupUserStateRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	user string,
	blockNumber uint32,
	balanceProof string,
	encryptedUserState string,
	signature string,
) (*node.BackupUserStateResponse_Data_Balance, error) {
	fmt.Printf("size of balanceProof: %d", len(balanceProof))
	ucInput := node.BackupUserStateRequest{
		UserAddress:        user,
		BlockNumber:        blockNumber,
		BalanceProof:       balanceProof,
		EncryptedUserState: encryptedUserState,
		AuthSignature:      signature,
	}

	bd, err := json.Marshal(&ucInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/user-state", cfg.API.DataStoreVaultUrl)

	r := resty.New().R()
	var resp *resty.Response
	resp, err = r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).SetBody(bd).Post(apiUrl)
	if err != nil {
		const msg = "failed to send user state backup request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "user state backup request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(node.BackupUserStateResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to backup user state: %s", response.Data)
	}

	return response.Data.Balance, nil
}

func GetBackupUserStateRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	user intMaxAcc.Address,
) (*BackupUserStateData, error) {
	userState, err := getBackupUserStateRawRequest(
		ctx, cfg, log,
		user.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user address: %w", err)
	}

	return convertGetBackupUserStateData(userState)
}

func getBackupUserStateRawRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	user string,
) (*node.GetBackupUserStateResponse_Data_Balance, error) {
	const (
		httpKey     = "http"
		httpsKey    = "https"
		contentType = "Content-Type"
		appJSON     = "application/json"
	)

	apiUrl := fmt.Sprintf("%s/v1/backups/user-state/%s", cfg.API.DataStoreVaultUrl, user)

	r := resty.New().R()
	var resp *resty.Response
	resp, err := r.SetContext(ctx).SetHeaders(map[string]string{
		contentType: appJSON,
	}).Get(apiUrl)
	if err != nil {
		const msg = "failed to fetch user state backup request: %w"
		return nil, fmt.Errorf(msg, err)
	}

	if resp == nil {
		const msg = "user state backup request error occurred"
		return nil, errors.New(msg)
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get response")
		log.WithFields(logger.Fields{
			"status_code": resp.StatusCode(),
			"response":    resp.String(),
		}).WithError(err).Errorf(unexpectedStatusCode)
		return nil, err
	}

	response := new(node.GetBackupUserStateResponse)
	if err = json.Unmarshal(resp.Body(), response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to backup user state: %s", response.Data)
	}

	return response.Data.Balance, nil
}

func convertGetBackupUserStateData(
	userState *node.GetBackupUserStateResponse_Data_Balance,
) (*BackupUserStateData, error) {
	userAddress, err := intMaxAcc.NewAddressFromHex(userState.UserAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user address: %w", err)
	}

	plonky2ProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(userState.BalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance proof: %w", err)
	}

	balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(plonky2ProofWithPis.PublicInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance proof: %w", err)
	}

	balanceProofWithPis := BalanceProofWithPublicInputs{
		Proof:        userState.BalanceProof,
		PublicInputs: balancePublicInputs,
	}

	return &BackupUserStateData{
		Id:                 userState.Id,
		UserAddress:        userAddress,
		BlockNumber:        userState.BlockNumber,
		BalanceProof:       &balanceProofWithPis,
		EncryptedUserState: []byte(userState.EncryptedUserState),
		AuthSignature:      userState.AuthSignature,
		CreatedAt:          userState.CreatedAt.AsTime(),
	}, nil
}

func convertBackupUserStateData(
	userState *node.BackupUserStateResponse_Data_Balance,
) (*BackupUserStateData, error) {
	userAddress, err := intMaxAcc.NewAddressFromHex(userState.UserAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user address: %w", err)
	}

	plonky2ProofWithPis, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(userState.BalanceProof)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance proof: %w", err)
	}

	balancePublicInputs, err := new(BalancePublicInputs).FromPublicInputs(plonky2ProofWithPis.PublicInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance proof: %w", err)
	}

	balanceProofWithPis := BalanceProofWithPublicInputs{
		Proof:        userState.BalanceProof,
		PublicInputs: balancePublicInputs,
	}

	return &BackupUserStateData{
		Id:                 userState.UserStateId,
		UserAddress:        userAddress,
		BlockNumber:        userState.BlockNumber,
		BalanceProof:       &balanceProofWithPis,
		EncryptedUserState: []byte(userState.EncryptedUserState),
		AuthSignature:      userState.AuthSignature,
		CreatedAt:          userState.CreatedAt.AsTime(),
	}, nil
}
