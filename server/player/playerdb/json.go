package playerdb

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"time"
)

func (p *Provider) fromJson(d jsonData, lookupWorld player.WorldLookup) (player.Config, *world.World, error) {
	dim, ok := world.DimensionByID(int(d.Dimension))
	if !ok {
		return player.Config{}, nil, fmt.Errorf("bilinmeyen oyuncu dimension kimliği: %d", d.Dimension)
	}
	w := lookupWorld(d.World, dim)
	if w == nil {
		if d.World == "" {
			return player.Config{}, nil, fmt.Errorf("eski oyuncu dünyası dimension %d içinde çözümlenemedi", d.Dimension)
		}
		return player.Config{}, nil, fmt.Errorf("oyuncu dünyası %q dimension %d içinde çözümlenemedi", d.World, d.Dimension)
	}
	mode, _ := world.GameModeByID(int(d.GameMode))
	name := d.LastKnownName
	if name == "" {
		name = d.Username
	}
	spawn := player.Spawn{}
	if d.SpawnValid {
		if spawnDim, ok := world.DimensionByID(int(d.SpawnDimension)); ok {
			if spawnWorld := lookupWorld(d.SpawnWorld, spawnDim); spawnWorld != nil {
				spawn = player.Spawn{Position: d.SpawnPosition, World: spawnWorld, Valid: true}
			}
		}
	}
	conf := player.Config{
		UUID:                uuid.MustParse(d.UUID),
		XUID:                d.XUID,
		Name:                name,
		Position:            d.Position,
		Rotation:            cube.Rotation{d.Yaw, d.Pitch},
		Velocity:            d.Velocity,
		Health:              d.Health,
		MaxHealth:           d.MaxHealth,
		Food:                d.Hunger,
		FoodTick:            d.FoodTick,
		Exhaustion:          d.ExhaustionLevel,
		Saturation:          d.SaturationLevel,
		Experience:          d.Experience,
		AirSupply:           d.AirSupply,
		MaxAirSupply:        d.MaxAirSupply,
		EnchantmentSeed:     d.EnchantmentSeed,
		GameMode:            mode,
		Effects:             dataToEffects(d.Effects),
		FireTicks:           d.FireTicks,
		FallDistance:        d.FallDistance,
		Inventory:           inventory.New(36, nil),
		EnderChestInventory: inventory.New(27, nil),
		OffHand:             inventory.New(1, nil),
		Armour:              inventory.NewArmour(nil),
		Spawn:               spawn,
	}
	echest := make([]item.Stack, 27)
	decodeItems(d.EnderChestInventory, echest)
	invData := dataToInv(d.Inventory)

	for slot, stack := range invData.Items {
		_ = conf.Inventory.SetItem(slot, stack)
	}
	_ = conf.OffHand.SetItem(0, invData.OffHand)
	conf.Armour.Set(invData.Helmet, invData.Chestplate, invData.Leggings, invData.Boots)
	conf.HeldSlot = int(invData.MainHandSlot)

	for slot, stack := range echest {
		_ = conf.EnderChestInventory.SetItem(slot, stack)
	}
	return conf, w, nil
}

func (p *Provider) toJson(d player.Config, w *world.World) jsonData {
	dim, _ := world.DimensionID(w.Dimension())
	mode, _ := world.GameModeID(d.GameMode)
	offHand, _ := d.OffHand.Item(0)
	data := jsonData{
		UUID:            d.UUID.String(),
		XUID:            d.XUID,
		LastKnownName:   d.Name,
		Position:        d.Position,
		Velocity:        d.Velocity,
		Yaw:             d.Rotation.Yaw(),
		Pitch:           d.Rotation.Pitch(),
		Health:          d.Health,
		MaxHealth:       d.MaxHealth,
		Hunger:          d.Food,
		FoodTick:        d.FoodTick,
		ExhaustionLevel: d.Exhaustion,
		SaturationLevel: d.Saturation,
		Experience:      d.Experience,
		AirSupply:       d.AirSupply,
		MaxAirSupply:    d.MaxAirSupply,
		EnchantmentSeed: d.EnchantmentSeed,
		GameMode:        uint8(mode),
		Effects:         effectsToData(d.Effects),
		FireTicks:       d.FireTicks,
		FallDistance:    d.FallDistance,
		Inventory: invToData(InventoryData{
			Items:        d.Inventory.Slots(),
			Boots:        d.Armour.Boots(),
			Leggings:     d.Armour.Leggings(),
			Chestplate:   d.Armour.Chestplate(),
			Helmet:       d.Armour.Helmet(),
			OffHand:      offHand,
			MainHandSlot: uint32(d.HeldSlot),
		}),
		EnderChestInventory: encodeItems(d.EnderChestInventory.Slots()),
		World:               w.Name(),
		Dimension:           uint8(dim),
	}
	if d.Spawn.Valid && d.Spawn.World != nil {
		spawnDim, _ := world.DimensionID(d.Spawn.World.Dimension())
		data.SpawnValid = true
		data.SpawnPosition = d.Spawn.Position
		data.SpawnWorld = d.Spawn.World.Name()
		data.SpawnDimension = uint8(spawnDim)
	}
	return data
}

type jsonData struct {
	UUID                             string
	XUID                             string
	LastKnownName                    string
	Username                         string `json:"Username,omitempty"`
	Position, Velocity               mgl64.Vec3
	Yaw, Pitch                       float64
	Health, MaxHealth                float64
	Hunger                           int
	FoodTick                         int
	ExhaustionLevel, SaturationLevel float64
	EnchantmentSeed                  int64
	Experience                       int
	AirSupply, MaxAirSupply          int
	GameMode                         uint8
	Inventory                        jsonInventoryData
	EnderChestInventory              []jsonSlot
	Effects                          []jsonEffect
	FireTicks                        int64
	FallDistance                     float64
	World                            string
	Dimension                        uint8
	SpawnValid                       bool
	SpawnPosition                    cube.Pos
	SpawnWorld                       string
	SpawnDimension                   uint8
}

type jsonInventoryData struct {
	Items        []jsonSlot
	Boots        []byte
	Leggings     []byte
	Chestplate   []byte
	Helmet       []byte
	OffHand      []byte
	MainHandSlot uint32
}

type jsonSlot struct {
	Item []byte
	Slot int
}

type jsonEffect struct {
	ID              int
	Level           int
	Duration        time.Duration
	Ambient         bool
	ParticlesHidden bool
	Infinite        bool
}
