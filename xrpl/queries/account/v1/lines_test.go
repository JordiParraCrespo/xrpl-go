package v1

import (
	"testing"

	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountLinesRequest(t *testing.T) {
	s := LinesRequest{
		Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Peer:        "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
		LedgerHash:  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
		LedgerIndex: common.LedgerIndex(256),
		Limit:       10,
		Marker:      map[string]interface{}{"abc": "def"},
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"ledger_hash": "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
	"ledger_index": 256,
	"peer": "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
	"limit": 10,
	"marker": {
		"abc": "def"
	}
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestLinesRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request LinesRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: LinesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing account",
			request: LinesRequest{},
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

func TestAccountLinesResponse(t *testing.T) {
	s := LinesResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Lines: []accounttypes.TrustLine{
			{
				Account:    "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Balance:    "123",
				Currency:   "USD",
				Limit:      "456",
				LimitPeer:  "10",
				QualityIn:  1,
				QualityOut: 2,
			},
		},
		LedgerCurrentIndex: 123,
		LedgerIndex:        345,
		LedgerHash:         "abc",
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"lines": [
		{
			"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"balance": "123",
			"currency": "USD",
			"limit": "456",
			"limit_peer": "10",
			"quality_in": 1,
			"quality_out": 2
		}
	],
	"ledger_current_index": 123,
	"ledger_index": 345,
	"ledger_hash": "abc"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
