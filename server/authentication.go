package server

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// validateAuthenticatedIdentity kimlik verisinin Xbox Live tarafından
// doğrulanmış bir oyuncuya ait olduğunu denetler. Özel listener'lardan gelenler
// dahil olmak üzere Server tarafından kabul edilen her bağlantı bu koşulu sağlamalıdır.
func validateAuthenticatedIdentity(identity login.IdentityData) error {
	if identity.XUID == "" {
		return fmt.Errorf("%s", i18n.R("%df.auth.error.empty_xuid"))
	}
	if err := identity.Validate(); err != nil {
		return fmt.Errorf("%s: %w", i18n.R("%df.auth.error.invalid_identity"), err)
	}
	return nil
}
