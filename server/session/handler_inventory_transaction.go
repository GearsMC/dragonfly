package session

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// InventoryTransactionHandler handles the InventoryTransaction packet.
type InventoryTransactionHandler struct{}

// Handle ...
func (h *InventoryTransactionHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) (err error) {
	pk := p.(*packet.InventoryTransaction)

	if len(pk.LegacySetItemSlots) > 2 {
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.too_many_sync"))
	}

	defer func() {
		// The client has requested the server to resend the specified slots even if they haven't changed server-side.
		// Handling these requests is necessary to ensure the client's inventory remains in sync with the server.
		for _, slot := range pk.LegacySetItemSlots {
			if len(slot.Slots) > 2 {
				err = errors.New(i18n.R("%df.session.handler.inventory_transaction.too_many_slots"))
				return
			}
			switch slot.ContainerID {
			case protocol.ContainerOffhand:
				s.sendInv(s.offHand, protocol.WindowIDOffHand)
			case protocol.ContainerInventory:
				for _, slot := range slot.Slots {
					if i, err := s.inv.Item(int(slot)); err == nil {
						s.sendItem(i, int(slot), protocol.WindowIDInventory)
					}
				}
			}
		}
	}()

	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		h.resendInventories(s)
		// Always resend inventories with normal transactions. Most of the time we do not use these
		// transactions, so we're best off making sure the client and server stay in sync.
		if err := h.handleNormalTransaction(pk, s, c); err != nil {
			s.conf.Log.Debug(i18n.R("%df.session.handler.inventory_transaction.verify_normal", err))
		}
		return
	case *protocol.MismatchTransactionData:
		// Just resend the inventory and don't do anything.
		h.resendInventories(s)
		return
	case *protocol.UseItemOnEntityTransactionData:
		if err = s.VerifyAndSetHeldSlot(int(data.HotBarSlot), stackToItem(s.br, data.HeldItem.Stack), c); err != nil {
			return
		}
		return h.handleUseItemOnEntityTransaction(data, s, tx, c)
	case *protocol.UseItemTransactionData:
		if err = s.VerifyAndSetHeldSlot(int(data.HotBarSlot), stackToItem(s.br, data.HeldItem.Stack), c); err != nil {
			return
		}
		return h.handleUseItemTransaction(data, s, c)
	case *protocol.ReleaseItemTransactionData:
		if err = s.VerifyAndSetHeldSlot(int(data.HotBarSlot), stackToItem(s.br, data.HeldItem.Stack), c); err != nil {
			return
		}
		return h.handleReleaseItemTransaction(c)
	}
	return fmt.Errorf("%s", i18n.R("%df.session.handler.inventory_transaction.unhandled_type", fmt.Sprintf("%T", pk.TransactionData)))
}

// resendInventories resends all inventories of the player.
func (h *InventoryTransactionHandler) resendInventories(s *Session) {
	s.sendInv(s.inv, protocol.WindowIDInventory)
	s.sendInv(s.ui, protocol.WindowIDUI)
	s.sendInv(s.offHand, protocol.WindowIDOffHand)
	s.sendInv(s.armour.Inventory(), protocol.WindowIDArmour)
}

