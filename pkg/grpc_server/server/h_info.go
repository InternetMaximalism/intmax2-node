package server

import (
	"context"
	errorsB "intmax2-node/internal/blockchain/errors"
	gpo "intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/pkg/grpc_server/utils"
	"math/big"
	"strconv"
)

func (s *Server) Info(
	ctx context.Context,
	_ *node.InfoRequest,
) (*node.InfoResponse, error) {
	resp := node.InfoResponse{}

	const (
		hName = "Handler Info"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	resp.Data = &node.DataInfoResponse{
		TransferFee: make(map[string]string),
		Difficulty:  uint32(s.config.PoW.Difficulty),
	}

	w, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(s.config.Blockchain.BuilderPrivateKeyHex)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "%s: %s"
		return &resp, utils.Internal(spanCtx, s.log, msg, errorsB.ErrWalletAddressNotRecognized, err)
	}

	list := []string{
		gpo.ScrollEthGPO,
	}

	for key := range list {
		var v *big.Int
		v, err = s.storageGPO.Value(spanCtx, list[key])
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get the gas price oracle value: %v"
			return &resp, utils.Internal(spanCtx, s.log, msg, err)
		}
		resp.Data.TransferFee[strconv.Itoa(key)] = v.String()
	}

	resp.Data.ScrollAddress = w.WalletAddress.String()
	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
