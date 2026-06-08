package console

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/text"
)

// LogHandler is a slog.Handler that formats server logs for the terminal.
type LogHandler struct {
	writer *Writer
	level  slog.Leveler
	attrs  []slog.Attr
	groups []string
}

// NewLogHandler creates a coloured terminal slog handler.
func NewLogHandler(writer *Writer, level slog.Leveler) *LogHandler {
	if level == nil {
		level = slog.LevelInfo
	}
	return &LogHandler{writer: writer, level: level}
}

// Enabled checks if the record level should be logged.
func (h *LogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle writes a log record to the terminal.
func (h *LogHandler) Handle(_ context.Context, record slog.Record) error {
	if h.writer == nil {
		return nil
	}

	var b strings.Builder
	t := record.Time
	if t.IsZero() {
		t = time.Now()
	}

	level, colour := levelLabel(record.Level)
	b.WriteString(text.Blue)
	b.WriteString(t.Format("15:04:05"))
	b.WriteString(" ")
	b.WriteString(colour)
	b.WriteString(level)
	b.WriteString(text.White)
	b.WriteString(" ")
	b.WriteString(record.Message)

	for _, attr := range h.attrs {
		appendAttr(&b, h.groups, attr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		appendAttr(&b, h.groups, attr)
		return true
	})
	b.WriteString(text.Reset)

	h.writer.Line(b.String())
	return nil
}

// WithAttrs returns a handler with the attributes passed attached.
func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *h
	cp.attrs = append(slicesClone(h.attrs), attrs...)
	return &cp
}

// WithGroup returns a handler with a group prefix attached to future attrs.
func (h *LogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	cp := *h
	cp.groups = append(slicesClone(h.groups), name)
	return &cp
}

func levelLabel(level slog.Level) (string, string) {
	switch {
	case level >= slog.LevelError:
		return "ERROR", text.Red
	case level >= slog.LevelWarn:
		return "WARN", text.Yellow
	case level <= slog.LevelDebug:
		return "DEBUG", text.Grey
	default:
		return "INFO", text.Orange
	}
}

func appendAttr(b *strings.Builder, groups []string, attr slog.Attr) {
	if attr.Key == "" {
		return
	}
	value := attr.Value.Resolve()

	b.WriteString(" ")
	b.WriteString(text.DarkGrey)
	if len(groups) > 0 {
		b.WriteString(strings.Join(groups, "."))
		b.WriteString(".")
	}
	b.WriteString(attr.Key)
	b.WriteString("=")
	b.WriteString(text.Grey)
	b.WriteString(formatValue(value))
	b.WriteString(text.White)
}

func formatValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339)
	case slog.KindDuration:
		return value.Duration().String()
	default:
		return fmt.Sprint(value.Any())
	}
}

func slicesClone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	cp := make([]T, len(s))
	copy(cp, s)
	return cp
}
