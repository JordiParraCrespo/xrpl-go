package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestLedgerDataRequest(t *testing.T) {
	s := DataRequest{
		LedgerIndex: common.Closed,
		Binary:      true,
		Limit:       5,
	}
	j := `{
	"ledger_index": "closed",
	"binary": true,
	"limit": 5
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

// LedgerDataRequest has no required fields.
func TestLedgerDataRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request DataRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: DataRequest{
				LedgerIndex: common.Validated,
			},
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
