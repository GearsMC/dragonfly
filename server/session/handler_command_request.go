package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// CommandRequestHandler handles the CommandRequest packet.
type CommandRequestHandler struct {
	origin protocol.CommandOrigin
}

// Handle ...
func (h *CommandRequestHandler) Handle(p packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.CommandRequest)
	if pk.Internal {
		return errors.New(i18n.R("%df.session.handler.command_request.internal_set"))
	}

	h.origin = pk.CommandOrigin
	c.ExecuteCommand(pk.CommandLine)
	return nil
}
