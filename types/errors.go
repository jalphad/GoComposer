package types

import "errors"

var (
	ErrInvalidArgument InvalidArgument = errors.New("invalid argument")
	ErrCompose         ComposeError    = errors.New("compose error")
)

type InvalidArgument error
type ComposeError error
