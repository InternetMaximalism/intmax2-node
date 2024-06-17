package http_response_modifier

import "errors"

// ErrCookieDomainInvalid error: cookie domain must be valid.
var ErrCookieDomainInvalid = errors.New(
	"cookie domain must be valid",
)

// ErrValueInvalid error: value must be valid
var ErrValueInvalid = errors.New("value must be valid")
