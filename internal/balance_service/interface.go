package balance_service

type GetBalancesResponse struct {
	// The list of deposits
	Deposits []*BackupDeposit `json:"deposits,omitempty"`
	// The list of transfers
	Transfers []*BackupTransfer `json:"transfers,omitempty"`
	// The list of transactions
	Transactions []*BackupTransaction `json:"transactions,omitempty"`
}

type BackupDeposit struct {
	Recipient        string `json:"recipient,omitempty"`
	EncryptedDeposit string `json:"encryptedDeposit,omitempty"`
	BlockNumber      string `json:"blockNumber,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
}

type BackupTransfer struct {
	EncryptedTransfer                string `json:"encryptedTransfer,omitempty"`
	Recipient                        string `json:"recipient,omitempty"`
	BlockNumber                      string `json:"blockNumber,omitempty"`
	SenderLastBalanceProofBody       string `json:"senderLastBalanceProofBody,omitempty"`
	SenderBalanceTransitionProofBody string `json:"senderBalanceTransitionProofBody,omitempty"`
	CreatedAt                        string `json:"createdAt,omitempty"`
}

type BackupTransaction struct {
	Sender          string `json:"sender,omitempty"`
	EncodingVersion uint32 `json:"encodingVersion"`
	EncryptedTx     string `json:"encryptedTx,omitempty"`
	BlockNumber     string `json:"blockNumber,omitempty"`
	CreatedAt       string `json:"createdAt,omitempty"`
}

type GetVerifyDepositConfirmationResponse struct {
	// Indicates if the verify deposit confirmation was successful
	Success bool ` json:"success,omitempty"`
	// Additional data related to the response
	Data *GetVerifyDepositConfirmationResponse_Data `json:"data,omitempty"`
}

type GetVerifyDepositConfirmationResponse_Data struct {
	// Indicates whether the deposit is confirmed
	Confirmed bool `json:"confirmed,omitempty"`
}
