// Package store keeps the maintainer bot's state in SQLite: every tracked
// issue and PR, who approved which commit for review, and the job queue.
// It is the seed of the maintenance CRM; GitHub stays the source of truth
// for content, the store tracks the maintenance workflow around it.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// Item kinds.
const (
	KindIssue = "issue"
	KindPull  = "pr"
)

// Item statuses.
const (
	StatusOpen             = "open"
	StatusDraft            = "draft"
	StatusTriaged          = "triaged"
	StatusAwaitingApproval = "awaiting-approval"
	StatusQueued           = "queued"
	StatusReviewing        = "reviewing"
	StatusReviewed         = "reviewed"
	StatusFailed           = "failed"
	StatusClosed           = "closed"
	StatusIgnored          = "ignored"
)

// Job kinds and statuses.
const (
	JobReview = "review"
	JobTriage = "triage"

	JobQueued     = "queued"
	JobRunning    = "running"
	JobDone       = "done"
	JobFailed     = "failed"
	JobSuperseded = "superseded"
	JobCanceled   = "canceled"
)

// Item is a tracked issue or pull request.
type Item struct {
	Kind            string
	Number          int
	Title           string
	Author          string
	URL             string
	Status          string
	Trusted         bool
	HeadSHA         string
	ApprovedSHA     string
	ApprovedBy      string
	LastReviewSHA   string
	Summary         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FirstResponseAt *time.Time
}

// Job is a queued unit of work.
type Job struct {
	ID         int64
	Kind       string
	Number     int
	SHA        string
	Status     string
	Attempts   int
	Error      string
	Result     string
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}

// Store wraps the database.
type Store struct {
	db *sql.DB
}

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

const schema = `
CREATE TABLE IF NOT EXISTS deliveries (
  id TEXT PRIMARY KEY,
  event TEXT NOT NULL,
  received_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
  kind TEXT NOT NULL,
  number INTEGER NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  author TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  trusted INTEGER NOT NULL DEFAULT 0,
  head_sha TEXT NOT NULL DEFAULT '',
  approved_sha TEXT NOT NULL DEFAULT '',
  approved_by TEXT NOT NULL DEFAULT '',
  last_review_sha TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  first_response_at INTEGER,
  PRIMARY KEY (kind, number)
);
CREATE TABLE IF NOT EXISTS jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL,
  number INTEGER NOT NULL,
  sha TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  result TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL,
  started_at INTEGER,
  finished_at INTEGER
);
CREATE INDEX IF NOT EXISTS jobs_status ON jobs (kind, status, id);
`

// Open opens (and migrates) the database at path.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// SeenDelivery records a webhook delivery ID and reports whether it was
// already processed. GitHub redelivers on timeouts.
func (s *Store) SeenDelivery(ctx context.Context, id, event string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO deliveries (id, event, received_at) VALUES (?, ?, ?)`,
		id, event, time.Now().Unix())
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 0, err
}

// UpsertItem inserts or updates an item's descriptive fields. Workflow
// fields (status, approval) are only set on insert; use the dedicated
// methods to change them.
func (s *Store) UpsertItem(ctx context.Context, it Item) error {
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx, `
INSERT INTO items (kind, number, title, author, url, status, trusted, head_sha, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (kind, number) DO UPDATE SET
  title = excluded.title, author = excluded.author, url = excluded.url,
  trusted = excluded.trusted,
  head_sha = CASE WHEN excluded.head_sha != '' THEN excluded.head_sha ELSE items.head_sha END,
  updated_at = excluded.updated_at`,
		it.Kind, it.Number, it.Title, it.Author, it.URL, it.Status, it.Trusted, it.HeadSHA, now, now)
	return err
}

// SetStatus sets an item's status.
func (s *Store) SetStatus(ctx context.Context, kind string, n int, status string) error {
	return s.exec1(ctx, `UPDATE items SET status = ?, updated_at = ? WHERE kind = ? AND number = ?`,
		status, time.Now().Unix(), kind, n)
}

// SetSummary stores a one-line summary (triage result or review verdict).
func (s *Store) SetSummary(ctx context.Context, kind string, n int, summary string) error {
	return s.exec1(ctx, `UPDATE items SET summary = ?, updated_at = ? WHERE kind = ? AND number = ?`,
		summary, time.Now().Unix(), kind, n)
}

// Approve records that approver allowed review of sha on PR n.
func (s *Store) Approve(ctx context.Context, n int, sha, approver string) error {
	return s.exec1(ctx, `UPDATE items SET approved_sha = ?, approved_by = ?, updated_at = ? WHERE kind = ? AND number = ?`,
		sha, approver, time.Now().Unix(), KindPull, n)
}

// RevokeApproval clears a PR's approval, e.g. after new commits.
func (s *Store) RevokeApproval(ctx context.Context, n int) error {
	return s.Approve(ctx, n, "", "")
}

// MarkReviewed records the SHA the last review covered.
func (s *Store) MarkReviewed(ctx context.Context, n int, sha, summary string) error {
	return s.exec1(ctx, `UPDATE items SET last_review_sha = ?, summary = ?, status = ?, updated_at = ? WHERE kind = ? AND number = ?`,
		sha, summary, StatusReviewed, time.Now().Unix(), KindPull, n)
}

// MarkResponded sets first_response_at once.
func (s *Store) MarkResponded(ctx context.Context, kind string, n int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE items SET first_response_at = COALESCE(first_response_at, ?) WHERE kind = ? AND number = ?`,
		time.Now().Unix(), kind, n)
	return err
}

