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
diff.Write(os.Stdout, edits, diff.WithGutter)
```

Given two versions of a function:

```go
// old                              // new
func fizzbuzz(n int) string {       func fizzbuzz(n int) string {
    if n%15 == 0 {                      switch {
                                        case n%15 == 0:
        return "FizzBuzz"                   return "FizzBuzz"
    } else if n%3 == 0 {                case n%3 == 0:
        return "Fizz"                       return "Fizz"
    } else if n%5 == 0 {                case n%5 == 0:
        return "Buzz"                       return "Buzz"
    }                                   default:
    return strconv.Itoa(n)                  return strconv.Itoa(n)
}                                       }
                                    }
```

Unified:

```
@@ -1,10 +1,12 @@
 func fizzbuzz(n int) string {
-	if n%15 == 0 {
+	switch {
+	case n%15 == 0:
 		return "FizzBuzz"
-	} else if n%3 == 0 {
+	case n%3 == 0:
 		return "Fizz"
-	} else if n%5 == 0 {
+	case n%5 == 0:
 		return "Buzz"
+	default:
+		return strconv.Itoa(n)
 	}
-	return strconv.Itoa(n)
 }
```

Gutter:

```
 1   │ func fizzbuzz(n int) string {
 2 - │ →if·n%15·==·0·{
   + │ →switch·{
   + │ →case·n%15·==·0:
 3   │ 		return "FizzBuzz"
 4 - │ →}·else·if·n%3·==·0·{
   + │ →case·n%3·==·0:
 5   │ 		return "Fizz"
 6 - │ →}·else·if·n%5·==·0·{
   + │ →case·n%5·==·0:
 7   │ 		return "Buzz"
   + │ →default:
   + │ →→return·strconv.Itoa(n)
 8   │ 	}
 9 - │ →return·strconv.Itoa(n)
10   │ }
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
