package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestMyersFramesDOTValid(t *testing.T) {
	oldLines := []string{"A", "B", "C", "A", "B", "B", "A"}
	newLines := []string{"C", "B", "A", "B", "A", "C"}
	anim := newAnimation(oldLines, newLines)

	frameNum := 0
	for {
		dot := anim.next()
		if dot == "" {
			break
		}
		cmd := exec.Command("neato", "-n", "-Tplain")
		cmd.Stdin = bytes.NewBufferString(dot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("frame %d DOT error: %v\nDOT:\n%s\nOutput:\n%s", frameNum, err, dot, out)
		}
		frameNum++
	}
	t.Logf("Generated %d frames", frameNum)

	summary := anim.summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	t.Logf("summary: %s", summary)
}

func TestMyersBacktrack(t *testing.T) {
	oldLines := []string{"A", "B", "C", "A", "B", "B", "A"}
	newLines := []string{"C", "B", "A", "B", "A", "C"}
	anim := newAnimation(oldLines, newLines)

	if len(anim.path) == 0 {
		t.Fatal("expected non-empty path")
	}

	first := anim.path[0]
	if first.x != 0 || first.y != 0 {
		t.Errorf("path should start at (0,0), got (%d,%d)", first.x, first.y)
	}
	last := anim.path[len(anim.path)-1]
	if last.x != anim.n || last.y != anim.m {
		t.Errorf("path should end at (%d,%d), got (%d,%d)", anim.n, anim.m, last.x, last.y)
	}

	// Verify each step is a legal edit graph move and count edits.
	var edits int
	for i := 1; i < len(anim.path); i++ {
		prev := anim.path[i-1]
		cur := anim.path[i]
		dx := cur.x - prev.x
		dy := cur.y - prev.y

		switch {
		case dx == 1 && dy == 0: // delete (right)
			edits++
		case dx == 0 && dy == 1: // insert (down)
			edits++
		case dx == 1 && dy == 1: // diagonal (match)
			if oldLines[prev.x] != newLines[prev.y] {
				t.Errorf("diagonal at (%d,%d)→(%d,%d) but old[%d]=%q != new[%d]=%q",
					prev.x, prev.y, cur.x, cur.y, prev.x, oldLines[prev.x], prev.y, newLines[prev.y])
			}
		default:
			t.Errorf("illegal move (%d,%d)→(%d,%d)", prev.x, prev.y, cur.x, cur.y)
		}
	}

	if edits != anim.edits {
		t.Errorf("path has %d edits, want %d", edits, anim.edits)
	}
}
