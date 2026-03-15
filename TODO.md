# TODO

* idea: reduce allocations by working with `[][]byte` instead of `[]string` and `mmap` files instead
of `os.ReadFile`. caveat: complicates API design since callers like assertive and tests want to pass
strings; would need two public entry points (`Lines`/`LinesBytes`) sharing one internal
implementation parameterized on the equality check

