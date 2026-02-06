# TODO

* use this library in assertive
  * implement
  * change api to just diff.Write defaulting to unified diff -U 3. add functional options to set
  tyle to gutter? replace whitespace, add color
  * look for bugs
  * git tag
  * depend  assertive
  * use new assertion in dot project

* cli: add color output support - use ANSI escape sequences (e.g., `\033[31m` for red, `\033[32m` for
  green, `\033[0m` to reset). Only emit colors when output is a terminal. Detect with stdlib:
  `fi, _ := os.Stdout.Stat(); isTerminal := (fi.Mode() & os.ModeCharDevice) != 0`. Allow override
  via `--color=always` or `--color=never` flag. Common convention: red for deletions, green for
  insertions

* feat: add collapsed context to gutter? what would that be good for?

* use dot/kitty image protocol to show an animation of it that works in ghostty

* refactor to linear space version (Section 4b of Myers paper) - current implementation uses O(D²)
  space for the trace; the linear space version uses divide-and-conquer to find the "middle snake"
  and only requires O(N) space. Also consider only cloning the active diagonal range [-d, d] per
  iteration instead of the full v slice to reduce per-clone cost from O(N+M) to O(d)
  * write benchmark and/or add cpu/memprofile flags

