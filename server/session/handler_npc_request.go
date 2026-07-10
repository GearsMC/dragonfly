package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/player/dialogue"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NPCRequestHandler handles the NPCRequest packet.
type NPCRequestHandler struct {
	dialogue        dialogue.Dialogue
	entityRuntimeID uint64
}

// Handle ...
func (h *NPCRequestHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.NPCRequest)
	if h.entityRuntimeID == 0 {
		// No dialogue is currently open for this session, so there is nothing to submit or close.
		return nil
	}
	switch pk.RequestType {
	case packet.NPCRequestActionExecuteAction:
		if err := h.dialogue.Submit(uint(pk.ActionType), c, tx); err != nil {
			return fmt.Errorf("%s: %w", i18n.R("%df.session.handler.npc_request.submit_error"), err)
		}
	case packet.NPCRequestActionExecuteClosingCommands:
		h.dialogue.Close(c, tx)
	}
	return nil
}
