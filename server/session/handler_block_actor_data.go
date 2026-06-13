package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
	"unicode/utf8"
)

// BlockActorDataHandler handles an incoming BlockActorData packet from the client, sent for some block entities like
// signs when they are edited.
type BlockActorDataHandler struct{}

// Handle ...
func (b BlockActorDataHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.BlockActorData)
	if id, ok := pk.NBTData["id"]; ok {
		pos := blockPosFromProtocol(pk.Position)
		if !canReach(c, pos.Vec3Middle()) {
			return fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.not_in_reach", pos))
		}
		if id == "Sign" {
			return b.handleSign(pk, pos, s, tx, c)
		}
		return fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.unhandled_id", id))
	}
	return fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.no_id_tag", pk.NBTData))
}

// handleSign handles the BlockActorData packet sent when editing a sign.
func (b BlockActorDataHandler) handleSign(pk *packet.BlockActorData, pos cube.Pos, s *Session, tx *world.Tx, co Controllable) error {
	if _, ok := tx.Block(pos).(block.Sign); !ok {
		s.conf.Log.Debug(i18n.R("%df.session.handler.block_actor_data.no_sign"), "pos", pos.String())
		return nil
	}

	frontText, err := b.textFromNBTData(pk.NBTData, true)
	if err != nil {
		return err
	}
	backText, err := b.textFromNBTData(pk.NBTData, false)
	if err != nil {
		return err
	}
	if err := co.EditSign(pos, frontText, backText); err != nil {
		return err
	}
	return nil
}

// textFromNBTData attempts to retrieve the text from the NBT data of specific sign from the BlockActorData packet.
func (b BlockActorDataHandler) textFromNBTData(data map[string]any, frontSide bool) (string, error) {
	var sideData map[string]any
	var side string
	if frontSide {
		frontSide, ok := data["FrontText"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.front_text_not_found", fmt.Sprintf("%#v", data["FrontText"])))
		}
		sideData = frontSide
		side = "front"
	} else {
		backSide, ok := data["BackText"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.back_text_not_found", fmt.Sprintf("%#v", data["BackText"])))
		}
		sideData = backSide
		side = "back"
	}
	var text string
	pkText, ok := sideData["Text"]
	if !ok {
		return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.no_text_tag", side))
	}
	if text, ok = pkText.(string); !ok {
		return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.text_not_string", side, fmt.Sprintf("%#v", pkText)))
	}

	// Verify that the text was valid. It must be valid UTF8 and not more than 100 characters long.
	text = strings.TrimRight(text, "\n")
	if len(text) > 256 {
		return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.text_too_long", side))
	}
	if !utf8.ValidString(text) {
		return "", fmt.Errorf("%s", i18n.R("%df.session.handler.block_actor_data.sign.text_not_utf8", side))
	}
	return text, nil
}

// canReach checks if a player can reach a position with its current range. The range depends on if the player
// is either survival or creative mode.
func canReach(c Controllable, pos mgl64.Vec3) bool {
	const (
		creativeRange = 14.0
		survivalRange = 8.0
	)
	if !c.GameMode().AllowsInteraction() {
		return false
	}

	eyes := entity.EyePosition(c)
	if c.GameMode().CreativeInventory() {
		return eyes.Sub(pos).Len() <= creativeRange && !c.Dead()
	}
	return eyes.Sub(pos).Len() <= survivalRange && !c.Dead()
}
