package clio

import "errors"

var (
	// ErrNoNFTokenID is returned when no NFToken ID is specified in a request.
	ErrNoNFTokenID = errors.New("no NFToken ID specified")
	// ErrNoIssuer is returned when no issuer is specified in a request.
	ErrNoIssuer = errors.New("no issuer specified")
)
