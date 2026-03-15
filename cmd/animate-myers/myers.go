package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/teleivo/diff/internal/myers"
)

const (
	colorPathFill   = "#00c853" // green — optimal path nodes
	colorPathBorder = "#007700" // dark green — optimal path node border
	colorDelete     = "#ff5252" // red — delete edges on path
	colorInsert     = "#2196f3" // blue — insert edges on path
	colorDiagonal   = "#00c853" // green — diagonal edges on path
	colorDiagonalBG = "#aaaaaa" // gray — diagonal edges not on path
	colorGridEdge   = "#cccccc" // light gray — default grid edges
	colorGridNode   = "#cccccc" // light gray — unvisited node border/font
	colorFrontier   = "#999999" // gray — frontier node border
	colorLabel      = "#333333" // dark gray — axis labels
)

// dColors maps edit distance d (mod len) to a fill color for frontier nodes.
var dColors = [...]string{
	"#e8f5e9", // d=0: very light green
	"#fff9c4", // d=1: light yellow
	"#ffe0b2", // d=2: light orange
	"#fce4ec", // d=3: light pink
	"#e3f2fd", // d=4: light blue
	"#f3e5f5", // d=5: light purple
	"#e0f7fa", // d=6: light cyan
	"#f9fbe7", // d=7: light lime
}

// animation animates the quadratic Myers diff algorithm on an edit graph.
type animation struct {
	oldLines []string
	newLines []string
	n        int                // len(oldLines)
	m        int                // len(newLines)
	edits    int                // shortest edit distance
	trace    [][]int            // trace[d] = V snapshot before processing edit distance d
	frontier map[point]int      // node -> d value when first reached
	path     []point            // optimal path nodes in order
	pathSet  map[point]struct{} // path nodes for O(1) lookup

	frameIdx int // current animation frame index
	// frames: 0 = initial grid, 1..len(trace) = d-passes, last = optimal path
}

type point struct {
	x, y int
}

func newAnimation(oldLines, newLines []string) *animation {
	n := len(oldLines)
	m := len(newLines)
	a := &animation{
		oldLines: oldLines,
		newLines: newLines,
		n:        n,
		m:        m,
	}
	a.trace, _ = myers.Trace(a.oldLines, a.newLines)
	a.edits = max(len(a.trace)-1, 0)
	a.frontier = a.buildFrontier()
	a.path = a.backtrack(a.trace)
	a.pathSet = make(map[point]struct{}, len(a.path))
	for _, p := range a.path {
		a.pathSet[p] = struct{}{}
	}
	return a
}

// buildFrontier precomputes which nodes are reached at each edit distance d.
func (a *animation) buildFrontier() map[point]int {
	n, m := a.n, a.m
	maxD := n + m
	frontier := make(map[point]int)
	for di, v := range a.trace {
		for k := -di; k <= di; k += 2 {
			if k > n || k < -m {
				continue
			}
			i := k + maxD
			x := v[i]

			var sx int
			if di > 0 {
				prevV := a.trace[di-1]
				if k == -di || (k != di && prevV[i-1] < prevV[i+1]) {
					sx = prevV[i+1]
				} else {
					sx = prevV[i-1] + 1
				}
			}
			sy := sx - k
			for sx <= x && sy <= x-k {
				if _, exists := frontier[point{sx, sy}]; !exists {
					frontier[point{sx, sy}] = di
				}
				sx++
				sy++
			}
		}
	}
	return frontier
}

// backtrack reconstructs the optimal path nodes by backtracking through the trace.
// Returns grid positions (x,y) on the optimal path, in order from (0,0) to (n,m).
func (a *animation) backtrack(trace [][]int) []point {
	n, m := a.n, a.m
	maxD := n + m
	x, y := n, m

	visited := make(map[point]struct{})
	visited[point{n, m}] = struct{}{}

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
		visited[point{prevX, prevY}] = struct{}{}
		visited[point{editEndX, editEndY}] = struct{}{}

		// Mark all snake nodes from editEnd to (x, y)
		sx, sy := editEndX, editEndY
		for sx <= x && sy <= y {
			visited[point{sx, sy}] = struct{}{}
			sx++
			sy++
		}

		x, y = prevX, prevY
	}

	// Initial snake from (0,0)
	visited[point{0, 0}] = struct{}{}
	for sx, sy := 0, 0; sx < n && sy < m && a.oldLines[sx] == a.newLines[sy]; {
		sx++
		sy++
		visited[point{sx, sy}] = struct{}{}
	}

	path := make([]point, 0, len(visited))
	for p := range visited {
		path = append(path, p)
	}
	slices.SortFunc(path, func(a, b point) int {
		return cmp.Or(cmp.Compare(a.x+a.y, b.x+b.y), cmp.Compare(a.x, b.x))
	})
	return path
}

