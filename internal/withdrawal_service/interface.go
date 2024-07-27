package withdrawal_service

type GenerateProofResponse struct {
	Success      bool   `json:"success"`
	Value        int    `json:"value"`
	ErrorMessage string `json:"error_message"`
}

type ProofValue struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
}

type ProofsResponse struct {
	Success      bool         `json:"success"`
	Values       []ProofValue `json:"values"`
	ErrorMessage string       `json:"error_message,omitempty"`
}

type ScrollMessengerResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Data    struct {
		Results []*ScrollMessengerResult `json:"results"`
		Total   int                      `json:"total"`
	} `json:"data"`
}

type ScrollMessengerResult struct {
	Hash               string   `json:"hash"`
	ReplayTxHash       string   `json:"replay_tx_hash"`
	RefundTxHash       string   `json:"refund_tx_hash"`
	MessageHash        string   `json:"message_hash"`
	TokenType          int      `json:"token_type"`
	TokenIds           []int    `json:"token_ids"`
	TokenAmounts       []string `json:"token_amounts"`
	MessageType        int      `json:"message_type"`
	L1TokenAddress     string   `json:"l1_token_address"`
	L2TokenAddress     string   `json:"l2_token_address"`
	BlockNumber        int      `json:"block_number"`
	TxStatus           int      `json:"tx_status"`
	CounterpartChainTx struct {
		Hash        string `json:"hash"`
		BlockNumber int    `json:"block_number"`
	} `json:"counterpart_chain_tx"`
	ClaimInfo       *ClaimInfo `json:"claim_info"`
	BlockTimestamp  int        `json:"block_timestamp"`
	BatchDepositFee string     `json:"batch_deposit_fee"`
}

type ClaimInfo struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Value   string `json:"value"`
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
	Proof   struct {
		BatchIndex  string `json:"batch_index"`
		MerkleProof string `json:"merkle_proof"`
	} `json:"proof"`
	Claimable bool `json:"claimable"`
}

type WithdrawalProverParameters struct {
	ID         string `json:"id"`
	Recipient  string `json:"recipient"`
	TokenIndex string `json:"token_index"`
	Amount     string `json:"amount"`
	Salt       string `json:"salt"`
	BlockHash  string `json:"block_hash"`
}
