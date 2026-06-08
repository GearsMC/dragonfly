package console

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Source is the command source used by the server terminal.
type Source struct {
	writer   *Writer
	position mgl64.Vec3
}

// NewSource creates a new console command source.
func NewSource(writer *Writer, position mgl64.Vec3) *Source {
	return &Source{writer: writer, position: position}
}

// Console marks this Source as the server console.
func (*Source) Console() bool {
	return true
}

// Position returns the command execution position of the console.
func (s *Source) Position() mgl64.Vec3 {
	return s.position
}

// HasCommandPermission grants every command permission to the server console.
func (*Source) HasCommandPermission(string) bool {
	return true
}

// SendCommandOutput writes command output to the terminal.
func (s *Source) SendCommandOutput(output *cmd.Output) {
	if output == nil || s.writer == nil {
		return
	}
	for _, message := range output.Messages() {
		s.writeLines(text.Green+"Command output | "+text.White+message.String()+text.Reset, false)
	}
	for _, err := range output.Errors() {
		s.writeLines(text.Red+"Command error | "+text.White+err.Error()+text.Reset, true)
	}
}

func (s *Source) writeLines(message string, trim bool) {
	if trim {
		message = strings.TrimSpace(message)
	}
	for _, line := range strings.Split(message, "\n") {
		if line == "" {
			continue
		}
		s.writer.Line(line)
	}
}
