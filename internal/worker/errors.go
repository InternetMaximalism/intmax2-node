package worker

import "errors"

// ErrMkdirTempFail error: failed to mkdir temp.
var ErrMkdirTempFail = errors.New("failed to mkdir temp")

// ErrMkdirFail error: failed to mkdir.
var ErrMkdirFail = errors.New("failed to mkdir")

// ErrRemoveAllFail error: failed to remove all.
var ErrRemoveAllFail = errors.New("failed to remove all")

// ErrCreateTempFail error: failed to create temp.
var ErrCreateTempFail = errors.New("failed to create temp")

// ErrCreateNewTempFileFail error: failed to create new temp file.
var ErrCreateNewTempFileFail = errors.New("failed to create new temp file")

// ErrCreateNewTempDirFail error: failed to create new temp directory.
var ErrCreateNewTempDirFail = errors.New("failed to create new temp directory")

// ErrStatCurrentFileFail error: failed to get stat of current file.
var ErrStatCurrentFileFail = errors.New("failed to get stat of current file")

// ErrReceiverWorkerEmpty error: the Receiver worker must not be empty.
var ErrReceiverWorkerEmpty = errors.New("the Receiver worker must not be empty")

// ErrReceiverWorkerDuplicate error: the Receiver worker must be unique.
var ErrReceiverWorkerDuplicate = errors.New("the Receiver worker must be unique")

// ErrRegisterReceiverFail error: failed to register Receiver.
var ErrRegisterReceiverFail = errors.New("failed to register Receiver")

// ErrOpenFileKvStoreFail error: failed to open file with key-value store.
var ErrOpenFileKvStoreFail = errors.New("failed to open file with key-value store")

// ErrKVStoreFail error: failed to get key-value store.
var ErrKVStoreFail = errors.New("failed to get key-value store")

// ErrTxBeginKVStoreFail error: failed to begin tx with key-value store.
var ErrTxBeginKVStoreFail = errors.New("failed to begin tx with key-value store")

// ErrTxCommitKVStoreFail error: failed to commit tx with key-value store.
var ErrTxCommitKVStoreFail = errors.New("failed to commit tx with key-value store")

// ErrMarshalFail error: failed to marshal.
var ErrMarshalFail = errors.New("failed to marshal")

// ErrUnmarshalFail error: failed to unmarshal.
var ErrUnmarshalFail = errors.New("failed to unmarshal")

// ErrPutBucketKVStoreFail error: failed to put into bucket with key-value store.
var ErrPutBucketKVStoreFail = errors.New("failed to put into bucket with key-value store")

// ErrCreateBucketKVStoreFail error: failed to create bucket with key-value store.
var ErrCreateBucketKVStoreFail = errors.New("failed to create bucket with key-value store")

// ErrNewTransferTreeFail error: failed to create new transfer tree.
var ErrNewTransferTreeFail = errors.New("failed to create new transfer tree")

// ErrCurrentRootCountAndSiblingsFromRW error: failed to get current root and count and siblings from the Receiver.
var ErrCurrentRootCountAndSiblingsFromRW = errors.New(
	"failed to get current root and count and siblings from the Receiver",
)

// ErrCreateTxMerkleProofsFail error: failed to create the tx merkle proofs.
var ErrCreateTxMerkleProofsFail = errors.New("failed to create the tx merkle proofs")
