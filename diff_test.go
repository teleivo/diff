package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestLines(t *testing.T) {
	tests := map[string]struct {
		a    []string
		b    []string
		want []Edit
	}{
		"BothEmpty": {
			a:    []string{},
			b:    []string{},
			want: []Edit{},
		},
		"FirstEmpty": {
			a: []string{},
			b: []string{"A", "B"},
			want: []Edit{
				{Op: Ins, NewLine: "A"},
				{Op: Ins, NewLine: "B"},
			},
		},
		"SecondEmpty": {
			a: []string{"A", "B"},
			b: []string{},
			want: []Edit{
				{Op: Del, OldLine: "A"},
				{Op: Del, OldLine: "B"},
			},
		},
		"Equal": {
			a: []string{"A", "B", "C"},
			b: []string{"A", "B", "C"},
			want: []Edit{
				{Op: Eq, OldLine: "A", NewLine: "A"},
				{Op: Eq, OldLine: "B", NewLine: "B"},
				{Op: Eq, OldLine: "C", NewLine: "C"},
			},
		},
		"CompletelyDifferent": {
			a: strings.Split("AB", ""),
			b: strings.Split("CD", ""),
			want: []Edit{
				{Op: Del, OldLine: "A"},
				{Op: Del, OldLine: "B"},
				{Op: Ins, NewLine: "C"},
				{Op: Ins, NewLine: "D"},
			},
		},
		"CommonPrefix": {
			a: strings.Split("ABCX", ""),
			b: strings.Split("ABCY", ""),
			want: []Edit{
				{Op: Eq, OldLine: "A", NewLine: "A"},
				{Op: Eq, OldLine: "B", NewLine: "B"},
				{Op: Eq, OldLine: "C", NewLine: "C"},
				{Op: Del, OldLine: "X"},
				{Op: Ins, NewLine: "Y"},
			},
		},
		"CommonSuffix": {
			a: strings.Split("XABC", ""),
			b: strings.Split("YABC", ""),
			want: []Edit{
				{Op: Del, OldLine: "X"},
				{Op: Ins, NewLine: "Y"},
				{Op: Eq, OldLine: "A", NewLine: "A"},
				{Op: Eq, OldLine: "B", NewLine: "B"},
				{Op: Eq, OldLine: "C", NewLine: "C"},
			},
		},
		"PaperExample": {
			a: strings.Split("ABCABBA", ""),
			b: strings.Split("CBABAC", ""),
			want: []Edit{
				{Op: Del, OldLine: "A"},
				{Op: Del, OldLine: "B"},
				{Op: Eq, OldLine: "C", NewLine: "C"},
				{Op: Ins, NewLine: "B"},
				{Op: Eq, OldLine: "A", NewLine: "A"},
				{Op: Eq, OldLine: "B", NewLine: "B"},
				{Op: Del, OldLine: "B"},
				{Op: Eq, OldLine: "A", NewLine: "A"},
				{Op: Ins, NewLine: "C"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Lines(test.a, test.b)
			if len(got) != len(test.want) {
				t.Fatalf("Lines(%v, %v) returned %d edits, want %d\ngot:  %v\nwant: %v",
					test.a, test.b, len(got), len(test.want), got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("Lines(%v, %v)[%d] = %v, want %v", test.a, test.b, i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestFiles(t *testing.T) {
	tests := map[string]struct {
		a       string
		b       string
		want    []Edit
		wantErr bool
	}{
		"BothEmpty": {
			a:    "testdata/empty.txt",
			b:    "testdata/empty.txt",
			want: nil,
		},
		"Identical": {
			a: "testdata/one_line.txt",
			b: "testdata/one_line.txt",
			want: []Edit{
				{Op: Eq, OldLine: "hello", NewLine: "hello"},
			},
		},
		"OneLineDifferent": {
			a: "testdata/one_line.txt",
			b: "testdata/one_line_different.txt",
			want: []Edit{
				{Op: Del, OldLine: "hello"},
				{Op: Ins, NewLine: "world"},
			},
		},
		"MultiLineMiddleChanged": {
			a: "testdata/multi_line_a.txt",
			b: "testdata/multi_line_b.txt",
			want: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Del, OldLine: "line2"},
				{Op: Ins, NewLine: "modified"},
				{Op: Eq, OldLine: "line3", NewLine: "line3"},
			},
		},
		"File1NotFound": {
			a:       "testdata/nonexistent.txt",
			b:       "testdata/empty.txt",
			wantErr: true,
		},
		"File2NotFound": {
			a:       "testdata/empty.txt",
			b:       "testdata/nonexistent.txt",
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Files(test.a, test.b)
			if test.wantErr {
				if err == nil {
					t.Fatalf("Files(%q, %q) expected error, got nil", test.a, test.b)
				}
				return
			}
			if err != nil {
				t.Fatalf("Files(%q, %q) unexpected error: %v", test.a, test.b, err)
			}
			if len(got) != len(test.want) {
				t.Fatalf("Files(%q, %q) returned %d edits, want %d\ngot:  %v\nwant: %v",
					test.a, test.b, len(got), len(test.want), got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Errorf("Files(%q, %q)[%d] = %v, want %v", test.a, test.b, i, got[i], test.want[i])
				}
			}
		})
	}
}

func TestWriteUnified(t *testing.T) {
	tests := map[string]struct {
		edits   []Edit
		context int
		want    string
	}{
		"Empty": {
			edits:   nil,
			context: 0,
			want:    "",
		},
		"OnlyEqual": {
			edits: []Edit{
				{Op: Eq, OldLine: "same", NewLine: "same"},
			},
			context: 0,
			want:    "",
		},
		"DelStartContext0": {
			edits: []Edit{
				{Op: Del, OldLine: "removed"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-removed\n",
		},
		"DelStartContext1": {
			edits: []Edit{
				{Op: Del, OldLine: "first"},
				{Op: Eq, OldLine: "second", NewLine: "second"},
				{Op: Eq, OldLine: "third", NewLine: "third"},
			},
			context: 1,
			want:    "@@ -1,2 +1 @@\n-first\n second\n",
		},
		"DelMiddleContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Del, OldLine: "line2"},
				{Op: Eq, OldLine: "line3", NewLine: "line3"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "before", NewLine: "before"},
				{Op: Del, OldLine: "removed"},
				{Op: Eq, OldLine: "after", NewLine: "after"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelMiddleContext5": {
			edits: []Edit{
				{Op: Eq, OldLine: "before", NewLine: "before"},
				{Op: Del, OldLine: "removed"},
				{Op: Eq, OldLine: "after", NewLine: "after"},
			},
			context: 5,
			want:    "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
		},
		"DelEndContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Del, OldLine: "line2"},
			},
			context: 0,
			want:    "@@ -2 +1,0 @@\n-line2\n",
		},
		"DelEndContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "first", NewLine: "first"},
				{Op: Eq, OldLine: "second", NewLine: "second"},
				{Op: Del, OldLine: "third"},
			},
			context: 1,
			want:    "@@ -2,2 +2 @@\n second\n-third\n",
		},
		"InsStartContext0": {
			edits: []Edit{
				{Op: Ins, NewLine: "added"},
			},
			context: 0,
			want:    "@@ -0,0 +1 @@\n+added\n",
		},
		"InsStartContext1": {
			edits: []Edit{
				{Op: Ins, NewLine: "first"},
				{Op: Eq, OldLine: "second", NewLine: "second"},
				{Op: Eq, OldLine: "third", NewLine: "third"},
			},
			context: 1,
			want:    "@@ -1 +1,2 @@\n+first\n second\n",
		},
		"InsMiddleContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Ins, NewLine: "line2"},
				{Op: Eq, OldLine: "line3", NewLine: "line3"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "before", NewLine: "before"},
				{Op: Ins, NewLine: "added"},
				{Op: Eq, OldLine: "after", NewLine: "after"},
			},
			context: 1,
			want:    "@@ -1,2 +1,3 @@\n before\n+added\n after\n",
		},
		"InsEndContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Ins, NewLine: "line2"},
			},
			context: 0,
			want:    "@@ -1,0 +2 @@\n+line2\n",
		},
		"InsEndContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "first", NewLine: "first"},
				{Op: Eq, OldLine: "second", NewLine: "second"},
				{Op: Ins, NewLine: "third"},
			},
			context: 1,
			want:    "@@ -2 +2,2 @@\n second\n+third\n",
		},
		"DelInsStartContext0": {
			edits: []Edit{
				{Op: Del, OldLine: "old"},
				{Op: Ins, NewLine: "new"},
			},
			context: 0,
			want:    "@@ -1 +1 @@\n-old\n+new\n",
		},
		"DelInsStartContext1": {
			edits: []Edit{
				{Op: Del, OldLine: "old"},
				{Op: Ins, NewLine: "new"},
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-old\n+new\n keep\n",
		},
		"DelInsMiddleContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep1", NewLine: "keep1"},
				{Op: Del, OldLine: "removed"},
				{Op: Ins, NewLine: "added"},
				{Op: Eq, OldLine: "keep2", NewLine: "keep2"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-removed\n+added\n",
		},
		"DelInsMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep1", NewLine: "keep1"},
				{Op: Del, OldLine: "removed"},
				{Op: Ins, NewLine: "added"},
				{Op: Eq, OldLine: "keep2", NewLine: "keep2"},
			},
			context: 1,
			want:    "@@ -1,3 +1,3 @@\n keep1\n-removed\n+added\n keep2\n",
		},
		"DelInsEndContext0": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
				{Op: Del, OldLine: "old"},
				{Op: Ins, NewLine: "new"},
			},
			context: 0,
			want:    "@@ -2 +2 @@\n-old\n+new\n",
		},
		"DelInsEndContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
				{Op: Del, OldLine: "old"},
				{Op: Ins, NewLine: "new"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n keep\n-old\n+new\n",
		},
		"ConsecDelMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
				{Op: Del, OldLine: "del1"},
				{Op: Del, OldLine: "del2"},
				{Op: Del, OldLine: "del3"},
				{Op: Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,5 +1,2 @@\n keep\n-del1\n-del2\n-del3\n end\n",
		},
		"ConsecInsMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
				{Op: Ins, NewLine: "ins1"},
				{Op: Ins, NewLine: "ins2"},
				{Op: Ins, NewLine: "ins3"},
				{Op: Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,2 +1,5 @@\n keep\n+ins1\n+ins2\n+ins3\n end\n",
		},
		"ConsecDelInsMiddleContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "keep", NewLine: "keep"},
				{Op: Del, OldLine: "del1"},
				{Op: Del, OldLine: "del2"},
				{Op: Ins, NewLine: "ins1"},
				{Op: Ins, NewLine: "ins2"},
				{Op: Eq, OldLine: "end", NewLine: "end"},
			},
			context: 1,
			want:    "@@ -1,4 +1,4 @@\n keep\n-del1\n-del2\n+ins1\n+ins2\n end\n",
		},
		"ConsecDelStartContext1": {
			edits: []Edit{
				{Op: Del, OldLine: "a"},
				{Op: Del, OldLine: "b"},
				{Op: Del, OldLine: "c"},
			},
			context: 1,
			want:    "@@ -1,3 +0,0 @@\n-a\n-b\n-c\n",
		},
		"ConsecInsStartContext1": {
			edits: []Edit{
				{Op: Ins, NewLine: "a"},
				{Op: Ins, NewLine: "b"},
				{Op: Ins, NewLine: "c"},
			},
			context: 1,
			want:    "@@ -0,0 +1,3 @@\n+a\n+b\n+c\n",
		},
		"TwoHunksSeparateContext1": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Del, OldLine: "del1"},
				{Op: Eq, OldLine: "line2", NewLine: "line2"},
				{Op: Eq, OldLine: "line3", NewLine: "line3"},
				{Op: Eq, OldLine: "line4", NewLine: "line4"},
				{Op: Eq, OldLine: "line5", NewLine: "line5"},
				{Op: Ins, NewLine: "ins1"},
				{Op: Eq, OldLine: "line6", NewLine: "line6"},
			},
			context: 1,
			want:    "@@ -1,3 +1,2 @@\n line1\n-del1\n line2\n@@ -6,2 +5,3 @@\n line5\n+ins1\n line6\n",
		},
		"TwoHunksMergedContext2": {
			edits: []Edit{
				{Op: Eq, OldLine: "line1", NewLine: "line1"},
				{Op: Del, OldLine: "del1"},
				{Op: Eq, OldLine: "line2", NewLine: "line2"},
				{Op: Eq, OldLine: "line3", NewLine: "line3"},
				{Op: Eq, OldLine: "line4", NewLine: "line4"},
				{Op: Eq, OldLine: "line5", NewLine: "line5"},
				{Op: Ins, NewLine: "ins1"},
				{Op: Eq, OldLine: "line6", NewLine: "line6"},
			},
			context: 2,
			want:    "@@ -1,7 +1,7 @@\n line1\n-del1\n line2\n line3\n line4\n line5\n+ins1\n line6\n",
		},
		"TwoHunksSeparateContext0": {
			edits: []Edit{
				{Op: Del, OldLine: "first"},
				{Op: Eq, OldLine: "middle", NewLine: "middle"},
				{Op: Ins, NewLine: "last"},
			},
			context: 0,
			want:    "@@ -1 +0,0 @@\n-first\n@@ -2,0 +2 @@\n+last\n",
		},
		"TwoHunksMergedContext1": {
			edits: []Edit{
				{Op: Del, OldLine: "first"},
				{Op: Eq, OldLine: "middle", NewLine: "middle"},
				{Op: Ins, NewLine: "last"},
			},
			context: 1,
			want:    "@@ -1,2 +1,2 @@\n-first\n middle\n+last\n",
		},
		"ThreeHunksSeparateContext1": {
			edits: []Edit{
				{Op: Del, OldLine: "del1"},
				{Op: Eq, OldLine: "a", NewLine: "a"},
				{Op: Eq, OldLine: "b", NewLine: "b"},
				{Op: Eq, OldLine: "c", NewLine: "c"},
				{Op: Eq, OldLine: "d", NewLine: "d"},
				{Op: Del, OldLine: "del2"},
				{Op: Eq, OldLine: "e", NewLine: "e"},
				{Op: Eq, OldLine: "f", NewLine: "f"},
				{Op: Eq, OldLine: "g", NewLine: "g"},
				{Op: Eq, OldLine: "h", NewLine: "h"},
				{Op: Ins, NewLine: "ins1"},
			},
			context: 1,
			want:    "@@ -1,2 +1 @@\n-del1\n a\n@@ -5,3 +4,2 @@\n d\n-del2\n e\n@@ -10 +8,2 @@\n h\n+ins1\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteUnified(&buf, test.edits, test.context)
			if err != nil {
				t.Fatalf("WriteUnified() error: %v", err)
			}
			got := buf.String()
			if got != test.want {
				t.Errorf("WriteUnified() =\n%q\nwant:\n%q", got, test.want)
			}
		})
	}
}
