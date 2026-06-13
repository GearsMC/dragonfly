package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	// smithingInputSlot is the slot index of the input item in the smithing table.
	smithingInputSlot = 0x33
	// smithingMaterialSlot is the slot index of the material in the smithing table.
	smithingMaterialSlot = 0x34
	// smithingTemplateSlot is the slot index of the template item in the smithing table.
	smithingTemplateSlot = 0x35
)

// handleSmithing handles a CraftRecipe stack request action made using a smithing table.
func (h *ItemStackRequestHandler) handleSmithing(a *protocol.CraftRecipeStackRequestAction, s *Session, tx *world.Tx) error {
	// First, check the recipe and ensure it is valid for the smithing table.
	craft, ok := s.recipes[a.RecipeNetworkID]
	if !ok {
		return errors.New(i18n.R("%df.session.handler.smithing.recipe_not_found", a.RecipeNetworkID))
	}
	if craft.Block() != "smithing_table" {
		return errors.New(i18n.R("%df.session.handler.smithing.not_smithing_table", a.RecipeNetworkID))
	}
	switch craft.(type) {
	case recipe.SmithingTransform, recipe.SmithingTrim:
	default:
		return errors.New(i18n.R("%df.session.handler.smithing.not_smithing", a.RecipeNetworkID))
	}

	// Check if the input item and material item match what the recipe requires.
	expectedInputs := craft.Input()
	input, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableInput},
		Slot:      smithingInputSlot,
	}, s, tx)
	if !matchingStacks(input, expectedInputs[0]) {
		return errors.New(i18n.R("%df.session.handler.smithing.input_mismatch"))
	}
	material, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableMaterial},
		Slot:      smithingMaterialSlot,
	}, s, tx)
	if !matchingStacks(material, expectedInputs[1]) {
		return errors.New(i18n.R("%df.session.handler.smithing.material_mismatch"))
	}
	template, _ := h.itemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableTemplate},
		Slot:      smithingTemplateSlot,
	}, s, tx)
	if !matchingStacks(template, expectedInputs[2]) {
		return errors.New(i18n.R("%df.session.handler.smithing.template_mismatch"))
	}

	// Create the output using the input stack as reference and the recipe's output item type.
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableInput},
		Slot:      smithingInputSlot,
	}, input.Grow(-1), s, tx)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableMaterial},
		Slot:      smithingMaterialSlot,
	}, material.Grow(-1), s, tx)
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerSmithingTableTemplate},
		Slot:      smithingTemplateSlot,
	}, template.Grow(-1), s, tx)

	if _, ok = craft.(recipe.SmithingTrim); ok {
		var trim item.ArmourTrim
		if t, ok := template.Item().(item.SmithingTemplate); ok {
			trim.Template = t.Template
		} else {
			return errors.New(i18n.R("%df.session.handler.smithing.not_template"))
		}
		if trim.Material, ok = material.Item().(item.ArmourTrimMaterial); !ok {
			return errors.New(i18n.R("%df.session.handler.smithing.not_trim_material"))
		}
		trimmable, ok := input.Item().(item.Trimmable)
		if !ok {
			return errors.New(i18n.R("%df.session.handler.smithing.not_trimmable"))
		}
		return h.createResults(s, tx, input.WithItem(trimmable.WithTrim(trim)))
	}
	return h.createResults(s, tx, input.WithItem(craft.Output()[0].Item()))
}
