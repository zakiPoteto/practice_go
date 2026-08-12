package service

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
