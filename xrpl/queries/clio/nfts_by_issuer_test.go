package clio

import (
	"testing"

	cliotypes "github.com/Peersyst/xrpl-go/xrpl/queries/clio/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestNFTsByIssuerRequest(t *testing.T) {
	s := NFTsByIssuerRequest{
		Issuer:   "abc",
		Marker:   "123",
		Limit:    10,
		NftTaxon: 1,
	}

	j := `{
	"issuer": "abc",
	"marker": "123",
	"limit": 10,
	"nft_taxon": 1
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestNFTsByIssuerRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request NFTsByIssuerRequest
		wantErr error
	}{
		{
			name: "pass - valid request",
			request: NFTsByIssuerRequest{
				Issuer: "rHVokeuSnjPjz718qdb47bGXBBHNMP3KDQ",
			},
			wantErr: nil,
		},
		{
			name:    "fail - missing issuer",
			request: NFTsByIssuerRequest{},
			wantErr: ErrNoIssuer,
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

func TestNFTsByIssuerResponse(t *testing.T) {
	s := NFTsByIssuerResponse{
		Issuer: "abc",
		NFTs: []cliotypes.NFToken{
			{
				NFTokenID:       "123",
				LedgerIndex:     1,
				Owner:           "abc",
				IsBurned:        false,
				Flags:           0,
				TransferFee:     0,
				Issuer:          "abc",
				NFTokenTaxon:    1,
				NFTokenSequence: 1,
				URI:             "abc",
			},
		},
		Marker:       "123",
		Limit:        10,
		NFTokenTaxon: 1,
	}

	j := `{
	"issuer": "abc",
	"nfts": [
		{
			"nft_id": "123",
			"ledger_index": 1,
			"owner": "abc",
			"is_burned": false,
			"flags": 0,
			"transfer_fee": 0,
			"issuer": "abc",
			"nft_taxon": 1,
			"nft_sequence": 1,
			"uri": "abc"
		}
	],
	"marker": "123",
	"limit": 10,
	"nft_taxon": 1
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
