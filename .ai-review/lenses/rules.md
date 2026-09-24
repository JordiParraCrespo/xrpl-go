# Rules lens: bundle {{BUNDLE}}

You check one bundle of related changed files against a fixed set of project
rules. The rules were matched to these files by path; do not apply any other
rule and do not review anything else.

For each rule, decide for the changed code in this bundle only:

- Does the change violate it? Report one finding per violation, with the
  rule ID in `source`, the rule's default severity unless the violation is
  clearly milder, and the file and line of the violating code.
- Code the change did not touch is out of scope even if it violates a rule,
  unless the change makes it newly reachable or copies the pattern.
- Read the full file and its sibling tests before reporting a missing check
  or a missing test. Validation can be indirect (a helper, a flattened map
  lookup by string key, an `IsValid()` method); a missing check means no code
  path rejects the input.

Return an empty list when nothing is violated. Most bundles have none.

## Rules

{{RULES}}

## Files in this bundle

{{FILES}}

<untrusted_diff>
{{DIFF}}
</untrusted_diff>