// handleNormalTransaction ...
func (h *InventoryTransactionHandler) handleNormalTransaction(pk *packet.InventoryTransaction, s *Session, c Controllable) error {
	if len(pk.Actions) != 2 {
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.drop_two_actions", len(pk.Actions)))
	}

	var (
		slot     int
		count    int
		expected item.Stack
	)
	for _, action := range pk.Actions {
		switch {
		case action.SourceType == protocol.InventoryActionSourceWorld && action.InventorySlot == 0:
			if old := stackToItem(s.br, action.OldItem.Stack); !old.Empty() {
				return fmt.Errorf("%s", i18n.R("%df.session.handler.inventory_transaction.non_empty_old", fmt.Sprintf("%#v", action.OldItem)))
			}
			count = int(action.NewItem.Stack.Count)
		case action.SourceType == protocol.InventoryActionSourceContainer && action.WindowID == protocol.WindowIDInventory:
			if expected = stackToItem(s.br, action.OldItem.Stack); expected.Empty() {
				return fmt.Errorf("%s", i18n.R("%df.session.handler.inventory_transaction.empty_old", fmt.Sprintf("%#v", action.OldItem)))
			}
			slot = int(action.InventorySlot)
		default:
			return errors.New(i18n.R("%df.session.handler.inventory_transaction.unexpected_action"))
		}
	}

	actual, _ := s.inv.Item(slot)
	if count < 1 {
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.no_drop", count))
	}
	if count > actual.Count() {
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.too_many_throw", count, actual.Count()))
	}
	if !expected.Equal(actual) {
		return fmt.Errorf("%s", i18n.R("%df.session.handler.inventory_transaction.different_thrown", fmt.Sprintf("%#v", expected), fmt.Sprintf("%#v", actual)))
	}

	// Explicitly don't re-use the thrown variable. This item was supplied by the user, and if some
	// logic in the Comparable() method was flawed, users would be able to cheat with item properties.
	// Only grow or shrink the held item to prevent any such issues.
	res := actual.Grow(count - actual.Count())
	if err := call(event.C(inventory.Holder(c)), int(*s.heldSlot), res, s.inv.Handler().HandleDrop); err != nil {
		return err
	}

	n := c.Drop(res)
	_ = s.inv.SetItem(slot, actual.Grow(-n))
	return nil
}

// handleUseItemOnEntityTransaction ...
func (h *InventoryTransactionHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData, s *Session, tx *world.Tx, c Controllable) error {
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	if data.TargetEntityRuntimeID == selfEntityRuntimeID {
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.interact_self"))
	}

	handle, ok := s.entityFromRuntimeID(data.TargetEntityRuntimeID)
	if !ok {
		// In some cases, for example when a falling block entity solidifies, latency may allow attacking an entity that
		// no longer exists server side. This is expected, so we shouldn't kick the player.
		s.conf.Log.Debug(i18n.R("%df.session.handler.inventory_transaction.no_entity"), "ID", data.TargetEntityRuntimeID)
		return nil
	}
	e, ok := handle.Entity(tx)
	if !ok {
		s.conf.Log.Debug(i18n.R("%df.session.handler.inventory_transaction.different_world"), "ID", data.TargetEntityRuntimeID)
		return nil
	}
	var valid bool
	switch data.ActionType {
	case protocol.UseItemOnEntityActionInteract:
		valid = c.UseItemOnEntity(e)
	case protocol.UseItemOnEntityActionAttack:
		valid = c.AttackEntity(e)
	default:
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.unhandled_use_item_on_entity", data.ActionType))
	}
	if !valid {
		slot := int(*s.heldSlot)
		it, _ := s.inv.Item(slot)
		s.sendItem(it, slot, protocol.WindowIDInventory)
	}
	return nil
}

// handleUseItemTransaction ...
func (h *InventoryTransactionHandler) handleUseItemTransaction(data *protocol.UseItemTransactionData, s *Session, c Controllable) error {
	pos := cube.Pos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	// We reset the inventory so that we can send the held item update without the client already
	// having done that client-side.
	// Because of the new inventory system, the client will expect a transaction confirmation, but instead of doing that
	// it's much easier to just resend the inventory.
	h.resendInventories(s)

	switch data.ActionType {
	case protocol.UseItemActionBreakBlock:
		c.BreakBlock(pos)
	case protocol.UseItemActionClickBlock:
		c.UseItemOnBlock(pos, cube.Face(data.BlockFace), vec32To64(data.ClickedPosition))
	case protocol.UseItemActionClickAir:
		c.UseItem()
	default:
		return errors.New(i18n.R("%df.session.handler.inventory_transaction.unhandled_use_item", data.ActionType))
	}
	return nil
}

// handleReleaseItemTransaction ...
func (h *InventoryTransactionHandler) handleReleaseItemTransaction(c Controllable) error {
	c.ReleaseItem()
	return nil
}
