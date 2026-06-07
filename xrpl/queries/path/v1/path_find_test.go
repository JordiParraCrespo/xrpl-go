package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestPathFindCloseRequest(t *testing.T) {
	s := FindCloseRequest{
		Subcommand: Close,
	}

	j := `{
	"subcommand": "close"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestPathFindCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request FindCreateRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: FindCreateRequest{
				Subcommand:         Create,
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			},
			wantErr: nil,
		},
		{
			name: "fail - missing subcommand",
			request: FindCreateRequest{
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			},
			wantErr: ErrNoSubcommand,
		},
		{
			name: "fail - invalid subcommand",
			request: FindCreateRequest{
				Subcommand:         Close,
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			},
			wantErr: ErrInvalidSubcommand,
		},
		{
			name: "fail - missing source_account",
			request: FindCreateRequest{
				Subcommand:         Create,
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			},
			wantErr: ErrNoSourceAccount,
		},
		{
			name: "fail - missing destination_account",
			request: FindCreateRequest{
				Subcommand:        Create,
				SourceAccount:     "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: types.XRPCurrencyAmount(1000000),
			},
			wantErr: ErrNoDestinationAccount,
		},
		{
			name: "fail - missing destination_amount",
			request: FindCreateRequest{
				Subcommand:         Create,
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
			},
			wantErr: ErrNoDestinationAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathFindCloseRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request FindCloseRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: FindCloseRequest{
				Subcommand: Close,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing subcommand",
			request: FindCloseRequest{},
			wantErr: ErrNoSubcommand,
		},
		{
			name: "fail - invalid subcommand",
			request: FindCloseRequest{
				Subcommand: Create,
			},
			wantErr: ErrInvalidSubcommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathFindStatusRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request FindStatusRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: FindStatusRequest{
				Subcommand: Status,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing subcommand",
			request: FindStatusRequest{},
			wantErr: ErrNoSubcommand,
		},
		{
			name: "fail - invalid subcommand",
			request: FindStatusRequest{
				Subcommand: Create,
			},
			wantErr: ErrInvalidSubcommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathFindStatusRequest(t *testing.T) {
	s := FindStatusRequest{
		Subcommand: Status,
	}

	j := `{
	"subcommand": "status"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
