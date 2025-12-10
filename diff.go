// Package diff implements the Myers diff algorithm for computing the shortest
// edit script (SES) between two sequences.
//
// The algorithm is described in Eugene W. Myers' paper "An O(ND) Difference
// Algorithm and Its Variations" (1986). It runs in O(ND) time where N is the
// sum of the lengths of the two sequences and D is the size of the minimum
// edit script.
package diff

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// OpType represents the type of edit operation.
type OpType int

const (
	// Ins indicates a line should be inserted from sequence b.
	Ins OpType = iota
	// Del indicates a line should be deleted from sequence a.
	Del
	// Eq indicates the line is equal in both sequences.
	Eq
)

func (op OpType) String() string {
	switch op {
	case Ins:
		return "+"
	case Del:
		return "-"
	case Eq:
		return " "
	default:
		panic("unknown OpType")
	}
}

// Edit represents a single edit operation in the diff.
type Edit struct {
	Op      OpType
	OldLine string // line from a (for Del and Eq)
	NewLine string // line from b (for Ins and Eq)
}

// Files computes the shortest edit script to transform file1 into file2.
// It reads both files, splits them into lines, and returns the edit operations.
func Files(file1, file2 string) ([]Edit, error) {
	a, err := readLines(file1)
	if err != nil {
		return nil, err
	}
	b, err := readLines(file2)
	if err != nil {
		return nil, err
	}
	return Lines(a, b), nil
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return strings.Split(string(data), "\n"), nil
}

// Lines computes the shortest edit script to transform sequence a into sequence b.
// It returns a slice of [Edit] operations that, when applied in order, convert a to b.
func Lines(a, b []string) []Edit {
	n := len(a)
	m := len(b)
	maxD := n + m
	if maxD == 0 {
		return nil
	}
	var edits []Edit
	trace := shortestEdit(a, b)
	x, y := n, m
	for d := len(trace) - 1; d >= 0; d-- {
		v := trace[d]
		k := x - y
		i := k + maxD
		var op OpType
		var prevK int
		var prevX, prevY int
		if k == -d || (k != d && v[i-1] < v[i+1]) {
			prevK = k + 1 // down i.e. insert
			op = Ins
		} else {
			prevK = k - 1 // right i.e. delete
			op = Del
		}
		prevX = v[prevK+maxD]
		prevY = prevX - prevK

		for x > prevX && y > prevY { // advance on snake i.e. diagonal
			edits = append(edits, Edit{Op: Eq, OldLine: a[x-1], NewLine: b[y-1]})
			x--
			y--
		}

		if d > 0 {
			if op == Ins {
				edits = append(edits, Edit{Op: Ins, NewLine: b[y-1]})
			} else {
				edits = append(edits, Edit{Op: Del, OldLine: a[x-1]})
			}
		}
		x, y = prevX, prevY
	}

	slices.Reverse(edits)
	return edits
}

// shortestEdit computes the trace of furthest reaching D-paths for transforming
// a into b. Each element in the returned slice represents the V array state
// before each iteration d, which is used to reconstruct the edit script.
func shortestEdit(a, b []string) [][]int {
	n := len(a)
	m := len(b)
	maxD := n + m
	var trace [][]int
	if maxD == 0 {
		return trace
	}
	v := make([]int, 2*maxD+1)

	for d := range maxD + 1 {
		trace = append(trace, slices.Clone(v))
		for k := -d; k <= d; k = k + 2 {
			if k > n || k < -m { // skip out of bounds diagonals
				continue
			}
			i := k + maxD
			var x int
			if k == -d || (k != d && v[i-1] < v[i+1]) {
				x = v[i+1] // down i.e. insert
			} else {
				x = v[i-1] + 1 // right i.e. delete
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] { // advance on snake i.e. diagonal
				x++
				y++
			}
			v[i] = x
			if x >= n && y >= m {
				return trace
			}
		}
	}
	return trace
}

type unifiedWriter struct {
	w       *bufio.Writer
	edits   []Edit
	context int
	eqCount int

	lineNew int
	lineOld int

	// hunk
	hunkStart int // 0 indexed
	hunkEnd   int // 0 indexed
	startNew  int // 1 indexed
	startOld  int // 1 indexed
	countNew  int
	countOld  int
}

// WriteUnified writes the edits in unified diff format to w.
// The context parameter specifies the number of unchanged lines to show around each change.
// With context=0, only deletions and insertions are written; equal lines are omitted.
func WriteUnified(w io.Writer, edits []Edit, context int) error {
	uw := &unifiedWriter{
		w:         bufio.NewWriter(w),
		edits:     edits,
		context:   context,
		hunkStart: -1,
		hunkEnd:   -1,
		// startOld:    -1,
		// startNew:    -1,
	}
	uw.write()
	return uw.w.Flush()
}

