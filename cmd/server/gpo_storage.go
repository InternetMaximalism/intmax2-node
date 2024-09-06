package server

import (
	"context"
	"math/big"
)

//go:generate mockgen -destination=mock_gpo_storage.go -package=server -source=gpo_storage.go

type GPOStorage interface {
	Init(ctx context.Context) (err error)
	Value(ctx context.Context, name string) (*big.Int, error)
	UpdValue(ctx context.Context, name string) (err error)
	UpdValues(ctx context.Context, name ...string) (err error)
}
