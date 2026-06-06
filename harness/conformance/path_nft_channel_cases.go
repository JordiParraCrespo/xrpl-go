package conformance

import (
	channel "github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	channelv1 "github.com/Peersyst/xrpl-go/xrpl/queries/channel/v1"
	nft "github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	nftv1 "github.com/Peersyst/xrpl-go/xrpl/queries/nft/v1"
	path "github.com/Peersyst/xrpl-go/xrpl/queries/path"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	pathv1 "github.com/Peersyst/xrpl-go/xrpl/queries/path/v1"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Conformance cases for the path, NFT, and channel query requests (API v1 + v2).
// Each case pairs a fully-valid request with the xrpl.js interface it must
// conform to.
func init() {
	// ---- path (API v2) ----

	Register(Case{
		Name:        "path.BookOffersRequest",
		JSInterface: "BookOffersRequest",
		Valid: func() Validatable {
			return &path.BookOffersRequest{
				TakerGets: pathtypes.BookOfferCurrency{Currency: "XRP"},
				TakerPays: pathtypes.BookOfferCurrency{
					Currency: "USD",
					Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				},
			}
		},
	})

	Register(Case{
		Name:        "path.DepositAuthorizedRequest",
		JSInterface: "DepositAuthorizedRequest",
		Valid: func() Validatable {
			return &path.DepositAuthorizedRequest{
				SourceAccount:      "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
				DestinationAccount: "rsUiUMpnrgxQp24dJYZDhmV4bE3aBtQyt8",
			}
		},
	})

	Register(Case{
		Name:        "path.FindCreateRequest",
		JSInterface: "PathFindCreateRequest",
		Valid: func() Validatable {
			return &path.FindCreateRequest{
				Subcommand:         path.Create,
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			}
		},
	})

	Register(Case{
		Name:        "path.FindCloseRequest",
		JSInterface: "PathFindCloseRequest",
		Valid: func() Validatable {
			return &path.FindCloseRequest{
				Subcommand: path.Close,
			}
		},
	})

	Register(Case{
		Name:        "path.FindStatusRequest",
		JSInterface: "PathFindStatusRequest",
		Valid: func() Validatable {
			return &path.FindStatusRequest{
				Subcommand: path.Status,
			}
		},
	})

	Register(Case{
		Name:        "path.RipplePathFindRequest",
		JSInterface: "RipplePathFindRequest",
		Valid: func() Validatable {
			return &path.RipplePathFindRequest{
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: types.IssuedCurrencyAmount{
					Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
					Currency: "USD",
					Value:    "0.001",
				},
			}
		},
	})

	// ---- path (API v1) ----

	Register(Case{
		Name:        "path_v1.BookOffersRequest",
		JSInterface: "BookOffersRequest",
		Valid: func() Validatable {
			return &pathv1.BookOffersRequest{
				TakerGets: pathtypes.BookOfferCurrency{Currency: "XRP"},
				TakerPays: pathtypes.BookOfferCurrency{
					Currency: "USD",
					Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
				},
			}
		},
	})

	Register(Case{
		Name:        "path_v1.DepositAuthorizedRequest",
		JSInterface: "DepositAuthorizedRequest",
		Valid: func() Validatable {
			return &pathv1.DepositAuthorizedRequest{
				SourceAccount:      "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
				DestinationAccount: "rsUiUMpnrgxQp24dJYZDhmV4bE3aBtQyt8",
			}
		},
	})

	Register(Case{
		Name:        "path_v1.FindCreateRequest",
		JSInterface: "PathFindCreateRequest",
		Valid: func() Validatable {
			return &pathv1.FindCreateRequest{
				Subcommand:         pathv1.Create,
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
				DestinationAmount:  types.XRPCurrencyAmount(1000000),
			}
		},
	})

	Register(Case{
		Name:        "path_v1.FindCloseRequest",
		JSInterface: "PathFindCloseRequest",
		Valid: func() Validatable {
			return &pathv1.FindCloseRequest{
				Subcommand: pathv1.Close,
			}
		},
	})

	Register(Case{
		Name:        "path_v1.FindStatusRequest",
		JSInterface: "PathFindStatusRequest",
		Valid: func() Validatable {
			return &pathv1.FindStatusRequest{
				Subcommand: pathv1.Status,
			}
		},
	})

	Register(Case{
		Name:        "path_v1.RipplePathFindRequest",
		JSInterface: "RipplePathFindRequest",
		Valid: func() Validatable {
			return &pathv1.RipplePathFindRequest{
				SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAmount: types.IssuedCurrencyAmount{
					Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
					Currency: "USD",
					Value:    "0.001",
				},
			}
		},
	})

	// ---- nft (API v2) ----

	Register(Case{
		Name:        "nft.NFTokenBuyOffersRequest",
		JSInterface: "NFTBuyOffersRequest",
		Valid: func() Validatable {
			return &nft.NFTokenBuyOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			}
		},
	})

	Register(Case{
		Name:        "nft.NFTokenSellOffersRequest",
		JSInterface: "NFTSellOffersRequest",
		Valid: func() Validatable {
			return &nft.NFTokenSellOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			}
		},
	})

	// ---- nft (API v1) ----

	Register(Case{
		Name:        "nft_v1.NFTokenBuyOffersRequest",
		JSInterface: "NFTBuyOffersRequest",
		Valid: func() Validatable {
			return &nftv1.NFTokenBuyOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			}
		},
	})

	Register(Case{
		Name:        "nft_v1.NFTokenSellOffersRequest",
		JSInterface: "NFTSellOffersRequest",
		Valid: func() Validatable {
			return &nftv1.NFTokenSellOffersRequest{
				NFTokenID: "00090000D0B007439B080E9B05BF62403911301A7B1F0CFAA048C0A200000007",
			}
		},
	})

	// ---- channel (API v2) ----

	Register(Case{
		Name:        "channel.VerifyRequest",
		JSInterface: "ChannelVerifyRequest",
		Valid: func() Validatable {
			return &channel.VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
				PublicKey: "aB44YfzW24VDEJQ2UuLPV2PvqcPCSoLnL7y5M1EzhdW4LnK5xMS3",
				Signature: "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064",
			}
		},
	})

	// ---- channel (API v1) ----

	Register(Case{
		Name:        "channel_v1.VerifyRequest",
		JSInterface: "ChannelVerifyRequest",
		Valid: func() Validatable {
			return &channelv1.VerifyRequest{
				Amount:    types.XRPCurrencyAmount(1000000),
				ChannelID: "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
				PublicKey: "aB44YfzW24VDEJQ2UuLPV2PvqcPCSoLnL7y5M1EzhdW4LnK5xMS3",
				Signature: "304402204EF0AFB78AC23ED1C472E74F4299C0C21F1B21D07EFC0A3838A420F76D783A400220154FB11B6F54320666E4C36CA7F686C16A3A0456800BBC43746F34AF50290064",
			}
		},
	})
}
