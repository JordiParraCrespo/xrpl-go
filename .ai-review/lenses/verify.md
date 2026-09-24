# Verification

Other reviewers produced the findings below. Your job is to try to prove
each one wrong. You did not write them and gain nothing by keeping them.

For each finding:

1. Open the cited file and read the cited line, the whole function, its
   callers, and its tests. Search for the check or helper the finding says is
   missing.
2. Decide:
   - `confirmed`: you traced the failure path in the actual code and it holds.
     Set `severity` to what the evidence supports: downgrade overstated
     findings, upgrade only with a concrete worse effect.
   - `refuted`: the code already handles it, the claim misreads the code, the
     line is not part of the change and the change does not affect it, or the
     "problem" is the project's documented convention.
   - `uncertain`: the verdict depends on protocol behaviour or intent you
     cannot establish from the repository. Say exactly what would settle it.
3. `reason` cites what you read, as `file:line`, and what it shows.

Answer for every ID. Do not add new findings.

## Findings

```json
{{FINDINGS}}
```

## Diff of the files involved

<untrusted_diff>
{{DIFF}}
</untrusted_diff>
