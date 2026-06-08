package console

import (
	"bufio"
	"io"
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Runner reads terminal commands and dispatches them on the server world.
type Runner struct {
	in     io.Reader
	writer *Writer
	source cmd.Source
	w      *world.World
	done   chan struct{}
}

// Start starts a terminal command runner.
func Start(in io.Reader, writer *Writer, source cmd.Source, w *world.World) *Runner {
	r := &Runner{
		in:     in,
		writer: writer,
		source: source,
		w:      w,
		done:   make(chan struct{}),
	}
	go r.run()
	return r
}

// Done returns a channel closed when the runner stops reading commands.
func (r *Runner) Done() <-chan struct{} {
	return r.done
}

func (r *Runner) run() {
	defer close(r.done)
	if r.in == nil || r.source == nil || r.w == nil {
		return
	}

	scanner := bufio.NewScanner(r.in)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		<-r.w.Exec(func(tx *world.Tx) {
			cmd.Dispatch(line, r.source, tx, nil)
		})
	}
	if err := scanner.Err(); err != nil && r.writer != nil {
		r.writer.Line(text.Red + "Console reader stopped: " + text.White + err.Error() + text.Reset)
	}
}
