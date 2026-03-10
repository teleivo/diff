package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// renderDOT converts DOT to PNG using graphviz.
// Uses neato -n to respect fixed node positions (no re-layout).
func renderDOT(dot string, dpi int) ([]byte, error) {
	cmd := exec.Command("neato", "-n", "-Tpng", fmt.Sprintf("-Gdpi=%d", dpi))
	cmd.Stdin = bytes.NewBufferString(dot)
	return cmd.Output()
}

// displayKitty displays PNG data using Kitty graphics protocol
func displayKitty(w io.Writer, png []byte) {
	b64 := base64.StdEncoding.EncodeToString(png)

	// Clear previous image and move cursor
	_, _ = fmt.Fprint(w, "\033[2J\033[H")

	// Send image in chunks (Kitty protocol limit)
	const chunkSize = 4096
	for i := 0; i < len(b64); i += chunkSize {
		end := min(i+chunkSize, len(b64))
		chunk := b64[i:end]

		m := 1 // more chunks coming
		if end >= len(b64) {
			m = 0 // last chunk
		}

		if i == 0 {
			_, _ = fmt.Fprintf(w, "\033_Ga=T,f=100,m=%d;%s\033\\", m, chunk)
		} else {
			_, _ = fmt.Fprintf(w, "\033_Gm=%d;%s\033\\", m, chunk)
		}
	}
	_, _ = fmt.Fprintln(w)
}

func renderFrame(dot string, dpi int) error {
	png, err := renderDOT(dot, dpi)
	if err != nil {
		return err
	}
	displayKitty(os.Stdout, png)
	return nil
}
