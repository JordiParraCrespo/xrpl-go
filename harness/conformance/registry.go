// Package conformance hosts the static conformance harness that checks every
// xrpl-go query-request Validate() against the field requirements declared by
// the matching xrpl.js request interface.
//
// The required/optional field sets are extracted from the pinned xrpl.js source
// by js/extract.mjs into jsspec.json (regenerate with `go generate ./...` once
// xrpl.js is checked out, or run the script directly). Each Go request type is
// registered here via Register, pairing a constructor for a fully-valid request
// with the name of the xrpl.js interface it must conform to.
//
// The conformance test then, for every registered case:
//   - asserts the valid instance passes Validate(); and
//   - for each field xrpl.js marks required, zeroes that field on a fresh valid
//     instance and asserts Validate() now reports an error.
//
// Precise guarantee and its limits. Because only the field under test is changed
// from a known-valid instance, an error there proves that field's presence is
// enforced. The harness therefore proves the required-field-subset direction:
// every xrpl.js-required field is enforced by Go. It deliberately does NOT
// prove the converse (that no optional field is over-validated), nor the
// identity of the returned error, nor any value-format rule — those are covered
// by the per-request table tests. Two blind spots to keep in mind when adding
// cases: a required field whose Go zero value is itself legal (so zeroing would
// not trigger an error) must be covered by a table test and listed in
// SkipRequired; and the harness only zeroes top-level fields, so nested
// object/element rules are out of its scope by design.
package conformance

import (
	_ "embed"
	"encoding/json"
)

// xrpl.js pinned at commit acdc1a206412077d59e243c9f4a03f3ba4d4bfa8 (v5.0.0).
// Regenerate jsspec.json with:
//
//	node harness/conformance/js/extract.mjs \
//	  <xrpl.js>/packages/xrpl/src/models/methods harness/conformance/jsspec.json
//
//go:embed jsspec.json
var jsSpecRaw []byte

// JSRequestSpec describes the field requirements of one xrpl.js request interface.
type JSRequestSpec struct {
	File     string   `json:"file"`
	Required []string `json:"required"`
	Optional []string `json:"optional"`
}

// jsSpec is the decoded jsspec.json, keyed by xrpl.js interface name.
var jsSpec = func() map[string]JSRequestSpec {
	var spec map[string]JSRequestSpec
	if err := json.Unmarshal(jsSpecRaw, &spec); err != nil {
		panic("conformance: invalid jsspec.json: " + err.Error())
	}
	return spec
}()

// Validatable is the contract every query request implements.
type Validatable interface {
	Validate() error
}

// Case pairs a Go request type with the xrpl.js interface it must conform to.
type Case struct {
	// Name is a human-readable identifier used in test output, typically the Go
	// type name including its package (e.g. "account.InfoRequest").
	Name string

	// JSInterface is the xrpl.js request interface name as it appears in
	// jsspec.json (e.g. "AccountInfoRequest"). When empty, the required-field
	// conformance check is skipped and only ValidPasses is exercised.
	JSInterface string

	// Valid returns a fresh, fully-valid request instance. Validate() must
	// return nil for it. It is called once per assertion so mutations do not
	// leak between checks.
	Valid func() Validatable

	// SkipRequired lists xrpl.js field names whose auto-generated missing-field
	// check should be skipped (for example, required fields whose Go zero value
	// is still a legal value, which are covered by hand-written table tests).
	SkipRequired []string
}

// cases is the global registry, populated by per-package init() functions.
var cases []Case

// Register adds a conformance case to the registry. It is intended to be called
// from init() in per-package case files so that parallel work on different
// query packages never edits a shared file.
func Register(c Case) {
	cases = append(cases, c)
}
