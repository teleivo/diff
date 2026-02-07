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
	"slices"
)

// OpType represents the type of edit operation.
type OpType int

const (
	// Ins indicates a line should be inserted from the new sequence.
	Ins OpType = iota
	// Del indicates a line should be deleted from the old sequence.
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

// Edit represents a single edit operation in the diff. Line values may include a trailing
// '\n' delimiter. A line without a trailing '\n' represents the last line of a sequence
// that has no final newline.
type Edit struct {
	Op      OpType
	OldLine string // line from the old sequence (for Del and Eq)
	NewLine string // line from the new sequence (for Ins and Eq)
}

// Lines computes the shortest edit script to transform oldLines into newLines.
// It returns a slice of [Edit] operations that, when applied in order, convert oldLines
// to newLines.
func Lines(oldLines, newLines []string) []Edit {
	n := len(oldLines)
	m := len(newLines)
	maxD := n + m
	if maxD == 0 {
		return nil
	}
	var edits []Edit
	trace := shortestEdit(oldLines, newLines)
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
			edits = append(edits, Edit{Op: Eq, OldLine: oldLines[x-1], NewLine: newLines[y-1]})
			x--
			y--
		}

		if d > 0 {
			if op == Ins {
				edits = append(edits, Edit{Op: Ins, NewLine: newLines[y-1]})
			} else {
				edits = append(edits, Edit{Op: Del, OldLine: oldLines[x-1]})
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

type config struct {
	context int
	gutter  bool
	color   bool
}

// Option configures how [Write] formats its output.
type Option func(*config)

// WithContext sets the number of unchanged lines to show around each change.
// It panics if lines is negative. The default is 3.
func WithContext(lines int) Option {
	if lines < 0 {
		panic("diff: negative context")
	}
	return func(conf *config) {
		conf.context = lines
	}
}

// WithGutter enables gutter format: each line is prefixed with a line number from the old
// sequence, an operation indicator, and a │ separator. Whitespace in changed lines is made
// visible (spaces as ·, tabs as →, trailing newlines as ↵). Runs of identical lines beyond
// the context window are collapsed into a summary line.
func WithGutter() Option {
	return func(conf *config) {
		conf.gutter = true
	}
}

// WithColor enables ANSI color output: deletions are red (\033[31m) and insertions are
// green (\033[32m). The caller is responsible for terminal detection and NO_COLOR handling.
func WithColor() Option {
	return func(conf *config) {
		conf.color = true
	}
}

// Write writes the edits to w. By default it produces unified diff output with hunk headers
// and 3 lines of context. Use [WithGutter] and [WithContext] to configure the output.
func Write(w io.Writer, edits []Edit, opts ...Option) error {
	conf := &config{context: 3}
	for _, opt := range opts {
		opt(conf)
	}
	hunks, maxOldLine := buildHunks(edits, conf.context)
	var lw int
	if conf.gutter {
		lw = 1
		for v := maxOldLine; v > 9; v /= 10 {
			lw++
		}
	}
	bw := bufio.NewWriter(w)
	if err := writeHunks(bw, edits, hunks, conf, lw); err != nil {
		return err
	}
	return bw.Flush()
}

// hunk represents a group of contiguous changes with surrounding context lines.
type hunk struct {
	startOld, startNew int // 1-indexed start line numbers
	countOld, countNew int // number of lines in the hunk per side
	start, end         int // index range into edits [start, end)
}

// buildHunks groups edits into hunks, merging hunks separated by fewer than 2*context equal
// lines. It also returns maxOldLine, the highest line number in the old sequence (for gutter
// line-number width).
func buildHunks(edits []Edit, context int) (hunks []hunk, maxOldLine int) {
	var lineOld int // current line number in the old sequence
	var lineNew int // current line number in the new sequence

	hunkStart := -1  // index into edits where the current hunk starts (0-indexed, -1 if no active hunk)
	var hunkEnd int  // index into edits where the current hunk ends (0-indexed, inclusive)
	var startOld int // start line in the old sequence for the current hunk (1-indexed)
	var startNew int // start line in the new sequence for the current hunk (1-indexed)
	var countOld int // number of old lines in the current hunk
	var countNew int // number of new lines in the current hunk
	var eqCount int // consecutive equal lines since the last change

	for i, edit := range edits {
		switch edit.Op {
		case Eq:
			lineNew++
			lineOld++
			maxOldLine++

			if hunkStart >= 0 {
				hunkEnd = i

				// set start line for the side that did not initiate the hunk
				if context > 0 {
					if startOld == 0 {
						startOld = lineOld
					} else if startNew == 0 {
						startNew = lineNew
					}
				} else {
					if startOld == 0 {
						startOld = lineOld - 1
					} else if startNew == 0 {
						startNew = lineNew - 1
					}
				}

				if eqCount+1 > 2*context { // hunk end
					// adjust for the extra eq we counted to wait for a possibly merged hunk
					if context > 0 && eqCount > context {
						adjust := eqCount - context
						countOld -= adjust
						countNew -= adjust
						hunkEnd -= adjust
					}

					hunks = append(hunks, hunk{
						startOld: startOld,
						countOld: countOld,
						startNew: startNew,
						countNew: countNew,
						start:    hunkStart,
						end:      hunkEnd,
					})
					hunkStart = -1
					hunkEnd = -1
					startNew = 0
					startOld = 0
					eqCount = 0
					countNew = 0
					countOld = 0
				} else {
					eqCount++
					countNew++
					countOld++
				}
			}
		case Ins:
			lineNew++
			countNew++
			eqCount = 0
			hunkEnd = i

			if hunkStart < 0 { // starting new hunk
				hunkStart = max(0, i-context)
				context := i - hunkStart
				countOld += context
				countNew += context
				if context > 0 {
					startOld = lineOld
				}
				startNew = lineNew - context
			} else { // part of an existing hunk
				if startNew == 0 {
					startNew = lineNew
				}
			}
		case Del:
			lineOld++
			maxOldLine++
			countOld++
			eqCount = 0
			hunkEnd = i

			if hunkStart < 0 { // starting new hunk
				hunkStart = max(0, i-context)
				context := i - hunkStart
				countOld += context
				countNew += context
				if context > 0 {
					startNew = lineNew
				}
				startOld = lineOld - context
			} else { // part of an existing hunk
				if startOld == 0 {
					startOld = lineOld
				}
			}
		}
	}

	// flush remaining hunk
	if hunkStart >= 0 {
		if startOld == 0 {
			startOld = lineOld
		} else if startNew == 0 {
			startNew = lineNew
		}
		// adjust for the extra eq we counted to wait for a possibly merged hunk
		if context > 0 && eqCount > context {
			adjust := eqCount - context
			countOld -= adjust
			countNew -= adjust
			hunkEnd -= adjust
		}

		hunks = append(hunks, hunk{
			startOld: startOld,
			countOld: countOld,
			startNew: startNew,
			countNew: countNew,
			start:    hunkStart,
			end:      hunkEnd + 1,
		})
	}
	return
}

func writeHunks(w *bufio.Writer, edits []Edit, hunks []hunk, conf *config, lineWidth int) error {
	for i, h := range hunks {
		if !conf.gutter {
			if err := writeHunkHeader(w, h.startOld, h.countOld, h.startNew, h.countNew); err != nil {
				return err
			}
		} else if i != 0 {
			collapsedEqs := h.start - hunks[i-1].end
			if _, err := fmt.Fprintf(w, "%*s───┼─── %d identical line(s) ───\n", lineWidth, "", collapsedEqs); err != nil {
				return err
			}
		}

		oldLine := h.startOld
		for j := h.start; j < h.end; j++ {
			e := edits[j]
			if err := writeEdit(w, e, oldLine, conf, lineWidth); err != nil {
				return err
			}
			if e.Op != Ins {
				oldLine++
			}
		}
	}
	return nil
}

// writeHunkHeader writes a hunk header in unified diff format.
// When count is 1, it is omitted (e.g., @@ -2 +2 @@ instead of @@ -2,1 +2,1 @@).
func writeHunkHeader(w io.Writer, oldStart, oldCount, newStart, newCount int) error {
	_, err := fmt.Fprintf(w, "@@ -%s +%s @@\n", hunkRange(oldStart, oldCount), hunkRange(newStart, newCount))
	return err
}

func hunkRange(start, count int) string {
	if count == 1 {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d,%d", start, count)
}

func writeEdit(w *bufio.Writer, e Edit, oldLine int, conf *config, lineWidth int) error {
	line := e.NewLine
	if e.Op == Del {
		line = e.OldLine
	}
	if conf.color && e.Op != Eq {
		var err error
		if e.Op == Del {
			_, err = w.WriteString("\033[31m")
		} else {
			_, err = w.WriteString("\033[32m")
		}
		if err != nil {
			return err
		}
	}
	if conf.gutter {
		if e.Op != Ins {
			if _, err := fmt.Fprintf(w, "%*d ", lineWidth, oldLine); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "%*s", lineWidth+1, ""); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(e.Op.String()); err != nil {
			return err
		}
		if _, err := w.WriteString(" │ "); err != nil {
			return err
		}
		if err := writeLine(w, line, e.Op != Eq, conf); err != nil {
			return err
		}
		return writeReset(w, e.Op, conf)
	}
	if _, err := w.WriteString(e.Op.String()); err != nil {
		return err
	}
	if err := writeLine(w, line, false, conf); err != nil {
		return err
	}
	return writeReset(w, e.Op, conf)
}

func writeReset(w *bufio.Writer, op OpType, conf *config) error {
	if conf.color && op != Eq {
		_, err := w.WriteString("\033[0m")
		return err
	}
	return nil
}

func writeLine(w *bufio.Writer, s string, showWhitespace bool, conf *config) error {
	hasNewline := len(s) > 0 && s[len(s)-1] == '\n'

	if showWhitespace {
		content := s
		if hasNewline {
			content = s[:len(s)-1]
		}
		for _, r := range content {
			switch r {
			case ' ':
				if _, err := w.WriteRune('·'); err != nil {
					return err
				}
			case '\t':
				if _, err := w.WriteRune('→'); err != nil {
					return err
				}
			default:
				if _, err := w.WriteRune(r); err != nil {
					return err
				}
			}
		}
		if hasNewline {
			if _, err := w.WriteRune('↵'); err != nil {
				return err
			}
		}
		if _, err := w.WriteString("\n"); err != nil {
			return err
		}
	} else {
		if _, err := w.WriteString(s); err != nil {
			return err
		}
		if !hasNewline {
			if conf.gutter {
				if _, err := w.WriteString("\n"); err != nil {
					return err
				}
			} else {
				if _, err := w.WriteString("\n\\ No newline at end of file\n"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
