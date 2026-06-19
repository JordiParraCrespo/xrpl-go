package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountOffersRequest(t *testing.T) {
	s := OffersRequest{
		Account:     "abc",
		LedgerIndex: common.LedgerIndex(10),
		Marker:      "123",
	}
	j := `{
	"account": "abc",
	"ledger_index": 10,
	"marker": "123"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}

}

func TestOffersRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request OffersRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: OffersRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing account",
			request: OffersRequest{},
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
