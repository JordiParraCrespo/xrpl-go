package nft

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestNFTokenSellOffersRequest(t *testing.T) {
	s := NFTokenSellOffersRequest{
		NFTokenID:   "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
		LedgerIndex: common.Validated,
	}

	j := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"ledger_index": "validated"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTokenSellOffersRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request NFTokenSellOffersRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: NFTokenSellOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing nft_id",
			request: NFTokenSellOffersRequest{},
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

func TestNFTokenSellOffersResponse(t *testing.T) {
	s := NFTokenSellOffersResponse{
		NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
		Offers: []nfttypes.NFTokenOffer{
			{
				Amount:            types.XRPCurrencyAmount(1000),
				Flags:             1,
				NFTokenOfferIndex: "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1",
				Owner:             "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt",
			},
		},
	}

	j := `{
	"nft_id": "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
	"offers": [
		{
			"amount": "1000",
			"flags": 1,
			"nft_offer_index": "9E28E366573187F8E5B85CE301F229E061A619EE5A589EF740088F8843BF10A1",
			"owner": "rLpSRZ1E8JHyNDZeHYsQs1R5cwDCB3uuZt"
		}
	]
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
