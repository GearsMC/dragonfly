package server

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestValidateAuthenticatedIdentity(t *testing.T) {
	t.Run("doğrulanmış kimlik", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			XUID:        "2533274790395904",
			Identity:    "494dca91-7aef-3068-8049-788897890039",
			DisplayName: "lexa5936",
		})
		if err != nil {
			t.Fatalf("doğrulanmış kimliğin geçerli olması bekleniyordu: %v", err)
		}
	})

	t.Run("eksik XUID", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			Identity:    "494dca91-7aef-3068-8049-788897890039",
			DisplayName: "lexa5936",
		})
		if err == nil {
			t.Fatal("XUID içermeyen kimliğin reddedilmesi bekleniyordu")
		}
	})

	t.Run("geçersiz kimlik", func(t *testing.T) {
		err := validateAuthenticatedIdentity(login.IdentityData{
			XUID:        "not-a-number",
			Identity:    "not-a-uuid",
			DisplayName: "lexa5936",
		})
		if err == nil {
			t.Fatal("geçersiz kimliğin reddedilmesi bekleniyordu")
		}
	})
}
