package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TextHandler handles the Text packet.
type TextHandler struct{}

// Handle ...
func (TextHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, c Controllable) error {
	pk := p.(*packet.Text)

	if pk.TextType != packet.TextTypeChat {
		return errors.New(i18n.R("%df.session.handler.text.bad_type", packet.TextTypeChat, pk.TextType))
	}
	if pk.SourceName != s.conn.IdentityData().DisplayName {
		return errors.New(i18n.R("%df.session.handler.text.source_name_mismatch"))
	}
	if pk.XUID != s.conn.IdentityData().XUID {
		return errors.New(i18n.R("%df.session.handler.text.xuid_mismatch"))
	}
	c.Chat(pk.Message)
	return nil
}
