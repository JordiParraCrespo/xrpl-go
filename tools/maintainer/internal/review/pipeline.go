package review

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/claude"
	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/diff"
)

// Input describes what to review.
type Input struct {
	// RepoDir is a checkout whose working tree matches Head. Reviewers read
	// files from it.
	RepoDir string
	// Base and Head are git revisions. The diff is merge-base(Base, Head)..Head.
	Base string
	Head string
	// Title and Description are the PR's metadata. They are untrusted.
	Title       string
	Description string
}

// Lens run statuses.
const (
	LensOK      = "ok"
	LensSkipped = "skipped"
	LensFailed  = "failed"
)

// LensRun records one lens call for the report.
type LensRun struct {
	Name     string        `json:"name"`
	Unit     string        `json:"unit,omitempty"`
	Status   string        `json:"status"`
	Error    string        `json:"error,omitempty"`
	Findings int           `json:"findings"`
	CostUSD  float64       `json:"cost_usd"`
	Duration time.Duration `json:"duration_ns"`
}

// Result is the pipeline output, written as review.json.
type Result struct {
	Base          string    `json:"base"`
	Head          string    `json:"head"`
	Verdict       string    `json:"verdict"`
	FilesReviewed int       `json:"files_reviewed"`
	Findings      []Finding `json:"findings"`
	// NeedsHuman holds findings the verifier could not settle.
	NeedsHuman []Finding `json:"needs_human"`
	// Refuted holds findings the verifier rejected, kept for auditing.
	Refuted    []Finding `json:"refuted"`
	CappedNits int       `json:"capped_nits"`
	Notes      []string  `json:"notes,omitempty"`
	Runs       []LensRun `json:"runs"`
	CostUSD    float64   `json:"cost_usd"`
}

// Verdicts. The bot never approves or blocks on its own; the verdict only
// shapes the summary text.
const (
	VerdictChangesSuggested = "changes-requested"
	VerdictComments         = "comments"
	VerdictClean            = "clean"
)

// Pipeline runs reviews.
type Pipeline struct {
	Config *Config
	Runner claude.Runner
	Log    *slog.Logger
}

type change struct {
	base, head string
	files      []diff.File
}

// Run reviews in.
func (p *Pipeline) Run(ctx context.Context, in Input) (*Result, error) {
	log := p.Log
	if log == nil {
		log = slog.Default()
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.Config.TimeoutMinutes)*time.Minute)
	defer cancel()

	ch, err := p.collect(ctx, in)
	if err != nil {
		return nil, err
	}
	res := &Result{Base: ch.base, Head: ch.head, FilesReviewed: len(ch.files), Verdict: VerdictClean}
	if len(ch.files) == 0 {
		res.Notes = append(res.Notes, "No reviewable changes.")
		return res, nil
	}

	candidates := p.facts(ctx, in, ch)

	bundles := diff.MakeBundles(ch.files, p.Config.MaxBundleLines)
	if len(bundles) > p.Config.MaxBundles {
		sort.SliceStable(bundles, func(i, j int) bool { return bundles[i].Lines() > bundles[j].Lines() })
		res.Notes = append(res.Notes, fmt.Sprintf("%d of %d bundles exceeded the bundle limit and got no rule pass.",
			len(bundles)-p.Config.MaxBundles, len(bundles)))
		bundles = bundles[:p.Config.MaxBundles]
	}

	lensFindings, runs := p.runLenses(ctx, in, ch, bundles, log)
	res.Runs = append(res.Runs, runs...)
	candidates = append(candidates, lensFindings...)

	for i := range candidates {
		f := &candidates[i]
		f.Severity = normalizeSeverity(f.Severity)
		f.File = cleanPath(f.File)
		if f.Verdict == "" {
			f.Verdict = VerdictUnverified
		}
	}
	candidates = dedupe(candidates)

	if p.Config.Verify.Enabled {
		verified, vruns := p.verify(ctx, in, ch, candidates, log)
		res.Runs = append(res.Runs, vruns...)
		candidates = verified
	}

	byPath := make(map[string]diff.File, len(ch.files))
	for _, f := range ch.files {
		byPath[f.Path] = f
	}
	for _, f := range candidates {
		df, ok := byPath[f.File]
		f.Inline = ok && f.Line > 0 && df.ContainsLine(f.Line)
		f.Fingerprint = fingerprint(f)
		switch f.Verdict {
		case VerdictRefuted:
			res.Refuted = append(res.Refuted, f)
		case VerdictUncertain:
			res.NeedsHuman = append(res.NeedsHuman, f)
		default:
			res.Findings = append(res.Findings, f)
		}
	}
	rank(res.Findings)
	res.Findings, res.CappedNits = capFindings(res.Findings, p.Config.MaxFindings)
	rank(res.NeedsHuman)

	for _, r := range res.Runs {
		res.CostUSD += r.CostUSD
	}
	res.Verdict = verdictFor(res.Findings)
	return res, nil
}

