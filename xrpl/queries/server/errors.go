package server

import "errors"

var (
	// ErrNoFeature is returned when no feature is specified in a FeatureOneRequest.
	ErrNoFeature = errors.New("no feature specified")
	// ErrNoPublicKey is returned when no public key is specified in a ManifestRequest.
	ErrNoPublicKey = errors.New("no public key specified")
)
