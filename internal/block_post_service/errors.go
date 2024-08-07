package block_post_service

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrTransactionByHashNotFound error: failed to get transaction by hash.
var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

// ErrTransactionIsStillPending error: transaction is still pending.
var ErrTransactionIsStillPending = errors.New("transaction is still pending")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrNewEthereumClientFail error: failed to create new Ethereum client.
var ErrNewEthereumClientFail = errors.New("failed to create new Ethereum client")

// ErrNewScrollClientFail error: failed to create new Scroll client.
var ErrNewScrollClientFail = errors.New("failed to create new Scroll client")

// ErrInstantiateLiquidityContractFail error: failed to instantiate a Liquidity contract.
var ErrInstantiateLiquidityContractFail = errors.New("failed to instantiate a Liquidity contract")

// ErrInstantiateRollupContractFail error: failed to instantiate a Rollup contract.
var ErrInstantiateRollupContractFail = errors.New("failed to instantiate a Rollup contract")

// ErrFilterLogsFail error: failed to filter logs.
var ErrFilterLogsFail = errors.New("failed to filter logs")

// ErrEncounteredWhileIterating error: encountered while iterating error occurred.
var ErrEncounteredWhileIterating = errors.New("encountered while iterating error occurred")

// ErrDecodeCallDataFail error: failed to decode calldata.
var ErrDecodeCallDataFail = errors.New("failed to decode calldata")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

// ErrSetTxRootFail error: failed to set tx tree root.
var ErrSetTxRootFail = errors.New("failed to set tx tree root")

// ErrRecoverAccountIDsFromBytesFail error: failed to recover account IDs from bytes.
var ErrRecoverAccountIDsFromBytesFail = errors.New("failed to recover account IDs from bytes")

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

// ErrFetchNewPostedBlocksFail error: failed to fetch new posted block.
var ErrFetchNewPostedBlocksFail = errors.New("failed to fetch new posted block")

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

// ErrRegisterPublicKeyFail error: failed to register public key.
var ErrRegisterPublicKeyFail = errors.New("failed to register public key")

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
