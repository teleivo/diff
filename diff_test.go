package diff_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/teleivo/diff"
)

func TestLines(t *testing.T) {
	tests := map[string]struct {
		a    []string
		b    []string
		want []diff.Edit
	}{
		"BothEmpty": {
			a:    []string{},
			b:    []string{},
			want: []diff.Edit{},
		},
		"FirstEmpty": {
			a: []string{},
			b: []string{"A", "B"},
			want: []diff.Edit{
				{Op: diff.Ins, NewLine: "A"},
				{Op: diff.Ins, NewLine: "B"},
			},
		},
		"SecondEmpty": {
			a: []string{"A", "B"},
			b: []string{},
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "A"},
				{Op: diff.Del, OldLine: "B"},
			},
		},
		"Equal": {
			a: []string{"A", "B", "C"},
			b: []string{"A", "B", "C"},
			want: []diff.Edit{
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
			},
		},
		"CompletelyDifferent": {
			a: strings.Split("AB", ""),
			b: strings.Split("CD", ""),
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "A"},
				{Op: diff.Del, OldLine: "B"},
				{Op: diff.Ins, NewLine: "C"},
				{Op: diff.Ins, NewLine: "D"},
			},
		},
		"CommonPrefix": {
			a: strings.Split("ABCX", ""),
			b: strings.Split("ABCY", ""),
			want: []diff.Edit{
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
				{Op: diff.Del, OldLine: "X"},
				{Op: diff.Ins, NewLine: "Y"},
			},
		},
		"CommonSuffix": {
			a: strings.Split("XABC", ""),
			b: strings.Split("YABC", ""),
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "X"},
				{Op: diff.Ins, NewLine: "Y"},
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
			},
		},
		"PaperExample": {
			a: strings.Split("ABCABBA", ""),
			b: strings.Split("CBABAC", ""),
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "A"},
				{Op: diff.Del, OldLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
				{Op: diff.Ins, NewLine: "B"},
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Del, OldLine: "B"},
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Ins, NewLine: "C"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := diff.Lines(test.a, test.b)
			if len(got) != len(test.want) {
				t.Fatalf("diff.Lines(%v, %v) returned %d edits, want %d\ngot:  %v\nwant: %v",
					test.a, test.b, len(got), len(test.want), got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("diff.Lines(%v, %v)[%d] = %v, want %v", test.a, test.b, i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestFiles(t *testing.T) {
	tests := map[string]struct {
		a        string
		b        string
		context  int
		wantDiff bool
		want     string
		wantErr  bool
	}{
		"BothEmpty": {
			a:        "testdata/empty.txt",
			b:        "testdata/empty.txt",
			context:  3,
			wantDiff: false,
			want:     "",
		},
		"Identical": {
			a:        "testdata/one_line.txt",
			b:        "testdata/one_line.txt",
			context:  3,
			wantDiff: false,
			want:     "",
		},
		"OneLineDifferent": {
			a:        "testdata/one_line.txt",
			b:        "testdata/one_line_different.txt",
			context:  3,
			wantDiff: true,
			want: `--- testdata/one_line.txt	2026-02-01 10:38:04.796973594 +0100
+++ testdata/one_line_different.txt	2026-02-01 10:38:10.508955843 +0100
@@ -1 +1 @@
-hello
\ No newline at end of file
+world
\ No newline at end of file
`,
		},
		"MultiLineMiddleChanged": {
			a:        "testdata/multi_line_a.txt",
			b:        "testdata/multi_line_b.txt",
			context:  3,
			wantDiff: true,
			want: `--- testdata/multi_line_a.txt	2026-02-04 09:12:40.837254955 +0100
+++ testdata/multi_line_b.txt	2026-02-04 09:12:40.837254955 +0100
@@ -1,3 +1,3 @@
 line1
-line2
+modified
 line3
`,
		},
		"File1NotFound": {
			a:       "testdata/nonexistent.txt",
			b:       "testdata/empty.txt",
			context: 3,
			wantErr: true,
		},
		"File2NotFound": {
			a:       "testdata/empty.txt",
			b:       "testdata/nonexistent.txt",
			context: 3,
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			hasDiff, err := diff.Files(&buf, test.a, test.b, test.context)
			if test.wantErr {
				if err == nil {
					t.Fatalf("Files() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Files() unexpected error: %v", err)
			}
			if hasDiff != test.wantDiff {
				t.Errorf("Files() hasDiff = %v, want %v", hasDiff, test.wantDiff)
			}
			got := buf.String()
			if got != test.want {
				t.Errorf("Files() =\n%q\nwant:\n%q", got, test.want)
			}
		})
	}
}

func TestWriteUnified(t *testing.T) {
	tests := map[string]struct {
		edits   []diff.Edit
		context int
		want    string
	}{
		"Empty": {
			edits:   nil,
			context: 0,
			want:    "",
		},
		"OnlyEqual": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "same", NewLine: "same"},
			},
			context: 0,
			want:    "",
		},
		"DelStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "removed"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-removed\n",
		},
		"DelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first"},
				{Op: diff.Eq, OldLine: "second", NewLine: "second"},
				{Op: diff.Eq, OldLine: "third", NewLine: "third"},
			},
			context: 1,
			want:    "@@ -1,2 +1 @@\n-first\n second\n",
		},
		"DelMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Del, OldLine: "line2"},
				{Op: diff.Eq, OldLine: "line3", NewLine: "line3"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before", NewLine: "before"},
				{Op: diff.Del, OldLine: "removed"},
				{Op: diff.Eq, OldLine: "after", NewLine: "after"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelMiddleContext5": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before", NewLine: "before"},
				{Op: diff.Del, OldLine: "removed"},
				{Op: diff.Eq, OldLine: "after", NewLine: "after"},
			},
			context: 5,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Del, OldLine: "line2"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first", NewLine: "first"},
				{Op: diff.Eq, OldLine: "second", NewLine: "second"},
				{Op: diff.Del, OldLine: "third"},
			},
			context: 1,
			want:    "@@ -2,2 +2 @@\n second\n-third\n",
		},
		"InsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "added"},
			},
			context: 0,
			want:    "@@ -0,0 +1 @@\n+added\n",
		},
		"InsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "first"},
				{Op: diff.Eq, OldLine: "second", NewLine: "second"},
				{Op: diff.Eq, OldLine: "third", NewLine: "third"},
			},
			context: 1,
			want:    "@@ -1 +1,2 @@\n+first\n second\n",
		},
		"InsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Ins, NewLine: "line2"},
				{Op: diff.Eq, OldLine: "line3", NewLine: "line3"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before", NewLine: "before"},
				{Op: diff.Ins, NewLine: "added"},
				{Op: diff.Eq, OldLine: "after", NewLine: "after"},
			},
			context: 1,
			want:    "@@ -1,2 +1,3 @@\n before\n+added\n after\n",
		},
		"InsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Ins, NewLine: "line2"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first", NewLine: "first"},
				{Op: diff.Eq, OldLine: "second", NewLine: "second"},
				{Op: diff.Ins, NewLine: "third"},
			},
			context: 1,
			want:    "@@ -2 +2,2 @@\n second\n+third\n",
		},
		"DelInsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old"},
				{Op: diff.Ins, NewLine: "new"},
			},
			context: 0,
			want:    "@@ -1 +1 @@\n-old\n+new\n",
		},
		"DelInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old"},
				{Op: diff.Ins, NewLine: "new"},
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-old\n+new\n keep\n",
		},
		"DelInsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1", NewLine: "keep1"},
				{Op: diff.Del, OldLine: "removed"},
				{Op: diff.Ins, NewLine: "added"},
				{Op: diff.Eq, OldLine: "keep2", NewLine: "keep2"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-removed\n+added\n",
		},
		"DelInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1", NewLine: "keep1"},
				{Op: diff.Del, OldLine: "removed"},
				{Op: diff.Ins, NewLine: "added"},
				{Op: diff.Eq, OldLine: "keep2", NewLine: "keep2"},
			},
			context: 1,
			want:    "@@ -1,3 +1,3 @@\n keep1\n-removed\n+added\n keep2\n",
		},
		"DelInsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
				{Op: diff.Del, OldLine: "old"},
				{Op: diff.Ins, NewLine: "new"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-old\n+new\n",
		},
		"DelInsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
				{Op: diff.Del, OldLine: "old"},
				{Op: diff.Ins, NewLine: "new"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n keep\n-old\n+new\n",
		},
		"ConsecDelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
				{Op: diff.Del, OldLine: "del1"},
				{Op: diff.Del, OldLine: "del2"},
				{Op: diff.Del, OldLine: "del3"},
				{Op: diff.Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,5 +1,2 @@\n keep\n-del1\n-del2\n-del3\n end\n",
		},
		"ConsecInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
				{Op: diff.Ins, NewLine: "ins1"},
				{Op: diff.Ins, NewLine: "ins2"},
				{Op: diff.Ins, NewLine: "ins3"},
				{Op: diff.Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,2 +1,5 @@\n keep\n+ins1\n+ins2\n+ins3\n end\n",
		},
		"ConsecDelInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep", NewLine: "keep"},
				{Op: diff.Del, OldLine: "del1"},
				{Op: diff.Del, OldLine: "del2"},
				{Op: diff.Ins, NewLine: "ins1"},
				{Op: diff.Ins, NewLine: "ins2"},
				{Op: diff.Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,4 +1,4 @@\n keep\n-del1\n-del2\n+ins1\n+ins2\n end\n",
		},
		"ConsecDelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "a"},
				{Op: diff.Del, OldLine: "b"},
				{Op: diff.Del, OldLine: "c"},
			},
			context: 1,
			want:    "@@ -1,3 +0,0 @@\n-a\n-b\n-c\n",
		},
		"ConsecInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "a"},
				{Op: diff.Ins, NewLine: "b"},
				{Op: diff.Ins, NewLine: "c"},
			},
			context: 1,
			want:    "@@ -0,0 +1,3 @@\n+a\n+b\n+c\n",
		},
		"TwoHunksSeparateContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Del, OldLine: "del1"},
				{Op: diff.Eq, OldLine: "line2", NewLine: "line2"},
				{Op: diff.Eq, OldLine: "line3", NewLine: "line3"},
				{Op: diff.Eq, OldLine: "line4", NewLine: "line4"},
				{Op: diff.Eq, OldLine: "line5", NewLine: "line5"},
				{Op: diff.Ins, NewLine: "ins1"},
				{Op: diff.Eq, OldLine: "line6", NewLine: "line6"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n line1\n-del1\n line2\n@@ -6,2 +5,3 @@\n line5\n+ins1\n line6\n",
		},
		"TwoHunksMergedContext2": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1", NewLine: "line1"},
				{Op: diff.Del, OldLine: "del1"},
				{Op: diff.Eq, OldLine: "line2", NewLine: "line2"},
				{Op: diff.Eq, OldLine: "line3", NewLine: "line3"},
				{Op: diff.Eq, OldLine: "line4", NewLine: "line4"},
				{Op: diff.Eq, OldLine: "line5", NewLine: "line5"},
				{Op: diff.Ins, NewLine: "ins1"},
				{Op: diff.Eq, OldLine: "line6", NewLine: "line6"},
			},
			context: 2,
			want:    "@@ -1,7 +1,7 @@\n line1\n-del1\n line2\n line3\n line4\n line5\n+ins1\n line6\n",
		},
		"TwoHunksSeparateContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first"},
				{Op: diff.Eq, OldLine: "middle", NewLine: "middle"},
				{Op: diff.Ins, NewLine: "last"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-first\n@@ -2,0 +2 @@\n+last\n",
		},
		"TwoHunksMergedContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first"},
				{Op: diff.Eq, OldLine: "middle", NewLine: "middle"},
				{Op: diff.Ins, NewLine: "last"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-first\n middle\n+last\n",
		},
		"ThreeHunksSeparateContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "del1"},
				{Op: diff.Eq, OldLine: "a", NewLine: "a"},
				{Op: diff.Eq, OldLine: "b", NewLine: "b"},
				{Op: diff.Eq, OldLine: "c", NewLine: "c"},
				{Op: diff.Eq, OldLine: "d", NewLine: "d"},
				{Op: diff.Del, OldLine: "del2"},
				{Op: diff.Eq, OldLine: "e", NewLine: "e"},
				{Op: diff.Eq, OldLine: "f", NewLine: "f"},
				{Op: diff.Eq, OldLine: "g", NewLine: "g"},
				{Op: diff.Eq, OldLine: "h", NewLine: "h"},
				{Op: diff.Ins, NewLine: "ins1"},
			},
			context: 1,
			want:    "@@ -1,2 +1 @@\n-del1\n a\n@@ -5,3 +4,2 @@\n d\n-del2\n e\n@@ -10 +8,2 @@\n h\n+ins1\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := diff.WriteUnified(&buf, test.edits, test.context)
			if err != nil {
				t.Fatalf("diff.WriteUnified() error: %v", err)
			}
			got := buf.String()
			if got != test.want {
				t.Errorf("diff.WriteUnified() =\n%q\nwant:\n%q", got, test.want)
			}
		})
	}
}

func TestWriteFileHeader(t *testing.T) {
	oldTime := time.Date(2026, 2, 4, 8, 12, 16, 2963487, time.FixedZone("CET", 3600))
	newTime := time.Date(2026, 2, 4, 9, 30, 45, 123456789, time.FixedZone("CET", 3600))
	want := "--- a.txt\t2026-02-04 08:12:16.002963487 +0100\n+++ b.txt\t2026-02-04 09:30:45.123456789 +0100\n"

	var buf bytes.Buffer
	err := diff.WriteFileHeader(&buf, "a.txt", oldTime, "b.txt", newTime)
	if err != nil {
		t.Fatalf("WriteFileHeader() error: %v", err)
	}
	got := buf.String()
	if got != want {
		t.Errorf("WriteFileHeader() =\n%q\nwant:\n%q", got, want)
	}
}
