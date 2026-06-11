package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// resolvePlayers, []cmd.Target hedef listesini *player.Player listesine dönüştürür.
// isim eşleşmesi case-insensitive yapılır.
// tx.Players() iterator'ü üzerinden geçerli oyuncular taranır.
func resolvePlayers(tx *world.Tx, targets []cmd.Target) []*player.Player {
	if tx == nil || len(targets) == 0 {
		return nil
	}

	// İsim → oyuncu eşleşme tablosu
	playerMap := make(map[string]*player.Player)
	for e := range tx.Players() {
		if p, ok := e.(*player.Player); ok {
			playerMap[strings.ToLower(p.Name())] = p
		}
	}

	// Hedefleri çözümle
	var resolved []*player.Player
	for _, t := range targets {
		if named, ok := t.(cmd.NamedTarget); ok {
			if p, ok := playerMap[strings.ToLower(named.Name())]; ok {
				resolved = append(resolved, p)
			}
		}
	}
	return resolved
}

// itemByName, eşya isminden world.Item nesnesine dönüşüm yapar.
// Yaygın eşya isimleri için hızlı bir eşleşme tablosu kullanır.
// Desteklenmeyen isimler için nil döndürür.
func itemByName(name string) world.Item {
	key := strings.ToLower(strings.ReplaceAll(name, "_", ""))

	switch key {
	// Değerli taşlar ve mineraller
	case "diamond":
		return item.Diamond{}
	case "emerald":
		return item.Emerald{}
	case "coal", "charcoal":
		return item.Coal{}
	case "ironingot", "iron":
		return item.IronIngot{}
	case "goldeningot", "gold":
		return item.GoldIngot{}
	case "netheriteingot", "netherite":
		return item.NetheriteIngot{}
	case "lapislazuli", "lapis":
		return item.LapisLazuli{}
	case "redstone":
		return item.RedstoneWire{}

	// Yiyecekler
	case "apple":
		return item.Apple{}
	case "goldenapple", "gapple":
		return item.GoldenApple{}
	case "enchantedapple", "notchapple":
		return item.EnchantedApple{}
	case "bread":
		return item.Bread{}
	case "cookedbeef", "steak":
		return item.Beef{}
	case "cookedporkchop":
		return item.Porkchop{}
	case "cookedchicken":
		return item.Chicken{}
	case "cookie":
		return item.Cookie{}
	case "bakedpotato":
		return item.BakedPotato{}
	case "carrot":
		return item.GoldenCarrot{}

	// Malzemeler
	case "stick":
		return item.Stick{}
	case "paper":
		return item.Paper{}
	case "leather":
		return item.Leather{}
	case "brick":
		return item.Brick{}
	case "clayball", "clay":
		return item.ClayBall{}
	case "glowstonedust":
		return item.GlowstoneDust{}
	case "gunpowder":
		return item.Gunpowder{}
	case "sugar":
		return item.Sugar{}
	case "flint":
		return item.Flint{}
	case "feather":
		return item.Feather{}
	case "string":
		return item.Stick{}
	case "bone":
		return item.Bone{}
	case "boneMeal", "bonemeal":
		return item.BoneMeal{}

	// Sıvı ve iksir
	case "glassbottle", "bottle":
		return item.GlassBottle{}
	case "potion":
		return item.Potion{}
	case "splashpotion":
		return item.SplashPotion{}
	case "lingeringpotion":
		return item.LingeringPotion{}

	// Diğer
	case "enderpearl", "pearl":
		return item.EnderPearl{}
	case "slimeball", "slime":
		return item.Slimeball{}
	case "book":
		return item.Book{}
	case "bow":
		return item.Bow{}
	case "arrow":
		return item.Arrow{}
	case "snowball":
		return item.Snowball{}
	case "egg":
		return item.Egg{}
	case "flintandsteel":
		return item.FlintAndSteel{}
	}

	// Desteklenmeyen eşya
	return nil
}
