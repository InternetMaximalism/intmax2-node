package intmax_block_service

type GetDataError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetINTMAXBlockInfoData struct {
	BlockNumber                 int64  `json:"blockNumber"`
	BlockHash                   string `json:"blockHash"`
	Status                      string `json:"status"`
	ExecutedBlockHashOnScroll   string `json:"executedBlockHashOnScroll"`
	ExecutedBlockHashOnEthereum string `json:"executedBlockHashOnEthereum"`
}

type GetINTMAXBlockInfoResponse struct {
	Success bool                    `json:"success"`
	Data    *GetINTMAXBlockInfoData `json:"data,omitempty"`
	Error   *GetDataError           `json:"error,omitempty"`
}
