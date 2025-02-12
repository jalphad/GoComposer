package types

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrCompose         = errors.New("compose error")
)
