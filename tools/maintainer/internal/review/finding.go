package review

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"unicode"
)

// Severity levels, shared with the code-review skill.
const (
	SeverityBlocker = "blocker"
	SeverityConcern = "concern"
	SeverityNit     = "nit"
)

// Verification outcomes.
const (
	VerdictConfirmed  = "confirmed"
	VerdictRefuted    = "refuted"
	VerdictUncertain  = "uncertain"
	VerdictUnverified = "unverified"
)

// Finding is one review comment candidate.
type Finding struct {
	ID         int      `json:"id"`
	Severity   string   `json:"severity"`
	Category   string   `json:"category,omitempty"`
	Title      string   `json:"title"`
	File       string   `json:"file,omitempty"`
	Line       int      `json:"line,omitempty"`
	Body       string   `json:"body"`
	Suggestion string   `json:"suggestion,omitempty"`
	Sources    []string `json:"sources"`
	// Verdict is the verification outcome; VerdictUnverified when the
	// verify pass is disabled or the finding is deterministic.
	Verdict    string `json:"verdict"`
	VerifyNote string `json:"verify_note,omitempty"`
	// Inline is true when File:Line falls inside a diff hunk, so GitHub
	// accepts it as an inline comment.
	Inline      bool   `json:"inline"`
	Fingerprint string `json:"fingerprint"`
}

func severityRank(s string) int {
	switch s {
	case SeverityBlocker:
		return 0
	case SeverityConcern:
		return 1
	default:
		return 2
	}
}

func normalizeSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "blocker", "critical", "high":
		return SeverityBlocker
	case "concern", "major", "medium", "warning":
		return SeverityConcern
	default:
		return SeverityNit
	}
}

// fingerprint identifies a finding across runs. It ignores the line number
// so a push that shifts code does not repost the same comment.
func fingerprint(f Finding) string {
	h := sha256.Sum256([]byte(f.File + "\x00" + strings.Join(titleTokens(f.Title), " ")))
	return hex.EncodeToString(h[:6])
}

func titleTokens(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	out := fields[:0]
	for _, f := range fields {
		if len(f) > 2 {
			out = append(out, f)
		}
	}
	return out
}

// similar reports whether two titles share at least half their tokens.
func similar(a, b string) bool {
	ta, tb := titleTokens(a), titleTokens(b)
	if len(ta) == 0 || len(tb) == 0 {
		return false
	}
	set := make(map[string]bool, len(ta))
	for _, t := range ta {
		set[t] = true
	}
	common := 0
	for _, t := range tb {
		if set[t] {
			common++
		}
	}
	return 2*common >= min(len(ta), len(tb))
}

// dedupe merges findings that point at the same place and say the same
// thing. The merged finding keeps the highest severity, the longest body,
// and the union of sources; agreement between lenses raises its rank.
func dedupe(in []Finding) []Finding {
	var out []Finding
	for _, f := range in {
		merged := false
		for i := range out {
			o := &out[i]
			if o.File != f.File || abs(o.Line-f.Line) > 3 {
				continue
			}
			if o.Category != f.Category && !similar(o.Title, f.Title) {
				continue
			}
			if severityRank(f.Severity) < severityRank(o.Severity) {
				o.Severity = f.Severity
			}
			if len(f.Body) > len(o.Body) {
				o.Body, o.Title = f.Body, f.Title
			}
			if o.Suggestion == "" {
				o.Suggestion = f.Suggestion
			}
			o.Sources = unionSorted(o.Sources, f.Sources)
			merged = true
			break
		}
		if !merged {
			out = append(out, f)
		}
	}
	return out
}

func unionSorted(a, b []string) []string {
	set := map[string]bool{}
	for _, s := range append(append([]string{}, a...), b...) {
		set[s] = true
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// rank orders findings by severity, then by how many lenses agree, then by
// location, and assigns stable IDs.
func rank(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		if severityRank(a.Severity) != severityRank(b.Severity) {
			return severityRank(a.Severity) < severityRank(b.Severity)
		}
		if len(a.Sources) != len(b.Sources) {
			return len(a.Sources) > len(b.Sources)
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	for i := range fs {
		fs[i].ID = i + 1
	}
}

// capFindings keeps at most n findings. Blockers and concerns are never
// dropped; nits fill whatever room is left.
func capFindings(fs []Finding, n int) (kept []Finding, dropped int) {
	for _, f := range fs {
		if f.Severity != SeverityNit || len(kept) < n {
			kept = append(kept, f)
			continue
		}
		dropped++
	}
	return kept, dropped
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
