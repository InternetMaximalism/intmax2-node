package block_validity_prover

type BlockValidityProver interface {
	SyncDepositTree() error
}
