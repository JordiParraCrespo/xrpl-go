package diff

import (
	"path"
	"sort"
	"strconv"
	"strings"
)

// Bundle is a group of related changed files reviewed together. Grouping
// follows Open Code Review's idea: a reviewer sees a package and its tests
// as one unit, and a large changeset becomes several bounded units.
type Bundle struct {
	Key   string
	Files []File
}

// Lines is the number of changed lines in the bundle.
func (b Bundle) Lines() int {
	n := 0
	for _, f := range b.Files {
		n += f.Added + f.Removed
	}
	return n
}

// Paths returns the bundle's file paths.
func (b Bundle) Paths() []string {
	out := make([]string, len(b.Files))
	for i, f := range b.Files {
		out[i] = f.Path
	}
	return out
}

// MakeBundles groups files by Go package directory (non-Go files by their
// top-level directory) and splits any group above maxLines changed lines.
func MakeBundles(files []File, maxLines int) []Bundle {
	groups := map[string][]File{}
	for _, f := range files {
		if f.Deleted || f.Binary {
			continue
		}
		groups[bundleKey(f.Path)] = append(groups[bundleKey(f.Path)], f)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []Bundle
	for _, k := range keys {
		fs := groups[k]
		sort.Slice(fs, func(i, j int) bool { return fs[i].Path < fs[j].Path })
		out = append(out, split(k, fs, maxLines)...)
	}
	return out
}

func bundleKey(p string) string {
	if strings.HasSuffix(p, ".go") {
		if d := path.Dir(p); d != "." {
			return d
		}
		return "(root)"
	}
	if i := strings.IndexByte(p, '/'); i > 0 {
		return p[:i]
	}
	return "(root)"
}

func split(key string, fs []File, maxLines int) []Bundle {
	if maxLines <= 0 {
		return []Bundle{{Key: key, Files: fs}}
	}
	var (
		out  []Bundle
		cur  []File
		size int
	)
	for _, f := range fs {
		n := f.Added + f.Removed
		if len(cur) > 0 && size+n > maxLines {
			out = append(out, Bundle{Files: cur})
			cur, size = nil, 0
		}
		cur = append(cur, f)
		size += n
	}
	if len(cur) > 0 {
		out = append(out, Bundle{Files: cur})
	}
	for i := range out {
		out[i].Key = key
		if len(out) > 1 {
			out[i].Key = key + "#" + strconv.Itoa(i+1)
		}
	}
	return out
}

// Match reports whether p matches pattern. Patterns use path.Match syntax per
// segment plus "**", which matches any number of segments.
func Match(pattern, p string) bool {
	return matchSegs(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			for i := 0; i <= len(segs); i++ {
				if matchSegs(rest, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, err := path.Match(pat[0], segs[0]); err != nil || !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// MatchAny reports whether p matches any of patterns.
func MatchAny(patterns []string, p string) bool {
	for _, pat := range patterns {
		if Match(pat, p) {
			return true
		}
	}
	return false
}
