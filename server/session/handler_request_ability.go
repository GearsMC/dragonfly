package session

import (
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestAbilityHandler handles the RequestAbility packet.
type RequestAbilityHandler struct{}

// Handle ...
func (a RequestAbilityHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.RequestAbility)
	if pk.Ability == packet.AbilityFlying {
		if !canFly(c, c.GameMode()) {
			s.conf.Log.Debug(i18n.R("%df.session.handler.request_ability.flying_flag"))
			s.SendAbilities(c)
			return nil
		}
		c.StartFlying()
	}
	return nil
}
