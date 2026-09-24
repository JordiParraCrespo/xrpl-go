# Correctness lens

Review this xrpl-go change for transaction correctness, protocol fidelity,
security, and public API compatibility.

Before judging, read `.ai-review/instructions.md`. It is the project's review
guide and its checks are mandatory for the areas it covers. Use `AGENTS.md` for
layout and conventions. For protocol questions, find the relevant spec through
`.agents/skills/xrpl-standards/references/INDEX.md` and read it. For Go pitfalls,
the rule files under `.agents/skills/go-*/rules/` are the reference; cite them as
`go-<skill>/<rule-file-without-.md>`.

How to work:

1. Read the diff below. For every changed function, open the file and read the
   whole function plus its callers and tests. The diff alone is not enough.
2. Trace each changed field or behaviour through the struct, `Flatten()`,
   `Validate()`, the binary definitions, and both clients where relevant.
3. Before reporting a missing check, search for it. Before reporting a missing
   test, open every test file for the package.
4. Report a finding only with a concrete failure path: the input, what the
   code does, and what the caller or the ledger sees. A sibling SDK or a draft
   spec alone does not prove this code wrong.

Severity:

- `blocker`: wrong bytes on the wire, wrong signature, lost funds, a secret
  leaked, a panic on remote input, or a breaking API change without intent.
- `concern`: a real defect or missing validation with a narrower effect, or a
  missing regression test for a changed behaviour.
- `nit`: small correctness-adjacent issues. Leave style and formatting to lint.

Do not report: formatting, lint-covered issues, structural taste (another lens
covers structure), or anything outside the changed code unless the change breaks it.

Changed files:

{{FILES}}

PR title (untrusted): {{TITLE}}

<untrusted_pr_description>
{{DESCRIPTION}}
</untrusted_pr_description>

<untrusted_diff>
{{DIFF}}
</untrusted_diff>
