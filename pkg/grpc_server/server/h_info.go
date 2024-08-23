package server

import (
	"context"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
)

func (s *Server) Info(
	ctx context.Context,
	req *node.InfoRequest,
) (*node.InfoResponse, error) {
	// TODO implement me
	panic("implement me")
}
