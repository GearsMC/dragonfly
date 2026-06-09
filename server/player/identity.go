package player

import (
	"time"

	"github.com/google/uuid"
)

// Identity, oyuncunun kalıcı hesap kimliğini ve son bilinen giriş bilgilerini temsil eder.
// Kalıcı referans XUID'dir; UUID ve isim yalnızca protokol, görüntüleme ve isimden XUID çözümü için tutulur.
type Identity struct {
	XUID          string
	UUID          uuid.UUID
	LastKnownName string
	LastSeen      time.Time
}

// IdentityProvider, isim/UUID gibi değişebilir giriş bilgilerini XUID'ye çözmek için kullanılan hafif indeks API'sidir.
// Bu API oyuncu verisini XUID dışındaki anahtarlarla saklamaz; sadece ters arama kayıtları tutar.
type IdentityProvider interface {
	// RememberIdentity, doğrulanmış bir oyuncunun güncel kimliğini indekse yazar.
	RememberIdentity(identity Identity) error
	// LookupIdentityByName, son bilinen oyuncu adından kimlik bilgisini çözer.
	LookupIdentityByName(name string) (Identity, bool, error)
	// LookupIdentityByXUID, XUID'den son bilinen kimlik bilgisini çözer.
	LookupIdentityByXUID(xuid string) (Identity, bool, error)
}
