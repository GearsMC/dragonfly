package playerdb

import (
	"errors"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
)

func TestWorldIdentityRoundTrip(t *testing.T) {
	w := testWorld(t, "island-42", world.Nether)
	spawnWorld := testWorld(t, "spawn-island", world.Overworld)
	provider := &Provider{}
	conf := testConfig()
	conf.Spawn = player.Spawn{Position: cube.Pos{12, 64, -5}, World: spawnWorld, Valid: true}
	data := provider.toJson(conf, w)

	if data.World != "island-42" {
		t.Fatalf("beklenmeyen dünya adı: %q", data.World)
	}
	if data.Dimension != 1 {
		t.Fatalf("beklenmeyen dimension kimliği: %d", data.Dimension)
	}
	if data.LastKnownName != "lexa5936" {
		t.Fatalf("beklenmeyen son bilinen isim: %q", data.LastKnownName)
	}
	if !data.SpawnValid || data.SpawnWorld != "spawn-island" || data.SpawnDimension != 0 || data.SpawnPosition != (cube.Pos{12, 64, -5}) {
		t.Fatalf("beklenmeyen spawn verisi: geçerli=%t dünya=%q dimension=%d pozisyon=%v", data.SpawnValid, data.SpawnWorld, data.SpawnDimension, data.SpawnPosition)
	}

	var lookedUpPlayerWorld, lookedUpSpawnWorld bool
	loaded, loadedWorld, err := provider.fromJson(data, func(name string, dimension world.Dimension) *world.World {
		switch {
		case name == "island-42" && dimension == world.Nether:
			lookedUpPlayerWorld = true
			return w
		case name == "spawn-island" && dimension == world.Overworld:
			lookedUpSpawnWorld = true
			return spawnWorld
		default:
			return nil
		}
	})
	if err != nil {
		t.Fatalf("oyuncu verisi yüklenemedi: %v", err)
	}
	if !lookedUpPlayerWorld || !lookedUpSpawnWorld {
		t.Fatalf("oyuncu ve spawn dünyalarının birlikte aranması bekleniyordu: oyuncu=%t spawn=%t", lookedUpPlayerWorld, lookedUpSpawnWorld)
	}
	if loadedWorld != w {
		t.Fatal("yüklenen dünya kaydedilen dünyayla eşleşmiyor")
	}
	if loaded.Spawn.World != spawnWorld || loaded.Spawn.Position != (cube.Pos{12, 64, -5}) || !loaded.Spawn.Valid {
		t.Fatalf("spawn verisi doğru yüklenmedi: %+v", loaded.Spawn)
	}
}

func TestLegacyUsernameLoadsAsLastKnownName(t *testing.T) {
	w := testWorld(t, "World", world.Overworld)
	provider := &Provider{}
	data := provider.toJson(testConfig(), w)
	data.LastKnownName = ""
	data.Username = "legacyName"

	conf, _, err := provider.fromJson(data, func(string, world.Dimension) *world.World {
		return w
	})
	if err != nil {
		t.Fatalf("eski isim alanı yüklenemedi: %v", err)
	}
	if conf.Name != "legacyName" {
		t.Fatalf("eski isim alanı son bilinen isim olarak yüklenmedi: %q", conf.Name)
	}
}

func TestLegacyWorldIdentityLookup(t *testing.T) {
	w := testWorld(t, "World", world.End)
	provider := &Provider{}
	data := provider.toJson(testConfig(), w)
	data.World = ""

	_, loadedWorld, err := provider.fromJson(data, func(name string, dimension world.Dimension) *world.World {
		if name == "" && dimension == world.End {
			return w
		}
		return nil
	})
	if err != nil {
		t.Fatalf("eski oyuncu verisi yüklenemedi: %v", err)
	}
	if loadedWorld != w {
		t.Fatal("eski oyuncu verisi dimension bilgisiyle çözümlenmedi")
	}
}

