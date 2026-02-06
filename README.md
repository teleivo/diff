# diff

A Myers diff algorithm implementation in Go. Includes `gdiff`, a minimal command-line diff tool.

## Install

```sh
go install github.com/teleivo/diff/cmd/gdiff@latest
```

## Library

```go
import "github.com/teleivo/diff"

edits := diff.Lines(oldLines, newLines)

// Write in unified diff format
diff.Write(os.Stdout, edits)

// Write with gutter format (line numbers, visible whitespace)
diff.Write(os.Stdout, edits, diff.WithGutter())
```

Given a DOT file before and after formatting with [dotx](https://github.com/teleivo/dot):

```dot
// old.dot                           // new.dot
digraph {                            digraph {
    3 -> 2 -> 4                          3 -> 2 -> 4
      [color="blue",len=2.6]              [color="blue",len=2.6]
        rank = same                      rank=same
                                     }
}
```

Unified (`gdiff old.dot new.dot`):

```
@@ -1,5 +2,4 @@
 digraph {
 	3 -> 2 -> 4 [color="blue",len=2.6]
-		rank = same
-
+	rank=same
 }
```

Gutter (`gdiff --gutter old.dot new.dot`):

```
1   │ digraph {
2   │ 	3 -> 2 -> 4 [color="blue",len=2.6]
3 - │ →→rank·=·same
4 - │ ↵
  + │ →rank=same
5   │ }
```

## CLI

```sh
gdiff file1.txt file2.txt
gdiff --gutter file1.txt file2.txt
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
