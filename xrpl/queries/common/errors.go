package common

import "errors"

// ErrInvalidLedgerSpecifier is returned when a ledger specifier string is not current, validated, or closed.
var ErrInvalidLedgerSpecifier = errors.New("decoding LedgerTitle: invalid string")
