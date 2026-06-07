package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountObjectsRequest(t *testing.T) {
	s := ObjectsRequest{
		Account:     "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
		Type:        SignerListObject,
		LedgerIndex: common.LedgerIndex(123),
	}

	j := `{
	"account": "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
	"type": "signer_list",
	"ledger_index": 123
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestObjectsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ObjectsRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: ObjectsRequest{
				Account:     "rsuHaTvJh1bDmDoxX9QcKP7HEBSBt4XsHx",
				LedgerIndex: common.Validated,
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing account",
			request: ObjectsRequest{},
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
