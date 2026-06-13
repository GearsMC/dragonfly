package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// stonecutterInputSlot is the slot index of the input item in the stonecutter.
const stonecutterInputSlot = 0x03

// handleStonecutting handles a CraftRecipe stack request action made using a stonecutter.
func (h *ItemStackRequestHandler) handleStonecutting(a *protocol.CraftRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return errors.New(i18n.R("%df.session.handler.stonecutter.recipe_not_found", a.RecipeNetworkID))
	}
	if _, shapeless := craft.(recipe.Shapeless); !shapeless {
		return errors.New(i18n.R("%df.session.handler.stonecutter.not_shapeless", a.RecipeNetworkID))
	}
	if craft.Block() != "stonecutter" {
		return errors.New(i18n.R("%df.session.handler.stonecutter.not_stonecutter", a.RecipeNetworkID))
	}

	timesCrafted := int(a.NumberOfCrafts)
	if timesCrafted < 1 {
		return errors.New(i18n.R("%df.session.handler.stonecutter.times_crafted_min"))
	}

	expectedInputs := craft.Input()
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerStonecutterInput},
		Slot:      stonecutterInputSlot,
	}, s, tx)
	if input.Count() < timesCrafted {
		return errors.New(i18n.R("%df.session.handler.stonecutter.input_count_low"))
	}
	if !matchingStacks(input, expectedInputs[0]) {
		return errors.New(i18n.R("%df.session.handler.stonecutter.input_mismatch"))
	}

	output := craft.Output()
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerStonecutterInput},
		Slot:      stonecutterInputSlot,
	}, input.Grow(-timesCrafted), s, tx)
	return h.createResults(s, tx, repeatStacks(output, timesCrafted)...)
}
