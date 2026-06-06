package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountTransactionsRequest(t *testing.T) {
	s := TransactionsRequest{
		Account:        "abc",
		LedgerIndexMin: 100,
		LedgerIndexMax: 120,
		LedgerHash:     "def",
		LedgerIndex:    common.LedgerIndex(10),
		Marker:         "123",
	}

	j := `{
	"account": "abc",
	"ledger_index_min": 100,
	"ledger_index_max": 120,
	"ledger_hash": "def",
	"ledger_index": 10,
	"marker": "123"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestTransactionsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request TransactionsRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: TransactionsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing account",
			request: TransactionsRequest{},
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
