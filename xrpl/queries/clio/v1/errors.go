package v1

import "errors"

var (
	// ErrNoNFTokenID is returned when no NFToken ID is specified in a request.
	ErrNoNFTokenID = errors.New("no NFToken ID specified")
)
