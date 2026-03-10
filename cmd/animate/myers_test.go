package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

func TestMyersFramesDOTValid(t *testing.T) {
	old := []string{"A", "B", "C", "A", "B", "B", "A"}
	new_ := []string{"C", "B", "A", "B", "A", "C"}
	my := NewMyers(old, new_)

	frameNum := 0
	for {
		dot := my.Next()
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

	summary := my.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	t.Logf("Summary: %s", summary)
}

func TestMyersBacktrack(t *testing.T) {
	old := []string{"A", "B", "C", "A", "B", "B", "A"}
	new_ := []string{"C", "B", "A", "B", "A", "C"}
	my := NewMyers(old, new_)

	// Path should start at (0,0) and end at (n,m)
	if len(my.path) == 0 {
		t.Fatal("expected non-empty path")
	}
	first := my.path[0]
	last := my.path[len(my.path)-1]
	if first.x != 0 || first.y != 0 {
		t.Errorf("path should start at (0,0), got (%d,%d)", first.x, first.y)
	}
	if last.x != my.n || last.y != my.m {
		t.Errorf("path should end at (%d,%d), got (%d,%d)", my.n, my.m, last.x, last.y)
	}
	fmt.Printf("Path: %v\n", my.path)
}
