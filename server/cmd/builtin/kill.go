package builtin

import (
	"math"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity"
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
		output.Error("Hedef bulunamadı.")
		return
	}

	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	for _, p := range players {
		p.Hurt(math.MaxFloat64, entity.VoidDamageSource{})
		output.Printf("%s öldürüldü.", p.Name())
	}
}

// init, kill komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("kill", "Oyuncuyu öldürür.",
		[]string{"suicide"},
		cmd.NewCommandTree(
			cmd.Argument("hedef", []cmd.Target{}).Optional().
				Executes(&KillCommand{}),
		),
	))
}
