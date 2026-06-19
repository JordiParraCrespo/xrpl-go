package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestNFTHistoryRequest(t *testing.T) {
	s := NFTHistoryRequest{
		NFTokenID:      "0000000000000000000000000000000000000000000000000000000000000000",
		LedgerIndexMin: 100,
		LedgerIndexMax: 200,
		Binary:         true,
		Forward:        true,
		Limit:          100,
		Marker:         "marker",
	}

	j := `{
	"nft_id": "0000000000000000000000000000000000000000000000000000000000000000",
	"ledger_index_min": 100,
	"ledger_index_max": 200,
	"binary": true,
	"forward": true,
	"limit": 100,
	"marker": "marker"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Fatal(err)
	}
}

func TestNFTHistoryRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request NFTHistoryRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: NFTHistoryRequest{
				NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing nft_id",
			request: NFTHistoryRequest{},
			wantErr: ErrNoNFTokenID,
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
