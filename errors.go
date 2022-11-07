package logng

import (
	"errors"
)

var (
	ErrInvalidSeverity = errors.New("invalid severity")
	ErrUnknownSeverity = errors.New("unknown severity")
)
