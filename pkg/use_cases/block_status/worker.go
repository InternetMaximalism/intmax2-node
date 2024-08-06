package block_status

//go:generate mockgen -destination=mock_worker_test.go -package=block_status_test -source=worker.go

type Worker interface {
	ExistsTxTreeRoot(txTreeRoot string) error
}
