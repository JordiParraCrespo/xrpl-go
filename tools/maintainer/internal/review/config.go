// Package review implements the headless review pipeline: collect the diff,
// bundle it, run the review lenses in parallel, verify each finding, position
// it on the diff, and render the result.
package review

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Lens kinds.
const (
	// KindDiff runs one call over the whole diff.
	KindDiff = "diff"
	// KindBundle runs one call per bundle that has at least one matching rule.
	KindBundle = "bundle"
	// KindCommand runs an external tool that prints findings as JSON.
	KindCommand = "command"
)

// Config is .ai-review/pipeline.yaml.
type Config struct {
	Model          string   `yaml:"model"`
	Concurrency    int      `yaml:"concurrency"`
	MaxBundleLines int      `yaml:"max_bundle_lines"`
	MaxBundles     int      `yaml:"max_bundles"`
	MaxFindings    int      `yaml:"max_findings"`
	MaxDiffBytes   int      `yaml:"max_diff_bytes"`
	TimeoutMinutes int      `yaml:"timeout_minutes"`
	Exclude        []string `yaml:"exclude"`
	// System is appended to every call's system prompt.
	System string       `yaml:"system"`
	Lenses []LensConfig `yaml:"lenses"`
	Verify VerifyConfig `yaml:"verify"`
	Facts  FactsConfig  `yaml:"facts"`

	// dir is the directory the config was loaded from; prompt and rule
	// paths resolve against it.
	dir string
}

// LensConfig is one reviewer.
type LensConfig struct {
	Name   string `yaml:"name"`
	Kind   string `yaml:"kind"`
	Prompt string `yaml:"prompt"`
	Model  string `yaml:"model"`
	// Rules is the rule file for bundle lenses.
	Rules string `yaml:"rules"`
	// Command is the argv for command lenses. {{base}}, {{head}} and
	// {{files}} are substituted.
	Command []string `yaml:"command"`
	// Optional lenses are skipped, not failed, when unavailable.
	Optional bool `yaml:"optional"`
	Disabled bool `yaml:"disabled"`
}

// VerifyConfig controls the adversarial verification pass.
type VerifyConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prompt  string `yaml:"prompt"`
	Model   string `yaml:"model"`
	// Batch is the maximum number of findings per verification call.
	Batch int `yaml:"batch"`
}

// FactsConfig controls deterministic checks that need no model.
type FactsConfig struct {
	// FileLineLimit flags files that grow past this many lines.
	FileLineLimit int `yaml:"file_line_limit"`
	// Changelog maps a changelog path to the source globs it covers.
	Changelog map[string][]string `yaml:"changelog"`
	// ChangelogIgnore lists globs that never require a changelog entry.
	ChangelogIgnore []string `yaml:"changelog_ignore"`
}

// ErrConfig is returned for an invalid pipeline configuration.
var ErrConfig = errors.New("invalid review config")

// LoadConfig reads pipeline.yaml from dir.
func LoadConfig(dir string) (*Config, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "pipeline.yaml"))
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfig, err)
	}
	cfg.dir = dir
	cfg.setDefaults()
	return cfg, cfg.validate()
}

// Dir returns the directory the config was loaded from.
func (c *Config) Dir() string { return c.dir }

func (c *Config) setDefaults() {
	if c.Concurrency <= 0 {
		c.Concurrency = 3
	}
	if c.MaxBundleLines <= 0 {
		c.MaxBundleLines = 600
	}
	if c.MaxBundles <= 0 {
		c.MaxBundles = 12
	}
	if c.MaxFindings <= 0 {
		c.MaxFindings = 25
	}
	if c.MaxDiffBytes <= 0 {
		c.MaxDiffBytes = 200_000
	}
	if c.TimeoutMinutes <= 0 {
		c.TimeoutMinutes = 20
	}
	if c.Verify.Batch <= 0 {
		c.Verify.Batch = 8
	}
}

func (c *Config) validate() error {
	seen := map[string]bool{}
	for _, l := range c.Lenses {
		if l.Name == "" {
			return fmt.Errorf("%w: lens without a name", ErrConfig)
		}
		if seen[l.Name] {
			return fmt.Errorf("%w: duplicate lens %q", ErrConfig, l.Name)
		}
		seen[l.Name] = true
		switch l.Kind {
		case KindDiff:
			if l.Prompt == "" {
				return fmt.Errorf("%w: lens %q needs a prompt", ErrConfig, l.Name)
			}
		case KindBundle:
			if l.Prompt == "" || l.Rules == "" {
				return fmt.Errorf("%w: lens %q needs a prompt and rules", ErrConfig, l.Name)
			}
		case KindCommand:
			if len(l.Command) == 0 {
				return fmt.Errorf("%w: lens %q needs a command", ErrConfig, l.Name)
			}
		default:
			return fmt.Errorf("%w: lens %q has unknown kind %q", ErrConfig, l.Name, l.Kind)
		}
	}
	if c.Verify.Enabled && c.Verify.Prompt == "" {
		return fmt.Errorf("%w: verify needs a prompt", ErrConfig)
	}
	return nil
}

func (c *Config) readFile(rel string) (string, error) {
	b, err := os.ReadFile(filepath.Join(c.dir, rel))
	return string(b), err
}

func (c *Config) modelFor(override string) string {
	if override != "" {
		return override
	}
	return c.Model
}
