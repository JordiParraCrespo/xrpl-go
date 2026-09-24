// Package diff parses unified diffs into per-file hunks and groups changed
// files into review bundles.
package diff

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Hunk is a changed range on the new side of a file.
type Hunk struct {
	NewStart int
	NewLines int
	Header   string
}

// Contains reports whether line falls inside the hunk on the new side.
func (h Hunk) Contains(line int) bool {
	return line >= h.NewStart && line < h.NewStart+h.NewLines
}

// File is one file in a unified diff.
type File struct {
	Path    string
	OldPath string
	Deleted bool
	Binary  bool
	Added   int
	Removed int
	Hunks   []Hunk
	// Patch is the file's section of the diff, headers included.
	Patch string
}

// ContainsLine reports whether line is inside a hunk of f, which is where
// GitHub accepts inline review comments.
func (f File) ContainsLine(line int) bool {
	for _, h := range f.Hunks {
		if h.Contains(line) {
			return true
		}
	}
	return false
}

// Parse parses the output of `git diff` (unified format).
func Parse(raw string) ([]File, error) {
	var (
		files []File
		cur   *File
		patch strings.Builder
	)
	flush := func() {
		if cur == nil {
			return
		}
		cur.Patch = patch.String()
		files = append(files, *cur)
		cur = nil
		patch.Reset()
	}

	sc := bufio.NewScanner(strings.NewReader(raw))
	sc.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			cur = &File{}
			a, b, ok := splitGitHeader(strings.TrimPrefix(line, "diff --git "))
			if ok {
				cur.OldPath, cur.Path = a, b
			}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "--- "):
			if p := stripPrefix(strings.TrimPrefix(line, "--- ")); p != "" {
				cur.OldPath = p
			}
		case strings.HasPrefix(line, "+++ "):
			p := stripPrefix(strings.TrimPrefix(line, "+++ "))
			if p == "" {
				cur.Deleted = true
			} else {
				cur.Path = p
			}
		case strings.HasPrefix(line, "Binary files "):
			cur.Binary = true
		case strings.HasPrefix(line, "deleted file mode"):
			cur.Deleted = true
		case strings.HasPrefix(line, "@@"):
			h, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cur.Path, err)
			}
			cur.Hunks = append(cur.Hunks, h)
		case strings.HasPrefix(line, "+"):
			cur.Added++
		case strings.HasPrefix(line, "-"):
			cur.Removed++
		}
		if cur != nil {
			patch.WriteString(line)
			patch.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	flush()
	return files, nil
}

// stripPrefix removes the a/ or b/ prefix and maps /dev/null to "".
func stripPrefix(p string) string {
	p = strings.TrimSpace(p)
	if i := strings.IndexByte(p, '\t'); i >= 0 {
		p = p[:i]
	}
	if p == "/dev/null" {
		return ""
	}
	if strings.HasPrefix(p, "a/") || strings.HasPrefix(p, "b/") {
		return p[2:]
	}
	return p
}

// splitGitHeader splits "a/x b/x". Paths with spaces are resolved by the
// ---/+++ lines that follow, so a best effort is enough here.
func splitGitHeader(s string) (string, string, bool) {
	i := strings.Index(s, " b/")
	if i < 0 {
		return "", "", false
	}
	return stripPrefix(s[:i]), stripPrefix(s[i+1:]), true
}

// parseHunkHeader parses "@@ -a,b +c,d @@ header".
func parseHunkHeader(line string) (Hunk, error) {
	parts := strings.SplitN(line, "@@", 3)
	if len(parts) < 3 {
		return Hunk{}, fmt.Errorf("malformed hunk header %q", line)
	}
	var h Hunk
	h.Header = strings.TrimSpace(parts[2])
	for _, f := range strings.Fields(parts[1]) {
		if !strings.HasPrefix(f, "+") {
			continue
		}
		start, count, err := parseRange(f[1:])
		if err != nil {
			return Hunk{}, fmt.Errorf("hunk header %q: %w", line, err)
		}
		h.NewStart, h.NewLines = start, count
	}
	return h, nil
}

func parseRange(s string) (int, int, error) {
	startStr, countStr, hasCount := strings.Cut(s, ",")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return 0, 0, err
	}
	if !hasCount {
		return start, 1, nil
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, 0, err
	}
	return start, count, nil
}
