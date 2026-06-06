#!/usr/bin/env bash
# validate_coverage.sh fails if any Validate() function in the query packages is
# below 100% statement coverage. Combined with the conformance harness, this is
# the objective "all branches tested" gate for the request validators.
#
# Usage: harness/validate_coverage.sh [packages...]
#   defaults to ./xrpl/queries/...
set -euo pipefail

cd "$(dirname "$0")/.."

pkgs=("$@")
if [ ${#pkgs[@]} -eq 0 ]; then
  pkgs=("./xrpl/queries/...")
fi

profile="$(mktemp)"
trap 'rm -f "$profile"' EXIT

go test -coverprofile="$profile" "${pkgs[@]}" >/dev/null

# Column layout of `go tool cover -func`: "<file>:<line>:" "<func>" "<pct>%".
fails="$(go tool cover -func="$profile" | awk '$2 == "Validate" && $3 != "100.0%" { print }')"

if [ -n "$fails" ]; then
  echo "Validate() functions below 100% coverage:"
  echo "$fails"
  exit 1
fi

count="$(go tool cover -func="$profile" | awk '$2 == "Validate"' | wc -l | tr -d ' ')"
echo "All ${count} Validate() functions at 100% coverage"
