# TODO

* use dot/kitty image protocol to show an animation of it that works in ghostty
  see ../algo-animate/ as inspiration

* idea: reduce allocations by working with `[][]byte` instead of `[]string` and `mmap` files instead
of `os.ReadFile`. caveat: complicates API design since callers like assertive and tests want to pass
strings; would need two public entry points (`Lines`/`LinesBytes`) sharing one internal
implementation parameterized on the equality check

