package playerdb

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
)

func TestProviderIdentityIndexResolvesNameAndXUID(t *testing.T) {
	provider, err := NewProvider(t.TempDir())
	if err != nil {
		t.Fatalf("provider açılamadı: %v", err)
	}
	t.Cleanup(func() {
		_ = provider.Close()
	})

	firstUUID := uuid.New()
	identity := player.Identity{
		XUID:          "2535457450295374",
		UUID:          firstUUID,
		LastKnownName: "Lexa5936",
		LastSeen:      time.Unix(100, 0).UTC(),
	}
	if err := provider.RememberIdentity(identity); err != nil {
		t.Fatalf("kimlik indeksi yazılamadı: %v", err)
	}

	byName, ok, err := provider.LookupIdentityByName("lexa5936")
	if err != nil {
		t.Fatalf("isimden kimlik çözülemedi: %v", err)
	}
	if !ok {
		t.Fatal("isimden kimlik bulunamadı")
	}
	if byName.XUID != identity.XUID || byName.UUID != firstUUID || byName.LastKnownName != "Lexa5936" {
		t.Fatalf("beklenmeyen isim çözümü: %+v", byName)
	}

	byXUID, ok, err := provider.LookupIdentityByXUID(identity.XUID)
	if err != nil {
		t.Fatalf("XUID'den kimlik çözülemedi: %v", err)
	}
	if !ok {
		t.Fatal("XUID'den kimlik bulunamadı")
	}
	if byXUID.XUID != identity.XUID || byXUID.UUID != firstUUID || byXUID.LastKnownName != "Lexa5936" {
		t.Fatalf("beklenmeyen XUID çözümü: %+v", byXUID)
	}

	secondUUID := uuid.New()
	identity.UUID = secondUUID
	identity.LastKnownName = "LexaNew"
	identity.LastSeen = time.Unix(200, 0).UTC()
	if err := provider.RememberIdentity(identity); err != nil {
		t.Fatalf("güncel kimlik indeksi yazılamadı: %v", err)
	}

	if _, ok, err := provider.LookupIdentityByName("lexa5936"); err != nil {
		t.Fatalf("eski isim sorgusu hata verdi: %v", err)
	} else if ok {
		t.Fatal("eski isim yeni XUID aramasında geçerli kalmamalıydı")
	}

	byName, ok, err = provider.LookupIdentityByName("lexanew")
	if err != nil {
		t.Fatalf("yeni isimden kimlik çözülemedi: %v", err)
	}
	if !ok {
		t.Fatal("yeni isimden kimlik bulunamadı")
	}
	if byName.XUID != identity.XUID || byName.UUID != secondUUID || byName.LastKnownName != "LexaNew" {
		t.Fatalf("güncel kimlik yanlış çözüldü: %+v", byName)
	}
}

func TestProviderIdentityIndexDoesNotDeleteReusedOldName(t *testing.T) {
	provider, err := NewProvider(t.TempDir())
	if err != nil {
		t.Fatalf("provider açılamadı: %v", err)
	}
	t.Cleanup(func() {
		_ = provider.Close()
	})

	first := player.Identity{XUID: "1", UUID: uuid.New(), LastKnownName: "Lexa"}
	second := player.Identity{XUID: "2", UUID: uuid.New(), LastKnownName: "Lexa"}
	if err := provider.RememberIdentity(first); err != nil {
		t.Fatalf("ilk kimlik yazılamadı: %v", err)
	}
	if err := provider.RememberIdentity(second); err != nil {
		t.Fatalf("ikinci kimlik yazılamadı: %v", err)
	}

	first.LastKnownName = "LexaNew"
	if err := provider.RememberIdentity(first); err != nil {
		t.Fatalf("ilk kimlik güncellenemedi: %v", err)
	}

	identity, ok, err := provider.LookupIdentityByName("lexa")
	if err != nil {
		t.Fatalf("yeniden kullanılan isim çözülemedi: %v", err)
	}
	if !ok {
		t.Fatal("yeniden kullanılan isim mapping'i silinmemeliydi")
	}
	if identity.XUID != second.XUID {
		t.Fatalf("yeniden kullanılan isim yanlış XUID'ye çözüldü: %+v", identity)
	}
}
