package utility

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestRandomResponse(t *testing.T) {
	s := RandomResponse{
		Random: "8ED765AEBBD6767603C2C9375B2679AEC76E6A8133EF59F04F9FC1AAA70E41AF",
	}

	j := `{
	"random": "8ED765AEBBD6767603C2C9375B2679AEC76E6A8133EF59F04F9FC1AAA70E41AF"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestRandomRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request RandomRequest
		wantErr error
	}{
		{
			// RandomRequest has no required fields in xrpl.js.
			name:    "pass - valid request",
			request: RandomRequest{},
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
