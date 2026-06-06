package conformance

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
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
}
