package transactions

import (
	"testing"
)

// TxRequest has no required fields: transaction, ctid, binary, min_ledger and
// max_ledger are all optional. Exactly one of transaction or ctid is expected
// in practice, but neither is mandatory and the Go struct does not model ctid,
// so Validate enforces nothing and only the valid path is exercised.
func TestTxRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request TxRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: TxRequest{
				Transaction: "E08D6E9754025BA2534A78707605E0601F03ACE063687A0CA1BDDACFCD1698C7",
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
