# TODO

* re-write using []Edit -> []Hunk pass
* look for bugs

* cli: add color output support - use ANSI escape sequences (e.g., `\033[31m` for red, `\033[32m` for
  green, `\033[0m` to reset). Only emit colors when output is a terminal. Detect with stdlib:
  `fi, _ := os.Stdout.Stat(); isTerminal := (fi.Mode() & os.ModeCharDevice) != 0`. Allow override
  via `--color=always` or `--color=never` flag. Common convention: red for deletions, green for
  insertions

* use dot/kitty image protocol to show an animation of it that works in ghostty

* refactor: to linear space version (Section 4b of Myers paper) - current implementation uses O(D²)
  space for the trace; the linear space version uses divide-and-conquer to find the "middle snake"
  and only requires O(N) space. Also consider only cloning the active diagonal range [-d, d] per
  iteration instead of the full v slice to reduce per-clone cost from O(N+M) to O(d)
  * write benchmark and/or add cpu/memprofile flags

