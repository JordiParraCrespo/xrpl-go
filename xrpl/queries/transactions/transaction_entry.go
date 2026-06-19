package transactions

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// ############################################################################
// Request
// ############################################################################

// EntryRequest is the request type for the transaction_entry command.
// It retrieves information on a single transaction from a specific ledger version.
type EntryRequest struct {
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	TxHash      string                 `json:"tx_hash"`
}

// Method returns the JSON-RPC method name for the EntryRequest.
func (*EntryRequest) Method() string {
	return "transaction_entry"
}

// Validate ensures the EntryRequest is valid. The transaction_entry command
// requires the tx_hash field; all other fields are optional.
func (r *EntryRequest) Validate() error {
	if r.TxHash == "" {
		return ErrNoTxHash
	}

	return nil
}

// ############################################################################
// Response
// ############################################################################

// EntryResponse is the response type returned by the transaction_entry command.
// It contains the ledger index, ledger hash, transaction metadata, and the transaction itself.
type EntryResponse struct {
	LedgerIndex common.LedgerIndex          `json:"ledger_index"`
	LedgerHash  common.LedgerHash           `json:"ledger_hash,omitempty"`
	Metadata    transaction.TxObjMeta       `json:"metadata"`
	Tx          transaction.FlatTransaction `json:"tx_json"`
}
