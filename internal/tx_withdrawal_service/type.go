package tx_withdrawal_service

import (
	"encoding/json"
	"intmax2-node/internal/tx_transfer_service"
)

type SimpleResponseData struct {
	Message string `json:"message"`
}

type SendTransactionResponse struct {
	Success bool               `json:"success"`
	Data    SimpleResponseData `json:"data"`
}

type GetTransferData struct {
	ID                string `json:"id"`
	Recipient         string `json:"recipient"`
	BlockNumber       string `json:"blockNumber"`
	EncryptedTransfer string `json:"encryptedTransfer"`
	CreatedAt         string `json:"createdAt"`
}

type GetTransfersListData struct {
	Transfers  []*GetTransferData `json:"transfers"`
	Pagination json.RawMessage    `json:"pagination"`
}

type GetTransfersListResponse struct {
	Success bool                   `json:"success"`
	Data    *GetTransfersListData  `json:"data,omitempty"`
	Error   *GetTransfersListError `json:"error,omitempty"`
}

type GetTxWithdrawalTransfersListData struct {
	Transfers []*tx_transfer_service.BackupWithdrawal `json:"transfers"`
}

type GetTransfersListError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetTxWithdrawalTransfersList struct {
	Success bool                              `json:"success"`
	Data    *GetTxWithdrawalTransfersListData `json:"data,omitempty"`
	GetTransfersListError
}
