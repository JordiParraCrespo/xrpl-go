package v1

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

func TestSubmitMultisignedRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request SubmitMultisignedRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: SubmitMultisignedRequest{
				Tx: transaction.FlatTransaction{
					"TransactionType": "Payment",
					"Account":         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				},
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing tx_json",
			request: SubmitMultisignedRequest{},
			wantErr: ErrNoTxJSON,
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