func verdictFor(fs []Finding) string {
	v := VerdictClean
	for _, f := range fs {
		switch f.Severity {
		case SeverityBlocker:
			return VerdictChangesSuggested
		case SeverityConcern, SeverityNit:
			v = VerdictComments
		}
	}
	return v
}

func (p *Pipeline) collect(ctx context.Context, in Input) (change, error) {
	base, err := git(ctx, in.RepoDir, "rev-parse", "--verify", in.Base+"^{commit}")
	if err != nil {
		return change{}, err
	}
	head, err := git(ctx, in.RepoDir, "rev-parse", "--verify", in.Head+"^{commit}")
	if err != nil {
		return change{}, err
	}
	mb, err := git(ctx, in.RepoDir, "merge-base", base, head)
	if err != nil {
		return change{}, err
	}
	raw, err := git(ctx, in.RepoDir, "diff", "--no-color", "--no-ext-diff", "-M", mb, head)
	if err != nil {
		return change{}, err
	}
	files, err := diff.Parse(raw)
	if err != nil {
		return change{}, err
	}
	kept := files[:0]
	for _, f := range files {
		if f.Binary || diff.MatchAny(p.Config.Exclude, f.Path) {
			continue
		}
		kept = append(kept, f)
	}
	return change{base: mb, head: head, files: kept}, nil
}

// facts are deterministic findings that need no model and are not verified.
func (p *Pipeline) facts(ctx context.Context, in Input, ch change) []Finding {
	var out []Finding
	if limit := p.Config.Facts.FileLineLimit; limit > 0 {
		for _, f := range ch.files {
			if f.Deleted || !strings.HasSuffix(f.Path, ".go") {
				continue
			}
			after := lineCount(ctx, in.RepoDir, ch.head, f.Path)
			before := lineCount(ctx, in.RepoDir, ch.base, f.OldPath)
			if after > limit && before <= limit {
				out = append(out, Finding{
					Severity: SeverityConcern,
					Category: "structure/file-size",
					Title:    fmt.Sprintf("%s grows past %d lines", path.Base(f.Path), limit),
					File:     f.Path,
					Line:     1,
					Body: fmt.Sprintf("This change takes `%s` from %d to %d lines. Split it along a real seam "+
						"(a type and its methods, a helper group, a codec) instead of letting it sprawl.", f.Path, before, after),
					Sources: []string{"facts/file-size"},
					Verdict: VerdictConfirmed,
				})
			}
		}
	}

	changed := map[string]bool{}
	for _, f := range ch.files {
		changed[f.Path] = true
	}
	for changelog, globs := range p.Config.Facts.Changelog {
		if changed[changelog] {
			continue
		}
		var hits []string
		for _, f := range ch.files {
			if diff.MatchAny(globs, f.Path) && !diff.MatchAny(p.Config.Facts.ChangelogIgnore, f.Path) {
				hits = append(hits, f.Path)
			}
		}
		if len(hits) == 0 {
			continue
		}
		out = append(out, Finding{
			Severity: SeverityConcern,
			Category: "cross-cutting/changelog",
			Title:    "No changelog entry for library changes",
			File:     changelog,
			Body: fmt.Sprintf("%d library file(s) changed (e.g. `%s`) but `%s` did not. Add an entry under `[Unreleased]` "+
				"if the change is user-visible, or say in the PR description why none is needed.", len(hits), hits[0], changelog),
			Sources: []string{"facts/changelog"},
			Verdict: VerdictConfirmed,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].File < out[j].File })
	return out
}

func lineCount(ctx context.Context, dir, rev, p string) int {
	if p == "" {
		return 0
	}
	out, err := git(ctx, dir, "show", rev+":"+p)
	if err != nil {
		return 0
	}
	return strings.Count(out, "\n")
}

// call is one unit of lens work.
type call struct {
	lens LensConfig
	unit string
	run  func(ctx context.Context) ([]Finding, float64, error)
}

