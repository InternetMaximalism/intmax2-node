package stop_block_builder

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=stop_block_builder -source=blockchain_service.go

type ServiceBlockchain interface {
	GenericCommandsSB
	WriteBlockchain
}

type GenericCommandsSB interface {
	CheckPrivateKey(ctx context.Context) (err error)
}
