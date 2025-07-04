package model

import "errors"

// Ошибки для бизнес-логики.
var (
	ErrLoginTaken           = errors.New("login already taken")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidOrderFormat   = errors.New("invalid order format")
	ErrOrderAlreadyUploaded = errors.New("order already uploaded")
	ErrOrderTakenByAnother  = errors.New("order taken by another user")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrOrderNotFound        = errors.New("order not found")
	ErrTooManyRequests      = errors.New("too many requests")
)
