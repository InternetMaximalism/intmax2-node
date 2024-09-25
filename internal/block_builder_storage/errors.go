package block_builder_storage

import "errors"

// ErrLastBlockNumberGeneratedValidityProofFail error: failed to get last block number generated validity proof.
var ErrLastBlockNumberGeneratedValidityProofFail = errors.New(
	"failed to get last block number generated validity proof",
)

// ErrScanBlockHashAndSendersFailed error: failed to scan block hash and senders.
var ErrScanBlockHashAndSendersFailed = errors.New(
	"failed to scan block hash and senders",
)

// ErrNewAccountTreeFail error: failed to create new account tree.
var ErrNewAccountTreeFail = errors.New("failed to create new account tree")

// ErrNewBlockHashTreeWithInitialLeavesFail error: failed to create new block hash tree with initial leaves.
var ErrNewBlockHashTreeWithInitialLeavesFail = errors.New(
	"failed to create new block hash tree with initial leaves",
)

// ErrNewKeccakMerkleTreeFail error: failed to create new keccak merkle tree.
var ErrNewKeccakMerkleTreeFail = errors.New("failed to create new keccak merkle tree")

// ErrBlockNumberNotFoundWithBlockHashAndSendersMap error: block number not found with map of block hash and senders.
var ErrBlockNumberNotFoundWithBlockHashAndSendersMap = errors.New(
	"block number not found with map of block hash and senders",
)

// ErrNewPublicKeyFromAddressHexFail error: failed to get public key from INTMAX address.
var ErrNewPublicKeyFromAddressHexFail = errors.New("failed to get public key from INTMAX address")

// ErrAddLeafToBlockHashTreeFail error: failed to add leaf to block hash tree.
var ErrAddLeafToBlockHashTreeFail = errors.New("failed to add leaf to block hash tree")

// ErrInsertAccountTreeFail error: failed to insert account tree.
var ErrInsertAccountTreeFail = errors.New("failed to insert account tree")

// ErrUpdateAccountTreeFail error: failed to update account tree.
var ErrUpdateAccountTreeFail = errors.New("failed to update account tree")

// ErrScanDepositsFail error: failed to scan deposits.
var ErrScanDepositsFail = errors.New("failed to scan deposits")

// ErrAddLeafToDepositTreeFail error: failed to add leaf to deposit tree.
var ErrAddLeafToDepositTreeFail = errors.New("failed to add leaf to deposit tree")

// ErrGenesisValidityProof error: genesis validity proof error.
var ErrGenesisValidityProof = errors.New("genesis validity proof error")

// ErrOfBlockContentByBlockNumber error: block content by block number error.
var ErrOfBlockContentByBlockNumber = errors.New("block content by block number error")

// ErrValidityProofByBlockNumberFail error: failed to validity proof by block number.
var ErrValidityProofByBlockNumberFail = errors.New("failed to validity proof by block number")

// ErrBlockTreeProofFail error: failed to get block tree proof.
var ErrBlockTreeProofFail = errors.New("failed to get block tree proof")

// ErrGetAccountMembershipProofFail error: failed to get account membership proof.
var ErrGetAccountMembershipProofFail = errors.New("failed to get account membership proof")

// ErrGetDepositLeafAndIndexByHashFail error: failed to get deposit leaf and index by hash.
var ErrGetDepositLeafAndIndexByHashFail = errors.New("failed to get deposit leaf and index by hash")

// ErrAccountTreeInsert error: account tree insert error.
var ErrAccountTreeInsert = errors.New("account tree insert error")

// ErrAccountTreeUpdate error: account tree update error.
var ErrAccountTreeUpdate = errors.New("account tree update error")

// ErrAccountTreeGetAccountID error: account id not found in the account tree.
var ErrAccountTreeGetAccountID = errors.New("account id not found in the account tree")

// ErrBlockTreeProve error: block tree prove error.
var ErrBlockTreeProve = errors.New("block tree prove error")

// ErrAppendAccountTreeLeaf error: append account tree leaf error.
var ErrAppendAccountTreeLeaf = errors.New("append account tree leaf error")

// ErrAccountTreeLeaf error; account tree leaf error.
var ErrAccountTreeLeaf = errors.New("account tree leaf error")

// ErrUpdateAccountTreeLeaf error: update account tree leaf error.
var ErrUpdateAccountTreeLeaf = errors.New("update account tree leaf error")

// ErrBlockNumberByDepositIndexFail error: failed to get block number by deposit index.
var ErrBlockNumberByDepositIndexFail = errors.New("failed to get block number by deposit index")

// ErrIsSynchronizedDepositIndexFail error: failed to check if deposit index is synchronized.
var ErrIsSynchronizedDepositIndexFail = errors.New("failed to check if deposit index is synchronized")

// ErrBlockTreeAddLeaf error: block tree add leaf error.
var ErrBlockTreeAddLeaf = errors.New("block tree add leaf error")

// ErrSetValidityWitnessFail error: failed to set validity witness.
var ErrSetValidityWitnessFail = errors.New("failed to set validity witness")

// ErrDepositTreeProof error: deposit tree proof error.
var ErrDepositTreeProof = errors.New("deposit tree proof error")

// ErrProveMembershipFail error: failed to prove membership.
var ErrProveMembershipFail = errors.New("failed to prove membership")

// ErrUpdateAccountFail error: failed to update account.
var ErrUpdateAccountFail = errors.New("failed to update account")

// ErrCreateAccountFail error: failed to create account.
var ErrCreateAccountFail = errors.New("failed to create account")

// ErrDecodeHexToPublicKeyFail error: failed to decode hex to public key.
var ErrDecodeHexToPublicKeyFail = errors.New("failed to decode hex to public key")
