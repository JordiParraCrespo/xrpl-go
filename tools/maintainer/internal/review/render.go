package review

import (
	"fmt"
	"regexp"
	"strings"
)

// SummaryMarker identifies the bot's sticky summary comment on a PR.
const SummaryMarker = "<!-- xrpl-go-ai-review:summary -->"

var fingerprintRe = regexp.MustCompile(`<!-- xrpl-go-ai-review:fp=([0-9a-f]+) -->`)

// FingerprintsIn extracts finding fingerprints from posted comment bodies.
func FingerprintsIn(bodies []string) map[string]bool {
	out := map[string]bool{}
	for _, b := range bodies {
		for _, m := range fingerprintRe.FindAllStringSubmatch(b, -1) {
			out[m[1]] = true
		}
	}
	return out
}

func marker(f Finding) string {
	return fmt.Sprintf("<!-- xrpl-go-ai-review:fp=%s -->", f.Fingerprint)
}

func icon(sev string) string {
	switch sev {
	case SeverityBlocker:
		return "🔴"
	case SeverityConcern:
		return "🟡"
	default:
		return "🔵"
	}
}

func label(sev string) string {
	switch sev {
	case SeverityBlocker:
		return "Blocker"
	case SeverityConcern:
		return "Concern"
	default:
		return "Nit"
	}
}

func counts(fs []Finding) string {
	var b, c, n int
	for _, f := range fs {
		switch f.Severity {
		case SeverityBlocker:
			b++
		case SeverityConcern:
			c++
		default:
			n++
		}
	}
	return fmt.Sprintf("%d blocker(s) · %d concern(s) · %d nit(s)", b, c, n)
}

// commentBody is the text posted for one finding. Internal metadata
// (sources, agreement, IDs) never appears in it.
func commentBody(f Finding) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s **%s: %s**\n\n%s\n", icon(f.Severity), label(f.Severity), f.Title, strings.TrimSpace(f.Body))
	if s := strings.TrimSpace(f.Suggestion); s != "" {
		fmt.Fprintf(&sb, "\n**Suggestion:** %s\n", s)
	}
	sb.WriteString("\n" + marker(f))
	return sb.String()
}

// InlineComment is one review comment anchored to a diff line.
type InlineComment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"`
	Body string `json:"body"`
}

// Publication is what the bot posts for a result.
type Publication struct {
	// Summary replaces the sticky summary comment.
	Summary string
	// Inline holds new inline comments; findings already posted on an
	// earlier run (same fingerprint) are skipped.
	Inline []InlineComment
}

// Publish builds the GitHub payload for res. posted holds fingerprints
// already on the PR.
func Publish(res *Result, posted map[string]bool) Publication {
	var pub Publication
	var body []Finding
	for _, f := range res.Findings {
		if !f.Inline {
			body = append(body, f)
			continue
		}
		if posted[f.Fingerprint] {
			continue
		}
		pub.Inline = append(pub.Inline, InlineComment{Path: f.File, Line: f.Line, Side: "RIGHT", Body: commentBody(f)})
	}

	var sb strings.Builder
	sb.WriteString(SummaryMarker + "\n")
	fmt.Fprintf(&sb, "### AI review of `%s`\n\n", short(res.Head))
	switch res.Verdict {
	case VerdictClean:
		sb.WriteString("No issues found.\n")
	default:
		fmt.Fprintf(&sb, "%s\n", counts(res.Findings))
	}

	if len(body) > 0 {
		sb.WriteString("\n")
		for _, f := range body {
			loc := ""
			if f.File != "" {
				loc = "`" + f.File
				if f.Line > 0 {
					loc += fmt.Sprintf(":%d", f.Line)
				}
				loc += "` (outside the diff): "
			}
			fmt.Fprintf(&sb, "%s **%s: %s**\n%s%s\n", icon(f.Severity), label(f.Severity), f.Title, loc, strings.TrimSpace(f.Body))
			if s := strings.TrimSpace(f.Suggestion); s != "" {
				fmt.Fprintf(&sb, "\n**Suggestion:** %s\n", s)
			}
			sb.WriteString("\n" + marker(f) + "\n\n")
		}
	}

	var inline []Finding
	for _, f := range res.Findings {
		if f.Inline {
			inline = append(inline, f)
		}
	}
	if len(inline) > 0 {
		fmt.Fprintf(&sb, "\n<details><summary>Inline comments (%d)</summary>\n\n", len(inline))
		for _, f := range inline {
			fmt.Fprintf(&sb, "- %s `%s:%d` %s\n", icon(f.Severity), f.File, f.Line, f.Title)
		}
		sb.WriteString("\n</details>\n")
	}
	if len(res.NeedsHuman) > 0 {
		fmt.Fprintf(&sb, "\n<details><summary>Needs a maintainer's judgement (%d)</summary>\n\n", len(res.NeedsHuman))
		sb.WriteString("The verification pass could not confirm or refute these.\n\n")
		for _, f := range res.NeedsHuman {
			fmt.Fprintf(&sb, "- `%s:%d` **%s**: %s\n", f.File, f.Line, f.Title, oneLine(f.VerifyNote))
		}
		sb.WriteString("\n</details>\n")
	}
	if res.CappedNits > 0 {
		fmt.Fprintf(&sb, "\n%d more nit(s) omitted.\n", res.CappedNits)
	}
	fmt.Fprintf(&sb, "\n<sub>%s</sub>\n", runLine(res))
	pub.Summary = sb.String()
	return pub
}

