package block_validity_prover

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")

// ErrDecodeCallDataFail error: failed to decode calldata.
var ErrDecodeCallDataFail = errors.New("failed to decode calldata")

// ErrRecoverAccountIDsFromBytesFail error: failed to recover account IDs from bytes.
var ErrRecoverAccountIDsFromBytesFail = errors.New("failed to recover account IDs from bytes")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

// ErrSetTxRootFail error: failed to set tx tree root.
var ErrSetTxRootFail = errors.New("failed to set tx tree root")

var ErrNewEthereumClientFail = errors.New("failed to create new Ethereum client")

var ErrScrollNetwrokChainLink = errors.New("failed to get Scroll network chain link")

var ErrNewScrollClientFail = errors.New("failed to create new Scroll client")

var ErrInstantiateLiquidityContractFail = errors.New("failed to instantiate a Liquidity contract")

var ErrInstantiateRollupContractFail = errors.New("failed to instantiate a Rollup contract")

// ErrTransactionByHashNotFound error: failed to get transaction by hash.
var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

// ErrTransactionIsStillPending error: transaction is still pending.
var ErrTransactionIsStillPending = errors.New("transaction is still pending")

// ErrFilterLogsFail error: failed to filter logs.
var ErrFilterLogsFail = errors.New("failed to filter logs")

// ErrEncounteredWhileIterating error: encountered while iterating error occurred.
var ErrEncounteredWhileIterating = errors.New("encountered while iterating error occurred")

// ErrFetchLatestBlockNumberFail error: failed to fetch latest block number.
var ErrFetchLatestBlockNumberFail = errors.New("failed to fetch latest block number")

// ErrInvalidSendersLength error: invalid senders length.
var ErrInvalidSendersLength = errors.New("invalid senders length")

// ErrTooManySenderPublicKeys error: too many sender public keys.
var ErrTooManySenderPublicKeys = errors.New("too many sender public keys")

// ErrUnmarshalSignatureFail error: failed to unmarshal signature.
var ErrUnmarshalSignatureFail = errors.New("failed to unmarshal signature")

// ErrInvalidSignature error: fail to verify signature.
var ErrInvalidSignature = errors.New("fail to verify signature")

// ErrNewBlockPostServiceFail error: failed to create new block post service.
var ErrNewBlockPostServiceFail = errors.New("failed to create new block post service")

// ErrFetchNewPostedBlocksFail error: fetch new posted blocks fail.
var ErrFetchNewPostedBlocksFail = errors.New("fetch new posted blocks fail")

// ErrFetchScrollCalldataByHashFail error: failed to fetch scroll calldate by hash.
var ErrFetchScrollCalldataByHashFail = errors.New("failed to fetch scroll calldate by hash")

// ErrFetchIntMaxBlockContentByCalldataFail error: failed to fetch IntMax block content by calldate.
var ErrFetchIntMaxBlockContentByCalldataFail = errors.New("failed to fetch IntMax block content by calldate")

// ErrProcessingBlocksFail error: failed to processing blocks.
var ErrProcessingBlocksFail = errors.New("failed to processing blocks")

// ErrNewCtrlEventBlockNumbersJobsFail error: failed to create new ctrl-event-block-numbers row.
var ErrNewCtrlEventBlockNumbersJobsFail = errors.New("failed to create new ctrl-event-block-numbers row")

// ErrCtrlEventBlockNumbersJobsFail error: failed to get info of ctrl-event-block-numbers row.
var ErrCtrlEventBlockNumbersJobsFail = errors.New("failed to get info of ctrl-event-block-numbers row")

// ErrSenderByAddressFail error: failed to get sender by address.
var ErrSenderByAddressFail = errors.New("failed to get sender by address")

// ErrCreateSendersFail error: failed to create sender.
var ErrCreateSendersFail = errors.New("failed to create sender")

// ErrCreateAccountFail error: failed to create account.
var ErrCreateAccountFail = errors.New("failed to create account")

// ErrAccountBySenderIDFail error: failed to get account by sender ID.
var ErrAccountBySenderIDFail = errors.New("failed to get account by sender ID")

// ErrAccountByAccountIDFail error: failed to get account by account ID.
var ErrAccountByAccountIDFail = errors.New("failed to get account by account ID")

// ErrSenderByIDFail error: failed to get sender by ID.
var ErrSenderByIDFail = errors.New("failed to get sender by ID")

// ErrDecodeHexToPublicKeyFail error: failed to decode hex to public key.
var ErrDecodeHexToPublicKeyFail = errors.New("failed to decode hex to public key")

// ErrDelAllAccountsFail error: failed to delete all accounts.
var ErrDelAllAccountsFail = errors.New("failed to delete all accounts")

// ErrResetSequenceByAccountsFail error: failed to reset sequence by accounts.
var ErrResetSequenceByAccountsFail = errors.New("failed to reset sequence by accounts")

// ErrInvalidRegistrationBlockContent error: invalid registration block content.
var ErrInvalidRegistrationBlockContent = errors.New("invalid registration block content")

// ErrInvalidNonRegistrationBlockContent error: invalid non-registration block content.
var ErrInvalidNonRegistrationBlockContent = errors.New("invalid non-registration block content")

// ErrCreateBlockContentFail error: failed to create block content.
var ErrCreateBlockContentFail = errors.New("failed to create block content")

// ErrRequestAndFetchBlockValidityProofFail error: request and fetch block validity proof fail.
var ErrRequestAndFetchBlockValidityProofFail = errors.New("request and fetch block validity proof fail")

// ErrLastSeenBlockPostedEventBlockNumberFail error: last seen block posted event block number fail.
var ErrLastSeenBlockPostedEventBlockNumberFail = errors.New("last seen block posted event block number fail")

// ErrNoLastValidityProof error: no last validity proof.
var ErrNoLastValidityProof = errors.New("no last validity proof")

// ErrGenesisValidityProof error: genesis validity proof error.
var ErrGenesisValidityProof = errors.New("genesis validity proof error")

// ErrNoValidityProofByBlockNumber error: no validity proof by block number.
var ErrNoValidityProofByBlockNumber = errors.New("no validity proof by block number")

// ErrAccountAlreadyExists error: account already exists.
var ErrAccountAlreadyExists = errors.New("account already exists")

// ErrAccountTreeGetAccountID error: account id not found in the account tree.
var ErrAccountTreeGetAccountID = errors.New("account id not found in the account tree")

// ErrBlockShouldBeValid error: the block should be valid if it is not an empty block.
var ErrBlockShouldBeValid = errors.New("the block should be valid if it is not an empty block")

// ErrBlockUnSynchronization error: not synchronized.
var ErrBlockUnSynchronization = errors.New("not synchronized")

// ErrAccountMembershipProof error: account membership proof error.
var ErrAccountMembershipProof = errors.New("account membership proof error")

// ErrBlockTreeProve error: block tree prove error.
var ErrBlockTreeProve = errors.New("block tree prove error")

// ErrRootBlockNumberNotFound error: root block number not found.
var ErrRootBlockNumberNotFound = errors.New("root block number not found")

// ErrLeafBlockNumberNotFound error: leaf block number not found.
var ErrLeafBlockNumberNotFound = errors.New("leaf block number not found")

// ErrBlockContentByBlockNumber error: block content by block number error
var ErrBlockContentByBlockNumber = errors.New("failed to fetch block content by block number")

// ErrDepositIndexNotFound error: no last deposit index found.
var ErrDepositIndexNotFound = errors.New("no last deposit index found")
