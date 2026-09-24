// Package publish posts a review result to a pull request: one sticky
// summary comment that is edited in place on every run, plus a COMMENT
// review carrying only inline comments not posted before.
package publish

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/github"
	"github.com/Peersyst/xrpl-go/tools/maintainer/internal/review"
)

// Outcome reports what was posted.
type Outcome struct {
	Inline         int
	SummaryUpdated bool
}

// ToPull publishes res on PR n.
func ToPull(ctx context.Context, gh *github.Client, n int, res *review.Result) (Outcome, error) {
	var out Outcome
	issueComments, err := gh.ListIssueComments(ctx, n)
	if err != nil {
		return out, err
	}
	reviewComments, err := gh.ListReviewComments(ctx, n)
	if err != nil {
		return out, err
	}
	bodies := make([]string, 0, len(issueComments)+len(reviewComments))
	for _, c := range reviewComments {
		bodies = append(bodies, c.Body)
	}
	var summary *github.Comment
	for i, c := range issueComments {
		if strings.Contains(c.Body, review.SummaryMarker) {
			summary = &issueComments[i]
			continue
		}
		bodies = append(bodies, c.Body)
	}
	posted := review.FingerprintsIn(bodies)

	pub := review.Publish(res, posted)
	if len(pub.Inline) > 0 {
		comments := make([]github.ReviewComment, len(pub.Inline))
		for i, c := range pub.Inline {
			comments[i] = github.ReviewComment{Path: c.Path, Line: c.Line, Side: c.Side, Body: c.Body}
		}
		err := gh.CreateReview(ctx, n, res.Head, "", "COMMENT", comments)
		var ae *github.APIError
		switch {
		case err == nil:
			out.Inline = len(comments)
		case errors.As(err, &ae) && ae.Status == http.StatusUnprocessableEntity:
			// A line GitHub cannot resolve (e.g. the head moved). Fall back
			// to listing everything in the summary.
			for i := range res.Findings {
				res.Findings[i].Inline = false
			}
			pub = review.Publish(res, posted)
		default:
			return out, err
		}
	}

	if summary != nil {
		if err := gh.UpdateIssueComment(ctx, summary.ID, pub.Summary); err == nil {
			out.SummaryUpdated = true
			return out, nil
		}
	}
	_, err = gh.CreateIssueComment(ctx, n, pub.Summary)
	return out, err
}
