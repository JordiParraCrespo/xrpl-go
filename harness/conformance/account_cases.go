package conformance

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accountv1 "github.com/Peersyst/xrpl-go/xrpl/queries/account/v1"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
)

// Conformance cases for the account_* query requests (API v2). Each case pairs
// a fully-valid request with the xrpl.js interface it must conform to.
func init() {
	Register(Case{
		Name:        "account.InfoRequest",
		JSInterface: "AccountInfoRequest",
		Valid: func() Validatable {
			return &account.InfoRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.ChannelsRequest",
		JSInterface: "AccountChannelsRequest",
		Valid: func() Validatable {
			return &account.ChannelsRequest{
				Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
				LedgerIndex:        common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.CurrenciesRequest",
		JSInterface: "AccountCurrenciesRequest",
		Valid: func() Validatable {
			return &account.CurrenciesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.GatewayBalancesRequest",
		JSInterface: "GatewayBalancesRequest",
		Valid: func() Validatable {
			return &account.GatewayBalancesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.LinesRequest",
		JSInterface: "AccountLinesRequest",
		Valid: func() Validatable {
			return &account.LinesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.NFTsRequest",
		JSInterface: "AccountNFTsRequest",
		Valid: func() Validatable {
			return &account.NFTsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.NoRippleCheckRequest",
		JSInterface: "NoRippleCheckRequest",
		Valid: func() Validatable {
			return &account.NoRippleCheckRequest{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Role:    "gateway",
			}
		},
	})

	Register(Case{
		Name:        "account.ObjectsRequest",
		JSInterface: "AccountObjectsRequest",
		Valid: func() Validatable {
			return &account.ObjectsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.OffersRequest",
		JSInterface: "AccountOffersRequest",
		Valid: func() Validatable {
			return &account.OffersRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account.TransactionsRequest",
		JSInterface: "AccountTxRequest",
		Valid: func() Validatable {
			return &account.TransactionsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	// API v1 mirrors. Same xrpl.js interfaces and required field sets as v2.
	Register(Case{
		Name:        "account_v1.CurrenciesRequest",
		JSInterface: "AccountCurrenciesRequest",
		Valid: func() Validatable {
			return &accountv1.CurrenciesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.InfoRequest",
		JSInterface: "AccountInfoRequest",
		Valid: func() Validatable {
			return &accountv1.InfoRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.LinesRequest",
		JSInterface: "AccountLinesRequest",
		Valid: func() Validatable {
			return &accountv1.LinesRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.NFTsRequest",
		JSInterface: "AccountNFTsRequest",
		Valid: func() Validatable {
			return &accountv1.NFTsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.NoRippleCheckRequest",
		JSInterface: "NoRippleCheckRequest",
		Valid: func() Validatable {
			return &accountv1.NoRippleCheckRequest{
				Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				Role:    "gateway",
			}
		},
	})

	Register(Case{
		Name:        "account_v1.ObjectsRequest",
		JSInterface: "AccountObjectsRequest",
		Valid: func() Validatable {
			return &accountv1.ObjectsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.OffersRequest",
		JSInterface: "AccountOffersRequest",
		Valid: func() Validatable {
			return &accountv1.OffersRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "account_v1.TransactionsRequest",
		JSInterface: "AccountTxRequest",
		Valid: func() Validatable {
			return &accountv1.TransactionsRequest{
				Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				LedgerIndex: common.Validated,
			}
		},
	})
}
