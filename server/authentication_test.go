package server

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestValidateAuthenticatedIdentity(t *testing.T) {
	t.Run("authenticated", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			XUID:        "2533274790395904",
			Identity:    "494dca91-7aef-3068-8049-788897890039",
			DisplayName: "lexa5936",
		})
		if err != nil {
			t.Fatalf("expected authenticated identity to be valid: %v", err)
		}
	})

	t.Run("missing XUID", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			Identity:    "494dca91-7aef-3068-8049-788897890039",
			DisplayName: "lexa5936",
		})
		if err == nil {
			t.Fatal("expected identity without XUID to be rejected")
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			XUID:        "not-a-number",
			Identity:    "not-a-uuid",
			DisplayName: "lexa5936",
		})
		if err == nil {
			t.Fatal("expected invalid identity to be rejected")
		}
	})
}