func (p *Pipeline) runLenses(ctx context.Context, in Input, ch change, bundles []diff.Bundle, log *slog.Logger) ([]Finding, []LensRun) {
	var calls []call
	var runs []LensRun
	for _, l := range p.Config.Lenses {
		if l.Disabled {
			continue
		}
		switch l.Kind {
		case KindDiff:
			calls = append(calls, p.diffCall(l, in, ch))
		case KindBundle:
			rules, err := p.Config.loadRules(l.Rules)
			if err != nil {
				runs = append(runs, LensRun{Name: l.Name, Status: LensFailed, Error: err.Error()})
				continue
			}
			for _, b := range bundles {
				matched := matchRules(rules, b)
				if len(matched) == 0 {
					continue
				}
				calls = append(calls, p.bundleCall(l, in, b, matched))
			}
		case KindCommand:
			if _, err := exec.LookPath(l.Command[0]); err != nil {
				status := LensFailed
				if l.Optional {
					status = LensSkipped
				}
				runs = append(runs, LensRun{Name: l.Name, Status: status, Error: l.Command[0] + " not installed"})
				continue
			}
			calls = append(calls, p.commandCall(l, in, ch))
		}
	}

	var (
		mu       sync.Mutex
		findings []Finding
		wg       sync.WaitGroup
		sem      = make(chan struct{}, p.Config.Concurrency)
	)
	for _, c := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			start := time.Now()
			fs, cost, err := c.run(ctx)
			run := LensRun{Name: c.lens.Name, Unit: c.unit, Status: LensOK, Findings: len(fs), CostUSD: cost, Duration: time.Since(start)}
			if err != nil {
				run.Status, run.Error = LensFailed, err.Error()
				log.Warn("lens failed", "lens", c.lens.Name, "unit", c.unit, "err", err)
			}
			mu.Lock()
			defer mu.Unlock()
			findings = append(findings, fs...)
			runs = append(runs, run)
		}()
	}
	wg.Wait()
	// Calls finish in any order; sort so IDs and merges are reproducible.
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Title < b.Title
	})
	sort.SliceStable(runs, func(i, j int) bool {
		if runs[i].Name != runs[j].Name {
			return runs[i].Name < runs[j].Name
		}
		return runs[i].Unit < runs[j].Unit
	})
	return findings, runs
}

func (p *Pipeline) diffCall(l LensConfig, in Input, ch change) call {
	return call{lens: l, unit: "diff", run: func(ctx context.Context) ([]Finding, float64, error) {
		tmpl, err := p.Config.readFile(l.Prompt)
		if err != nil {
			return nil, 0, err
		}
		prompt := fill(tmpl, map[string]string{
			"FILES":       fileList(ch.files),
			"DIFF":        p.inlineDiff(ch.files),
			"TITLE":       in.Title,
			"DESCRIPTION": in.Description,
		})
		return p.ask(ctx, l, "lens:"+l.Name, in.RepoDir, prompt)
	}}
}

func (p *Pipeline) bundleCall(l LensConfig, in Input, b diff.Bundle, rules []Rule) call {
	ids := make([]string, len(rules))
	for i, r := range rules {
		ids[i] = r.ID
	}
	return call{lens: l, unit: b.Key, run: func(ctx context.Context) ([]Finding, float64, error) {
		tmpl, err := p.Config.readFile(l.Prompt)
		if err != nil {
			return nil, 0, err
		}
		prompt := fill(tmpl, map[string]string{
			"BUNDLE": b.Key,
			"FILES":  fileList(b.Files),
			"DIFF":   p.inlineDiff(b.Files),
			"RULES":  renderRules(rules),
		})
		fs, cost, err := p.ask(ctx, l, "lens:"+l.Name+":"+b.Key, in.RepoDir, prompt)
		for i := range fs {
			// Keep citations to rules that were actually offered.
			if len(fs[i].Sources) == 1 && !contains(ids, strings.TrimPrefix(fs[i].Sources[0], l.Name+"/")) {
				fs[i].Sources = []string{l.Name}
			}
		}
		return fs, cost, err
	}}
}

func (p *Pipeline) commandCall(l LensConfig, in Input, ch change) call {
	return call{lens: l, unit: "command", run: func(ctx context.Context) ([]Finding, float64, error) {
		paths := make([]string, len(ch.files))
		for i, f := range ch.files {
			paths[i] = f.Path
		}
		r := strings.NewReplacer("{{base}}", ch.base, "{{head}}", ch.head, "{{files}}", strings.Join(paths, ","))
		argv := make([]string, len(l.Command))
		for i, a := range l.Command {
			argv[i] = r.Replace(a)
		}
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		cmd.Dir = in.RepoDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			return nil, 0, fmt.Errorf("%s: %v: %s", argv[0], err, strings.TrimSpace(stderr.String()))
		}
		fs, err := parseExternal(stdout.Bytes(), l.Name)
		return fs, 0, err
	}}
}

