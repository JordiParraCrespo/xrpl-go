package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

// Ledger closed request does not have any fields to test

func TestLedgerClosedRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ClosedRequest
		wantErr error
	}{
		{
			name:    "pass - valid request",
			request: ClosedRequest{},
			wantErr: nil,
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

func TestLedgerClosedResponse(t *testing.T) {
	s := ClosedResponse{
		LedgerHash:  "abc",
		LedgerIndex: 123,
	}
	j := `{
	"ledger_hash": "abc",
	"ledger_index": 123
}`
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
