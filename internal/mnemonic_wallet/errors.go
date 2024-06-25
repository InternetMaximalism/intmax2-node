package mnemonic_wallet

import "errors"

// ErrMnemonicRequired error: mnemonic is required.
var ErrMnemonicRequired = errors.New("mnemonic is required")

// ErrMnemonicInvalid error: mnemonic is invalid.
var ErrMnemonicInvalid = errors.New("mnemonic is invalid")

// ErrNewSeedFail error: failed to create new seed.
var ErrNewSeedFail = errors.New("failed to create new seed")

// ErrNewMasterFromSeedFail error: failed to create new master from seed.
var ErrNewMasterFromSeedFail = errors.New("failed to create new master from seed")

// ErrParseDerivationPathFail error: failed to parse derivation path.
var ErrParseDerivationPathFail = errors.New("failed to parse derivation path")

// ErrHDPathDeriveFail error: failed to derive HD path.
var ErrHDPathDeriveFail = errors.New("failed to derive HD path")

// ErrECPrivateKeyFail error: failed to get EC private key.
var ErrECPrivateKeyFail = errors.New("failed to get EC private key")

// ErrPrivateKeyToHexFail error: failed to convert private key to HEX.
var ErrPrivateKeyToHexFail = errors.New("failed to convert private key to HEX")

// ErrPrivateKeyHexInvalid error: the HEX private key must be valid.
var ErrPrivateKeyHexInvalid = errors.New("the HEX private key must be valid")

// ErrHexEncodeFail error: failed to encode to HEX.
var ErrHexEncodeFail = errors.New("failed to encode to HEX")

// ErrHexDecodeFail error: failed to decode to HEX.
var ErrHexDecodeFail = errors.New("failed to decode to HEX")

// ErrConvertPrivateKeyToECDSAFail error: failed to convert private key to ECDSA.
var ErrConvertPrivateKeyToECDSAFail = errors.New("failed to convert private key to ECDSA")

// ErrDecodePrivateKeyFromHEXToECDSAFail error: failed to decode private key from HEX to ECDSA.
var ErrDecodePrivateKeyFromHEXToECDSAFail = errors.New("failed to decode private key from HEX to ECDSA")

// ErrProcessingPrivateKeyToHexFail error: failed to processing private key to HEX.
var ErrProcessingPrivateKeyToHexFail = errors.New("failed to processing private key to HEX")

// ErrPublicKeyFail error: failed to get public key.
var ErrPublicKeyFail = errors.New("failed to get public key")

// ErrNewMnemonicFail error: failed to create new mnemonic.
var ErrNewMnemonicFail = errors.New("failed to create new mnemonic")

// ErrWalletFromMnemonicFail error: failed to recover wallet from mnemonic.
var ErrWalletFromMnemonicFail = errors.New("failed to recover wallet from mnemonic")

// ErrNewINTMAXAccountFromECDSAKeyFail error: failed to create new INTMAX account from ECDSA key.
var ErrNewINTMAXAccountFromECDSAKeyFail = errors.New("failed to create new INTMAX account from ECDSA key")
