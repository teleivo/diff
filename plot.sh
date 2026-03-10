#!/usr/bin/env bash
set -euo pipefail

# Parse Go benchmark output and plot time + space complexity with gnuplot.
#
# Usage: ./plot.sh results.bench
#
# Produces: bench_time.svg and bench_space.svg next to the input file.

if [[ $# -ne 1 ]]; then
    echo "usage: $0 <benchmark-output-file>" >&2
    exit 1
fi

input="$1"
outdir=$(cd "$(dirname "$input")" && pwd)
tmpdir=$(mktemp --directory)
trap 'rm -rf "$tmpdir"' EXIT

# Parse benchmark lines into per-series data files: N ns_op B_op
# One file per (algo, D) combination.
awk -v tmpdir="$tmpdir" '/^BenchmarkLines/ {
    split($1, parts, "/")
    split(parts[2], n_parts, "=")
    split(parts[3], d_parts, "=")
    split(parts[4], algo_parts, "-")
    n = n_parts[2]
    d = d_parts[2]
    algo = algo_parts[1]
    ns_op = $3
    b_op = $5
    dn = (d == n) ? "N" : d
    file = tmpdir "/" algo "_D" dn ".dat"
    print n, ns_op, b_op >> file
}' "$input"

gpfile="$tmpdir/plot.gp"

style='set logscale xy
set grid
set key left top
set lmargin 14

set style line 1 lc rgb "#2196F3" lw 2 pt 7 ps 1.2
set style line 2 lc rgb "#F44336" lw 2 pt 7 ps 1.2
set style line 3 lc rgb "#2196F3" lw 2 pt 5 ps 1.2 dt 2
set style line 4 lc rgb "#F44336" lw 2 pt 5 ps 1.2 dt 2
set style line 5 lc rgb "#555555" lw 1.5 dt 3
set style line 6 lc rgb "#555555" lw 1.5 dt 4'

# Plot time complexity (ns/op)
# Both algos are O(N*D). With D=N that is O(N^2); with D=10 that is O(N).
# Fit a power law c*x^a to each series so reference curves are derived from data.
cat > "$gpfile" <<EOF
set terminal svg size 900,500 font "sans,12" background "white"
set output "$outdir/bench_time.svg"
set title "Myers Diff: Time Complexity O(ND)"
set xlabel "N (sequence length)"
set ylabel "ns/op"
$style

# Fit O(N^2) reference to Linear D=N series (column 2 = ns/op)
f_n2(x) = a_n2 * x**2
a_n2 = 1.0
fit f_n2(x) "$tmpdir/Linear_DN.dat" using 1:2 via a_n2

# Fit O(N) reference to Linear D=10 series
f_n(x) = a_n * x
a_n = 1.0
fit f_n(x) "$tmpdir/Linear_D10.dat" using 1:2 via a_n

plot "$tmpdir/Linear_DN.dat" using 1:2 with linespoints ls 1 title "O(N+D) space, D=N", \
     "$tmpdir/Quadratic_DN.dat" using 1:2 with linespoints ls 2 title "O(ND) space, D=N", \
     "$tmpdir/Linear_D10.dat" using 1:2 with linespoints ls 3 title "O(N+D) space, D=10", \
     "$tmpdir/Quadratic_D10.dat" using 1:2 with linespoints ls 4 title "O(ND) space, D=10", \
     f_n2(x) with lines ls 5 title "O(N²)", \
     f_n(x) with lines ls 6 title "O(N)"
EOF
gnuplot "$gpfile"

# Plot space complexity (B/op)
# Quadratic: O(N*D) = O(N^2) when D=N; Linear: O(N+D) = O(N).
# Fit reference curves to the D=N series of each algo.
cat > "$gpfile" <<EOF
set terminal svg size 900,500 font "sans,12" background "white"
set output "$outdir/bench_space.svg"
set title "Myers Diff: Space Complexity O(N+D) vs O(ND)"
set xlabel "N (sequence length)"
set ylabel "B/op"
$style

# Fit O(N^2) reference to Quadratic D=N series (column 3 = B/op)
g_n2(x) = b_n2 * x**2
b_n2 = 1.0
fit g_n2(x) "$tmpdir/Quadratic_DN.dat" using 1:3 via b_n2

# Fit O(N) reference to Linear D=N series
g_n(x) = b_n * x
b_n = 1.0
fit g_n(x) "$tmpdir/Linear_DN.dat" using 1:3 via b_n

plot "$tmpdir/Linear_DN.dat" using 1:3 with linespoints ls 1 title "O(N+D) space, D=N", \
     "$tmpdir/Quadratic_DN.dat" using 1:3 with linespoints ls 2 title "O(ND) space, D=N", \
     "$tmpdir/Linear_D10.dat" using 1:3 with linespoints ls 3 title "O(N+D) space, D=10", \
     "$tmpdir/Quadratic_D10.dat" using 1:3 with linespoints ls 4 title "O(ND) space, D=10", \
     g_n2(x) with lines ls 5 title "O(N²)", \
     g_n(x) with lines ls 6 title "O(N)"
EOF
gnuplot "$gpfile"

echo "Generated $outdir/bench_time.svg and $outdir/bench_space.svg"
