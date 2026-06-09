package playerdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
)

type identityRecord struct {
	XUID          string
	UUID          string
	LastKnownName string
	LastSeen      time.Time
}

// RememberIdentity, oyuncunun güncel ismini ve UUID bilgisini XUID merkezli ters arama indeksine yazar.
func (p *Provider) RememberIdentity(identity player.Identity) error {
	if identity.XUID == "" {
		return fmt.Errorf("xuid boş olamaz")
	}
	identity.LastKnownName = strings.TrimSpace(identity.LastKnownName)
	if identity.LastSeen.IsZero() {
		identity.LastSeen = time.Now()
	}

	previous, previousOK, err := p.LookupIdentityByXUID(identity.XUID)
	if err != nil {
		return err
	}

	record := identityRecordFrom(identity)
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	batch.Put(identityXUIDKey(identity.XUID), b)
	if identity.LastKnownName != "" {
		batch.Put(identityNameKey(identity.LastKnownName), b)
	}
	if previousOK && previous.LastKnownName != "" && !strings.EqualFold(previous.LastKnownName, identity.LastKnownName) {
		current, ok, err := p.LookupIdentityByName(previous.LastKnownName)
		if err != nil {
			return err
		}
		if ok && current.XUID == identity.XUID {
			batch.Delete(identityNameKey(previous.LastKnownName))
		}
	}
	return p.db.Write(batch, nil)
}

// LookupIdentityByName, son bilinen oyuncu adını XUID kimliğine çözer.
func (p *Provider) LookupIdentityByName(name string) (player.Identity, bool, error) {
	name = normalizeIdentityName(name)
	if name == "" {
		return player.Identity{}, false, nil
	}
	return p.lookupIdentity(identityNameKey(name))
}

// LookupIdentityByXUID, XUID için son bilinen oyuncu kimliğini döndürür.
func (p *Provider) LookupIdentityByXUID(xuid string) (player.Identity, bool, error) {
	if xuid == "" {
		return player.Identity{}, false, nil
	}
	return p.lookupIdentity(identityXUIDKey(xuid))
}

func (p *Provider) lookupIdentity(key []byte) (player.Identity, bool, error) {
	b, err := p.db.Get(key, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return player.Identity{}, false, nil
	}
	if err != nil {
		return player.Identity{}, false, err
	}

	var record identityRecord
	if err := json.Unmarshal(b, &record); err != nil {
		return player.Identity{}, false, err
	}
	identity, err := record.identity()
	if err != nil {
		return player.Identity{}, false, err
	}
	return identity, true, nil
}

func identityRecordFrom(identity player.Identity) identityRecord {
	return identityRecord{
		XUID:          identity.XUID,
		UUID:          identity.UUID.String(),
		LastKnownName: identity.LastKnownName,
		LastSeen:      identity.LastSeen,
	}
}

func (record identityRecord) identity() (player.Identity, error) {
	if record.XUID == "" {
		return player.Identity{}, fmt.Errorf("identity xuid boş olamaz")
	}
	id := uuid.Nil
	if record.UUID != "" {
		parsed, err := uuid.Parse(record.UUID)
		if err != nil {
			return player.Identity{}, err
		}
		id = parsed
	}
	return player.Identity{
		XUID:          record.XUID,
		UUID:          id,
		LastKnownName: record.LastKnownName,
		LastSeen:      record.LastSeen,
	}, nil
}

func identityXUIDKey(xuid string) []byte {
	return []byte("identity:xuid:" + xuid)
}

func identityNameKey(name string) []byte {
	return []byte("identity:name:" + normalizeIdentityName(name))
}

func normalizeIdentityName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
