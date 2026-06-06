package conformance

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/clio"
	cliov1 "github.com/Peersyst/xrpl-go/xrpl/queries/clio/v1"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	ledgerv1 "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/v1"
)

// Conformance cases for the ledger_* and CLIO NFT query requests (API v2 and
// v1). Each case pairs a fully-valid request with the xrpl.js interface it must
// conform to. The ledger requests have no required fields in xrpl.js, so their
// only assertion is that the valid instance passes Validate().
func init() {
	// ------------------------------------------------------------------
	// ledger (API v2)
	// ------------------------------------------------------------------
	Register(Case{
		Name:        "ledger.Request",
		JSInterface: "LedgerRequest",
		Valid: func() Validatable {
			return &ledger.Request{
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "ledger.ClosedRequest",
		JSInterface: "LedgerClosedRequest",
		Valid: func() Validatable {
			return &ledger.ClosedRequest{}
		},
	})

	Register(Case{
		Name:        "ledger.CurrentRequest",
		JSInterface: "LedgerCurrentRequest",
		Valid: func() Validatable {
			return &ledger.CurrentRequest{}
		},
	})

	Register(Case{
		Name:        "ledger.DataRequest",
		JSInterface: "LedgerDataRequest",
		Valid: func() Validatable {
			return &ledger.DataRequest{
				LedgerIndex: common.Validated,
			}
		},
	})

	// ------------------------------------------------------------------
	// ledger (API v1)
	// ------------------------------------------------------------------
	Register(Case{
		Name:        "ledger_v1.Request",
		JSInterface: "LedgerRequest",
		Valid: func() Validatable {
			return &ledgerv1.Request{
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "ledger_v1.ClosedRequest",
		JSInterface: "LedgerClosedRequest",
		Valid: func() Validatable {
			return &ledgerv1.ClosedRequest{}
		},
	})

	Register(Case{
		Name:        "ledger_v1.CurrentRequest",
		JSInterface: "LedgerCurrentRequest",
		Valid: func() Validatable {
			return &ledgerv1.CurrentRequest{}
		},
	})

	Register(Case{
		Name:        "ledger_v1.DataRequest",
		JSInterface: "LedgerDataRequest",
		Valid: func() Validatable {
			return &ledgerv1.DataRequest{
				LedgerIndex: common.Validated,
			}
		},
	})

	// ------------------------------------------------------------------
	// clio (API v2)
	// ------------------------------------------------------------------
	Register(Case{
		Name:        "clio.NFTHistoryRequest",
		JSInterface: "NFTHistoryRequest",
		Valid: func() Validatable {
			return &clio.NFTHistoryRequest{
				NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
			}
		},
	})

	Register(Case{
		Name:        "clio.NFTInfoRequest",
		JSInterface: "NFTInfoRequest",
		Valid: func() Validatable {
			return &clio.NFTInfoRequest{
				NFTokenID:   "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
				LedgerIndex: common.Validated,
			}
		},
	})

	Register(Case{
		Name:        "clio.NFTsByIssuerRequest",
		JSInterface: "NFTsByIssuerRequest",
		Valid: func() Validatable {
			return &clio.NFTsByIssuerRequest{
				Issuer: "rHVokeuSnjPjz718qdb47bGXBBHNMP3KDQ",
			}
		},
	})

	// ------------------------------------------------------------------
	// clio (API v1)
	// ------------------------------------------------------------------
	Register(Case{
		Name:        "clio_v1.NFTHistoryRequest",
		JSInterface: "NFTHistoryRequest",
		Valid: func() Validatable {
			return &cliov1.NFTHistoryRequest{
				NFTokenID: "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
			}
		},
	})

	Register(Case{
		Name:        "clio_v1.NFTInfoRequest",
		JSInterface: "NFTInfoRequest",
		Valid: func() Validatable {
			return &cliov1.NFTInfoRequest{
				NFTokenID:   "00080000B4F4AFC5FBCBD76873F18006173D2193467D3EE70000099B00000000",
				LedgerIndex: common.Validated,
			}
		},
	})
}
