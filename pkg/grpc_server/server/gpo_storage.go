package server

import (
	"context"
	"math/big"
)

//go:generate mockgen -destination=mock_gpo_storage_test.go -package=server_test -source=gpo_storage.go

type GPOStorage interface {
	Value(ctx context.Context, name string) (*big.Int, error)
}
