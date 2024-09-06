package block_info

import (
	"context"
	"errors"
	"intmax2-node/configs"
	gpo "intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	blockInfo "intmax2-node/internal/use_cases/block_info"
	"math/big"
	"strconv"
)

// uc describes use case
type uc struct {
	cfg        *configs.Config
	storageGPO GPOStorage
}

func New(
	cfg *configs.Config,
	storageGPO GPOStorage,
) blockInfo.UseCaseBlockInfo {
	return &uc{
		cfg:        cfg,
		storageGPO: storageGPO,
	}
}

func (u *uc) Do(
	ctx context.Context,
) (*blockInfo.UCBlockInfo, error) {
	const (
		hName = "UseCase BlockInfo"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	info := blockInfo.UCBlockInfo{
		TransferFee: make(map[string]string),
		Difficulty:  int64(u.cfg.PoW.Difficulty),
	}

	w, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(u.cfg.Blockchain.BuilderPrivateKeyHex)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrInvalidPrivateKey, err)
	}
	info.ScrollAddress = w.WalletAddress.String()
	info.IntMaxAddress = w.IntMaxWalletAddress

	list := []string{
		gpo.ScrollEthGPO,
	}

	for key := range list {
		var v *big.Int
		v, err = u.storageGPO.Value(spanCtx, list[key])
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(ErrStorageGPOValueFail, err)
		}
		info.TransferFee[strconv.Itoa(key)] = v.String()
	}

	return &info, nil
}
