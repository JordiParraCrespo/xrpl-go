package review

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/claude"
)

// fakeRunner answers by label prefix and records every request.
type fakeRunner struct {
	mu      sync.Mutex
	answers map[string]string
	// verdicts maps a finding title to the verifier's verdict.
	verdicts map[string]string
	reqs     []claude.Request
}

var findingItemRe = regexp.MustCompile(`"id": (\d+),\s*"file": "[^"]*",\s*"line": \d+,\s*"severity": "[^"]*",\s*"title": "([^"]*)"`)

func (f *fakeRunner) Run(_ context.Context, req claude.Request) (claude.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reqs = append(f.reqs, req)
	if strings.HasPrefix(req.Label, "verify") && f.verdicts != nil {
		var results []string
		for _, m := range findingItemRe.FindAllStringSubmatch(req.Prompt, -1) {
			if v, ok := f.verdicts[m[2]]; ok {
				results = append(results, `{"id":`+m[1]+`,"verdict":"`+v+`","reason":"checked"}`)
			}
		}
		return claude.Response{Output: json.RawMessage(`{"results":[` + strings.Join(results, ",") + `]}`)}, nil
	}
	for prefix, out := range f.answers {
		if strings.HasPrefix(req.Label, prefix) {
			return claude.Response{Output: json.RawMessage(out), CostUSD: 0.01}, nil
		}
	}
	return claude.Response{Output: json.RawMessage(`{"findings":[]}`)}, nil
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setupRepo builds a repo with a base commit and a feature branch that
// changes one transaction file and grows another past the line limit.
func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "xrpl/transaction/payment.go", "package transaction\n\nfunc a() {}\n")
	write(t, dir, "xrpl/transaction/big.go", "package transaction\n")
	write(t, dir, "CHANGELOG.md", "# Changelog\n")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "base")
	gitRun(t, dir, "checkout", "-q", "-b", "feature")
	write(t, dir, "xrpl/transaction/payment.go", "package transaction\n\nfunc a() {}\n\nfunc b() { panic(1) }\n")
	write(t, dir, "xrpl/transaction/big.go", "package transaction\n"+strings.Repeat("// x\n", 30))
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "feature")
	return dir
}

func setupConfig(t *testing.T, verify bool) *Config {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "lenses/correctness.md", "Review.\n<untrusted_diff>\n{{DIFF}}\n</untrusted_diff>\n")
	write(t, dir, "lenses/rules.md", "Rules:\n{{RULES}}\nFiles:\n{{FILES}}\n{{DIFF}}")
	write(t, dir, "lenses/verify.md", "Verify:\n{{FINDINGS}}\n{{DIFF}}")
	write(t, dir, "rules.yaml", `rules:
  - id: TX-01
    title: no panics
    severity: blocker
    paths: ["xrpl/transaction/*.go"]
    exclude: ["**/*_test.go"]
    text: Never panic.
  - id: DOC-01
    title: docs only
    paths: ["docs/**"]
    text: Unrelated.
`)
	pipeline := `model: sonnet
facts:
  file_line_limit: 20
  changelog:
    CHANGELOG.md: ["xrpl/**/*.go"]
lenses:
  - name: correctness
    kind: diff
    prompt: lenses/correctness.md
  - name: rules
    kind: bundle
    prompt: lenses/rules.md
    rules: rules.yaml
  - name: typesafe
    kind: command
    command: ["definitely-not-installed-gorev"]
    optional: true
verify:
  enabled: ` + map[bool]string{true: "true", false: "false"}[verify] + `
  prompt: lenses/verify.md
`
	write(t, dir, "pipeline.yaml", pipeline)
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestPipelineRun(t *testing.T) {
	repo := setupRepo(t)
	cfg := setupConfig(t, true)
	runner := &fakeRunner{answers: map[string]string{
		"lens:correctness": `{"findings":[
			{"file":"xrpl/transaction/payment.go","line":5,"severity":"concern","category":"errors","title":"b panics on every call","body":"panic(1) crashes callers."},
			{"file":"xrpl/transaction/payment.go","line":3,"severity":"nit","category":"style","title":"a is empty","body":"Remove it."}
		]}`,
		"lens:rules": `{"findings":[
			{"file":"./xrpl/transaction/payment.go","line":5,"severity":"blocker","category":"errors","title":"Function b panics","body":"TX-01 forbids panics in library code.","source":"TX-01"}
		]}`,
	}, verdicts: map[string]string{
		"Function b panics":      VerdictConfirmed,
		"b panics on every call": VerdictConfirmed,
		"a is empty":             VerdictRefuted,
	}}
	p := &Pipeline{Config: cfg, Runner: runner}
	res, err := p.Run(context.Background(), Input{RepoDir: repo, Base: "main", Head: "feature"})
	if err != nil {
		t.Fatal(err)
	}

	// The rules lens only received the matching rule.
	for _, r := range runner.reqs {
		if strings.HasPrefix(r.Label, "lens:rules") {
			if !strings.Contains(r.Prompt, "TX-01") || strings.Contains(r.Prompt, "DOC-01") {
				t.Errorf("rules prompt has wrong rules:\n%s", r.Prompt)
			}
		}
		if r.Model != "sonnet" {
			t.Errorf("%s used model %q", r.Label, r.Model)
		}
	}

	if res.Verdict != VerdictChangesSuggested {
		t.Errorf("verdict = %s", res.Verdict)
	}
	// Two lens findings on line 5 merge into one blocker with both sources.
	var merged *Finding
	for i, f := range res.Findings {
		if f.File == "xrpl/transaction/payment.go" && f.Line == 5 {
			merged = &res.Findings[i]
		}
	}
	if merged == nil {
		t.Fatalf("merged finding missing: %+v", res.Findings)
	}
	if merged.Severity != SeverityBlocker || len(merged.Sources) != 2 || !merged.Inline || merged.Verdict != VerdictConfirmed {
		t.Errorf("merged finding: %+v", *merged)
	}
	if res.Findings[0].ID != 1 || res.Findings[0].Severity != SeverityBlocker {
		t.Errorf("blocker not ranked first: %+v", res.Findings[0])
	}
	if len(res.Refuted) != 1 || res.Refuted[0].Title != "a is empty" {
		t.Errorf("refuted = %+v", res.Refuted)
	}

	// Deterministic facts: file size and changelog.
	var sawSize, sawChangelog bool
	for _, f := range res.Findings {
		switch f.Category {
		case "structure/file-size":
			sawSize = f.File == "xrpl/transaction/big.go"
		case "cross-cutting/changelog":
			sawChangelog = f.File == "CHANGELOG.md" && !f.Inline
		}
	}
	if !sawSize || !sawChangelog {
		t.Errorf("facts missing: size=%v changelog=%v in %+v", sawSize, sawChangelog, res.Findings)
	}

	var skipped bool
	for _, r := range res.Runs {
		if r.Name == "typesafe" && r.Status == LensSkipped {
			skipped = true
		}
	}
	if !skipped {
		t.Errorf("optional missing command should be skipped: %+v", res.Runs)
	}
}

