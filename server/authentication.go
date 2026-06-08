package server

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// validateAuthenticatedIdentity verifies that identity data belongs to an
// Xbox Live authenticated player. Every connection accepted by Server must
// satisfy this invariant, including connections supplied by custom listeners.
func validateAuthenticatedIdentity(identity login.IdentityData) error {
	if identity.XUID == "" {
		return fmt.Errorf("Xbox Live authenticated identity must have an XUID")
	}
	if err := identity.Validate(); err != nil {
		return fmt.Errorf("invalid Xbox Live identity: %w", err)
	}
	return nil
}
