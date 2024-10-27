package data_store_vault

import (
	"math/big"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	types "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
)

type DepositDetails struct {
	Recipient         *intMaxAcc.PublicKey
	TokenIndex        uint32
	Amount            *big.Int
	Salt              *intMaxGP.PoseidonHashOut
	RecipientSaltHash common.Hash
	DepositID         uint32
	DepositHash       common.Hash
}

type TransferDetails struct {
	TransferWitness              *types.TransferWitness
	TxTreeRoot                   intMaxGP.PoseidonHashOut
	TxIndex                      uint32
	TxMerkleProof                []intMaxGP.PoseidonHashOut
	SenderEnoughBalanceProofUUID string
}

type TransferDetailProof struct {
	SenderLastBalanceProof       *balance_prover_service.BalanceProofWithPublicInputs
	SenderBalanceTransitionProof *balance_prover_service.SpentProofWithPublicInputs
}

type TxDetails struct {
	Tx            *types.Tx
	TxTreeRoot    intMaxGP.PoseidonHashOut
	TxIndex       uint32
	TxMerkleProof []intMaxGP.PoseidonHashOut
}
