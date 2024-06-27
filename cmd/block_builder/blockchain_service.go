package block_builder

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=block_builder -source=blockchain_service.go

type ServiceBlockchain interface {
	GenericCommandsSB
}

type GenericCommandsSB interface {
	CheckScrollPrivateKey(ctx context.Context) (err error)
}
