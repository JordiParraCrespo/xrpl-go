package channel

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestChannelVerifyRequest(t *testing.T) {
	s := VerifyRequest{
		Amount:    types.XRPCurrencyAmount(1000000),
		ChannelID: "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
		PublicKey: "aB44YfzW24VDEJQ2UuLPV2PvqcPCSoLnL7y5M1EzhdW4LnK5xMS3",
		Signature: "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064",
	}

	j := `{
	"amount": "1000000",
	"channel_id": "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
	"public_key": "aB44YfzW24VDEJQ2UuLPV2PvqcPCSoLnL7y5M1EzhdW4LnK5xMS3",
	"signature": "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestChannelVerifyRequest_Validate(t *testing.T) {
	const (
		channelID = "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3"
		publicKey = "aB44YfzW24VDEJQ2UuLPV2PvqcPCSoLnL7y5M1EzhdW4LnK5xMS3"
		signature = "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064"
	)
	tests := []struct {
		name    string
		request VerifyRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: channelID,
				PublicKey: publicKey,
				Signature: signature,
			},
			wantErr: nil,
		},
		{
			name: "fail - missing amount",
			request: VerifyRequest{
				ChannelID: channelID,
				PublicKey: publicKey,
				Signature: signature,
			},
			wantErr: ErrNoAmount,
		},
		{
			name: "fail - missing channel_id",
			request: VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				PublicKey: publicKey,
				Signature: signature,
			},
			wantErr: ErrNoChannelID,
		},
		{
			name: "fail - missing public_key",
			request: VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: channelID,
				Signature: signature,
			},
			wantErr: ErrNoPublicKey,
		},
		{
			name: "fail - missing signature",
			request: VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: channelID,
				PublicKey: publicKey,
			},
			wantErr: ErrNoSignature,
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

func TestChannelVerifyResponse(t *testing.T) {
	s := VerifyResponse{
		SignatureVerified: false,
	}

	j := `{
	"signature_verified": false
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}

}
