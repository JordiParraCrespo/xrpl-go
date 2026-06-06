package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

// Ledger Current request has no required fields.

func TestLedgerCurrentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CurrentRequest
		wantErr error
	}{
		{
			name:    "pass - valid request",
			request: CurrentRequest{},
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

func TestLedgerCurrentResponse(t *testing.T) {
	s := CurrentResponse{
		LedgerCurrentIndex: 123,
	}
	j := `{
	"ledger_current_index": 123
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