// GetItem returns one item.
func (s *Store) GetItem(ctx context.Context, kind string, n int) (*Item, error) {
	rows, err := s.db.QueryContext(ctx, itemSelect+` WHERE kind = ? AND number = ?`, kind, n)
	if err != nil {
		return nil, err
	}
	items, err := scanItems(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return &items[0], nil
}

// ListItems returns items, open ones first, most recently updated first.
func (s *Store) ListItems(ctx context.Context, kind string, includeClosed bool, limit int) ([]Item, error) {
	q := itemSelect + ` WHERE kind = ?`
	if !includeClosed {
		q += ` AND status NOT IN ('closed', 'ignored')`
	}
	q += ` ORDER BY updated_at DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, q, kind, limit)
	if err != nil {
		return nil, err
	}
	return scanItems(rows)
}

const itemSelect = `SELECT kind, number, title, author, url, status, trusted, head_sha, approved_sha, approved_by,
  last_review_sha, summary, created_at, updated_at, first_response_at FROM items`

func scanItems(rows *sql.Rows) ([]Item, error) {
	defer rows.Close()
	var out []Item
	for rows.Next() {
		var (
			it            Item
			created, upd  int64
			firstResponse sql.NullInt64
		)
		if err := rows.Scan(&it.Kind, &it.Number, &it.Title, &it.Author, &it.URL, &it.Status, &it.Trusted,
			&it.HeadSHA, &it.ApprovedSHA, &it.ApprovedBy, &it.LastReviewSHA, &it.Summary, &created, &upd, &firstResponse); err != nil {
			return nil, err
		}
		it.CreatedAt, it.UpdatedAt = time.Unix(created, 0), time.Unix(upd, 0)
		if firstResponse.Valid {
			t := time.Unix(firstResponse.Int64, 0)
			it.FirstResponseAt = &t
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Enqueue adds a job. A queued review for the same PR is superseded, so
// only the newest approved commit is reviewed.
func (s *Store) Enqueue(ctx context.Context, kind string, n int, sha string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit
	if kind == JobReview {
		if _, err := tx.ExecContext(ctx, `UPDATE jobs SET status = ?, finished_at = ? WHERE kind = ? AND number = ? AND status = ?`,
			JobSuperseded, time.Now().Unix(), kind, n, JobQueued); err != nil {
			return 0, err
		}
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO jobs (kind, number, sha, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		kind, n, sha, JobQueued, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// CancelQueued cancels queued jobs of kind for item n.
func (s *Store) CancelQueued(ctx context.Context, kind string, n int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jobs SET status = ?, finished_at = ? WHERE kind = ? AND number = ? AND status = ?`,
		JobCanceled, time.Now().Unix(), kind, n, JobQueued)
	return err
}

// Claim takes the oldest queued job of kind and marks it running. It
// returns ErrNotFound when the queue is empty.
func (s *Store) Claim(ctx context.Context, kind string) (*Job, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit
	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM jobs WHERE kind = ? AND status = ? ORDER BY id LIMIT 1`, kind, JobQueued).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jobs SET status = ?, attempts = attempts + 1, started_at = ? WHERE id = ?`,
		JobRunning, time.Now().Unix(), id); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetJob(ctx, id)
}

// Finish records a job's outcome.
func (s *Store) Finish(ctx context.Context, id int64, status, result, errMsg string) error {
	return s.exec1(ctx, `UPDATE jobs SET status = ?, result = ?, error = ?, finished_at = ? WHERE id = ?`,
		status, result, errMsg, time.Now().Unix(), id)
}

// RequeueRunning puts jobs left running by a crash back in the queue.
func (s *Store) RequeueRunning(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jobs SET status = ? WHERE status = ?`, JobQueued, JobRunning)
	return err
}

// GetJob returns one job.
func (s *Store) GetJob(ctx context.Context, id int64) (*Job, error) {
	jobs, err := s.queryJobs(ctx, jobSelect+` WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, ErrNotFound
	}
	return &jobs[0], nil
}

// RecentJobs returns the newest jobs.
func (s *Store) RecentJobs(ctx context.Context, limit int) ([]Job, error) {
	return s.queryJobs(ctx, jobSelect+` ORDER BY id DESC LIMIT ?`, limit)
}

const jobSelect = `SELECT id, kind, number, sha, status, attempts, error, result, created_at, started_at, finished_at FROM jobs`

func (s *Store) queryJobs(ctx context.Context, q string, args ...any) ([]Job, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		var (
			j                 Job
			created           int64
			started, finished sql.NullInt64
		)
		if err := rows.Scan(&j.ID, &j.Kind, &j.Number, &j.SHA, &j.Status, &j.Attempts, &j.Error, &j.Result,
			&created, &started, &finished); err != nil {
			return nil, err
		}
		j.CreatedAt = time.Unix(created, 0)
		if started.Valid {
			t := time.Unix(started.Int64, 0)
			j.StartedAt = &t
		}
		if finished.Valid {
			t := time.Unix(finished.Int64, 0)
			j.FinishedAt = &t
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *Store) exec1(ctx context.Context, q string, args ...any) error {
	res, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
