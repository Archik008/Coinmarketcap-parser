package app_errors

import "errors"

var (
	ErrNonPositiveValue = errors.New("value must be positive")
	ErrValueOutOfRange  = errors.New("value must be in range [0, 100]")
	ErrInvalidLabel     = errors.New("invalid label")
)
