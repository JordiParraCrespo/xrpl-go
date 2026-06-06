package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountCurrenciesRequest(t *testing.T) {
	s := CurrenciesRequest{
		Account:     "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		Strict:      true,
		LedgerIndex: common.LedgerIndex(1234),
	}

	j := `{
	"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"ledger_index": 1234,
	"strict": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestCurrenciesRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CurrenciesRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: CurrenciesRequest{
				Account:     "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				LedgerIndex: common.Validated,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing account",
			request: CurrenciesRequest{},
			wantErr: ErrNoAccountID,
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

func TestAccountCurrenciesResponse(t *testing.T) {
	s := CurrenciesResponse{
		LedgerHash:  "abc",
		LedgerIndex: 123,
		ReceiveCurrencies: []string{
			"USD",
			"JPY",
		},
		SendCurrencies: []string{
			"USD",
			"CAD",
		},
		Validated: true,
	}
	j := `{
	"ledger_hash": "abc",
	"ledger_index": 123,
	"receive_currencies": [
		"USD",
		"JPY"
	],
	"send_currencies": [
		"USD",
		"CAD"
	],
	"validated": true
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
