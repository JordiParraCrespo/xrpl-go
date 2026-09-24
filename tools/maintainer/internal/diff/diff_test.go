package diff

import (
	"testing"
)

const sample = `diff --git a/xrpl/transaction/payment.go b/xrpl/transaction/payment.go
index 111..222 100644
--- a/xrpl/transaction/payment.go
+++ b/xrpl/transaction/payment.go
@@ -10,6 +10,8 @@ func (p *Payment) Validate() (bool, error) {
 	ctx
 	ctx
-	old
+	new1
+	new2
+	new3
 	ctx
 	ctx
@@ -40 +42 @@ func other() {
-	x
+	y
diff --git a/old.md b/new.md
similarity 90%
rename from old.md
rename to new.md
--- a/old.md
+++ b/new.md
@@ -1,2 +1,2 @@
-a
+b
 c
diff --git a/gone.go b/gone.go
deleted file mode 100644
--- a/gone.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package x
-var y
diff --git a/img.png b/img.png
Binary files a/img.png and b/img.png differ
`

func TestParse(t *testing.T) {
	files, err := Parse(sample)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Fatalf("got %d files, want 4", len(files))
	}

	p := files[0]
	if p.Path != "xrpl/transaction/payment.go" || p.Added != 4 || p.Removed != 2 {
		t.Errorf("payment: %+v", p)
	}
	if len(p.Hunks) != 2 || p.Hunks[0].NewStart != 10 || p.Hunks[0].NewLines != 8 || p.Hunks[1].NewStart != 42 || p.Hunks[1].NewLines != 1 {
		t.Errorf("payment hunks: %+v", p.Hunks)
	}
	for _, tc := range []struct {
		line int
		want bool
	}{{9, false}, {10, true}, {17, true}, {18, false}, {42, true}, {43, false}} {
		if got := p.ContainsLine(tc.line); got != tc.want {
			t.Errorf("ContainsLine(%d) = %v, want %v", tc.line, got, tc.want)
		}
	}

	if r := files[1]; r.Path != "new.md" || r.OldPath != "old.md" {
		t.Errorf("rename: %+v", r)
	}
	if d := files[2]; !d.Deleted || d.Path != "gone.go" {
		t.Errorf("delete: %+v", d)
	}
	if b := files[3]; !b.Binary {
		t.Errorf("binary: %+v", b)
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		pattern, path string
		want          bool
	}{
		{"xrpl/transaction/*.go", "xrpl/transaction/payment.go", true},
		{"xrpl/transaction/*.go", "xrpl/transaction/types/amount.go", false},
		{"xrpl/**/*.go", "xrpl/transaction/types/amount.go", true},
		{"xrpl/**/*.go", "xrpl/a.go", true},
		{"**/*_test.go", "a/b/c_test.go", true},
		{"**/*_test.go", "c_test.go", true},
		{"**", "anything/at/all", true},
		{"docs/**", "xrpl/docs/x.md", false},
		{"CHANGELOG.md", "CHANGELOG.md", true},
	}
	for _, tt := range tests {
		if got := Match(tt.pattern, tt.path); got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}

func TestMakeBundles(t *testing.T) {
	files := []File{
		{Path: "xrpl/transaction/payment.go", Added: 300},
		{Path: "xrpl/transaction/payment_test.go", Added: 200},
		{Path: "xrpl/transaction/escrow.go", Added: 200},
		{Path: "docs/a.md", Added: 5},
		{Path: "README.md", Added: 1},
		{Path: "gone.go", Deleted: true},
	}
	bundles := MakeBundles(files, 600)
	keys := map[string][]string{}
	for _, b := range bundles {
		keys[b.Key] = b.Paths()
	}
	want := map[string]int{"(root)": 1, "docs": 1, "xrpl/transaction#1": 2, "xrpl/transaction#2": 1}
	if len(keys) != len(want) {
		t.Fatalf("bundles %v", keys)
	}
	for k, n := range want {
		if len(keys[k]) != n {
			t.Errorf("bundle %s = %v, want %d files", k, keys[k], n)
		}
	}
}
