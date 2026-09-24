package review

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/diff"
)

// Rule is one entry of a rule file. Rules are matched to bundles by path
// before any model call, so a reviewer only sees the rules that apply to the
// files in front of it.
type Rule struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Severity string   `yaml:"severity"`
	Paths    []string `yaml:"paths"`
	Exclude  []string `yaml:"exclude"`
	Text     string   `yaml:"text"`
}

type ruleFile struct {
	Rules []Rule `yaml:"rules"`
}

func (c *Config) loadRules(rel string) ([]Rule, error) {
	raw, err := c.readFile(rel)
	if err != nil {
		return nil, err
	}
	var rf ruleFile
	if err := yaml.Unmarshal([]byte(raw), &rf); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrConfig, rel, err)
	}
	for _, r := range rf.Rules {
		if r.ID == "" || len(r.Paths) == 0 || strings.TrimSpace(r.Text) == "" {
			return nil, fmt.Errorf("%w: %s: rule %q needs id, paths and text", ErrConfig, rel, r.ID)
		}
	}
	return rf.Rules, nil
}

// Applies reports whether the rule covers path p.
func (r Rule) Applies(p string) bool {
	return diff.MatchAny(r.Paths, p) && !diff.MatchAny(r.Exclude, p)
}

// matchRules returns the rules that apply to at least one file of b.
func matchRules(rules []Rule, b diff.Bundle) []Rule {
	var out []Rule
	for _, r := range rules {
		for _, f := range b.Files {
			if r.Applies(f.Path) {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

func renderRules(rules []Rule) string {
	var sb strings.Builder
	for _, r := range rules {
		fmt.Fprintf(&sb, "### %s: %s (default severity: %s)\n\nApplies to: %s\n\n%s\n\n",
			r.ID, r.Title, normalizeSeverity(r.Severity), strings.Join(r.Paths, ", "), strings.TrimSpace(r.Text))
	}
	return sb.String()
}
