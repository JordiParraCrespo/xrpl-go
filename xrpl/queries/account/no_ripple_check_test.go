package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestNoRippleCheckRequest(t *testing.T) {
	s := NoRippleCheckRequest{
		Account:      types.Address("r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH"),
		Role:         "gateway",
		Transactions: true,
		Limit:        10,
		LedgerHash:   common.LedgerHash("0000000000000000000000000000000000000000000000000000000000000000"),
		LedgerIndex:  common.Validated,
	}

	j := `{
	"account": "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
	"role": "gateway",
	"transactions": true,
	"limit": 10,
	"ledger_hash": "0000000000000000000000000000000000000000000000000000000000000000",
	"ledger_index": "validated"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNoRippleCheckRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request NoRippleCheckRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: NoRippleCheckRequest{
				Account: "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
				Role:    "gateway",
			},
			wantErr: nil,
		},
		{
			name: "pass - valid request with user role",
			request: NoRippleCheckRequest{
				Account: "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
				Role:    "user",
			},
			wantErr: nil,
		},
		{
			name: "fail - missing account",
			request: NoRippleCheckRequest{
				Role: "gateway",
			},
			wantErr: ErrNoAccountID,
		},
		{
			name: "fail - missing role",
			request: NoRippleCheckRequest{
				Account: "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
			},
			wantErr: ErrNoRole,
		},
		{
			name: "fail - invalid role",
			request: NoRippleCheckRequest{
				Account: "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
				Role:    "issuer",
			},
			wantErr: ErrInvalidRole,
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

func TestNoRippleCheckResponse(t *testing.T) {
	s := NoRippleCheckResponse{
		LedgerCurrentIndex: common.LedgerIndex(1234567890),
		Problems:           []string{"problem1", "problem2"},
		Transactions: []transaction.FlatTransaction{
			{
				"Account":         "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
				"Fee":             "1000000000",
				"Sequence":        1,
				"TransactionType": "Payment",
				"Amount":          "1000000000",
			},
		},
	}

	j := `{
	"ledger_current_index": 1234567890,
	"problems": [
		"problem1",
		"problem2"
	],
	"transactions": [
		{
			"Account": "r9cZA1mLK5R5Am25ArfXF7tRp1PeperEvH",
			"Amount": "1000000000",
			"Fee": "1000000000",
			"Sequence": 1,
			"TransactionType": "Payment"
		}
	]
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
