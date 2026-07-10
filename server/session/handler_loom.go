package session

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// loomInputSlot is the slot index of the input item in the loom table.
	loomInputSlot = 0x09
	// loomDyeSlot is the slot index of the dye item in the loom table.
	loomDyeSlot = 0x0a
	// loomPatternSlot is the slot index of the pattern item in the loom table.
	loomPatternSlot = 0x0b
)

// handleLoomCraft handles a CraftLoomRecipe stack request action made using a loom table.
func (h *ItemStackRequestHandler) handleLoomCraft(a *protocol.CraftLoomRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	// First check if there actually is a loom opened.
	if _, ok := tx.Block(*s.openedPos.Load()).(block.Loom); !ok || !s.containerOpened.Load() {
		return errors.New(i18n.R("%df.session.handler.loom.no_container"))
	}
	timesCrafted := int(a.TimesCrafted)
	if timesCrafted < 1 {
		return errors.New(i18n.R("%df.session.handler.loom.times_crafted_min"))
	}

	// Next, check if the input slot has a valid banner item.
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomInput},
		Slot:      loomInputSlot,
	}, s, tx)
	if input.Count() < timesCrafted {
		return errors.New(i18n.R("%df.session.handler.loom.input_count_low"))
	}
	b, ok := input.Item().(block.Banner)
	if !ok {
		return errors.New(i18n.R("%df.session.handler.loom.not_banner"))
	}
	if b.Illager {
		return errors.New(i18n.R("%df.session.handler.loom.illager_banner"))
	}

	// Do the same with the input dye.
	dye, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomDye},
		Slot:      loomDyeSlot,
	}, s, tx)
	if dye.Count() < timesCrafted {
		return errors.New(i18n.R("%df.session.handler.loom.dye_count_low"))
	}
	d, ok := dye.Item().(item.Dye)
	if !ok {
		return errors.New(i18n.R("%df.session.handler.loom.not_dye"))
	}

	// The action contains the pattern that the client wanted to apply, so parse the ID and check if it is a valid
	// pattern.
	expectedPattern, exists := block.BannerPatternByID(a.Pattern)
	if !exists {
		return fmt.Errorf("unknown banner pattern id %q", a.Pattern)
	}

	// Some banner patterns have equivalent banner pattern items that are required to craft the pattern. If the expected
	// pattern has a pattern item, check if the player input the correct pattern item.
	pattern, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomMaterial},
		Slot:      loomPatternSlot,
	}, s, tx)
	if expectedPatternItem, hasPatternItem := expectedPattern.Item(); hasPatternItem {
		if pattern.Empty() {
			return errors.New(i18n.R("%df.session.handler.loom.pattern_empty"))
		}
		p, ok := pattern.Item().(item.BannerPattern)
		if !ok {
			return errors.New(i18n.R("%df.session.handler.loom.not_pattern"))
		}
		if expectedPatternItem != p.Type {
			return errors.New(i18n.R("%df.session.handler.loom.pattern_mismatch"))
		}
	}

	// Add a new pattern layer onto the banner, and create the result.
	b.Patterns = append(b.Patterns, block.BannerPatternLayer{
		Type:   expectedPattern,
		Colour: d.Colour,
	})
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomInput},
		Slot:      loomInputSlot,
	}, input.Grow(-timesCrafted), s, tx)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerLoomDye},
		Slot:      loomDyeSlot,
	}, dye.Grow(-timesCrafted), s, tx)
	return h.createResults(s, tx, input.WithItem(b))
}
