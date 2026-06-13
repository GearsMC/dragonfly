package builtin

import (
	"math"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/world"
)

// KillCommand, /kill komutu.
// Belirtilen oyuncuyu veya kendini öldürür.
type KillCommand struct {
	Target cmd.Optional[[]cmd.Target]
}

// Run, /kill komutunu çalıştırır.
func (k KillCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var targets []cmd.Target
	if p, ok := k.Target.Load(); ok {
		targets = p
	} else if t, ok := src.(cmd.Target); ok {
		targets = []cmd.Target{t}
	}
	if len(targets) == 0 {
		output.Errort(cmd.MessageNoTargets)
		return
	}

	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Errorm(src, "%df.generic.target.notfound")
		return
	}

	for _, p := range players {
		p.Hurt(math.MaxFloat64, entity.VoidDamageSource{})
		output.Printt(i18n.T("%commands.kill.successful", 1), p.Name())
	}
}

// init, kill komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("kill", i18n.D("%df.cmd.kill.description"),
		[]string{"suicide"},
		cmd.NewCommandTree(
			cmd.Argument("hedef", []cmd.Target{}).Optional().
				Executes(&KillCommand{}),
		),
	))
}
