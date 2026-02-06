package diff_test

import (
	"bytes"
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
			a: []string{"A", "B"},
			b: []string{"C", "D"},
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "A"},
				{Op: diff.Del, OldLine: "B"},
				{Op: diff.Ins, NewLine: "C"},
				{Op: diff.Ins, NewLine: "D"},
			},
		},
		"CommonPrefix": {
			a: []string{"A", "B", "C", "X"},
			b: []string{"A", "B", "C", "Y"},
			want: []diff.Edit{
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
				{Op: diff.Del, OldLine: "X"},
				{Op: diff.Ins, NewLine: "Y"},
			},
		},
		"CommonSuffix": {
			a: []string{"X", "A", "B", "C"},
			b: []string{"Y", "A", "B", "C"},
			want: []diff.Edit{
				{Op: diff.Del, OldLine: "X"},
				{Op: diff.Ins, NewLine: "Y"},
				{Op: diff.Eq, OldLine: "A", NewLine: "A"},
				{Op: diff.Eq, OldLine: "B", NewLine: "B"},
				{Op: diff.Eq, OldLine: "C", NewLine: "C"},
			},
		},
		"PaperExample": {
			a: []string{"A", "B", "C", "A", "B", "B", "A"},
			b: []string{"C", "B", "A", "B", "A", "C"},
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

func TestWrite(t *testing.T) {
	tests := map[string]struct {
		edits       []diff.Edit
		context     int
		wantUnified string
		wantGutter  string
	}{
		"Empty": {
			edits:       nil,
			wantUnified: "",
			wantGutter:  "",
		},
		"OnlyEqual": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "same\n", NewLine: "same\n"},
			},
			wantUnified: "",
			wantGutter:  "",
		},
		"DelStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "removed\n"},
			},
			context:     0,
			wantUnified: "@@ -1 +0,0 @@\n-removed\n",
			wantGutter:  "1 - │ removed\n",
		},
		"DelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Eq, OldLine: "third\n", NewLine: "third\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1 @@\n-first\n second\n",
			wantGutter: "1 - │ first\n" +
				"2   │ second\n",
		},
		"DelMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
			},
			context:     0,
			wantUnified: "@@ -2 +1,0 @@\n-line2\n",
			wantGutter:  "2 - │ line2\n",
		},
		"DelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
			wantGutter: "1   │ before\n" +
				"2 - │ removed\n" +
				"3   │ after\n",
		},
		"DelMiddleContext5": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context:     5,
			wantUnified: "@@ -1,3 +1,2 @@\n before\n-removed\n after\n",
			wantGutter: "1   │ before\n" +
				"2 - │ removed\n" +
				"3   │ after\n",
		},
		"DelEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Del, OldLine: "line2\n"},
			},
			context:     0,
			wantUnified: "@@ -2 +1,0 @@\n-line2\n",
			wantGutter:  "2 - │ line2\n",
		},
		"DelEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first\n", NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Del, OldLine: "third\n"},
			},
			context:     1,
			wantUnified: "@@ -2,2 +2 @@\n second\n-third\n",
			wantGutter: "2   │ second\n" +
				"3 - │ third\n",
		},
		"InsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "added\n"},
			},
			context:     0,
			wantUnified: "@@ -0,0 +1 @@\n+added\n",
			wantGutter:  "  + │ added\n",
		},
		"InsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Eq, OldLine: "third\n", NewLine: "third\n"},
			},
			context:     1,
			wantUnified: "@@ -1 +1,2 @@\n+first\n second\n",
			wantGutter: "  + │ first\n" +
				"1   │ second\n",
		},
		"InsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Ins, NewLine: "line2\n"},
				{Op: diff.Eq, OldLine: "line3\n", NewLine: "line3\n"},
			},
			context:     0,
			wantUnified: "@@ -1,0 +2 @@\n+line2\n",
			wantGutter:  "  + │ line2\n",
		},
		"InsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "before\n", NewLine: "before\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "after\n", NewLine: "after\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1,3 @@\n before\n+added\n after\n",
			wantGutter: "1   │ before\n" +
				"  + │ added\n" +
				"2   │ after\n",
		},
		"InsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "line1\n", NewLine: "line1\n"},
				{Op: diff.Ins, NewLine: "line2\n"},
			},
			context:     0,
			wantUnified: "@@ -1,0 +2 @@\n+line2\n",
			wantGutter:  "  + │ line2\n",
		},
		"InsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "first\n", NewLine: "first\n"},
				{Op: diff.Eq, OldLine: "second\n", NewLine: "second\n"},
				{Op: diff.Ins, NewLine: "third\n"},
			},
			context:     1,
			wantUnified: "@@ -2 +2,2 @@\n second\n+third\n",
			wantGutter: "2   │ second\n" +
				"  + │ third\n",
		},
		"DelInsStartContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context:     0,
			wantUnified: "@@ -1 +1 @@\n-old\n+new\n",
			wantGutter: "1 - │ old\n" +
				"  + │ new\n",
		},
		"DelInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1,2 @@\n-old\n+new\n keep\n",
			wantGutter: "1 - │ old\n" +
				"  + │ new\n" +
				"2   │ keep\n",
		},
		"DelInsMiddleContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1\n", NewLine: "keep1\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "keep2\n", NewLine: "keep2\n"},
			},
			context:     0,
			wantUnified: "@@ -2 +2 @@\n-removed\n+added\n",
			wantGutter: "2 - │ removed\n" +
				"  + │ added\n",
		},
		"DelInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep1\n", NewLine: "keep1\n"},
				{Op: diff.Del, OldLine: "removed\n"},
				{Op: diff.Ins, NewLine: "added\n"},
				{Op: diff.Eq, OldLine: "keep2\n", NewLine: "keep2\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,3 @@\n keep1\n-removed\n+added\n keep2\n",
			wantGutter: "1   │ keep1\n" +
				"2 - │ removed\n" +
				"  + │ added\n" +
				"3   │ keep2\n",
		},
		"DelInsEndContext0": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context:     0,
			wantUnified: "@@ -2 +2 @@\n-old\n+new\n",
			wantGutter: "2 - │ old\n" +
				"  + │ new\n",
		},
		"DelInsEndContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1,2 @@\n keep\n-old\n+new\n",
			wantGutter: "1   │ keep\n" +
				"2 - │ old\n" +
				"  + │ new\n",
		},
		"ConsecDelMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Del, OldLine: "del1\n"},
				{Op: diff.Del, OldLine: "del2\n"},
				{Op: diff.Del, OldLine: "del3\n"},
				{Op: diff.Eq, OldLine: "end\n", NewLine: "end\n"},
			},
			context:     1,
			wantUnified: "@@ -1,5 +1,2 @@\n keep\n-del1\n-del2\n-del3\n end\n",
			wantGutter: "1   │ keep\n" +
				"2 - │ del1\n" +
				"3 - │ del2\n" +
				"4 - │ del3\n" +
				"5   │ end\n",
		},
		"ConsecInsMiddleContext1": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "keep\n", NewLine: "keep\n"},
				{Op: diff.Ins, NewLine: "ins1\n"},
				{Op: diff.Ins, NewLine: "ins2\n"},
				{Op: diff.Ins, NewLine: "ins3\n"},
				{Op: diff.Eq, OldLine: "end\n", NewLine: "end\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1,5 @@\n keep\n+ins1\n+ins2\n+ins3\n end\n",
			wantGutter: "1   │ keep\n" +
				"  + │ ins1\n" +
				"  + │ ins2\n" +
				"  + │ ins3\n" +
				"2   │ end\n",
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
			context:     1,
			wantUnified: "@@ -1,4 +1,4 @@\n keep\n-del1\n-del2\n+ins1\n+ins2\n end\n",
			wantGutter: "1   │ keep\n" +
				"2 - │ del1\n" +
				"3 - │ del2\n" +
				"  + │ ins1\n" +
				"  + │ ins2\n" +
				"4   │ end\n",
		},
		"ConsecDelStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "a\n"},
				{Op: diff.Del, OldLine: "b\n"},
				{Op: diff.Del, OldLine: "c\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +0,0 @@\n-a\n-b\n-c\n",
			wantGutter: "1 - │ a\n" +
				"2 - │ b\n" +
				"3 - │ c\n",
		},
		"ConsecInsStartContext1": {
			edits: []diff.Edit{
				{Op: diff.Ins, NewLine: "a\n"},
				{Op: diff.Ins, NewLine: "b\n"},
				{Op: diff.Ins, NewLine: "c\n"},
			},
			context:     1,
			wantUnified: "@@ -0,0 +1,3 @@\n+a\n+b\n+c\n",
			wantGutter: "  + │ a\n" +
				"  + │ b\n" +
				"  + │ c\n",
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
			context:     1,
			wantUnified: "@@ -1,3 +1,2 @@\n line1\n-del1\n line2\n@@ -6,2 +5,3 @@\n line5\n+ins1\n line6\n",
			// 4 equal lines between changes, 4 > 2*1=2 so hunks separate
			wantGutter: "1   │ line1\n" +
				"2 - │ del1\n" +
				"3   │ line2\n" +
				" ───┼─── 2 identical lines ───\n" +
				"6   │ line5\n" +
				"  + │ ins1\n" +
				"7   │ line6\n",
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
			context:     2,
			wantUnified: "@@ -1,7 +1,7 @@\n line1\n-del1\n line2\n line3\n line4\n line5\n+ins1\n line6\n",
			// 4 equal lines between changes, 4 = 2*2=4 so hunks merge
			wantGutter: "1   │ line1\n" +
				"2 - │ del1\n" +
				"3   │ line2\n" +
				"4   │ line3\n" +
				"5   │ line4\n" +
				"6   │ line5\n" +
				"  + │ ins1\n" +
				"7   │ line6\n",
		},
		"TwoHunksSeparateContext0": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "middle\n", NewLine: "middle\n"},
				{Op: diff.Ins, NewLine: "last\n"},
			},
			context:     0,
			wantUnified: "@@ -1 +0,0 @@\n-first\n@@ -2,0 +2 @@\n+last\n",
			// 1 equal line between changes, 1 > 2*0=0 so hunks separate
			wantGutter: "1 - │ first\n" +
				" ───┼─── 1 identical line ───\n" +
				"  + │ last\n",
		},
		"TwoHunksMergedContext1": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "first\n"},
				{Op: diff.Eq, OldLine: "middle\n", NewLine: "middle\n"},
				{Op: diff.Ins, NewLine: "last\n"},
			},
			context:     1,
			wantUnified: "@@ -1,2 +1,2 @@\n-first\n middle\n+last\n",
			// 1 equal line between changes, 1 < 2*1=2 so hunks merge
			wantGutter: "1 - │ first\n" +
				"2   │ middle\n" +
				"  + │ last\n",
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
			context:     1,
			wantUnified: "@@ -1,2 +1 @@\n-del1\n a\n@@ -5,3 +4,2 @@\n d\n-del2\n e\n@@ -10 +8,2 @@\n h\n+ins1\n",
			// 4 eq between changes, 4 > 2*1=2 so three separate hunks
			wantGutter: " 1 - │ del1\n" +
				" 2   │ a\n" +
				"  ───┼─── 2 identical lines ───\n" +
				" 5   │ d\n" +
				" 6 - │ del2\n" +
				" 7   │ e\n" +
				"  ───┼─── 2 identical lines ───\n" +
				"10   │ h\n" +
				"   + │ ins1\n",
		},
		"GutterExtraSpaces": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "func foo(a int) {\n", NewLine: "func foo(a int) {\n"},
				{Op: diff.Del, OldLine: "    fmt.Println(item)\n"},
				{Op: diff.Ins, NewLine: "    fmt.Println( item )\n"},
				{Op: diff.Eq, OldLine: "}\n", NewLine: "}\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,3 @@\n func foo(a int) {\n-    fmt.Println(item)\n+    fmt.Println( item )\n }\n",
			wantGutter: "1   │ func foo(a int) {\n" +
				"2 - │ ····fmt.Println(item)\n" +
				"  + │ ····fmt.Println(·item·)\n" +
				"3   │ }\n",
		},
		"GutterTabsToSpaces": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "func main() {\n", NewLine: "func main() {\n"},
				{Op: diff.Del, OldLine: "\tfmt.Println(\"hello\")\n"},
				{Op: diff.Ins, NewLine: "    fmt.Println(\"hello\")\n"},
				{Op: diff.Eq, OldLine: "}\n", NewLine: "}\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,3 @@\n func main() {\n-\tfmt.Println(\"hello\")\n+    fmt.Println(\"hello\")\n }\n",
			wantGutter: "1   │ func main() {\n" +
				"2 - │ →fmt.Println(\"hello\")\n" +
				"  + │ ····fmt.Println(\"hello\")\n" +
				"3   │ }\n",
		},
		"GutterTrailingWhitespace": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "func main() {\n", NewLine: "func main() {\n"},
				{Op: diff.Del, OldLine: "\tfmt.Println(\"hello\")   \n"},
				{Op: diff.Ins, NewLine: "\tfmt.Println(\"hello\")\n"},
				{Op: diff.Eq, OldLine: "}\n", NewLine: "}\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,3 @@\n func main() {\n-\tfmt.Println(\"hello\")   \n+\tfmt.Println(\"hello\")\n }\n",
			wantGutter: "1   │ func main() {\n" +
				"2 - │ →fmt.Println(\"hello\")···\n" +
				"  + │ →fmt.Println(\"hello\")\n" +
				"3   │ }\n",
		},
		"GutterMissingFinalNewline": {
			// want side has no trailing newline, got side has one
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "func main() {\n", NewLine: "func main() {\n"},
				{Op: diff.Eq, OldLine: "\tfmt.Println(\"hello\")\n", NewLine: "\tfmt.Println(\"hello\")\n"},
				{Op: diff.Del, OldLine: "}"},
				{Op: diff.Ins, NewLine: "}\n"},
			},
			context:     1,
			wantUnified: "@@ -2,2 +2,2 @@\n \tfmt.Println(\"hello\")\n-}\n\\ No newline at end of file\n+}\n",
			wantGutter: "2   │ \tfmt.Println(\"hello\")\n" +
				"3 - │ }\n" +
				"  + │ }↵\n",
		},
		"GutterExtraBlankLines": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "foo()\n", NewLine: "foo()\n"},
				{Op: diff.Ins, NewLine: "\n"},
				{Op: diff.Ins, NewLine: "\n"},
			},
			context:     1,
			wantUnified: "@@ -1 +1,3 @@\n foo()\n+\n+\n",
			wantGutter: "1   │ foo()\n" +
				"  + │ ↵\n" +
				"  + │ ↵\n",
		},
		"GutterBlankLineRemoved": {
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "a\n", NewLine: "a\n"},
				{Op: diff.Del, OldLine: "\n"},
				{Op: diff.Eq, OldLine: "b\n", NewLine: "b\n"},
			},
			context:     1,
			wantUnified: "@@ -1,3 +1,2 @@\n a\n-\n b\n",
			wantGutter: "1   │ a\n" +
				"2 - │ ↵\n" +
				"3   │ b\n",
		},
		"GutterContextLinesNoMarkers": {
			// Context lines with tabs should NOT have whitespace markers
			edits: []diff.Edit{
				{Op: diff.Eq, OldLine: "prefix\n", NewLine: "prefix\n"},
				{Op: diff.Eq, OldLine: "\tindented\n", NewLine: "\tindented\n"},
				{Op: diff.Del, OldLine: "old\n"},
				{Op: diff.Ins, NewLine: "new\n"},
				{Op: diff.Eq, OldLine: "suffix\n", NewLine: "suffix\n"},
			},
			context:     3,
			wantUnified: "@@ -1,4 +2,4 @@\n prefix\n \tindented\n-old\n+new\n suffix\n",
			wantGutter: "1   │ prefix\n" +
				"2   │ \tindented\n" +
				"3 - │ old\n" +
				"  + │ new\n" +
				"4   │ suffix\n",
		},
		"GutterCollapsedContext": {
			edits: []diff.Edit{
				{Op: diff.Del, OldLine: "func foo() {\n"},
				{Op: diff.Ins, NewLine: "func foo()  {\n"},
				{Op: diff.Eq, OldLine: "    a\n", NewLine: "    a\n"},
				{Op: diff.Eq, OldLine: "    b\n", NewLine: "    b\n"},
				{Op: diff.Eq, OldLine: "    c\n", NewLine: "    c\n"},
				{Op: diff.Eq, OldLine: "    d\n", NewLine: "    d\n"},
				{Op: diff.Eq, OldLine: "    e\n", NewLine: "    e\n"},
				{Op: diff.Eq, OldLine: "    f\n", NewLine: "    f\n"},
				{Op: diff.Eq, OldLine: "    g\n", NewLine: "    g\n"},
				{Op: diff.Eq, OldLine: "    h\n", NewLine: "    h\n"},
				{Op: diff.Eq, OldLine: "    i\n", NewLine: "    i\n"},
				{Op: diff.Eq, OldLine: "    j\n", NewLine: "    j\n"},
				{Op: diff.Del, OldLine: "}  \n"},
				{Op: diff.Ins, NewLine: "}\n"},
			},
			context: 3,
			// 10 equal lines between changes, context=3: show 3 after first, 3 before second, collapse 4
			wantUnified: "@@ -1,4 +1,4 @@\n-func foo() {\n+func foo()  {\n     a\n     b\n     c\n@@ -9,4 +11,4 @@\n     h\n     i\n     j\n-}  \n+}\n",
			wantGutter: " 1 - │ func·foo()·{\n" +
				"   + │ func·foo()··{\n" +
				" 2   │     a\n" +
				" 3   │     b\n" +
				" 4   │     c\n" +
				"  ───┼─── 4 identical lines ───\n" +
				" 9   │     h\n" +
				"10   │     i\n" +
				"11   │     j\n" +
				"12 - │ }··\n" +
				"   + │ }\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Run("Unified", func(t *testing.T) {
				var buf bytes.Buffer
				err := diff.WriteUnified(&buf, test.edits, test.context)
				if err != nil {
					t.Fatalf("WriteUnified() error: %v", err)
				}
				got := buf.String()
				if got != test.wantUnified {
					t.Errorf("WriteUnified() =\n%q\nwant:\n%q", got, test.wantUnified)
				}
			})
			t.Run("Gutter", func(t *testing.T) {
				var buf bytes.Buffer
				err := diff.WriteGutter(&buf, test.edits, test.context)
				if err != nil {
					t.Fatalf("WriteGutter() error: %v", err)
				}
				got := buf.String()
				if got != test.wantGutter {
					t.Errorf("WriteGutter() =\n%q\nwant:\n%q", got, test.wantGutter)
				}
			})
		})
	}
}
