package v1

import (
	"testing"
)

func TestPingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request PingRequest
		wantErr error
	}{
		{
			// PingRequest has no required fields in xrpl.js.
			name:    "pass - valid request",
			request: PingRequest{},
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
