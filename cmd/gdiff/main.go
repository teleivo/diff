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
	unified := flags.Bool("u", false, "output 3 lines of unified context")
	context := flags.Int("U", 3, "output NUM lines of unified context")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(wErr, "gdiff computes the shortest edit script between two files")
		_, _ = fmt.Fprintln(wErr, "")
		_, _ = fmt.Fprintln(wErr, "usage: gdiff [-u] [-U NUM] file1 file2")
		_, _ = fmt.Fprintln(wErr, "")
		flags.PrintDefaults()
	}
	_ = unified // -u just uses default context of 3, same as -U 3

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

	oldFile := flags.Arg(0)
	newFile := flags.Arg(1)

	hasDiff, err := diff.Files(w, oldFile, newFile, *context)
	if err != nil {
		return 2, err
	}
	if hasDiff {
		return 1, nil
	}
	return 0, nil
}
