// Command review runs the xrpl-go review pipeline on a local checkout and
// writes review.json and review.md. With -publish it posts the result to a
// pull request.
//
//	go run ./cmd/review -repo ../.. -base origin/main -out ../../reviews/my-branch
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/claude"
	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/github"
	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/publish"
	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/review"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "review:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		repo      = flag.String("repo", ".", "checkout to review; its working tree must match -head")
		base      = flag.String("base", "origin/main", "base revision")
		head      = flag.String("head", "HEAD", "head revision")
		configDir = flag.String("config", "", "directory holding pipeline.yaml (default <repo>/.ai-review)")
		out       = flag.String("out", "", "output directory for review.json and review.md (default: stdout summary only)")
		model     = flag.String("model", "", "override the model for every lens")
		noVerify  = flag.Bool("no-verify", false, "skip the verification pass")
		only      = flag.String("only", "", "comma-separated lens names to run")
		pr        = flag.String("publish", "", "post to a pull request, as owner/repo#number (needs GITHUB_TOKEN)")
		claudeBin = flag.String("claude", "claude", "claude CLI binary")
	)
	flag.Parse()

	if *configDir == "" {
		*configDir = filepath.Join(*repo, ".ai-review")
	}
	cfg, err := review.LoadConfig(*configDir)
	if err != nil {
		return err
	}
	if *model != "" {
		cfg.Model = *model
		cfg.Verify.Model = *model
		for i := range cfg.Lenses {
			cfg.Lenses[i].Model = ""
		}
	}
	if *noVerify {
		cfg.Verify.Enabled = false
	}
	if *only != "" {
		keep := map[string]bool{}
		for _, n := range strings.Split(*only, ",") {
			keep[strings.TrimSpace(n)] = true
		}
		for i := range cfg.Lenses {
			cfg.Lenses[i].Disabled = !keep[cfg.Lenses[i].Name]
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := &review.Pipeline{Config: cfg, Runner: claude.CLI{Bin: *claudeBin}, Log: log}
	log.Info("reviewing", "repo", *repo, "base", *base, "head", *head)
	res, err := p.Run(ctx, review.Input{RepoDir: *repo, Base: *base, Head: *head})
	if err != nil {
		return err
	}

	if *out != "" {
		if err := os.MkdirAll(*out, 0o755); err != nil {
			return err
		}
		j, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*out, "review.json"), j, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*out, "review.md"), []byte(review.Markdown(res)), 0o644); err != nil {
			return err
		}
		log.Info("wrote review", "dir", *out)
	} else {
		fmt.Println(review.Markdown(res))
	}

	if *pr != "" {
		owner, name, n, err := parsePR(*pr)
		if err != nil {
			return err
		}
		tok := os.Getenv("GITHUB_TOKEN")
		if tok == "" {
			return fmt.Errorf("-publish needs GITHUB_TOKEN")
		}
		o, err := publish.ToPull(ctx, github.NewClient(owner, name, github.StaticToken(tok)), n, res)
		if err != nil {
			return err
		}
		log.Info("published", "pr", *pr, "inline", o.Inline, "summary_updated", o.SummaryUpdated)
	}
	fmt.Fprintf(os.Stderr, "verdict=%s findings=%d needs_human=%d refuted=%d cost=$%.2f\n",
		res.Verdict, len(res.Findings), len(res.NeedsHuman), len(res.Refuted), res.CostUSD)
	return nil
}

func parsePR(s string) (string, string, int, error) {
	repo, num, ok := strings.Cut(s, "#")
	owner, name, ok2 := strings.Cut(repo, "/")
	n, err := strconv.Atoi(num)
	if !ok || !ok2 || err != nil || owner == "" || name == "" {
		return "", "", 0, fmt.Errorf("-publish wants owner/repo#number, got %q", s)
	}
	return owner, name, n, nil
}