// externalFinding is the code-review skill's finding contract, which
// external tools such as gorev emit.
type externalFinding struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Severity   string `json:"severity"`
	Source     string `json:"source"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

func parseExternal(raw []byte, lens string) ([]Finding, error) {
	raw = bytes.TrimSpace(raw)
	var list []externalFinding
	if len(raw) > 0 && raw[0] == '{' {
		var wrapped struct {
			Findings []externalFinding `json:"findings"`
		}
		if err := json.Unmarshal(raw, &wrapped); err != nil {
			return nil, err
		}
		list = wrapped.Findings
	} else if len(raw) > 0 {
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, err
		}
	}
	out := make([]Finding, 0, len(list))
	for _, e := range list {
		title := e.Title
		if title == "" {
			title = firstSentence(e.Message)
		}
		src := lens
		if e.Source != "" {
			src = lens + "/" + e.Source
		}
		out = append(out, Finding{
			Severity: e.Severity, Category: e.Source, Title: title, File: e.File, Line: e.Line,
			Body: e.Message, Suggestion: e.Suggestion, Sources: []string{src},
		})
	}
	return out, nil
}

// lensOutput is what a model lens returns.
type lensOutput struct {
	Findings []struct {
		File       string `json:"file"`
		Line       int    `json:"line"`
		Severity   string `json:"severity"`
		Category   string `json:"category"`
		Title      string `json:"title"`
		Body       string `json:"body"`
		Suggestion string `json:"suggestion"`
		Source     string `json:"source"`
	} `json:"findings"`
}

func (p *Pipeline) ask(ctx context.Context, l LensConfig, label, dir, prompt string) ([]Finding, float64, error) {
	resp, err := p.Runner.Run(ctx, claude.Request{
		Label:              label,
		Prompt:             prompt,
		AppendSystemPrompt: p.Config.System,
		Dir:                dir,
		Model:              p.Config.modelFor(l.Model),
		Schema:             lensSchema,
	})
	if err != nil {
		return nil, 0, err
	}
	var out lensOutput
	if err := json.Unmarshal(resp.Output, &out); err != nil {
		return nil, resp.CostUSD, fmt.Errorf("%s: decode findings: %w", label, err)
	}
	fs := make([]Finding, 0, len(out.Findings))
	for _, f := range out.Findings {
		src := l.Name
		if f.Source != "" {
			src = l.Name + "/" + f.Source
		}
		cat := f.Category
		if cat == "" {
			cat = l.Name
		}
		fs = append(fs, Finding{
			Severity: f.Severity, Category: cat, Title: f.Title, File: f.File, Line: f.Line,
			Body: f.Body, Suggestion: f.Suggestion, Sources: []string{src},
		})
	}
	return fs, resp.CostUSD, nil
}

// verifyOutput is what the verifier returns.
type verifyOutput struct {
	Results []struct {
		ID       int    `json:"id"`
		Verdict  string `json:"verdict"`
		Reason   string `json:"reason"`
		Severity string `json:"severity"`
	} `json:"results"`
}

// verify asks an independent reviewer to confirm or refute each finding
// against the code. Findings are grouped by file so one call reads the
// relevant code once. A failed call leaves its findings unverified rather
// than dropping them.
func (p *Pipeline) verify(ctx context.Context, in Input, ch change, fs []Finding, log *slog.Logger) ([]Finding, []LensRun) {
	tmpl, err := p.Config.readFile(p.Config.Verify.Prompt)
	if err != nil {
		return fs, []LensRun{{Name: "verify", Status: LensFailed, Error: err.Error()}}
	}
	for i := range fs {
		fs[i].ID = i + 1
	}
	var batches [][]int
	byFile := map[string][]int{}
	var order []string
	for i, f := range fs {
		if f.Verdict == VerdictConfirmed {
			continue
		}
		if _, ok := byFile[f.File]; !ok {
			order = append(order, f.File)
		}
		byFile[f.File] = append(byFile[f.File], i)
	}
	var cur []int
	for _, file := range order {
		for _, idx := range byFile[file] {
			if len(cur) == p.Config.Verify.Batch {
				batches = append(batches, cur)
				cur = nil
			}
			cur = append(cur, idx)
		}
	}
	if len(cur) > 0 {
		batches = append(batches, cur)
	}

	byPath := map[string]diff.File{}
	for _, f := range ch.files {
		byPath[f.Path] = f
	}

	var (
		mu   sync.Mutex
		runs []LensRun
		wg   sync.WaitGroup
		sem  = make(chan struct{}, p.Config.Concurrency)
	)
	for n, batch := range batches {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			type item struct {
				ID         int    `json:"id"`
				File       string `json:"file"`
				Line       int    `json:"line"`
				Severity   string `json:"severity"`
				Title      string `json:"title"`
				Body       string `json:"body"`
				Suggestion string `json:"suggestion,omitempty"`
			}
			items := make([]item, len(batch))
			var files []diff.File
			seen := map[string]bool{}
			for i, idx := range batch {
				f := fs[idx]
				items[i] = item{f.ID, f.File, f.Line, f.Severity, f.Title, f.Body, f.Suggestion}
				if df, ok := byPath[f.File]; ok && !seen[f.File] {
					files = append(files, df)
					seen[f.File] = true
				}
			}
			list, _ := json.MarshalIndent(items, "", "  ")
			prompt := fill(tmpl, map[string]string{
				"FINDINGS": string(list),
				"DIFF":     p.inlineDiff(files),
			})
			label := fmt.Sprintf("verify:%d", n+1)
			start := time.Now()
			resp, err := p.Runner.Run(ctx, claude.Request{
				Label: label, Prompt: prompt, AppendSystemPrompt: p.Config.System,
				Dir: in.RepoDir, Model: p.Config.modelFor(p.Config.Verify.Model), Schema: verifySchema,
			})
			run := LensRun{Name: "verify", Unit: label, Status: LensOK, Findings: len(batch), CostUSD: resp.CostUSD, Duration: time.Since(start)}
			var out verifyOutput
			if err == nil {
				err = json.Unmarshal(resp.Output, &out)
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				run.Status, run.Error = LensFailed, err.Error()
				log.Warn("verify failed", "batch", label, "err", err)
				runs = append(runs, run)
				return
			}
			runs = append(runs, run)
			valid := map[int]int{}
			for _, idx := range batch {
				valid[fs[idx].ID] = idx
			}
			for _, r := range out.Results {
				idx, ok := valid[r.ID]
				if !ok {
					continue
				}
				switch r.Verdict {
				case VerdictConfirmed, VerdictRefuted, VerdictUncertain:
					fs[idx].Verdict = r.Verdict
				default:
					continue
				}
				fs[idx].VerifyNote = r.Reason
				if r.Verdict == VerdictConfirmed && r.Severity != "" {
					fs[idx].Severity = normalizeSeverity(r.Severity)
				}
			}
		}()
	}
	wg.Wait()
	sort.Slice(runs, func(i, j int) bool { return runs[i].Unit < runs[j].Unit })
	return fs, runs
}

// inlineDiff returns the patches of files, stopping at the configured byte
// budget and listing the rest for the reviewer to Read.
func (p *Pipeline) inlineDiff(files []diff.File) string {
	var sb strings.Builder
	var omitted []string
	for _, f := range files {
		if sb.Len()+len(f.Patch) > p.Config.MaxDiffBytes {
			omitted = append(omitted, f.Path)
			continue
		}
		sb.WriteString(f.Patch)
	}
	if len(omitted) > 0 {
		fmt.Fprintf(&sb, "\n[diff truncated: patches for %d file(s) omitted; Read them directly: %s]\n",
			len(omitted), strings.Join(omitted, ", "))
	}
	return sb.String()
}

func fileList(files []diff.File) string {
	var sb strings.Builder
	for _, f := range files {
		state := ""
		if f.Deleted {
			state = " (deleted)"
		} else if f.OldPath != "" && f.OldPath != f.Path {
			state = " (renamed from " + f.OldPath + ")"
		}
		fmt.Fprintf(&sb, "- %s (+%d/-%d)%s\n", f.Path, f.Added, f.Removed, state)
	}
	return sb.String()
}

// fill substitutes {{KEY}} placeholders. Values are inserted once, so
// placeholder-like text inside the diff is never expanded.
func fill(tmpl string, vals map[string]string) string {
	pairs := make([]string, 0, 2*len(vals))
	for k, v := range vals {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl)
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	if strings.HasPrefix(p, "a/") || strings.HasPrefix(p, "b/") {
		p = p[2:]
	}
	return p
}

func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, ".\n"); i > 0 {
		s = s[:i]
	}
	if len(s) > 100 {
		s = s[:100] + "…"
	}
	return s
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
