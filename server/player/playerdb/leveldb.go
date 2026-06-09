package playerdb

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"os"
)

// Provider is a player data provider that uses a LevelDB database to store data. The data passed on
// will first be converted to make sure it can be marshaled into JSON. This JSON (in bytes) will then
// be stored in the database under a key derived from the player's XUID.
type Provider struct {
	db *leveldb.DB
}

// NewProvider creates a new player data provider that saves and loads data using
// a LevelDB database.
func NewProvider(path string) (*Provider, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0777)
	}
	db, err := leveldb.OpenFile(path, &opt.Options{Compression: opt.SnappyCompression})
	if err != nil {
		return nil, err
	}
	return &Provider{db: db}, nil
}

// Save ...
func (p *Provider) Save(xuid string, d player.Config, w *world.World) error {
	if xuid == "" {
		return fmt.Errorf("xuid boş olamaz")
	}
	d.XUID = xuid
	b, err := json.Marshal(p.toJson(d, w))
	if err != nil {
		return err
	}
	return p.db.Put(playerKey(xuid), b, nil)
}

// Load ...
func (p *Provider) Load(xuid string, lookupWorld player.WorldLookup) (player.Config, *world.World, error) {
	if xuid == "" {
		return player.Config{}, nil, fmt.Errorf("xuid boş olamaz")
	}
	b, err := p.db.Get(playerKey(xuid), nil)
	if err != nil {
		return player.Config{}, nil, err
	}
	var d jsonData
	err = json.Unmarshal(b, &d)
	if err != nil {
		return player.Config{}, nil, err
	}
	return p.fromJson(d, lookupWorld)
}

func playerKey(xuid string) []byte {
	return []byte("xuid:" + xuid)
}

// Close ...
func (p *Provider) Close() error {
	return p.db.Close()
}
