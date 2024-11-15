package block_validity_prover

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

var ErrNewEthereumClientFail = errors.New("failed to create new Ethereum client")

var ErrScrollNetwrokChainLink = errors.New("failed to get Scroll network chain link")

var ErrNewScrollClientFail = errors.New("failed to create new Scroll client")

var ErrInstantiateLiquidityContractFail = errors.New("failed to instantiate a Liquidity contract")

var ErrInstantiateRollupContractFail = errors.New("failed to instantiate a Rollup contract")

// ErrTransactionByHashNotFound error: failed to get transaction by hash.
var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

// ErrTransactionIsStillPending error: transaction is still pending.
var ErrTransactionIsStillPending = errors.New("transaction is still pending")

// ErrFetchLatestBlockNumberFail error: failed to fetch latest block number.
var ErrFetchLatestBlockNumberFail = errors.New("failed to fetch latest block number")

// ErrFetchNewPostedBlocksFail error: fetch new posted blocks fail.
var ErrFetchNewPostedBlocksFail = errors.New("fetch new posted blocks fail")

// ErrFetchScrollCalldataByHashFail error: failed to fetch scroll calldate by hash.
var ErrFetchScrollCalldataByHashFail = errors.New("failed to fetch scroll calldate by hash")

// ErrFetchIntMaxBlockContentByCalldataFail error: failed to fetch IntMax block content by calldate.
var ErrFetchIntMaxBlockContentByCalldataFail = errors.New("failed to fetch IntMax block content by calldate")

// ErrProcessingBlocksFail error: failed to processing blocks.
var ErrProcessingBlocksFail = errors.New("failed to processing blocks")

// ErrSenderByAddressFail error: failed to get sender by address.
var ErrSenderByAddressFail = errors.New("failed to get sender by address")

// ErrCreateSendersFail error: failed to create sender.
var ErrCreateSendersFail = errors.New("failed to create sender")

// ErrCreateAccountFail error: failed to create account.
var ErrCreateAccountFail = errors.New("failed to create account")

// ErrAccountBySenderIDFail error: failed to get account by sender ID.
var ErrAccountBySenderIDFail = errors.New("failed to get account by sender ID")

// ErrDecodeHexToPublicKeyFail error: failed to decode hex to public key.
var ErrDecodeHexToPublicKeyFail = errors.New("failed to decode hex to public key")

// ErrCreateBlockContentFail error: failed to create block content.
var ErrCreateBlockContentFail = errors.New("failed to create block content")

// ErrRequestAndFetchBlockValidityProofFail error: request and fetch block validity proof fail.
var ErrRequestAndFetchBlockValidityProofFail = errors.New("request and fetch block validity proof fail")

// ErrGenesisValidityProof error: genesis validity proof error.
var ErrGenesisValidityProof = errors.New("genesis validity proof error")

// ErrNoValidityProofByBlockNumber error: no validity proof by block number.
var ErrNoValidityProofByBlockNumber = errors.New("no validity proof by block number")

// ErrValidityProofByBlockNumberFail error: failed to get validity proof by block number.
var ErrValidityProofByBlockNumberFail = errors.New("failed to get validity proof by block number")

// ErrAccountAlreadyExists error: account already exists.
var ErrAccountAlreadyExists = errors.New("account already exists")

// ErrAccountTreeGetAccountID error: account id not found in the account tree.
var ErrAccountTreeGetAccountID = errors.New("account id not found in the account tree")

// ErrBlockUnSynchronization error: not synchronized.
var ErrBlockUnSynchronization = errors.New("not synchronized")

// ErrAccountMembershipProof error: account membership proof error.
var ErrAccountMembershipProof = errors.New("account membership proof error")

// ErrRootBlockNumberNotFound error: root block number not found.
var ErrRootBlockNumberNotFound = errors.New("root block number not found")

// ErrRootBlockNumberLessThenLeafBlockNumber error: root block number should be greater than or equal to leaf block number.
var ErrRootBlockNumberLessThenLeafBlockNumber = errors.New(
	"root block number should be greater than or equal to leaf block number",
)

// ErrLeafBlockNumberNotFound error: leaf block number not found.
var ErrLeafBlockNumberNotFound = errors.New("leaf block number not found")

// ErrBlockContentByBlockNumber error: block content by block number error
var ErrBlockContentByBlockNumber = errors.New("failed to fetch block content by block number")

// ErrCurrentBlockNumberNotFound error: current block number not found.
var ErrCurrentBlockNumberNotFound = errors.New("current block number not found")

// ErrBlockNumberInvalid error: block number must be valid.
var ErrBlockNumberInvalid = errors.New("block number must be valid")

// ErrBlockNumberOutOfRange error: block number is out of range.
var ErrBlockNumberOutOfRange = errors.New("block number is out of range")

// ErrDepositTreeProof error: deposit tree proof error.
var ErrDepositTreeProof = errors.New("deposit tree proof error")

// ErrBlockTreeProofFail error: failed to get block tree proof.
var ErrBlockTreeProofFail = errors.New("failed to get block tree proof")

// ErrValidityPublicInputsByBlockNumberFail error: failed to get validity public inputs by block number.
var ErrValidityPublicInputsByBlockNumberFail = errors.New("failed to get validity public inputs by block number")

// ErrDelAllAccountsFail error: failed to delete all accounts.
var ErrDelAllAccountsFail = errors.New("failed to delete all accounts")

// ErrResetSequenceByAccountsFail error: failed to reset sequence by accounts.
var ErrResetSequenceByAccountsFail = errors.New("failed to reset sequence by accounts")
