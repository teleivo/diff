package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/teleivo/diff/internal/myers"
)

// Myers animates the quadratic Myers diff algorithm on an edit graph.
type Myers struct {
	old   []string
	new   []string
	n     int // len(old)
	m     int // len(new)
	maxD  int
	trace [][]int // trace[d] = V snapshot before processing edit distance d
	path  []pathStep

	frameIdx int // current animation frame index
	// frames: 0 = initial grid, 1..len(trace) = d-passes, last = optimal path
}

type pathStep struct {
	x, y int
}

func NewMyers(old, new []string) *Myers {
	my := &Myers{
		old:  old,
		new:  new,
		n:    len(old),
		m:    len(new),
		maxD: len(old) + len(new),
	}
	my.trace, my.path = my.compute()
	return my
}

// compute runs the quadratic Myers algorithm and returns the trace and optimal path.
func (my *Myers) compute() ([][]int, []pathStep) {
	if my.maxD == 0 {
		return nil, nil
	}
	trace, maxD := myers.Trace(my.old, my.new)
	my.maxD = maxD
	path := my.buildPath(trace)
	return trace, path
}

// buildPath reconstructs the optimal path nodes by backtracking through the trace.
// Returns grid positions (x,y) on the optimal path, in order from (0,0) to (n,m).
func (my *Myers) buildPath(trace [][]int) []pathStep {
	n, m, maxD := my.n, my.m, my.maxD
	x, y := n, m

	pathSet := make(map[pathStep]bool)
	pathSet[pathStep{n, m}] = true

	for d := len(trace) - 1; d >= 1; d-- {
		v := trace[d]
		k := x - y
		i := k + maxD

		var prevK int
		if k == -d || (k != d && v[i-1] < v[i+1]) {
			prevK = k + 1 // insert (down)
		} else {
			prevK = k - 1 // delete (right)
		}
		prevX := v[prevK+maxD]
		prevY := prevX - prevK

		// The edit step (right or down) from (prevX, prevY) to snake start
		var editEndX, editEndY int
		if prevK == k+1 {
			// insert: y increases by 1
			editEndX = prevX
			editEndY = prevY + 1
		} else {
			// delete: x increases by 1
			editEndX = prevX + 1
			editEndY = prevY
		}
		pathSet[pathStep{prevX, prevY}] = true
		pathSet[pathStep{editEndX, editEndY}] = true

		// Mark all snake nodes from editEnd to (x, y)
		sx, sy := editEndX, editEndY
		for sx <= x && sy <= y {
			pathSet[pathStep{sx, sy}] = true
			sx++
			sy++
		}

		x, y = prevX, prevY
	}

	// Mark initial snake from (0,0)
	pathSet[pathStep{0, 0}] = true
	sx, sy := 0, 0
	for sx < n && sy < m && my.old[sx] == my.new[sy] {
		sx++
		sy++
		pathSet[pathStep{sx, sy}] = true
	}

	// Collect and sort path nodes by (x+y, x)
	steps := make([]pathStep, 0, len(pathSet))
	for ps := range pathSet {
		steps = append(steps, ps)
	}
	slices.SortFunc(steps, func(a, b pathStep) int {
		da, db := a.x+a.y, b.x+b.y
		if da != db {
			return da - db
		}
		return a.x - b.x
	})
	return steps
}

// Next returns the next DOT frame, or "" when done.
func (my *Myers) Next() string {
	// Frames: 0 = initial grid, 1..len(trace)-1 = d-passes, len(trace) = final path
	total := len(my.trace) + 1 // +1 for the final path frame
	if my.frameIdx >= total {
		return ""
	}
	frame := my.frameIdx
	my.frameIdx++

	if frame == 0 {
		return my.toDOT(-1, false)
	}
	d := frame - 1
	if d < len(my.trace)-1 {
		return my.toDOT(d, false)
	}
	return my.toDOT(-1, true)
}

// Summary returns a summary string after animation.
func (my *Myers) Summary() string {
	d := len(my.trace) - 2 // number of edit passes (trace includes d=0)
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("Done: %d edits to transform %v → %v", d, my.old, my.new)
}

