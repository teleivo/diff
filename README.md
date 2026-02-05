# diff

A Myers diff algorithm implementation in Go. Includes `gdiff`, a minimal command-line diff tool.

## Install

```sh
go install github.com/teleivo/diff/cmd/gdiff@latest
```

## Library

```go
import "github.com/teleivo/diff"

// Compute the shortest edit script between two sequences
edits := diff.Lines(
	[]string{"a", "b", "c"},
	[]string{"a", "x", "c"},
)

// Write the edit script in unified diff format
err := diff.WriteUnified(os.Stdout, edits, 3)
```

## CLI

```sh
gdiff file1.txt file2.txt
```

Exit codes: 0 (identical), 1 (differences found), 2 (error)

## Acknowledgments

This implementation is based on James Coglan's great blog series ["The Myers Diff Algorithm"](https://blog.jcoglan.com/2017/02/12/the-myers-diff-algorithm-part-1/)
which walks through Eugene Myers' 1986 paper ["An O(ND) Difference Algorithm and Its Variations"](http://www.xmailserver.org/diff2.pdf).

## Disclaimer

I wrote this library for my personal projects and it is provided as-is without warranty. It is
tailored to my needs and my intention is not to adjust it to someone else's liking. Feel free to use
it!

See [LICENSE](LICENSE) for full license terms.
