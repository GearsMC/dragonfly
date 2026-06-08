package server

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// validateAuthenticatedIdentity kimlik verisinin Xbox Live tarafından
// doğrulanmış bir oyuncuya ait olduğunu denetler. Özel listener'lardan gelenler
// dahil olmak üzere Server tarafından kabul edilen her bağlantı bu koşulu sağlamalıdır.
func validateAuthenticatedIdentity(identity login.IdentityData) error {
	if identity.XUID == "" {
		return fmt.Errorf("doğrulanmış Xbox Live kimliğinin XUID değeri olmalıdır")
	}
	if err := identity.Validate(); err != nil {
		return fmt.Errorf("geçersiz Xbox Live kimliği: %w", err)
	}
	return nil
}