// toDOT generates a DOT string for the given state.
// d: current edit distance being highlighted (-1 = none)
// showPath: if true, draw the optimal path in green
func (my *Myers) toDOT(d int, showPath bool) string {
	n, m, maxD := my.n, my.m, my.maxD
	const scale = 72.0 // neato uses points (72pt = 1 inch); space nodes 1 inch apart

	var sb strings.Builder
	sb.WriteString("graph G {\n")
	sb.WriteString("  graph [bgcolor=white, splines=false];\n")
	sb.WriteString("  node [shape=circle, width=0.3, fixedsize=true, fontsize=10];\n")
	sb.WriteString("  edge [color=\"#cccccc\"];\n")

	// Determine which nodes are on the frontier up to current d,
	// and which are on the optimal path.
	type nodeKey struct{ x, y int }
	frontierD := make(map[nodeKey]int) // node -> d value when first reached
	if d >= 0 {
		for di := 0; di <= d && di < len(my.trace); di++ {
			v := my.trace[di]
			for k := -di; k <= di; k += 2 {
				if k > n || k < -m {
					continue
				}
				i := k + maxD
				x := v[i]
				if di > 0 {
					// x from after snakes; compute snake endpoints
					// find start of snake
					var sx int
					prevV := my.trace[di-1]
					if k == -di || (k != di && prevV[i-1] < prevV[i+1]) {
						sx = prevV[i+1]
					} else {
						sx = prevV[i-1] + 1
					}
					sy := sx - k
					// mark snake nodes
					for sx <= x && sy <= x-k {
						nk := nodeKey{sx, sy}
						if _, exists := frontierD[nk]; !exists {
							frontierD[nk] = di
						}
						sx++
						sy++
					}
				} else {
					// d=0: starting snake from (0,0)
					sx, sy := 0, 0
					for sx <= x && sy <= x-k {
						nk := nodeKey{sx, sy}
						if _, exists := frontierD[nk]; !exists {
							frontierD[nk] = 0
						}
						sx++
						sy++
					}
				}
			}
		}
	}

	pathNodes := make(map[nodeKey]bool)
	if showPath {
		for _, s := range my.path {
			pathNodes[nodeKey{s.x, s.y}] = true
		}
	}

	// Colors for d values (cycle through a palette)
	dColors := []string{
		"#e8f5e9", // d=0: very light green
		"#fff9c4", // d=1: light yellow
		"#ffe0b2", // d=2: light orange
		"#fce4ec", // d=3: light pink
		"#e3f2fd", // d=4: light blue
		"#f3e5f5", // d=5: light purple
		"#e0f7fa", // d=6: light cyan
		"#f9fbe7", // d=7: light lime
	}

	// Emit axis labels: old sequence along the top (x-axis), new along the left (y-axis)
	for x := 0; x < n; x++ {
		px := (float64(x) + 0.5) * scale
		py := 0.6 * scale // above the top row
		fmt.Fprintf(&sb, "  xlabel_old%d [label=%q, pos=\"%.2f,%.2f!\", shape=plaintext, fontsize=12, fontcolor=\"#333333\"];\n",
			x, my.old[x], px, py)
	}
	for y := 0; y < m; y++ {
		px := -0.6 * scale // left of the leftmost column
		py := float64(-y-1)*scale + scale*0.5
		fmt.Fprintf(&sb, "  xlabel_new%d [label=%q, pos=\"%.2f,%.2f!\", shape=plaintext, fontsize=12, fontcolor=\"#333333\"];\n",
			y, my.new[y], px, py)
	}

	// Emit nodes
	for y := 0; y <= m; y++ {
		for x := 0; x <= n; x++ {
			nk := nodeKey{x, y}
			px := float64(x) * scale
			py := float64(-y) * scale

			var fillColor, borderColor, fontColor string
			borderWidth := 1.0

			if showPath && pathNodes[nk] {
				fillColor = "#00c853"
				borderColor = "#007700"
				fontColor = "white"
				borderWidth = 2.0
			} else if di, ok := frontierD[nk]; ok {
				idx := di % len(dColors)
				fillColor = dColors[idx]
				borderColor = "#999999"
				fontColor = "black"
			} else {
				fillColor = "white"
				borderColor = "#cccccc"
				fontColor = "#cccccc"
			}

			label := fmt.Sprintf("%d,%d", x, y)
			fmt.Fprintf(&sb, "  n%d_%d [label=%q, pos=\"%.2f,%.2f!\", style=filled, fillcolor=%q, color=%q, fontcolor=%q, penwidth=%.1f];\n",
				x, y, label, px, py, fillColor, borderColor, fontColor, borderWidth)
		}
	}

	// Emit grid edges (right = delete, down = insert, diagonal = equal)
	// Right edges
	for y := 0; y <= m; y++ {
		for x := 0; x < n; x++ {
			if showPath && pathNodes[nodeKey{x, y}] && pathNodes[nodeKey{x + 1, y}] {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=\"#ff5252\", penwidth=2.0, style=dashed];\n",
					x, y, x+1, y)
			} else {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [style=dashed];\n", x, y, x+1, y)
			}
		}
	}

	// Down edges
	for y := 0; y < m; y++ {
		for x := 0; x <= n; x++ {
			if showPath && pathNodes[nodeKey{x, y}] && pathNodes[nodeKey{x, y + 1}] {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=\"#2196f3\", penwidth=2.0, style=dotted];\n",
					x, y, x, y+1)
			} else {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [style=dotted];\n", x, y, x, y+1)
			}
		}
	}

	// Diagonal edges (only where old[x] == new[y])
	for y := 0; y < m; y++ {
		for x := 0; x < n; x++ {
			if my.old[x] == my.new[y] {
				if showPath && pathNodes[nodeKey{x, y}] && pathNodes[nodeKey{x + 1, y + 1}] {
					fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=\"#00c853\", penwidth=2.5];\n",
						x, y, x+1, y+1)
				} else {
					fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=\"#aaaaaa\"];\n",
						x, y, x+1, y+1)
				}
			}
		}
	}

	// Label
	var labelText string
	if showPath {
		numEdits := len(my.trace) - 2
		if numEdits < 0 {
			numEdits = 0
		}
		labelText = fmt.Sprintf("Done: %d edits (optimal path in green)", numEdits)
	} else if d >= 0 {
		labelText = fmt.Sprintf("d=%d: exploring edit distance %d", d, d)
	} else {
		labelText = "Edit graph (right=delete, down=insert, diagonal=equal)"
	}
	fmt.Fprintf(&sb, "  label=%q;\n", labelText)
	sb.WriteString("  labelloc=t;\n")
	sb.WriteString("  fontsize=14;\n")

	sb.WriteString("}\n")
	return sb.String()
}
