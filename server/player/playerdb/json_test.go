package playerdb

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

func TestWorldIdentityRoundTrip(t *testing.T) {
	w := testWorld(t, "island-42", world.Nether)
	provider := &Provider{}
	data := provider.toJson(testConfig(), w)

	if data.World != "island-42" {
		t.Fatalf("beklenmeyen dünya adı: %q", data.World)
	}
	if data.Dimension != 1 {
		t.Fatalf("beklenmeyen dimension kimliği: %d", data.Dimension)
	}

	var lookupName string
	var lookupDimension world.Dimension
	_, loadedWorld, err := provider.fromJson(data, func(name string, dimension world.Dimension) *world.World {
		lookupName, lookupDimension = name, dimension
		return w
	})
	if err != nil {
		t.Fatalf("oyuncu verisi yüklenemedi: %v", err)
	}
	if lookupName != "island-42" || lookupDimension != world.Nether {
		t.Fatalf("beklenmeyen dünya araması: ad=%q dimension=%v", lookupName, lookupDimension)
	}
	if loadedWorld != w {
		t.Fatal("yüklenen dünya kaydedilen dünyayla eşleşmiyor")
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
		Name:                "lexa5936",
		GameMode:            world.GameModeSurvival,
		Inventory:           inventory.New(36, nil),
		EnderChestInventory: inventory.New(27, nil),
		OffHand:             inventory.New(1, nil),
		Armour:              inventory.NewArmour(nil),
	}
}
