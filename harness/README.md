# Query-request validation harness

This harness implements and verifies the `Validate()` methods on every XRPL
query-request type in `xrpl/queries/...`. It makes "Go validation conforms to
xrpl.js and rippled" a machine-checked invariant instead of a manual review.

## Layout

```
harness/
  conformance/
    js/extract.mjs        # parses xrpl.js request interfaces -> jsspec.json
    jsspec.json           # pinned required/optional field sets (xrpl.js v5.0.0,
                          # commit acdc1a206412077d59e243c9f4a03f3ba4d4bfa8)
    registry.go           # Case registry + jsspec loader + Validatable contract
    conformance_test.go   # reflection-based required-field conformance test
    <pkg>_cases.go        # per-query-package registrations (one file per pkg)
  validate_coverage.sh    # fails if any Validate() is < 100% coverage
  README.md
```

## What "validated against the references" means

xrpl.js request models are **TypeScript interfaces** (no runtime validator), so
the parity check is *static*:

- **xrpl.js** defines which fields are **required** (un-`?`-marked) vs optional.
  `extract.mjs` distills this into `jsspec.json`. The conformance test zeroes
  each required field on a valid instance and asserts `Validate()` rejects it.
- **rippled** RPC handlers (`src/.../rpc/handlers/*.cpp`) are the authority on
  field **formats / ranges / mutual-exclusivity** and error behavior. These are
  encoded as hand-written table-test cases per request and are what the coverage
  gate measures.

## The contract for adding a request validator

For each query-request type with a `Validate()` stub:

1. **Implement `Validate()`** in its `.go` file. Enforce every field that
   xrpl.js marks required (see `jsspec.json`), plus the format/range rules from
   the matching rippled handler. Return a typed error from the package's
   `errors.go` (add one if needed; reuse existing errors where possible).
2. **Register a conformance case** in `harness/conformance/<pkg>_cases.go`
   (create the file if it does not exist — one file per query package so
   parallel work never edits a shared file). Provide a fully-valid instance and
   the xrpl.js interface name from `jsspec.json`. Use `SkipRequired` only for a
   required field whose Go zero value is itself legal (rare), and cover it with
   a hand-written test instead.
3. **Add a `TestXxxRequest_Validate` table test** in the package's `_test.go`,
   covering the valid path and every error branch (each missing required field,
   each format/range failure). This is what drives coverage to 100%.

Run the gates:

```
make validate-conformance   # required-field parity vs xrpl.js
make validate-coverage      # 100% statement coverage of every Validate()
```

Regenerate the JS spec when bumping the pinned xrpl.js version:

```
make validate-jsspec XRPLJS=/path/to/xrpl.js
```
