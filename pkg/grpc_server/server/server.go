package server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/dimiro1/health"
)

type Server struct {
	log              logger.Logger
	config           *configs.Config
	dbApp            SQLDriverApp
	commands         Commands
	cookieForAuthUse bool
	hc               *health.Handler
}

// New initializes a new Server struct.
func New(
	log logger.Logger,
	config *configs.Config,
	dbApp SQLDriverApp,
	commands Commands,
	cookieForAuthUse bool,
	hc *health.Handler,
) *Server {
	const (
		srv  = "server"
		name = "users"
	)

	return &Server{
		log:              log.WithFields(logger.Fields{srv: name}),
		config:           config,
		dbApp:            dbApp,
		commands:         commands,
		cookieForAuthUse: cookieForAuthUse,
		hc:               hc,
	}
}

func (s *Server) Commands() Commands {
	return s.commands
}