func runLine(res *Result) string {
	type agg struct{ ok, skipped, failed int }
	order := []string{}
	byName := map[string]*agg{}
	for _, r := range res.Runs {
		a, ok := byName[r.Name]
		if !ok {
			a = &agg{}
			byName[r.Name] = a
			order = append(order, r.Name)
		}
		switch r.Status {
		case LensOK:
			a.ok++
		case LensSkipped:
			a.skipped++
		default:
			a.failed++
		}
	}
	parts := make([]string, 0, len(order))
	for _, n := range order {
		a := byName[n]
		s := n
		switch {
		case a.failed > 0:
			s += fmt.Sprintf(" (%d failed)", a.failed)
		case a.skipped > 0 && a.ok == 0:
			s += " (skipped)"
		case a.ok > 1:
			s += fmt.Sprintf(" ×%d", a.ok)
		}
		parts = append(parts, s)
	}
	return fmt.Sprintf("Lenses: %s · %d file(s) · %d refuted by verification",
		strings.Join(parts, ", "), res.FilesReviewed, len(res.Refuted))
}

// Markdown renders a full local report, metadata included, for the
// multi-review walkthrough and for humans reading review.md.
func Markdown(res *Result) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Review of `%s` against `%s`\n\n", short(res.Head), short(res.Base))
	fmt.Fprintf(&sb, "Verdict: **%s** · %s · %d file(s)\n\n", res.Verdict, counts(res.Findings), res.FilesReviewed)
	for _, n := range res.Notes {
		fmt.Fprintf(&sb, "> %s\n", n)
	}
	if len(res.Notes) > 0 {
		sb.WriteString("\n")
	}
	section := func(title string, fs []Finding) {
		if len(fs) == 0 {
			return
		}
		fmt.Fprintf(&sb, "## %s\n\n", title)
		for _, f := range fs {
			fmt.Fprintf(&sb, "### %d. %s %s: %s\n\n", f.ID, icon(f.Severity), label(f.Severity), f.Title)
			if f.File != "" {
				where := "outside the diff"
				if f.Inline {
					where = "inline"
				}
				fmt.Fprintf(&sb, "`%s:%d` (%s) · sources: %s · verification: %s\n\n", f.File, f.Line, where, strings.Join(f.Sources, ", "), f.Verdict)
			}
			sb.WriteString(strings.TrimSpace(f.Body) + "\n\n")
			if f.Suggestion != "" {
				fmt.Fprintf(&sb, "**Suggestion:** %s\n\n", strings.TrimSpace(f.Suggestion))
			}
			if f.VerifyNote != "" {
				fmt.Fprintf(&sb, "_Verifier:_ %s\n\n", oneLine(f.VerifyNote))
			}
		}
	}
	section("Findings", res.Findings)
	section("Needs human judgement", res.NeedsHuman)
	section("Refuted by verification", res.Refuted)
	sb.WriteString("## Runs\n\n| lens | unit | status | findings | cost | time |\n|---|---|---|---|---|---|\n")
	for _, r := range res.Runs {
		status := r.Status
		if r.Error != "" {
			status += ": " + oneLine(r.Error)
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | %d | $%.2f | %s |\n", r.Name, r.Unit, status, r.Findings, r.CostUSD, r.Duration.Round(1e9))
	}
	fmt.Fprintf(&sb, "\nTotal cost: $%.2f\n", res.CostUSD)
	return sb.String()
}

func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 300 {
		s = s[:300] + "…"
	}
	return s
}
