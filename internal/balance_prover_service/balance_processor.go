package balance_prover_service

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
)

type BalanceProcessor struct{}

func (s *BalanceProcessor) ProveSend(
	publicKey *intMaxAcc.PublicKey,
	sendWitness *SendWitness,
	updateWitness *UpdateWitness,
	lastBalanceProof *intMaxTypes.Plonky2Proof,
) (*intMaxTypes.Plonky2Proof, error) {
	// request balance prover
	return nil, errors.New("not implemented")
}

func (s *BalanceProcessor) ProveUpdate(
	publicKey *intMaxAcc.PublicKey,
	updateWitness *UpdateWitness,
	lastBalanceProof *intMaxTypes.Plonky2Proof,
) (*intMaxTypes.Plonky2Proof, error) {
	// request balance prover
	return nil, errors.New("not implemented")
}

func (s *BalanceProcessor) ProveReceiveDeposit(
	publicKey *intMaxAcc.PublicKey,
	receiveDepositWitness *ReceiveDepositWitness,
	lastBalanceProof *intMaxTypes.Plonky2Proof,
) (*intMaxTypes.Plonky2Proof, error) {
	// request balance prover
	return nil, errors.New("not implemented")
}

func (s *BalanceProcessor) ProveReceiveTransfer(
	publicKey *intMaxAcc.PublicKey,
	receiveTransferWitness *ReceiveTransferWitness,
	lastBalanceProof *intMaxTypes.Plonky2Proof,
) (*intMaxTypes.Plonky2Proof, error) {
	// request balance prover
	return nil, errors.New("not implemented")
}
