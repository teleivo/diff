package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/teleivo/diff"
)

// errFlagParse is a sentinel error indicating flag parsing failed.
// The flag package already printed the error, so main should not print again.
var errFlagParse = errors.New("flag parse error")

func main() {
	code, err := run(os.Args, os.Stdout, os.Stderr)
	if err != nil && err != errFlagParse {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(code)
}

func run(args []string, w io.Writer, wErr io.Writer) (int, error) {
	flags := flag.NewFlagSet("gdiff", flag.ContinueOnError)
	flags.SetOutput(wErr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(wErr, "gdiff computes the shortest edit script between two files")
		_, _ = fmt.Fprintln(wErr, "")
		_, _ = fmt.Fprintln(wErr, "usage: gdiff file1 file2")
	}

	err := flags.Parse(args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			return 0, nil
		}
		return 2, errFlagParse
	}

	if flags.NArg() != 2 {
		flags.Usage()
		return 2, nil
	}

	file1 := flags.Arg(0)
	file2 := flags.Arg(1)

	edits, err := diff.Files(file1, file2)
	if err != nil {
		return 2, err
	}

	hasDiff := false
	for _, e := range edits {
		if e.Op != diff.Eq {
			hasDiff = true
			break
		}
	}

	if !hasDiff {
		return 0, nil
	}

	if err := diff.WriteUnified(w, edits, 0); err != nil {
		return 2, err
	}
	return 1, nil
}
