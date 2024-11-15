package block_validity_prover

//go:generate mockgen -destination=mock_block_validity_prover.go -package=block_validity_prover -source=block_validity_prover.go

type BlockValidityProver interface {
	SyncDepositedEvents() error
	SyncDepositTree(endBlock *uint64, depositIndex uint32) error
	SyncBlockContent() (lastEventSeenBlockNumber uint64, err error)
	SyncBlockValidityProof() error
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
}
