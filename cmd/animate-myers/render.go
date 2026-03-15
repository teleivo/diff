package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os/exec"
)

func renderFrame(w io.Writer, dot string, dpi int) error {
	png, err := renderDOT(dot, dpi)
	if err != nil {
		return err
	}
	return displayKitty(w, png)
}

// renderDOT converts DOT to PNG using graphviz.
// Uses neato -n to respect fixed node positions (no re-layout).
func renderDOT(dot string, dpi int) ([]byte, error) {
	cmd := exec.Command("neato", "-n", "-Tpng", fmt.Sprintf("-Gdpi=%d", dpi))
	cmd.Stdin = bytes.NewBufferString(dot)
	return cmd.Output()
}

const (
	clearScreen = "\033[2J\033[H" // clear screen and move cursor home
	apcStart    = "\033_G"        // Kitty graphics protocol APC start
	apcEnd      = "\033\\"        // APC end (ST - String Terminator)
)

// displayKitty displays PNG data using Kitty graphics protocol.
func displayKitty(w io.Writer, png []byte) error {
	b64 := base64.StdEncoding.EncodeToString(png)

	// Clear previous image and move cursor
	if _, err := fmt.Fprint(w, clearScreen); err != nil {
		return err
	}

	// Send image in chunks (Kitty protocol limit)
	const chunkSize = 4096
	for i := 0; i < len(b64); i += chunkSize {
		end := min(i+chunkSize, len(b64))
		chunk := b64[i:end]

		var more int
		if end != len(b64) {
			more = 1
		}

		if i == 0 {
			// a=T: transmit and display, f=100: PNG format, m: more chunks flag
			_, err := fmt.Fprintf(w, "%sa=T,f=100,m=%d;%s%s", apcStart, more, chunk, apcEnd)
			if err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "%sm=%d;%s%s", apcStart, more, chunk, apcEnd); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}
