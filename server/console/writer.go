package console

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Writer serialises writes to the terminal and converts Minecraft formatting
// codes to ANSI escape sequences when the terminal supports colours.
type Writer struct {
	out    io.Writer
	colour bool
	mu     sync.Mutex
}

// NewWriter creates a terminal writer using out. If out is nil, os.Stdout is
// used.
func NewWriter(out io.Writer, colour bool) *Writer {
	if out == nil {
		out = os.Stdout
	}
	return &Writer{out: out, colour: colour}
}

// SupportsColour checks if the file passed is likely to support ANSI colours.
func SupportsColour(file *os.File) bool {
	if file == nil || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "" {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// Line writes one formatted line to the terminal.
func (w *Writer) Line(line string) {
	if w == nil {
		return
	}
	if w.colour {
		line = text.ANSI(line)
	} else {
		line = text.Clean(line)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = fmt.Fprintln(w.out, line)
}
