package diff_test

import (
	"bytes"
	"strings"
	"testing"

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
				{Op: diff.Eq, OldLine: "same\n", NewLine: "same\n"},
			},
			context: 0,
			want:    "",
		},
		"DelStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "removed\n"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-removed\n",
		},
		"DelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Eq, OldLine: "third\n", NewLine: "third\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1 @@\n-first\n second\n",
		},
		"DelMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelMiddleContext5": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context: 5,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "line2\n"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first\n", NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Del, OldLine: "third\n"},
			},
			context: 1,
			want:    "@@ -2,2 +2 @@\n second\n-third\n",
		},
		"InsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "added\n"},
			},
			context: 0,
			want:    "@@ -0,0 +1 @@\n+added\n",
		},
		"InsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Eq, OldLine: "third\n", NewLine: "third\n"},
			},
			context: 1,
			want:    "@@ -1 +1,2 @@\n+first\n second\n",
		},
		"InsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Ins, NewLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1,3 @@\n before\n+added\n after\n",
		},
		"InsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Ins, NewLine: "line2\n"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first\n", NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Ins, NewLine: "third\n"},
			},
			context: 1,
			want:    "@@ -2 +2,2 @@\n second\n+third\n",
		},
		"DelInsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context: 0,
			want:    "@@ -1 +1 @@\n-old\n+new\n",
		},
		"DelInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-old\n+new\n keep\n",
		},
		"DelInsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1\n", NewLine: "keep1\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "keep2\n", NewLine: "keep2\n"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-removed\n+added\n",
		},
		"DelInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1\n", NewLine: "keep1\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "keep2\n", NewLine: "keep2\n"},
			},
			context: 1,
			want:    "@@ -1,3 +1,3 @@\n keep1\n-removed\n+added\n keep2\n",
		},
		"DelInsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-old\n+new\n",
		},
		"DelInsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n keep\n-old\n+new\n",
		},
		"ConsecDelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Del, OldLine: "del2\n"},
				{Op: diff.Del, OldLine: "del3\n"},
				{Op: diff.Eq, OldLine: "end\n", NewLine: "end\n"},
			},
			context: 1,
			want:    "@@ -1,5 +1,2 @@\n keep\n-del1\n-del2\n-del3\n end\n",
		},
		"ConsecInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
				{Op: diff.Ins, NewLine: "ins2\n"},
				{Op: diff.Ins, NewLine: "ins3\n"},
				{Op: diff.Eq, OldLine: "end\n", NewLine: "end\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1,5 @@\n keep\n+ins1\n+ins2\n+ins3\n end\n",
		},
		"ConsecDelInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Del, OldLine: "del2\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
				{Op: diff.Ins, NewLine: "ins2\n"},
				{Op: diff.Eq, OldLine: "end\n", NewLine: "end\n"},
			},
			context: 1,
			want:    "@@ -1,4 +1,4 @@\n keep\n-del1\n-del2\n+ins1\n+ins2\n end\n",
		},
		"ConsecDelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "a\n"},
				{Op: diff.Del, OldLine: "b\n"},
				{Op: diff.Del, OldLine: "c\n"},
			},
			context: 1,
			want:    "@@ -1,3 +0,0 @@\n-a\n-b\n-c\n",
		},
		"ConsecInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "a\n"},
				{Op: diff.Ins, NewLine: "b\n"},
				{Op: diff.Ins, NewLine: "c\n"},
			},
			context: 1,
			want:    "@@ -0,0 +1,3 @@\n+a\n+b\n+c\n",
		},
		"TwoHunksSeparateContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Eq, OldLine: "line2\n", NewLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
				{Op: diff.Eq, OldLine: "line4\n", NewLine: "line4\n"},
				{Op: diff.Eq, OldLine: "line5\n", NewLine: "line5\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
				{Op: diff.Eq, OldLine: "line6\n", NewLine: "line6\n"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n line1\n-del1\n line2\n@@ -6,2 +5,3 @@\n line5\n+ins1\n line6\n",
		},
		"TwoHunksMergedContext2": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Eq, OldLine: "line2\n", NewLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
				{Op: diff.Eq, OldLine: "line4\n", NewLine: "line4\n"},
				{Op: diff.Eq, OldLine: "line5\n", NewLine: "line5\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
				{Op: diff.Eq, OldLine: "line6\n", NewLine: "line6\n"},
			},
			context: 2,
			want:    "@@ -1,7 +1,7 @@\n line1\n-del1\n line2\n line3\n line4\n line5\n+ins1\n line6\n",
		},
		"TwoHunksSeparateContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "middle\n", NewLine: "middle\n"},
				{Op: diff.Ins, NewLine: "last\n"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-first\n@@ -2,0 +2 @@\n+last\n",
		},
		"TwoHunksMergedContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "middle\n", NewLine: "middle\n"},
				{Op: diff.Ins, NewLine: "last\n"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-first\n middle\n+last\n",
		},
		"ThreeHunksSeparateContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Eq, OldLine: "a\n", NewLine: "a\n"},
				{Op: diff.Eq, OldLine: "b\n", NewLine: "b\n"},
				{Op: diff.Eq, OldLine: "c\n", NewLine: "c\n"},
				{Op: diff.Eq, OldLine: "d\n", NewLine: "d\n"},
				{Op: diff.Del, OldLine: "del2\n"},
				{Op: diff.Eq, OldLine: "e\n", NewLine: "e\n"},
				{Op: diff.Eq, OldLine: "f\n", NewLine: "f\n"},
				{Op: diff.Eq, OldLine: "g\n", NewLine: "g\n"},
				{Op: diff.Eq, OldLine: "h\n", NewLine: "h\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
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
