package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RespawnHandler handles the Respawn packet.
type RespawnHandler struct{}

// Handle ...
func (*RespawnHandler) Handle(p packet.Packet, _ *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.Respawn)
	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return errSelfRuntimeID
	}
	if pk.State != packet.RespawnStateClientReadyToSpawn {
		return errors.New(i18n.R("%df.session.handler.respawn.bad_state", packet.RespawnStateClientReadyToSpawn, pk.State))
	}
	c.Respawn()
	return nil
}
