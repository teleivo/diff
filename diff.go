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

type config struct {
	context int
	gutter  bool
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
// visible (spaces as ·, tabs as →, newline-only lines as ↵). Runs of identical lines beyond
// the context window are collapsed into a summary line.
func WithGutter(conf *config) {
	conf.gutter = true
}

// Write writes the edits to w. By default it produces unified diff output with hunk headers
// and 3 lines of context. Use [WithGutter] and [WithContext] to configure the output.
func Write(w io.Writer, edits []Edit, opts ...Option) error {
	conf := &config{context: 3}
	for _, opt := range opts {
		opt(conf)
	}
	uw := &unifiedWriter{w: bufio.NewWriter(w), conf: conf, edits: edits, hunkStart: -1, hunkEnd: -1}
	if err := uw.write(); err != nil {
		return err
	}
	return uw.w.Flush()
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

// unifiedWriter writes edits as unified diff hunks. It groups changes into hunks, merging
// hunks that are separated by fewer than 2*context equal lines.
type unifiedWriter struct {
	w       *bufio.Writer
	conf    *config
	edits   []Edit
	eqCount int // consecutive equal lines since the last change

	lineOld int // current line number in the old sequence
	lineNew int // current line number in the new sequence

	hunkStart int // index into edits where the current hunk starts (0-indexed, -1 if no active hunk)
	hunkEnd   int // index into edits where the current hunk ends (0-indexed, inclusive)
	startOld  int // start line in the old sequence for the current hunk (1-indexed)
	startNew  int // start line in the new sequence for the current hunk (1-indexed)
	countOld  int // number of old lines in the current hunk
	countNew  int // number of new lines in the current hunk
}

func (uw *unifiedWriter) write() error {
	for i := 0; i < len(uw.edits); i++ {
		switch uw.edits[i].Op {
		case Eq:
			uw.lineNew++
			uw.lineOld++

			if uw.hunkStart >= 0 {
				uw.hunkEnd = i

				// set start line for the side that did not initiate the hunk
				if uw.conf.context > 0 {
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

				if uw.eqCount+1 > 2*uw.conf.context { // hunk end
					// adjust for the extra eq we counted to wait for a possibly merged hunk
					if uw.conf.context > 0 && uw.eqCount > uw.conf.context {
						adjust := uw.eqCount - uw.conf.context
						uw.countOld -= adjust
						uw.countNew -= adjust
						uw.hunkEnd -= adjust
					}

					if err := uw.writeHunk(uw.hunkEnd); err != nil {
						return err
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
				uw.hunkStart = max(0, i-uw.conf.context)
				context := i - uw.hunkStart
				uw.countOld += context
				uw.countNew += context
				if context > 0 {
					uw.startOld = uw.lineOld
				}
				uw.startNew = uw.lineNew - context
			} else { // part of an existing hunk
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
				uw.hunkStart = max(0, i-uw.conf.context)
				context := i - uw.hunkStart
				uw.countOld += context
				uw.countNew += context
				if context > 0 {
					uw.startNew = uw.lineNew
				}
				uw.startOld = uw.lineOld - context
			} else { // part of an existing hunk
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
		if uw.conf.context > 0 && uw.eqCount > uw.conf.context {
			adjust := uw.eqCount - uw.conf.context
			uw.countOld -= adjust
			uw.countNew -= adjust
			uw.hunkEnd -= adjust
		}

		if err := uw.writeHunk(uw.hunkEnd + 1); err != nil {
			return err
		}
	}
	return nil
}

// writeHunk writes the hunk header and edits from hunkStart up to but not including end.
func (uw *unifiedWriter) writeHunk(end int) error {
	if !uw.conf.gutter {
		if err := writeHunkHeader(uw.w, uw.startOld, uw.countOld, uw.startNew, uw.countNew); err != nil {
			return err
		}
	}

	oldLine := uw.startOld
	for j := uw.hunkStart; j < end; j++ {
		e := uw.edits[j]
		// Detect missing-final-newline asymmetry in Del/Ins pairs.
		// When one side has a trailing newline and the other doesn't,
		// the side with the newline gets a ↵ appended.
		var showNewlineMark bool
		if uw.conf.gutter && (e.Op == Del || e.Op == Ins) {
			var pair *Edit
			if e.Op == Del && j+1 < end && uw.edits[j+1].Op == Ins {
				pair = &uw.edits[j+1]
			} else if e.Op == Ins && j-1 >= uw.hunkStart && uw.edits[j-1].Op == Del {
				pair = &uw.edits[j-1]
			}
			if pair != nil {
				line := e.NewLine
				pairLine := pair.NewLine
				if e.Op == Del {
					line = e.OldLine
					pairLine = pair.NewLine
				} else {
					pairLine = pair.OldLine
				}
				lineHasNL := len(line) > 0 && line[len(line)-1] == '\n'
				pairHasNL := len(pairLine) > 0 && pairLine[len(pairLine)-1] == '\n'
				showNewlineMark = lineHasNL && !pairHasNL
			}
		}
		if err := uw.writeEdit(e, oldLine, showNewlineMark); err != nil {
			return err
		}
		if e.Op != Ins {
			oldLine++
		}
	}
	return nil
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

func (uw *unifiedWriter) writeEdit(e Edit, oldLine int, showNewlineMark bool) error {
	if uw.conf.gutter {
		if e.Op != Ins {
			if _, err := fmt.Fprintf(uw.w, "%d ", oldLine); err != nil {
				return err
			}
		} else {
			if _, err := uw.w.WriteString("  "); err != nil {
				return err
			}
		}
		if _, err := uw.w.WriteString(e.Op.String()); err != nil {
			return err
		}
		if _, err := uw.w.WriteString(" │ "); err != nil {
			return err
		}
		replace := e.Op != Eq
		line := e.NewLine
		if e.Op == Del {
			line = e.OldLine
		}
		return uw.writeLine(line, replace, showNewlineMark)
	}
	if _, err := uw.w.WriteString(e.Op.String()); err != nil {
		return err
	}
	line := e.NewLine
	if e.Op == Del {
		line = e.OldLine
	}
	return uw.writeLine(line, false, false)
}

func (uw *unifiedWriter) writeLine(s string, replace bool, showNewlineMark bool) error {
	hasNewline := len(s) > 0 && s[len(s)-1] == '\n'

	if replace {
		content := s
		if hasNewline {
			content = s[:len(s)-1]
		}
		if len(content) == 0 && hasNewline {
			// Newline-only line (blank line): render as ↵
			if _, err := uw.w.WriteRune('↵'); err != nil {
				return err
			}
		} else {
			for _, r := range content {
				switch r {
				case ' ':
					if _, err := uw.w.WriteRune('·'); err != nil {
						return err
					}
				case '\t':
					if _, err := uw.w.WriteRune('→'); err != nil {
						return err
					}
				default:
					if _, err := uw.w.WriteRune(r); err != nil {
						return err
					}
				}
			}
			if showNewlineMark {
				if _, err := uw.w.WriteRune('↵'); err != nil {
					return err
				}
			}
		}
		if _, err := uw.w.WriteString("\n"); err != nil {
			return err
		}
	} else {
		if _, err := uw.w.WriteString(s); err != nil {
			return err
		}
		if !hasNewline {
			if uw.conf.gutter {
				if _, err := uw.w.WriteString("\n"); err != nil {
					return err
				}
			} else {
				if _, err := uw.w.WriteString("\n\\ No newline at end of file\n"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
