package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// KickCommand, /kick komutu.
// Belirtilen oyuncuyu sunucudan atar.
type KickCommand struct {
	Player []cmd.Target
	Reason cmd.Optional[cmd.Varargs]
}

// Run, /kick komutunu çalıştırır.
func (k KickCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	players := resolvePlayers(k.Player)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	reason := "Sunucudan atıldı."
	if r, ok := k.Reason.Load(); ok {
		reason = string(r)
	}

	for _, p := range players {
		p.Disconnect(reason)
		output.Printf("%s sunucudan atıldı: %s", p.Name(), reason)
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandKick)
}

// init, kick komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("kick", "Oyuncuyu sunucudan atar.",
		nil,
		cmd.NewCommandTree(
			cmd.Argument("oyuncu", []cmd.Target{}).
				Then(
					cmd.GreedyText("sebep").Optional().
						Executes(&KickCommand{}),
				),
		),
	).WithPermissions(permission.CommandKick))
}
