package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	context := flags.Int("U", 3, "output NUM lines of unified context")
	gutter := flags.Bool("gutter", false, "show line numbers and visible whitespace")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(wErr, "gdiff computes the shortest edit script between two files")
		_, _ = fmt.Fprintln(wErr, "")
		_, _ = fmt.Fprintln(wErr, "usage: gdiff [-U NUM] [-gutter] file1 file2")
		_, _ = fmt.Fprintln(wErr, "")
		flags.PrintDefaults()
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

	oldFile := flags.Arg(0)
	newFile := flags.Arg(1)

	hasDiff, err := files(w, oldFile, newFile, *context, *gutter)
	if err != nil {
		return 2, err
	}
	if hasDiff {
		return 1, nil
	}
	return 0, nil
}

func files(w io.Writer, oldFile, newFile string, context int, gutter bool) (bool, error) {
	oldStat, err := os.Stat(oldFile)
	if err != nil {
		return false, err
	}
	newStat, err := os.Stat(newFile)
	if err != nil {
		return false, err
	}

	a, err := readLines(oldFile)
	if err != nil {
		return false, err
	}
	b, err := readLines(newFile)
	if err != nil {
		return false, err
	}

	edits := diff.Lines(a, b)

	hasDiff := false
	for _, e := range edits {
		if e.Op != diff.Eq {
			hasDiff = true
			break
		}
	}
	if !hasDiff {
		return false, nil
	}

	opts := []diff.Option{diff.WithContext(context)}
	if gutter {
		opts = append(opts, diff.WithGutter())
	} else {
		if err := writeFileHeader(w, oldFile, oldStat.ModTime(), newFile, newStat.ModTime()); err != nil {
			return false, err
		}
	}
	if err := diff.Write(w, edits, opts...); err != nil {
		return false, err
	}
	return true, nil
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	// SplitAfter keeps the delimiter on each element. Files ending in "\n"
	// produce a trailing empty string that is not a real line.
	lines := strings.SplitAfter(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

func writeFileHeader(w io.Writer, oldName string, oldTime time.Time, newName string, newTime time.Time) error {
	const timeFormat = "2006-01-02 15:04:05.000000000 -0700"
	_, err := fmt.Fprintf(w, "--- %s\t%s\n+++ %s\t%s\n",
		oldName, oldTime.Format(timeFormat),
		newName, newTime.Format(timeFormat))
	return err
}
