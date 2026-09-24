// Package github is a small GitHub REST client covering what the maintainer
// bot needs: pull requests, issues, comments, reviews, labels, permissions,
// GitHub App authentication, and webhook signature checks.
package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the public GitHub API.
const DefaultBaseURL = "https://api.github.com"

// TokenSource returns a token for API calls.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// StaticToken is a personal access token or GITHUB_TOKEN.
type StaticToken string

// Token implements TokenSource.
func (t StaticToken) Token(context.Context) (string, error) { return string(t), nil }

// Client calls the GitHub REST API for one repository.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Tokens  TokenSource
	Owner   string
	Repo    string
}

// NewClient returns a client for owner/repo.
func NewClient(owner, repo string, tokens TokenSource) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		Tokens:  tokens,
		Owner:   owner,
		Repo:    repo,
	}
}

// APIError is a non-2xx response.
type APIError struct {
	Status  int
	Method  string
	Path    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("github %s %s: %d %s", e.Method, e.Path, e.Status, e.Message)
}

// IsNotFound reports whether err is a 404 from the API.
func IsNotFound(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.Status == http.StatusNotFound
}

func (c *Client) repoPath(format string, args ...any) string {
	return fmt.Sprintf("/repos/%s/%s", c.Owner, c.Repo) + fmt.Sprintf(format, args...)
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Tokens != nil {
		tok, err := c.Tokens.Token(ctx)
		if err != nil {
			return fmt.Errorf("github token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &e)
		return &APIError{Status: resp.StatusCode, Method: method, Path: path, Message: e.Message}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// getAll pages through a list endpoint.
func getAll[T any](ctx context.Context, c *Client, path string, maxPages int) ([]T, error) {
	var all []T
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	for page := 1; page <= maxPages; page++ {
		var batch []T
		if err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s%sper_page=100&page=%d", path, sep, page), nil, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

// User is a GitHub account.
type User struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// Ref is one side of a pull request.
type Ref struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	} `json:"repo"`
}

// PullRequest is the subset of a PR the bot uses.
type PullRequest struct {
	Number            int    `json:"number"`
	State             string `json:"state"`
	Title             string `json:"title"`
	Body              string `json:"body"`
	Draft             bool   `json:"draft"`
	HTMLURL           string `json:"html_url"`
	User              User   `json:"user"`
	AuthorAssociation string `json:"author_association"`
	Head              Ref    `json:"head"`
	Base              Ref    `json:"base"`
	Labels            []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// Issue is the subset of an issue the bot uses.
type Issue struct {
	Number            int    `json:"number"`
	State             string `json:"state"`
	Title             string `json:"title"`
	Body              string `json:"body"`
	HTMLURL           string `json:"html_url"`
	User              User   `json:"user"`
	AuthorAssociation string `json:"author_association"`
	PullRequest       *struct {
		URL string `json:"url"`
	} `json:"pull_request,omitempty"`
}

// Comment is an issue or review comment.
type Comment struct {
	ID      int64  `json:"id"`
	Body    string `json:"body"`
	User    User   `json:"user"`
	HTMLURL string `json:"html_url"`
}

// GetPull fetches a pull request.
func (c *Client) GetPull(ctx context.Context, n int) (*PullRequest, error) {
	var pr PullRequest
	return &pr, c.do(ctx, http.MethodGet, c.repoPath("/pulls/%d", n), nil, &pr)
}

// ListIssueComments lists the conversation comments on an issue or PR.
func (c *Client) ListIssueComments(ctx context.Context, n int) ([]Comment, error) {
	return getAll[Comment](ctx, c, c.repoPath("/issues/%d/comments", n), 10)
}

// ListReviewComments lists inline review comments on a PR.
func (c *Client) ListReviewComments(ctx context.Context, n int) ([]Comment, error) {
	return getAll[Comment](ctx, c, c.repoPath("/pulls/%d/comments", n), 10)
}

// CreateIssueComment posts a comment on an issue or PR.
func (c *Client) CreateIssueComment(ctx context.Context, n int, body string) (*Comment, error) {
	var out Comment
	return &out, c.do(ctx, http.MethodPost, c.repoPath("/issues/%d/comments", n), map[string]string{"body": body}, &out)
}

// UpdateIssueComment edits a comment.
func (c *Client) UpdateIssueComment(ctx context.Context, id int64, body string) error {
	return c.do(ctx, http.MethodPatch, c.repoPath("/issues/comments/%d", id), map[string]string{"body": body}, nil)
}

// ReviewComment is an inline comment in a review submission.
type ReviewComment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"`
	Body string `json:"body"`
}

// CreateReview submits a review pinned to commitID.
func (c *Client) CreateReview(ctx context.Context, n int, commitID, body, event string, comments []ReviewComment) error {
	in := map[string]any{"commit_id": commitID, "body": body, "event": event, "comments": comments}
	return c.do(ctx, http.MethodPost, c.repoPath("/pulls/%d/reviews", n), in, nil)
}

// AddLabels adds labels to an issue or PR.
func (c *Client) AddLabels(ctx context.Context, n int, labels []string) error {
	return c.do(ctx, http.MethodPost, c.repoPath("/issues/%d/labels", n), map[string][]string{"labels": labels}, nil)
}

// RemoveLabel removes a label; a missing label is not an error.
func (c *Client) RemoveLabel(ctx context.Context, n int, label string) error {
	err := c.do(ctx, http.MethodDelete, c.repoPath("/issues/%d/labels/%s", n, url.PathEscape(label)), nil, nil)
	if IsNotFound(err) {
		return nil
	}
	return err
}

// ListLabels lists the repository's label names.
func (c *Client) ListLabels(ctx context.Context) ([]string, error) {
	ls, err := getAll[struct {
		Name string `json:"name"`
	}](ctx, c, c.repoPath("/labels"), 5)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = l.Name
	}
	return out, nil
}

// Permission returns a user's permission on the repository: admin,
// maintain, write, triage, read, or none.
func (c *Client) Permission(ctx context.Context, login string) (string, error) {
	var out struct {
		Permission string `json:"permission"`
		RoleName   string `json:"role_name"`
	}
	err := c.do(ctx, http.MethodGet, c.repoPath("/collaborators/%s/permission", url.PathEscape(login)), nil, &out)
	if IsNotFound(err) {
		return "none", nil
	}
	if err != nil {
		return "", err
	}
	if out.RoleName == "maintain" {
		return "maintain", nil
	}
	return out.Permission, nil
}

// SearchIssues runs an issue search scoped to this repository.
func (c *Client) SearchIssues(ctx context.Context, query string, limit int) ([]Issue, error) {
	q := url.QueryEscape(fmt.Sprintf("repo:%s/%s %s", c.Owner, c.Repo, query))
	var out struct {
		Items []Issue `json:"items"`
	}
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/search/issues?q=%s&per_page=%d", q, limit), nil, &out)
	return out.Items, err
}

// VerifySignature checks a webhook's X-Hub-Signature-256 header.
func VerifySignature(secret, body []byte, header string) bool {
	sig, ok := strings.CutPrefix(header, "sha256=")
	if !ok || len(secret) == 0 {
		return false
	}
	got, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}
