package block_validity_prover_tx_root_status

import (
	"context"
	"intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_tx_root_status.go -package=mocks -source=block_validity_prover_tx_root_status.go

type UCBlockValidityProverTxRootStatus struct {
	IsRegistrationBlock bool
	TxTreeRoot          common.Hash
	PrevBlockHash       common.Hash
	BlockNumber         uint32
	DepositRoot         common.Hash
	SignatureHash       common.Hash
	MessagePoint        *bn254.G2Affine
	AggregatedPublicKey *accounts.PublicKey
	AggregatedSignature *bn254.G2Affine
	Senders             []intMaxTypes.Sender
}

type TxRootError struct {
	Message string `json:"-"`
}

type UCBlockValidityProverTxRootStatusInput struct {
	TxRoots       []string                `json:"txRoots"`
	ConvertTxRoot []common.Hash           `json:"-"`
	TxRootErrors  map[string]*TxRootError `json:"-"`
}

type UseCaseBlockValidityProverTxRootStatus interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverTxRootStatusInput,
	) (map[string]*UCBlockValidityProverTxRootStatus, error)
}
