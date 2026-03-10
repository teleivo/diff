// Package myers implements the quadratic space variant of the Myers diff algorithm,
// exposing the internal trace for use by both the diff package and visualization tools.
package myers

import "slices"

// Trace runs the quadratic Myers algorithm and returns the V-array snapshots,
// one per edit distance d (starting from d=0). The returned maxD is n+m.
// old and new are the sequences being compared.
func Trace(old, new []string) (trace [][]int, maxD int) {
	n := len(old)
	m := len(new)
	maxD = n + m
	if maxD == 0 {
		return nil, 0
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
			for x < n && y < m && old[x] == new[y] { // advance on diagonal
				x++
				y++
			}
			v[i] = x
			if x >= n && y >= m {
				return trace, maxD
			}
		}
	}
	return trace, maxD
}
