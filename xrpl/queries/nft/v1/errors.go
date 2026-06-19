package v1

import "errors"

var (
	// ErrNoNFTokenID is returned when no nft_id is specified in a request.
	ErrNoNFTokenID = errors.New("no nft_id specified")
)
