package tx_deposit_service

import (
	"encoding/json"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
)

type GetDepositData struct {
	ID               string `json:"id"`
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

type GetDepositsListResponse struct {
	Success bool                 `json:"success"`
	Data    *GetDepositsListData `json:"data"`
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

type GetDepositsList struct {
	Deposits []*Deposit `json:"deposits"`
}

type GetDepositByHashIncomingData struct {
	Deposit *GetDepositData `json:"deposit"`
}

type GetDepositByHashIncomingResponse struct {
	Success bool                          `json:"success"`
	Data    *GetDepositByHashIncomingData `json:"data"`
}
