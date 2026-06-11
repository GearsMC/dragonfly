package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// XPCommand, /xp komutu.
// Oyuncuya deneyim puanı verir veya seviyesini ayarlar.
type XPCommand struct {
	Amount int32
	Player cmd.Optional[[]cmd.Target]
}

// Run, /xp komutunu çalıştırır.
func (x XPCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var targets []cmd.Target
	if p, ok := x.Player.Load(); ok {
		targets = p
	} else if t, ok := src.(cmd.Target); ok {
		targets = []cmd.Target{t}
	}
	if len(targets) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	for _, p := range players {
		p.AddExperience(int(x.Amount))
		output.Printf("%s oyuncusuna %d deneyim verildi.", p.Name(), x.Amount)
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandXP)
}

// init, xp komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("xp", "Oyuncuya deneyim verir.",
		[]string{"experience"},
		cmd.NewCommandTree(
			cmd.Argument("miktar", int32(0)).
				Then(
					cmd.Argument("oyuncu", []cmd.Target{}).Optional().
						Executes(&XPCommand{}),
				),
		),
	).WithPermissions(permission.CommandXP))
}
