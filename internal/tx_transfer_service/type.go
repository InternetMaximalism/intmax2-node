package tx_transfer_service

import (
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

type GetTransactionsListDataMeta struct {
	StartBlockNumber string `json:"startBlockNumber"`
	EndBlockNumber   string `json:"endBlockNumber"`
}

type GetTransactionsListData struct {
	Transactions []*GetTransactionData        `json:"transactions"`
	Meta         *GetTransactionsListDataMeta `json:"meta"`
}

type GetTransactionsListError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GetTransactionsListResponse struct {
	Success bool                      `json:"success"`
	Data    *GetTransactionsListData  `json:"data"`
	Error   *GetTransactionsListError `json:"error"`
}

type GetTransactionsListTransaction struct {
	BlockNumber string `json:"blockNumber"`
	TxHash      string `json:"txHash"`
	CreatedAt   string `json:"createdAt"`
}

type GetTransactionsList struct {
	Transactions []*GetTransactionsListTransaction `json:"transactions"`
	Meta         *GetTransactionsListDataMeta      `json:"meta"`
}

type GetTransactionByHashData struct {
	Transaction *GetTransactionData `json:"transaction"`
}

type GetTransactionByHashResponse struct {
	Success bool                      `json:"success"`
	Data    *GetTransactionByHashData `json:"data"`
}
