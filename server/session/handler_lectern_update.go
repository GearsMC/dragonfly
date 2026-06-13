package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LecternUpdateHandler handles the LecternUpdate packet, sent when a player interacts with a lectern.
type LecternUpdateHandler struct{}

// Handle ...
func (LecternUpdateHandler) Handle(p packet.Packet, _ *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.LecternUpdate)
	pos := blockPosFromProtocol(pk.Position)
	if !canReach(c, pos.Vec3Middle()) {
		return errors.New(i18n.R("%df.session.handler.lectern_update.not_in_reach", pos))
	}
	if _, ok := tx.Block(pos).(block.Lectern); !ok {
		return errors.New(i18n.R("%df.session.handler.lectern_update.not_lectern", pos))
	}
	return c.TurnLecternPage(pos, int(pk.Page))
}
