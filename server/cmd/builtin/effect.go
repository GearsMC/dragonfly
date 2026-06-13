package builtin

import (
	"strconv"
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// EffectCommand, /effect komutu.
// Oyuncuya durum efekti verir veya kaldırır.
// Efekt listesi Effect enum'u ile client'a gönderilir.
type EffectCommand struct {
	Player  []cmd.Target
	Effect  string
	Seconds cmd.Optional[int32]
	Level   cmd.Optional[int32]
	Hide    cmd.Optional[string]
}

// Run, /effect komutunu çalıştırır.
func (e EffectCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	players := resolvePlayers(e.Player)
	if len(players) == 0 {
		output.Errorm(src, "%df.generic.target.notfound")
		return
	}

	if e.Effect == "clear" {
		for _, p := range players {
			for _, ef := range p.Effects() {
				p.RemoveEffect(ef.Type())
			}
		}
		output.Printm(src, "%df.cmd.effect.clear", len(players))
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandEffect)
		return
	}

	effType := effectTypeByName(e.Effect)
	if effType == nil {
		output.Errort(i18n.T("%commands.effect.notFound", 1), e.Effect)
		return
	}

	seconds := int32(30)
	if s, ok := e.Seconds.Load(); ok && s > 0 {
		seconds = s
	}
	level := 1
	if l, ok := e.Level.Load(); ok && l > 0 {
		level = int(l)
	}

	for _, p := range players {
		if lasting, ok := effType.(effect.LastingType); ok {
			p.AddEffect(effect.New(lasting, level, time.Duration(seconds)*time.Second))
		} else {
			p.AddEffect(effect.NewInstant(effType, level))
		}
	}

	if len(players) == 1 {
		output.Printm(src, "%df.cmd.effect.success", players[0].Name(), e.Effect, level, seconds)
	} else {
		output.Printm(src, "%df.cmd.effect.success.multi", len(players), e.Effect, level, seconds)
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandEffect)
}

// effectTypeByName, efekt ismini effect.Type'a dönüştürür.
func effectTypeByName(name string) effect.Type {
	id, err := strconv.Atoi(name)
	if err == nil {
		if t, ok := effect.ByID(id); ok {
			return t
		}
	}
	switch name {
	case "speed":
		return effect.Speed
	case "slowness":
		return effect.Slowness
	case "haste":
		return effect.Haste
	case "mining_fatigue":
		return effect.MiningFatigue
	case "strength":
		return effect.Strength
	case "instant_health":
		return effect.InstantHealth
	case "instant_damage":
		return effect.InstantDamage
	case "jump_boost":
		return effect.JumpBoost
	case "nausea":
		return effect.Nausea
	case "regeneration":
		return effect.Regeneration
	case "resistance":
		return effect.Resistance
	case "fire_resistance":
		return effect.FireResistance
	case "water_breathing":
		return effect.WaterBreathing
	case "invisibility":
		return effect.Invisibility
	case "blindness":
		return effect.Blindness
	case "night_vision":
		return effect.NightVision
	case "hunger":
		return effect.Hunger
	case "weakness":
		return effect.Weakness
	case "poison":
		return effect.Poison
	case "wither":
		return effect.Wither
	case "health_boost":
		return effect.HealthBoost
	case "absorption":
		return effect.Absorption
	case "saturation":
		return effect.Saturation
	case "levitation":
		return effect.Levitation
	case "fatal_poison":
		return effect.FatalPoison
	case "conduit_power":
		return effect.ConduitPower
	case "slow_falling":
		return effect.SlowFalling
	case "darkness":
		return effect.Darkness
	}
	return nil
}

// init, effect komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("effect", i18n.D("%df.cmd.effect.description"),
		nil,
		cmd.NewCommandTree(
			cmd.Argument("oyuncu", []cmd.Target{}).
				Then(
					cmd.Argument("efekt", "", cmd.ArgumentSuggestions("Effect", func(_ cmd.Source) []string {
						return []string{"speed", "slowness", "haste", "mining_fatigue", "strength",
							"instant_health", "instant_damage", "jump_boost", "nausea", "regeneration",
							"resistance", "fire_resistance", "water_breathing", "invisibility", "blindness",
							"night_vision", "hunger", "weakness", "poison", "wither", "health_boost",
							"absorption", "saturation", "levitation", "fatal_poison", "conduit_power",
							"slow_falling", "darkness", "clear"}
					})).Then(
						cmd.Argument("süre", int32(30)).Optional().
							Then(
								cmd.Argument("seviye", int32(1)).Optional().
									Then(
										cmd.Argument("gizle", "").Optional().
											Executes(&EffectCommand{}),
									),
							),
					),
				),
		),
	).WithPermissions(permission.CommandEffect))
}
