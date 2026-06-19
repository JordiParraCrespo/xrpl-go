package oracle

import "errors"

var (
	// ErrNoBaseAsset is returned when no base_asset is specified in a
	// GetAggregatePriceRequest.
	ErrNoBaseAsset = errors.New("no base_asset specified")
	// ErrNoQuoteAsset is returned when no quote_asset is specified in a
	// GetAggregatePriceRequest.
	ErrNoQuoteAsset = errors.New("no quote_asset specified")
	// ErrNoOracles is returned when no oracles are specified in a
	// GetAggregatePriceRequest.
	ErrNoOracles = errors.New("no oracles specified")
)
