package account

import "errors"

var (
	// ErrNoAccountID is returned when no account ID is specified in a request.
	ErrNoAccountID = errors.New("no account ID specified")
	// ErrNoRole is returned when no role is specified in a noripple_check request.
	ErrNoRole = errors.New("no role specified")
	// ErrInvalidRole is returned when the role specified in a noripple_check
	// request is not one of the allowed values ("gateway" or "user").
	ErrInvalidRole = errors.New("invalid role specified: must be 'gateway' or 'user'")
)
