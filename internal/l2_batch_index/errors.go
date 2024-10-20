package l2_batch_index

import "errors"

// ErrNeedRepeatAction error: need repeat action.
var ErrNeedRepeatAction = errors.New("need repeat action")

// ErrCreateCtrlProcessingJobsFail error: failed to create new ctrl-processing-job by the mask name.
var ErrCreateCtrlProcessingJobsFail = errors.New("failed to create new ctrl-processing-job by the mask name")

// ErrCtrlProcessingJobsByMaskNameFail error: failed to get ctrl-processing-job by the mask name.
var ErrCtrlProcessingJobsByMaskNameFail = errors.New("failed to get ctrl-processing-job by the mask name")

// ErrJSONUnmarshalFail error: failed to unmarshal JSON.
var ErrJSONUnmarshalFail = errors.New("failed to unmarshal JSON")

// ErrUpdateAtOfCtrlProcessingJobByNameFail error: failed to update the updateAt value by name of ctrl-processing-job.
var ErrUpdateAtOfCtrlProcessingJobByNameFail = errors.New(
	"failed to update the updateAt value by name of ctrl-processing-job",
)

// ErrProcessingL2BlockNumberFail error: failed to processing L2 block number.
var ErrProcessingL2BlockNumberFail = errors.New("failed to processing L2 block number")

// ErrDeleteCtrlProcessingJobByNameFail error: failed to delete ctrl-processing-job by name.
var ErrDeleteCtrlProcessingJobByNameFail = errors.New("failed to delete ctrl-processing-job by name")

// ErrApplyTickerL2BlockNumberWithDBAppFail error: failed to apply ticker L2 block number with DB App.
var ErrApplyTickerL2BlockNumberWithDBAppFail = errors.New("failed to apply ticker L2 block number with DB App")

// ErrBlockContentIDByL2BlockNumberFail error: failed to get the block content ID value by l2-block-number.
var ErrBlockContentIDByL2BlockNumberFail = errors.New(
	"failed to get the block content ID value by l2-block-number",
)

// ErrScrollNetworkChainLinkEvmJSONRPCFail error: failed to get the scroll network chain link evm JSON RPC.
var ErrScrollNetworkChainLinkEvmJSONRPCFail = errors.New(
	"failed to get the scroll network chain link evm JSON RPC",
)

// ErrCreateNewClientOfRPCEthFail error: failed to create new RPC Eth client.
var ErrCreateNewClientOfRPCEthFail = errors.New(
	"failed to create new RPC Eth client",
)

// ErrBlockByNumberWithClientOfRPCEthFail error: failed to get block number with RPC Eth client.
var ErrBlockByNumberWithClientOfRPCEthFail = errors.New(
	"failed to get block number with RPC Eth client",
)

// ErrScrollNetworkChainLinkRollupExplorerFail error: failed to get the rollup explorer link of scroll network chain.
var ErrScrollNetworkChainLinkRollupExplorerFail = errors.New(
	"failed to get the rollup explorer link of scroll network chain",
)

// ErrCreateL2BatchIndexWithDBFail error: failed to create l2-batch-index with DB.
var ErrCreateL2BatchIndexWithDBFail = errors.New("failed to create l2-batch-index with DB")

// ErrCreateRelationshipL2BatchIndexAndBlockContentIDWithDBFail error: failed to create relationship of l2-batch-index and block-content-id with DB.
var ErrCreateRelationshipL2BatchIndexAndBlockContentIDWithDBFail = errors.New(
	"failed to create relationship of l2-batch-index and block-content-id with DB",
)

// ErrProcessingL2BatchIndexFail error: failed to processing L2 batch index.
var ErrProcessingL2BatchIndexFail = errors.New("failed to processing L2 batch index")

// ErrL2BatchIndexFail error: failed to get l2 batch index.
var ErrL2BatchIndexFail = errors.New("failed to get l2 batch index")

// ErrUpdOptionsOfBatchIndexWithDBFail error: failed to update options of batch index with DB.
var ErrUpdOptionsOfBatchIndexWithDBFail = errors.New("failed to update options of batch index with DB")

// ErrUpdL1VerifiedBatchTxHashOfBatchIndexWithDBFail error: failed to update l1_verified_batch_tx_hash of batch index with DB.
var ErrUpdL1VerifiedBatchTxHashOfBatchIndexWithDBFail = errors.New(
	"failed to update l1_verified_batch_tx_hash of batch index with DB",
)
