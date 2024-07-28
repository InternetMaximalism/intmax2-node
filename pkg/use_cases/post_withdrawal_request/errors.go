package post_withdrawal_request

import "errors"

// ErrUCInputEmpty error: uc-input must not be empty.
var ErrUCInputEmpty = errors.New("uc-input must not be empty")