func TestMissingWorldIdentityRejected(t *testing.T) {
	w := testWorld(t, "deleted-island", world.Overworld)
	data := (&Provider{}).toJson(testConfig(), w)

	_, _, err := (&Provider{}).fromJson(data, func(string, world.Dimension) *world.World {
		return nil
	})
	if err == nil {
		t.Fatal("bulunmayan oyuncu dünyasının reddedilmesi bekleniyordu")
	}
}

func TestMissingSpawnWorldClearsSpawnOnly(t *testing.T) {
	w := testWorld(t, "World", world.Overworld)
	spawnWorld := testWorld(t, "deleted-spawn", world.Overworld)
	conf := testConfig()
	conf.Spawn = player.Spawn{Position: cube.Pos{5, 70, 5}, World: spawnWorld, Valid: true}
	data := (&Provider{}).toJson(conf, w)

	loaded, loadedWorld, err := (&Provider{}).fromJson(data, func(name string, dimension world.Dimension) *world.World {
		if name == "World" && dimension == world.Overworld {
			return w
		}
		return nil
	})
	if err != nil {
		t.Fatalf("spawn dünyası silinen oyuncu verisi yüklenemedi: %v", err)
	}
	if loadedWorld != w {
		t.Fatal("ana oyuncu dünyası yüklenmedi")
	}
	if loaded.Spawn.Valid || loaded.Spawn.World != nil {
		t.Fatalf("silinmiş spawn dünyası geçerli kalmamalıydı: %+v", loaded.Spawn)
	}
}

func TestProviderStoresPlayerDataByXUID(t *testing.T) {
	provider, err := NewProvider(t.TempDir())
	if err != nil {
		t.Fatalf("provider açılamadı: %v", err)
	}
	t.Cleanup(func() {
		_ = provider.Close()
	})

	w := testWorld(t, "World", world.Overworld)
	conf := testConfig()
	conf.XUID = "2535457450295374"
	conf.Name = "lexa5936"
	oldUUID := conf.UUID

	if err := provider.Save(conf.XUID, conf, w); err != nil {
		t.Fatalf("oyuncu verisi XUID ile kaydedilemedi: %v", err)
	}
	if _, err := provider.db.Get(oldUUID[:], nil); !errors.Is(err, leveldb.ErrNotFound) {
		t.Fatalf("oyuncu verisi UUID key altında yazılmamalıydı: %v", err)
	}

	conf.UUID = uuid.New()
	conf.Name = "lexaNew"
	if err := provider.Save(conf.XUID, conf, w); err != nil {
		t.Fatalf("isim değişen oyuncu verisi XUID ile güncellenemedi: %v", err)
	}

	loaded, loadedWorld, err := provider.Load(conf.XUID, func(name string, dimension world.Dimension) *world.World {
		if name == "World" && dimension == world.Overworld {
			return w
		}
		return nil
	})
	if err != nil {
		t.Fatalf("oyuncu verisi XUID ile yüklenemedi: %v", err)
	}
	if loadedWorld != w {
		t.Fatal("oyuncu verisi yanlış dünyaya yüklendi")
	}
	if loaded.XUID != conf.XUID {
		t.Fatalf("beklenmeyen XUID: %q", loaded.XUID)
	}
	if loaded.UUID != conf.UUID {
		t.Fatalf("beklenmeyen UUID: %v", loaded.UUID)
	}
	if loaded.Name != "lexaNew" {
		t.Fatalf("son bilinen isim güncellenmedi: %q", loaded.Name)
	}
}

func testWorld(t *testing.T, name string, dimension world.Dimension) *world.World {
	t.Helper()
	w := world.Config{
		Dim:      dimension,
		Provider: world.NopProvider{Set: &world.Settings{Name: name}},
	}.New()
	t.Cleanup(func() {
		_ = w.Close()
	})
	return w
}

func testConfig() player.Config {
	return player.Config{
		UUID:                uuid.New(),
		XUID:                "2535457450295374",
		Name:                "lexa5936",
		GameMode:            world.GameModeSurvival,
		Inventory:           inventory.New(36, nil),
		EnderChestInventory: inventory.New(27, nil),
		OffHand:             inventory.New(1, nil),
		Armour:              inventory.NewArmour(nil),
	}
}
