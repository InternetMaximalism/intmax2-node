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

type GetTransactionByHashData struct {
	Transaction *GetTransactionData `json:"transaction"`
}

type GetTransactionByHashResponse struct {
	Success bool                      `json:"success"`
	Data    *GetTransactionByHashData `json:"data"`
}

type GetTransactionsListResponse struct {
	Success bool                     `json:"success"`
	Data    *GetTransactionsListData `json:"data"`
}

type GetTransactionsListData struct {
	Transactions []*GetTransactionData `json:"transactions"`
	Pagination   json.RawMessage       `json:"pagination"`
}

type GetTransactionsList struct {
	TxHashes []string `json:"txHashes"`
}
