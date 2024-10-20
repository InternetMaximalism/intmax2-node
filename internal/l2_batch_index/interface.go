package l2_batch_index

import "context"

const (
	L2BlockNumberJobMask = "l2_block_number_"
	L2BatchIndexJobMask  = "l2_batch_index_"
)

type L2IndexIndex interface {
	Start(ctx context.Context) error
}
