package balance_prover_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
)

type balanceProverService struct {
	ctx context.Context
	cfg *configs.Config
	log logger.Logger
}

func NewBalanceProverService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *balanceProverService {
	return &balanceProverService{
		ctx,
		cfg,
		log,
	}
}

func Start() error {
	privateKey, err := intMaxAcc.NewPrivateKeyFromString("7397927abf5b7665c4667e8cb8b92e929e287625f79264564bb66c1fa2232b2c")
	if err != nil {
		return err
	}
	fmt.Printf("private key: %v", privateKey)

	return nil
}