func TestPipelineNoVerifyKeepsFindings(t *testing.T) {
	repo := setupRepo(t)
	cfg := setupConfig(t, false)
	runner := &fakeRunner{answers: map[string]string{
		"lens:correctness": `{"findings":[{"file":"xrpl/transaction/payment.go","line":99,"severity":"concern","title":"outside hunk","body":"x"}]}`,
	}}
	res, err := (&Pipeline{Config: cfg, Runner: runner}).Run(context.Background(), Input{RepoDir: repo, Base: "main", Head: "feature"})
	if err != nil {
		t.Fatal(err)
	}
	var f *Finding
	for i := range res.Findings {
		if res.Findings[i].Title == "outside hunk" {
			f = &res.Findings[i]
		}
	}
	if f == nil || f.Inline || f.Verdict != VerdictUnverified {
		t.Errorf("finding = %+v", f)
	}
	for _, r := range runner.reqs {
		if strings.HasPrefix(r.Label, "verify") {
			t.Error("verify ran while disabled")
		}
	}
}

func TestParseExternal(t *testing.T) {
	raw := `[{"file":"a.go","line":3,"severity":"concern","source":"TX-06","message":"Validate never checks InvoiceID. It passes locally.","suggestion":"Use IsHex256."}]`
	fs, err := parseExternal([]byte(raw), "typesafe")
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].Title != "Validate never checks InvoiceID" || fs[0].Sources[0] != "typesafe/TX-06" {
		t.Errorf("got %+v", fs)
	}
	wrapped, err := parseExternal([]byte(`{"findings":`+raw+`}`), "typesafe")
	if err != nil || len(wrapped) != 1 {
		t.Errorf("wrapped: %v %+v", err, wrapped)
	}
}

func TestPublishSkipsPostedFingerprints(t *testing.T) {
	res := &Result{Head: "abcdef0123456789", Verdict: VerdictComments, Findings: []Finding{
		{ID: 1, Severity: SeverityConcern, Title: "one", File: "a.go", Line: 3, Body: "b", Inline: true, Fingerprint: "aaa111"},
		{ID: 2, Severity: SeverityNit, Title: "two", File: "a.go", Line: 9, Body: "b", Inline: true, Fingerprint: "bbb222"},
		{ID: 3, Severity: SeverityConcern, Title: "three", File: "CHANGELOG.md", Body: "b", Fingerprint: "ccc333"},
	}}
	posted := FingerprintsIn([]string{"old comment\n<!-- xrpl-go-ai-review:fp=aaa111 -->"})
	pub := Publish(res, posted)
	if len(pub.Inline) != 1 || pub.Inline[0].Line != 9 {
		t.Errorf("inline = %+v", pub.Inline)
	}
	if !strings.HasPrefix(pub.Summary, SummaryMarker) || !strings.Contains(pub.Summary, "three") || !strings.Contains(pub.Summary, "fp=ccc333") {
		t.Errorf("summary:\n%s", pub.Summary)
	}
	if strings.Contains(pub.Inline[0].Body, "sources") {
		t.Error("inline comment leaks metadata")
	}
}

func TestFill(t *testing.T) {
	got := fill("A {{DIFF}} B {{FILES}}", map[string]string{"DIFF": "has {{FILES}} inside", "FILES": "f"})
	if got != "A has {{FILES}} inside B f" {
		t.Errorf("fill expanded placeholders inside values: %q", got)
	}
}

func TestLoadConfigRejectsBadLens(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pipeline.yaml", "lenses:\n  - name: x\n    kind: nope\n")
	if _, err := LoadConfig(dir); err == nil {
		t.Error("expected error for unknown lens kind")
	}
}
