package channel

import "errors"

var (
	// ErrNoAmount is returned when no amount is specified in a channel_verify request.
	ErrNoAmount = errors.New("no amount specified")
	// ErrNoChannelID is returned when no channel_id is specified in a channel_verify request.
	ErrNoChannelID = errors.New("no channel_id specified")
	// ErrNoPublicKey is returned when no public_key is specified in a channel_verify request.
	ErrNoPublicKey = errors.New("no public_key specified")
	// ErrNoSignature is returned when no signature is specified in a channel_verify request.
	ErrNoSignature = errors.New("no signature specified")
)
