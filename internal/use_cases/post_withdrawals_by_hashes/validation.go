package post_withdrawals_by_hashes

import (
	"errors"
	"regexp"

	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

const (
	Base10        = 10
	MaxHashLength = 66 // 0x prefix + 64 hex characters
	MaxHashCount  = 10
	HexPattern    = `^0x[0-9a-fA-F]{64}$`
)

func (input *UCPostWithdrawalsByHashesInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TransferHashes, validation.Required, validation.By(validateTransferHashes)),
	)
}

func validateTransferHashes(value interface{}) error {
	transferHashes, ok := value.([]string)
	if !ok {
		return ErrValueInvalid
	}

	if len(transferHashes) == 0 {
		return errors.New("transfer_hashes must not be empty")
	}

	if len(transferHashes) > MaxHashCount {
		return errors.New("transfer_hashes must not contain more than 10 hashes")
	}

	hexPattern := regexp.MustCompile(HexPattern)
	for _, hash := range transferHashes {
		if len(hash) > MaxHashLength {
			return errors.New("each transfer_hash must not exceed 66 characters")
		}
		if !hexPattern.MatchString(hash) {
			return errors.New("each transfer_hash must be a valid 64-character hex string prefixed with '0x'")
		}
	}
	return nil
}
