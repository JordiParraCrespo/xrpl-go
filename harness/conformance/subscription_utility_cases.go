package conformance

import (
	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	subscribev1 "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/v1"
	"github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	utilityv1 "github.com/Peersyst/xrpl-go/xrpl/queries/utility/v1"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Conformance cases for the subscription_* and utility_* query requests. Every
// one of these xrpl.js interfaces declares zero required fields, so each valid
// instance must pass Validate() and the harness asserts no field is required.
func init() {
	// ----- subscription (API v2) -----
	Register(Case{
		Name:        "subscription.SubscribeRequest",
		JSInterface: "SubscribeRequest",
		Valid: func() Validatable {
			return &subscribe.Request{
				Streams:  []string{"ledger", "transactions"},
				Accounts: []types.Address{"rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn"},
			}
		},
	})

	Register(Case{
		Name:        "subscription.UnsubscribeRequest",
		JSInterface: "UnsubscribeRequest",
		Valid: func() Validatable {
			return &subscribe.UnsubscribeRequest{
				Streams:  []string{"ledger", "transactions"},
				Accounts: []types.Address{"rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"},
			}
		},
	})

	// ----- subscription (API v1) -----
	Register(Case{
		Name:        "subscription_v1.SubscribeRequest",
		JSInterface: "SubscribeRequest",
		Valid: func() Validatable {
			return &subscribev1.Request{
				Streams:  []string{"ledger", "transactions"},
				Accounts: []types.Address{"rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn"},
			}
		},
	})

	Register(Case{
		Name:        "subscription_v1.UnsubscribeRequest",
		JSInterface: "UnsubscribeRequest",
		Valid: func() Validatable {
			return &subscribev1.UnsubscribeRequest{
				Streams:  []string{"ledger", "transactions"},
				Accounts: []types.Address{"rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"},
			}
		},
	})

	// ----- utility (API v2) -----
	Register(Case{
		Name:        "utility.PingRequest",
		JSInterface: "PingRequest",
		Valid: func() Validatable {
			return &utility.PingRequest{}
		},
	})

	Register(Case{
		Name:        "utility.RandomRequest",
		JSInterface: "RandomRequest",
		Valid: func() Validatable {
			return &utility.RandomRequest{}
		},
	})

	// ----- utility (API v1) -----
	Register(Case{
		Name:        "utility_v1.PingRequest",
		JSInterface: "PingRequest",
		Valid: func() Validatable {
			return &utilityv1.PingRequest{}
		},
	})

	Register(Case{
		Name:        "utility_v1.RandomRequest",
		JSInterface: "RandomRequest",
		Valid: func() Validatable {
			return &utilityv1.RandomRequest{}
		},
	})
}
