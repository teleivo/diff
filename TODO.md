# TODO

* look for bugs in diff algo/Write algo

* use dot/kitty image protocol to show an animation of it that works in ghostty

* refactor: to linear space version (Section 4b of Myers paper) - current implementation uses O(D²)
  space for the trace; the linear space version uses divide-and-conquer to find the "middle snake"
  and only requires O(N) space. Also consider only cloning the active diagonal range [-d, d] per
  iteration instead of the full v slice to reduce per-clone cost from O(N+M) to O(d)
  * write benchmark and/or add cpu/memprofile flags

