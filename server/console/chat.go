package console

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ChatSubscriber forwards global chat messages to the terminal writer.
type ChatSubscriber struct {
	writer *Writer
	id     uuid.UUID
}

// NewChatSubscriber creates a terminal chat subscriber.
func NewChatSubscriber(writer *Writer) ChatSubscriber {
	return ChatSubscriber{writer: writer, id: uuid.New()}
}

// UUID returns the stable subscriber UUID.
func (s ChatSubscriber) UUID() uuid.UUID {
	return s.id
}

// Message writes a chat message to the terminal.
func (s ChatSubscriber) Message(a ...any) {
	if s.writer == nil {
		return
	}
	parts := make([]string, len(a))
	for i, part := range a {
		parts[i] = fmt.Sprint(part)
	}
	s.writer.Line(strings.Join(parts, " "))
}
