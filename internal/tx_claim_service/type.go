package tx_claim_service

type SimpleResponseData struct {
	Message string `json:"message"`
}

type SendTransactionResponse struct {
	Success bool               `json:"success"`
	Data    SimpleResponseData `json:"data"`
}
