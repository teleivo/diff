package main

import (
	"bytes"
	"testing"
	"time"
)

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
			want: `--- testdata/one_line.txt	2026-02-05 07:06:29.205156380 +0100
+++ testdata/one_line_different.txt	2026-02-05 07:06:29.205156380 +0100
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
			want: `--- testdata/multi_line_a.txt	2026-02-05 07:06:29.205156380 +0100
+++ testdata/multi_line_b.txt	2026-02-05 07:06:29.205156380 +0100
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
			hasDiff, err := files(&buf, test.a, test.b, test.context, false)
			if test.wantErr {
				if err == nil {
					t.Fatalf("files() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("files() unexpected error: %v", err)
			}
			if hasDiff != test.wantDiff {
				t.Errorf("files() hasDiff = %v, want %v", hasDiff, test.wantDiff)
			}
			got := buf.String()
			if got != test.want {
				t.Errorf("files() =\n%q\nwant:\n%q", got, test.want)
			}
		})
	}
}

func TestWriteFileHeader(t *testing.T) {
	oldTime := time.Date(2026, 2, 4, 8, 12, 16, 2963487, time.FixedZone("CET", 3600))
	newTime := time.Date(2026, 2, 4, 9, 30, 45, 123456789, time.FixedZone("CET", 3600))
	want := "--- a.txt\t2026-02-04 08:12:16.002963487 +0100\n+++ b.txt\t2026-02-04 09:30:45.123456789 +0100\n"

	var buf bytes.Buffer
	err := writeFileHeader(&buf, "a.txt", oldTime, "b.txt", newTime)
	if err != nil {
		t.Fatalf("writeFileHeader() error: %v", err)
	}
	got := buf.String()
	if got != want {
		t.Errorf("writeFileHeader() =\n%q\nwant:\n%q", got, want)
	}
}