func (uw *unifiedWriter) write() {
	for i := 0; i < len(uw.edits); i++ {
		switch uw.edits[i].Op {
		case Eq:
			uw.lineNew++
			uw.lineOld++

			if uw.hunkStart >= 0 {
				uw.hunkEnd = i

				if uw.context > 0 {
					// set start line for the file that had no context before the change
					if uw.startOld == 0 {
						uw.startOld = uw.lineOld
					} else if uw.startNew == 0 {
						uw.startNew = uw.lineNew
					}
				} else {
					if uw.startOld == 0 {
						uw.startOld = uw.lineOld - 1
					} else if uw.startNew == 0 {
						uw.startNew = uw.lineNew - 1
					}
				}

				if uw.eqCount+1 > 2*uw.context { // hunk end
					// adjust for the extra eq we counted to wait for a possibly merged hunk
					if uw.context > 0 && uw.eqCount > uw.context {
						adjust := uw.eqCount - uw.context
						uw.countOld -= adjust
						uw.countNew -= adjust
						uw.hunkEnd -= adjust
					}

					_ = writeHunkHeader(uw.w, uw.startOld, uw.countOld, uw.startNew, uw.countNew)
					for j := uw.hunkStart; j < uw.hunkEnd; j++ {
						uw.writeEdit(uw.edits[j])
					}
					uw.hunkStart = -1
					uw.hunkEnd = -1
					uw.startNew = 0
					uw.startOld = 0
					uw.eqCount = 0
					uw.countNew = 0
					uw.countOld = 0
				} else {
					uw.eqCount++
					uw.countNew++
					uw.countOld++
				}
			}
		case Ins:
			uw.lineNew++
			uw.countNew++
			uw.eqCount = 0
			uw.hunkEnd = i

			if uw.hunkStart < 0 { // starting new hunk
				uw.hunkStart = max(0, i-uw.context)

				var context int
				if i > 0 { // context before
					context = min(i, uw.context)
					uw.countOld += context
					uw.countNew += context
					if context > 0 {
						uw.startOld = uw.lineOld
					}
				}
				uw.startNew = uw.lineNew - context
			} else { // part of an existing hunk
				// set start line for the file that had no context before the change
				if uw.startNew == 0 {
					uw.startNew = uw.lineNew
				}
			}
		case Del:
			uw.lineOld++
			uw.countOld++
			uw.eqCount = 0
			uw.hunkEnd = i

			if uw.hunkStart < 0 { // starting new hunk
				uw.hunkStart = max(0, i-uw.context)

				var context int
				if i > 0 { // context before
					context = min(i, uw.context)
					uw.countOld += context
					uw.countNew += context
					if context > 0 {
						uw.startNew = uw.lineNew
					}
				}
				uw.startOld = uw.lineOld - context
			} else { // part of an existing hunk
				// set start line for the file that had no context before the change
				if uw.startOld == 0 {
					uw.startOld = uw.lineOld
				}
			}
		}
	}

	// flush remaining hunk
	if uw.hunkStart >= 0 {
		if uw.startOld == 0 {
			uw.startOld = uw.lineOld
		} else if uw.startNew == 0 {
			uw.startNew = uw.lineNew
		}
		// adjust for the extra eq we counted to wait for a possibly merged hunk
		if uw.context > 0 && uw.eqCount > uw.context {
			adjust := uw.eqCount - uw.context
			uw.countOld -= adjust
			uw.countNew -= adjust
			uw.hunkEnd -= adjust
		}

		_ = writeHunkHeader(uw.w, uw.startOld, uw.countOld, uw.startNew, uw.countNew)
		for j := uw.hunkStart; j <= uw.hunkEnd; j++ {
			uw.writeEdit(uw.edits[j])
		}
	}
}

func (uw *unifiedWriter) writeEdit(e Edit) {
	_, _ = uw.w.WriteString(e.Op.String())
	if e.Op == Del {
		_, _ = uw.w.WriteString(e.OldLine)
	} else {
		_, _ = uw.w.WriteString(e.NewLine)
	}
	_ = uw.w.WriteByte('\n')
}

// writeHunkHeader writes a hunk header in unified diff format.
// When count is 1, it is omitted (e.g., @@ -2 +2 @@ instead of @@ -2,1 +2,1 @@).
func writeHunkHeader(w io.Writer, oldStart, oldCount, newStart, newCount int) error {
	var err error
	if oldCount != 1 && newCount != 1 {
		_, err = fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount)
	} else if oldCount == 1 && newCount == 1 {
		_, err = fmt.Fprintf(w, "@@ -%d +%d @@\n", oldStart, newStart)
	} else if oldCount == 1 {
		_, err = fmt.Fprintf(w, "@@ -%d +%d,%d @@\n", oldStart, newStart, newCount)
	} else {
		_, err = fmt.Fprintf(w, "@@ -%d,%d +%d @@\n", oldStart, oldCount, newStart)
	}
	return err
}
