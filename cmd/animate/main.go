package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Options controls animation behavior.
type Options struct {
	dpi   int
	delay time.Duration
	step  bool
}

func (o *Options) wait() {
	if o.step {
		_, _ = os.Stdin.Read(make([]byte, 1))
	} else {
		time.Sleep(o.delay)
	}
}

// Algorithm produces DOT strings frame by frame.
type Algorithm interface {
	Next() string
	Summary() string
}

// Animate runs the algorithm and renders each frame.
func Animate(algo Algorithm, opts *Options) error {
	for {
		dot := algo.Next()
		if dot == "" {
			break
		}
		if err := renderFrame(dot, opts.dpi); err != nil {
			return err
		}
		opts.wait()
	}
	result := algo.Summary()
	if result != "" {
		fmt.Println(result)
	}
	return nil
}

func main() {
	speed := flag.Float64("speed", 1.0, "animation speed multiplier")
	dpi := flag.Int("dpi", 150, "image resolution")
	step := flag.Bool("step", false, "step mode: press Enter to advance each frame")
	flag.Parse()

	// Classic Myers paper example sequences
	old := []string{"A", "B", "C", "A", "B", "B", "A"}
	new := []string{"C", "B", "A", "B", "A", "C"}

	opts := &Options{
		dpi:   *dpi,
		delay: time.Duration(float64(800*time.Millisecond) / *speed),
		step:  *step,
	}

	// Hide cursor during animation
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	if *step {
		fmt.Println("Step mode: press Enter to advance each frame")
	}

	algo := NewMyers(old, new)
	if err := Animate(algo, opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Println("\nPress Enter to exit...")
	_, _ = os.Stdin.Read(make([]byte, 1))
}
