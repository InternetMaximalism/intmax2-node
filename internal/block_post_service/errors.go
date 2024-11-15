package block_post_service

import "errors"

// ErrInvalidSendersLength error: invalid senders length.
var ErrInvalidSendersLength = errors.New("invalid senders length")

// ErrTooManySenderPublicKeys error: too many sender public keys.
var ErrTooManySenderPublicKeys = errors.New("too many sender public keys")

// ErrUnmarshalSignatureFail error: failed to unmarshal signature.
var ErrUnmarshalSignatureFail = errors.New("failed to unmarshal signature")

// ErrInvalidSignature error: fail to verify signature.
var ErrInvalidSignature = errors.New("fail to verify signature")

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

// ErrInitFail error: failed to init.
var ErrInitFail = errors.New("failed to init")

// ErrRegisterPublicKeyFail error: failed to register public key.
var ErrRegisterPublicKeyFail = errors.New("failed to register public key")
