package block_validity_prover_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/dimiro1/health"
)

type BlockValidityProverServer struct {
	log              logger.Logger
	config           *configs.Config
	dbApp            SQLDriverApp
	commands         Commands
	cookieForAuthUse bool
	hc               *health.Handler
	sb               ServiceBlockchain
}

// New initializes a new Server struct.
func New(
	log logger.Logger,
	config *configs.Config,
	dbApp SQLDriverApp,
	commands Commands,
	cookieForAuthUse bool,
	hc *health.Handler,
	sb ServiceBlockchain,
) *BlockValidityProverServer {
	const (
		srv  = "blockValidityProverServer"
		name = "blockvalidityproverserver"
	)

	return &BlockValidityProverServer{
		log:              log.WithFields(logger.Fields{srv: name}),
		config:           config,
		dbApp:            dbApp,
		commands:         commands,
		cookieForAuthUse: cookieForAuthUse,
		hc:               hc,
		sb:               sb,
	}
}

func (s *BlockValidityProverServer) Commands() Commands {
	return s.commands
}
