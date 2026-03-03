package domain

import "errors"

var (
	ErrInvalidURL        = errors.New("invalid url")
	ErrInvalidTransition = errors.New("invalid status transition")
)
