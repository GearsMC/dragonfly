package session

import (
	"errors"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// BookEditHandler handles the BookEdit packet.
type BookEditHandler struct{}

// Handle ...
func (b BookEditHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, _ Controllable) error {
	pk := p.(*packet.BookEdit)

	it, err := s.inv.Item(int(pk.InventorySlot))
	if err != nil {
		return errors.New(i18n.R("%df.session.handler.book_edit.invalid_slot", pk.InventorySlot))
	}
	book, ok := it.Item().(item.BookAndQuill)
	if !ok {
		return errors.New(i18n.R("%df.session.handler.book_edit.not_writable_book", pk.InventorySlot))
	}

	page := int(pk.PageNumber)
	if page >= 50 || page < 0 {
		return errors.New(i18n.R("%df.session.handler.book_edit.page_out_of_bounds", pk.PageNumber))
	}
	if len(pk.Text) > 256 {
		return errors.New(i18n.R("%df.session.handler.book_edit.text_too_long"))
	}

	slot := int(pk.InventorySlot)
	switch pk.ActionType {
	case packet.BookActionReplacePage:
		book = book.SetPage(page, pk.Text)
	case packet.BookActionAddPage:
		if len(book.Pages) >= 50 {
			return errors.New(i18n.R("%df.session.handler.book_edit.page_limit"))
		}
		if page >= len(book.Pages) && page <= len(book.Pages)+2 {
			book = book.SetPage(page, "")
			break
		}
		if _, ok := book.Page(page); !ok {
			return errors.New(i18n.R("%df.session.handler.book_edit.insert_page", pk.PageNumber))
		}
		book = book.InsertPage(page, pk.Text)
	case packet.BookActionDeletePage:
		if _, ok := book.Page(page); !ok {
			// We break here instead of returning an error because the client can be a page or two ahead in the UI then
			// the actual pages representation server side. The client still sends the deletion indexes.
			break
		}
		book = book.DeletePage(page)
	case packet.BookActionSwapPages:
		if pk.SecondaryPageNumber >= 50 {
			return errors.New(i18n.R("%df.session.handler.book_edit.swap_out_of_bounds"))
		}
		_, ok := book.Page(page)
		_, ok2 := book.Page(int(pk.SecondaryPageNumber))
		if !ok || !ok2 {
			// We break here instead of returning an error because the client can try to swap pages that don't exist.
			// This happens as a result of the client being a page or two ahead in the UI then the actual pages
			// representation server side. The client still sends the swap indexes.
			break
		}
		book = book.SwapPages(page, int(pk.SecondaryPageNumber))
	case packet.BookActionSign:
		_ = s.inv.SetItem(slot, it.WithItem(item.WrittenBook{Title: pk.Title, Author: pk.Author, Pages: book.Pages, Generation: item.OriginalGeneration()}))
		return nil
	}
	_ = s.inv.SetItem(slot, it.WithItem(book))
	return nil
}
