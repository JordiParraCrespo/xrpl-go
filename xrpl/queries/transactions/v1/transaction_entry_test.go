package v1

import (
	"testing"
)

func TestEntryRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request EntryRequest
		wantErr error
	}{
		{
			name:    "pass - valid request",
			request: EntryRequest{TxHash: "E08D6E9754025BA2534A78707605E0601F03ACE063687A0CA1BDDACFCD1698C7"},
			wantErr: nil,
		},
		{
			name:    "fail - missing tx_hash",
			request: EntryRequest{},
			wantErr: ErrNoTxHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.request.Validate(); err != tt.wantErr {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
