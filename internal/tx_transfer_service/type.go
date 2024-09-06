package tx_transfer_service

import (
	"encoding/json"
	intMaxTypes "intmax2-node/internal/types"
)

type SimpleResponseData struct {
	Message string `json:"message"`
}

type SendTransactionResponse struct {
	Success bool               `json:"success"`
	Data    SimpleResponseData `json:"data"`
}

type GetTransactionData struct {
	ID          string `json:"id"`
	Sender      string `json:"sender"`
	Signature   string `json:"signature"`
	BlockNumber string `json:"blockNumber"`
	EncryptedTx string `json:"encryptedTx"`
	CreatedAt   string `json:"createdAt"`
}

type GetTransactionTxData struct {
	ID          string                 `json:"id"`
	Sender      string                 `json:"sender"`
	Signature   string                 `json:"signature"`
	BlockNumber string                 `json:"blockNumber"`
	TxDetails   *intMaxTypes.TxDetails `json:"txDetails"`
	CreatedAt   string                 `json:"createdAt"`
}

type GetTransactionTxResponse struct {
	Success bool                  `json:"success"`
	Data    *GetTransactionTxData `json:"data,omitempty"`
	GetTransactionsListError
}

type GetTransactionByHashData struct {
	Transaction *GetTransactionData `json:"transaction"`
}

type GetTransactionByHashResponse struct {
	Success bool                      `json:"success"`
	Data    *GetTransactionByHashData `json:"data,omitempty"`
	Error   *GetTransactionsListError `json:"error,omitempty"`
}

type GetTransactionsListResponse struct {
	Success bool                      `json:"success"`
	Data    *GetTransactionsListData  `json:"data,omitempty"`
	Error   *GetTransactionsListError `json:"error,omitempty"`
}

type GetTransactionsListError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetTransactionsListData struct {
	Transactions []*GetTransactionData `json:"transactions"`
	Pagination   json.RawMessage       `json:"pagination"`
}

type GetTxTransactionsListData struct {
	TxHashes []string `json:"txHashes"`
}

type GetTransactionsList struct {
	Success bool                       `json:"success"`
	Data    *GetTxTransactionsListData `json:"data,omitempty"`
	GetTransactionsListError
}
