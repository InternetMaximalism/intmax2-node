package tx_deposit_service

import (
	"encoding/json"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
)

type GetDepositData struct {
	ID               string `json:"id"`
	DepositHash      string `json:"depositHash"`
	Recipient        string `json:"recipient"`
	BlockNumber      string `json:"blockNumber"`
	EncryptedDeposit string `json:"encryptedDeposit"`
	CreatedAt        string `json:"createdAt"`
}

type GetDepositTxData struct {
	ID          string               `json:"id"`
	Recipient   string               `json:"recipient"`
	BlockNumber string               `json:"blockNumber"`
	Deposit     *intMaxTypes.Deposit `json:"deposit"`
	CreatedAt   string               `json:"createdAt"`
}

type GetTransactionByHashData struct {
	Transaction *GetDepositData `json:"transaction"`
}

type GetTransactionByHashError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetDepositsListResponse struct {
	Success bool                       `json:"success"`
	Data    *GetDepositsListData       `json:"data"`
	Error   *GetTransactionByHashError `json:"error"`
}

type GetDepositsListData struct {
	Deposits   []*GetDepositData `json:"deposits"`
	Pagination json.RawMessage   `json:"pagination"`
}

type Deposit struct {
	Hash       string
	Recipient  string
	TokenIndex uint32
	Amount     *big.Int
	Salt       string
}

type GetTxDepositByHashIncomingData struct {
	Deposits []*Deposit `json:"deposits,omitempty"`
}

type GetDepositsList struct {
	Success bool                            `json:"success"`
	Data    *GetTxDepositByHashIncomingData `json:"data,omitempty"`
	GetDepositByHashIncomingError
}

type GetDepositByHashIncomingError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetDepositByHashIncomingData struct {
	Deposit *GetDepositData `json:"deposit,omitempty"`
}

type GetDepositByHashIncomingResponse struct {
	Success bool                           `json:"success"`
	Data    *GetDepositByHashIncomingData  `json:"data,omitempty"`
	Error   *GetDepositByHashIncomingError `json:"error,omitempty"`
}

type GetDepositTxByHashIncomingResponse struct {
	Success bool              `json:"success"`
	Data    *GetDepositTxData `json:"data,omitempty"`
	GetDepositByHashIncomingError
}
