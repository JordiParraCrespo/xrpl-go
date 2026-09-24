// Package claude runs Claude Code headless (`claude -p`) and returns its
// structured output.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ReadOnlyTools is the tool set used for every review and triage call. Nothing
// here can run code, write files, or reach the network, so untrusted PR and
// issue content cannot turn the model into an actor.
var ReadOnlyTools = []string{"Read", "Grep", "Glob"}

// Request is one headless Claude call.
type Request struct {
	// Label names the call in logs and errors, e.g. "lens:correctness".
	Label string
	// Prompt is sent on stdin.
	Prompt string
	// AppendSystemPrompt is appended to Claude Code's system prompt.
	AppendSystemPrompt string
	// Dir is the working directory; file tools are confined to it.
	Dir string
	// Model is a model alias or full ID; empty uses the CLI default.
	Model string
	// Schema is a JSON Schema the final answer must satisfy.
	Schema json.RawMessage
	// Tools lists the built-in tools available; nil means ReadOnlyTools.
	Tools []string
	// Timeout bounds the call; zero means no extra bound beyond ctx.
	Timeout time.Duration
}

// Response is the parsed result of a call.
type Response struct {
	// Output is the structured output validated against Request.Schema.
	Output json.RawMessage
	// Text is the final text answer.
	Text     string
	CostUSD  float64
	NumTurns int
	Duration time.Duration
}

// Runner runs Claude calls. The CLI implementation is used in production and
// a fake in tests.
type Runner interface {
	Run(ctx context.Context, req Request) (Response, error)
}

// CLI runs the `claude` binary.
type CLI struct {
	// Bin is the claude binary; empty means "claude" on PATH.
	Bin string
	// Env is appended to the process environment, e.g. CLAUDE_CODE_OAUTH_TOKEN.
	Env []string
}

// ErrCall is returned when the CLI reports a failed call.
var ErrCall = errors.New("claude call failed")

type cliResult struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"`
	TotalCostUSD     float64         `json:"total_cost_usd"`
	NumTurns         int             `json:"num_turns"`
}

// Args returns the CLI arguments for req. Exported for tests and dry runs.
func Args(req Request) []string {
	tools := req.Tools
	if tools == nil {
		tools = ReadOnlyTools
	}
	args := []string{
		"-p",
		"--output-format", "json",
		// Restricted mode drops code-running tools, confines file tools to
		// Dir, and ignores project settings a PR could have modified.
		"--restricted",
		"--strict-mcp-config",
		"--tools", strings.Join(tools, ","),
		"--permission-mode", "dontAsk",
		"--permission-prompts", "none",
		"--no-session-persistence",
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if len(req.Schema) > 0 {
		args = append(args, "--json-schema", string(req.Schema))
	}
	if req.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", req.AppendSystemPrompt)
	}
	return args
}

// Run implements Runner.
func (c CLI) Run(ctx context.Context, req Request) (Response, error) {
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	bin := c.Bin
	if bin == "" {
		bin = "claude"
	}
	cmd := exec.CommandContext(ctx, bin, Args(req)...)
	cmd.Dir = req.Dir
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Env = append(os.Environ(), c.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	var res cliResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		if runErr != nil {
			return Response{}, fmt.Errorf("%w: %s: %v: %s", ErrCall, req.Label, runErr, tail(stderr.String()))
		}
		return Response{}, fmt.Errorf("%w: %s: decode output: %v", ErrCall, req.Label, err)
	}
	if res.IsError || res.Subtype != "success" {
		return Response{}, fmt.Errorf("%w: %s: %s: %s", ErrCall, req.Label, res.Subtype, tail(res.Result))
	}
	out := Response{
		Output:   res.StructuredOutput,
		Text:     res.Result,
		CostUSD:  res.TotalCostUSD,
		NumTurns: res.NumTurns,
		Duration: elapsed,
	}
	if len(req.Schema) > 0 && len(out.Output) == 0 {
		// Older CLIs return the JSON only as text.
		if !json.Valid([]byte(strings.TrimSpace(res.Result))) {
			return Response{}, fmt.Errorf("%w: %s: no structured output", ErrCall, req.Label)
		}
		out.Output = json.RawMessage(strings.TrimSpace(res.Result))
	}
	return out, nil
}

func tail(s string) string {
	const n = 500
	s = strings.TrimSpace(s)
	if len(s) > n {
		return "…" + s[len(s)-n:]
	}
	return s
}