// next returns the next DOT frame, or "" when done.
func (a *animation) next() string {
	// Frames: 0 = initial grid, 1..len(trace)-1 = d-passes, len(trace) = final path
	total := len(a.trace) + 1 // +1 for the final path frame
	if a.frameIdx >= total {
		return ""
	}
	frame := a.frameIdx
	a.frameIdx++

	if frame == 0 {
		return a.toDOT(-1, false)
	}
	if d := frame - 1; d < len(a.trace)-1 {
		return a.toDOT(d, false)
	}
	return a.toDOT(-1, true)
}

func (a *animation) onPath(p point) bool {
	_, ok := a.pathSet[p]
	return ok
}

// toDOT generates a DOT string for the given state.
// d: current edit distance being highlighted (-1 = none)
// showPath: if true, draw the optimal path in green
func (a *animation) toDOT(d int, showPath bool) string {
	n, m := a.n, a.m
	const scale = 72.0 // neato uses points (72pt = 1 inch); space nodes 1 inch apart

	var sb strings.Builder
	sb.WriteString(`graph G {
  graph [bgcolor=white, splines=false];
  node [shape=circle, width=0.3, fixedsize=true, fontsize=10];
  edge [color="#cccccc"];
`)

	// Emit axis labels: old sequence along the top (x-axis), new along the left (y-axis)
	for x := range n {
		px := (float64(x) + 0.5) * scale
		py := 0.6 * scale // above the top row
		fmt.Fprintf(&sb, "  xlabel_old%d [label=%q, pos=\"%.2f,%.2f!\", shape=plaintext, fontsize=12, fontcolor=%q];\n",
			x, a.oldLines[x], px, py, colorLabel)
	}
	for y := range m {
		px := -0.6 * scale // left of the leftmost column
		py := float64(-y-1)*scale + scale*0.5
		fmt.Fprintf(&sb, "  xlabel_new%d [label=%q, pos=\"%.2f,%.2f!\", shape=plaintext, fontsize=12, fontcolor=%q];\n",
			y, a.newLines[y], px, py, colorLabel)
	}

	// Emit nodes
	for y := 0; y <= m; y++ {
		for x := 0; x <= n; x++ {
			nk := point{x, y}
			px := float64(x) * scale
			py := float64(-y) * scale

			var fillColor, borderColor, fontColor string
			borderWidth := 1.0

			if showPath && a.onPath(nk) {
				fillColor = colorPathFill
				borderColor = colorPathBorder
				fontColor = "white"
				borderWidth = 2.0
			} else if di, ok := a.frontier[nk]; ok && d >= 0 && di <= d {
				idx := di % len(dColors)
				fillColor = dColors[idx]
				borderColor = colorFrontier
				fontColor = "black"
			} else {
				fillColor = "white"
				borderColor = colorGridNode
				fontColor = colorGridNode
			}

			label := fmt.Sprintf("%d,%d", x, y)
			fmt.Fprintf(&sb, "  n%d_%d [label=%q, pos=\"%.2f,%.2f!\", style=filled, fillcolor=%q, color=%q, fontcolor=%q, penwidth=%.1f];\n",
				x, y, label, px, py, fillColor, borderColor, fontColor, borderWidth)
		}
	}

	// Emit grid edges (right = delete, down = insert, diagonal = equal)
	// Right edges
	for y := 0; y <= m; y++ {
		for x := range n {
			if showPath && a.onPath(point{x, y}) && a.onPath(point{x + 1, y}) {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=%q, penwidth=2.0, style=dashed];\n",
					x, y, x+1, y, colorDelete)
			} else {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [style=dashed];\n", x, y, x+1, y)
			}
		}
	}

	// Down edges
	for y := range m {
		for x := 0; x <= n; x++ {
			if showPath && a.onPath(point{x, y}) && a.onPath(point{x, y + 1}) {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=%q, penwidth=2.0, style=dotted];\n",
					x, y, x, y+1, colorInsert)
			} else {
				fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [style=dotted];\n", x, y, x, y+1)
			}
		}
	}

	// Diagonal edges (only where old[x] == new[y])
	for y := range m {
		for x := range n {
			if a.oldLines[x] == a.newLines[y] {
				if showPath && a.onPath(point{x, y}) && a.onPath(point{x + 1, y + 1}) {
					fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=%q, penwidth=2.5];\n",
						x, y, x+1, y+1, colorDiagonal)
				} else {
					fmt.Fprintf(&sb, "  n%d_%d -- n%d_%d [color=%q];\n",
						x, y, x+1, y+1, colorDiagonalBG)
				}
			}
		}
	}

	// Label
	var labelText string
	if showPath {
		labelText = fmt.Sprintf("Done: %d edits (optimal path in green)", a.edits)
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

// summary returns a summary string after animation.
func (a *animation) summary() string {
	return fmt.Sprintf("Done: %d edits to transform %v → %v", a.edits, a.oldLines, a.newLines)
}
