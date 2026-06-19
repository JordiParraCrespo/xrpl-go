package path

import "errors"

var (
	// ErrNoTakerGets is returned when no taker_gets is specified in a book_offers request.
	ErrNoTakerGets = errors.New("no taker_gets specified")
	// ErrNoTakerPays is returned when no taker_pays is specified in a book_offers request.
	ErrNoTakerPays = errors.New("no taker_pays specified")
	// ErrNoSourceAccount is returned when no source_account is specified in a request.
	ErrNoSourceAccount = errors.New("no source_account specified")
	// ErrNoDestinationAccount is returned when no destination_account is specified in a request.
	ErrNoDestinationAccount = errors.New("no destination_account specified")
	// ErrNoDestinationAmount is returned when no destination_amount is specified in a request.
	ErrNoDestinationAmount = errors.New("no destination_amount specified")
	// ErrNoSubcommand is returned when no subcommand is specified in a path_find request.
	ErrNoSubcommand = errors.New("no subcommand specified")
	// ErrInvalidSubcommand is returned when the subcommand specified in a path_find
	// request is not one of the allowed values ("create", "close" or "status").
	ErrInvalidSubcommand = errors.New("invalid subcommand specified: must be 'create', 'close' or 'status'")
)
