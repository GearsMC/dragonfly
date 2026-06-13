package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// EnchantCommand, /enchant komutu.
// Oyuncunun elindeki eşyaya büyü ekler.
// Büyü listesi enchantmentName enum'u ile client'a gönderilir.
type EnchantCommand struct {
	Player []cmd.Target
	Ench   string
	Level  cmd.Optional[int32]
}

// Run, /enchant komutunu çalıştırır.
func (e EnchantCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	players := resolvePlayers(e.Player)
	if len(players) == 0 {
		output.Errorm(src, "%df.generic.target.notfound")
		return
	}

	enchType := enchantTypeByName(e.Ench)
	if enchType == nil {
		output.Errort(i18n.T("%commands.enchant.notFound", 1), e.Ench)
		return
	}

	level := 1
	if l, ok := e.Level.Load(); ok && l > 0 {
		level = int(l)
	}

	for _, p := range players {
		mainHand, _ := p.HeldItems()
		if mainHand.Empty() {
			output.Errort(i18n.S("%commands.enchant.noItem"))
			continue
		}
		enchanted := mainHand.WithEnchantments(item.NewEnchantment(enchType, level))
		p.SetHeldItems(enchanted, enchanted)
		output.Printm(src, "%df.cmd.enchant.success", p.Name(), e.Ench, level)
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandEnchant)
}

// enchantTypeByName, büyü ismini item.EnchantmentType'a dönüştürür.
func enchantTypeByName(name string) item.EnchantmentType {
	switch name {
	case "protection":
		return enchantment.Protection
	case "fire_protection":
		return enchantment.FireProtection
	case "feather_falling":
		return enchantment.FeatherFalling
	case "blast_protection":
		return enchantment.BlastProtection
	case "projectile_protection":
		return enchantment.ProjectileProtection
	case "thorns":
		return enchantment.Thorns
	case "respiration":
		return enchantment.Respiration
	case "depth_strider":
		return enchantment.DepthStrider
	case "aqua_affinity":
		return enchantment.AquaAffinity
	case "sharpness":
		return enchantment.Sharpness
	case "knockback":
		return enchantment.Knockback
	case "fire_aspect":
		return enchantment.FireAspect
	case "efficiency":
		return enchantment.Efficiency
	case "silk_touch":
		return enchantment.SilkTouch
	case "unbreaking":
		return enchantment.Unbreaking
	case "fortune":
		return enchantment.Fortune
	case "power":
		return enchantment.Power
	case "punch":
		return enchantment.Punch
	case "flame":
		return enchantment.Flame
	case "infinity":
		return enchantment.Infinity
	case "mending":
		return enchantment.Mending
	case "curse_of_vanishing":
		return enchantment.CurseOfVanishing
	case "multishot":
		return enchantment.Multishot
	case "piercing":
		return enchantment.Piercing
	case "quick_charge":
		return enchantment.QuickCharge
	case "soul_speed":
		return enchantment.SoulSpeed
	case "swift_sneak":
		return enchantment.SwiftSneak
	}
	return nil
}

// init, enchant komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("enchant", i18n.D("%df.cmd.enchant.description"),
		nil,
		cmd.NewCommandTree(
			cmd.Argument("oyuncu", []cmd.Target{}).
				Then(
					cmd.Argument("büyü", "", cmd.ArgumentSuggestions("enchantmentName", func(_ cmd.Source) []string {
						return []string{"protection", "fire_protection", "feather_falling", "blast_protection",
							"projectile_protection", "thorns", "respiration", "depth_strider",
							"aqua_affinity", "sharpness", "knockback", "fire_aspect",
							"efficiency", "silk_touch", "unbreaking", "fortune", "power",
							"punch", "flame", "infinity", "mending", "curse_of_vanishing",
							"multishot", "piercing", "quick_charge", "soul_speed", "swift_sneak"}
					})).Then(
						cmd.Argument("seviye", int32(1)).Optional().
							Executes(&EnchantCommand{}),
					),
				),
		),
	).WithPermissions(permission.CommandEnchant))
}
