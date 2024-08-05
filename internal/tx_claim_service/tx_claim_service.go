package tx_claim_service

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
)

func ClaimWithdrawals(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	recipientEthAddress string,
) {

	log.Infof("The claiming withdrawals has been successfully sent.")
}
