package domain

import "errors"

var (
	ErrInvalidURL        = errors.New("invalid url")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrNotFound          = errors.New("url not found")
	ErrAlreadyExists     = errors.New("url already exists")
	ErrInvalidURLStatus  = errors.New("invalid url status")
)
