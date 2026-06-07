package v1

import (
	"testing"
)

// TxRequest models xrpl.js TxRequest. Per jsspec.json (and the .ts interface),
// every field is optional: `transaction`, `ctid`, `binary`, `min_ledger`,
// `max_ledger`. xrpl.js documents in prose that "exactly one of transaction or
// ctid must be specified", but neither is marked required in the interface, and
// the Go struct does not even model `ctid`. Per the parity-not-creativity rule
// we encode no hard-required field, so only the valid path is exercised.
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
