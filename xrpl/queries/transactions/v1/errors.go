package v1

import "errors"

var (
	// ErrNoTxBlob is returned when no TxBlob is defined in the SubmitRequest.
	ErrNoTxBlob = errors.New("no TxBlob defined")
	// ErrNoTxJSON is returned when no tx_json is defined in the
	// SubmitMultisignedRequest.
	ErrNoTxJSON = errors.New("no tx_json defined")
)
