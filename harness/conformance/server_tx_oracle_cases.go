package conformance

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	oracletypes "github.com/Peersyst/xrpl-go/xrpl/queries/oracle/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	transactionsv1 "github.com/Peersyst/xrpl-go/xrpl/queries/transactions/v1"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// Conformance cases for the server_*, transaction (tx / submit_multisigned) and
// oracle (get_aggregate_price) query requests. Each case pairs a fully-valid
// request with the xrpl.js interface it must conform to.
func init() {
	// ########################################################################
	// server
	// ########################################################################

	// FeatureAllRequest models xrpl.js FeatureAllRequest, which has no required
	// fields (feature is `feature?: never`).
	Register(Case{
		Name:        "server.FeatureAllRequest",
		JSInterface: "FeatureAllRequest",
		Valid: func() Validatable {
			return &server.FeatureAllRequest{}
		},
	})

	// FeatureOneRequest models xrpl.js FeatureOneRequest, which requires feature.
	Register(Case{
		Name:        "server.FeatureOneRequest",
		JSInterface: "FeatureOneRequest",
		Valid: func() Validatable {
			return &server.FeatureOneRequest{
				Feature: "foo",
			}
		},
	})

	// FeeRequest models xrpl.js FeeRequest, which has no required fields.
	Register(Case{
		Name:        "server.FeeRequest",
		JSInterface: "FeeRequest",
		Valid: func() Validatable {
			return &server.FeeRequest{}
		},
	})

	// ManifestRequest models xrpl.js ManifestRequest, which requires public_key.
	Register(Case{
		Name:        "server.ManifestRequest",
		JSInterface: "ManifestRequest",
		Valid: func() Validatable {
			return &server.ManifestRequest{
				PublicKey: "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
			}
		},
	})

	// InfoRequest models xrpl.js ServerInfoRequest, which has no required fields.
	Register(Case{
		Name:        "server.InfoRequest",
		JSInterface: "ServerInfoRequest",
		Valid: func() Validatable {
			return &server.InfoRequest{}
		},
	})

	// StateRequest models xrpl.js ServerStateRequest, which has no required fields.
	Register(Case{
		Name:        "server.StateRequest",
		JSInterface: "ServerStateRequest",
		Valid: func() Validatable {
			return &server.StateRequest{}
		},
	})

	// ########################################################################
	// transactions (API v2)
	// ########################################################################

	// SubmitMultisignedRequest models xrpl.js SubmitMultisignedRequest, which
	// requires tx_json. The Go field is a transaction.FlatTransaction (a map);
	// its reflect.Zero value (nil map) is cleanly rejectable by Validate().
	Register(Case{
		Name:        "transactions.SubmitMultisignedRequest",
		JSInterface: "SubmitMultisignedRequest",
		Valid: func() Validatable {
			return &transactions.SubmitMultisignedRequest{
				Tx: transaction.FlatTransaction{
					"TransactionType": "Payment",
					"Account":         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				},
			}
		},
	})

	// TxRequest models xrpl.js TxRequest. Per jsspec.json every field is
	// optional (transaction/ctid/binary/min_ledger/max_ledger); xrpl.js only
	// documents the "exactly one of transaction or ctid" rule in prose, and the
	// Go struct does not even model ctid, so no hard-required field is encoded.
	Register(Case{
		Name:        "transactions.TxRequest",
		JSInterface: "TxRequest",
		Valid: func() Validatable {
			return &transactions.TxRequest{
				Transaction: "E08D6E9754025BA2534A78707605E0601F03ACE063687A0CA1BDDACFCD1698C7",
			}
		},
	})

	// ########################################################################
	// transactions (API v1 mirrors)
	// ########################################################################

	Register(Case{
		Name:        "transactions_v1.SubmitMultisignedRequest",
		JSInterface: "SubmitMultisignedRequest",
		Valid: func() Validatable {
			return &transactionsv1.SubmitMultisignedRequest{
				Tx: transaction.FlatTransaction{
					"TransactionType": "Payment",
					"Account":         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				},
			}
		},
	})

	Register(Case{
		Name:        "transactions_v1.TxRequest",
		JSInterface: "TxRequest",
		Valid: func() Validatable {
			return &transactionsv1.TxRequest{
				Transaction: "E08D6E9754025BA2534A78707605E0601F03ACE063687A0CA1BDDACFCD1698C7",
			}
		},
	})

	// ########################################################################
	// oracle
	// ########################################################################

	// GetAggregatePriceRequest models xrpl.js GetAggregatePriceRequest, which
	// requires base_asset, quote_asset and oracles. oracles is a slice whose
	// reflect.Zero value (nil) is cleanly rejectable by Validate().
	Register(Case{
		Name:        "oracle.GetAggregatePriceRequest",
		JSInterface: "GetAggregatePriceRequest",
		Valid: func() Validatable {
			return &oracle.GetAggregatePriceRequest{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				Oracles: []oracletypes.Oracle{
					{
						Account:          "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
						OracleDocumentID: "123",
					},
				},
			}
		},
	})
}
