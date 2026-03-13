package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

// options controls animation behavior.
type options struct {
	dpi   int
	delay time.Duration
	step  bool
}

// errFlagParse is a sentinel error indicating flag parsing failed.
// The flag package already printed the error, so main should not print again.
var errFlagParse = errors.New("flag parse error")

func main() {
	code, err := run(os.Args, os.Stdout, os.Stdin, os.Stderr)
	if err != nil && err != errFlagParse {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(code)
}

func run(args []string, w io.Writer, in io.Reader, wErr io.Writer) (int, error) {
	flags := flag.NewFlagSet("animate-myers", flag.ContinueOnError)
	flags.SetOutput(wErr)
	speed := flags.Float64("speed", 1.0, "animation speed multiplier")
	dpi := flags.Int("dpi", 150, "image resolution")
	step := flags.Bool("step", false, "step mode: press any key to advance each frame")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(wErr, "animate-myers animates the Myers diff edit graph")
		_, _ = fmt.Fprintln(wErr, "")
		_, _ = fmt.Fprintln(wErr, "usage: animate-myers [-speed N] [-dpi N] [-step]")
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

	_, _ = fmt.Fprint(wErr, hideCursor)
	defer fmt.Fprint(wErr, showCursor) //nolint:errcheck

	// Classic Myers paper example sequences
	oldLines := []string{"A", "B", "C", "A", "B", "B", "A"}
	newLines := []string{"C", "B", "A", "B", "A", "C"}
	anim := newAnimation(oldLines, newLines)
	opts := &options{
		dpi:   *dpi,
		delay: time.Duration(float64(time.Second) / *speed),
		step:  *step,
	}
	if err := animate(w, in, anim, opts); err != nil {
		return 2, err
	}

	_, _ = fmt.Fprintln(wErr, "\nPress any key to exit...")
	_, _ = in.Read(make([]byte, 1))
	return 0, nil
}

// animate runs the animation and renders each frame.
func animate(w io.Writer, in io.Reader, anim *animation, opts *options) error {
	for {
		dot := anim.next()
		if dot == "" {
			break
		}
		if err := renderFrame(w, dot, opts.dpi); err != nil {
			return err
		}
		if opts.step {
			_, _ = in.Read(make([]byte, 1))
		} else {
			time.Sleep(opts.delay)
		}
	}
	if result := anim.summary(); result != "" {
		if _, err := fmt.Fprintln(w, result); err != nil {
			return err
		}
	}
	return nil
}
