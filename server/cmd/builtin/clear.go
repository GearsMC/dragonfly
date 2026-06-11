package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// ClearCommand, /clear komutu.
// Oyuncunun envanterini temizler.
type ClearCommand struct {
	Player cmd.Optional[[]cmd.Target]
	Item   cmd.Optional[string]
	Data   cmd.Optional[int32]
}

// Run, /clear komutunu çalıştırır.
func (c ClearCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var targets []cmd.Target
	if p, ok := c.Player.Load(); ok {
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
		p.Inventory().Clear()
		output.Printf("%s oyuncusunun envanteri temizlendi.", p.Name())
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandClear)
}

// init, clear komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("clear", "Oyuncu envanterini temizler.",
		nil,
		cmd.NewCommandTree(
			cmd.Argument("oyuncu", []cmd.Target{}).Optional().
				Then(
					cmd.Argument("eşya", "", cmd.ArgumentSuggestions("Item", func(_ cmd.Source) []string {
						return nil
					})).Optional().
						Then(
							cmd.Argument("veri", int32(0)).Optional().
								Executes(&ClearCommand{}),
						),
				),
		),
	).WithPermissions(permission.CommandClear))
}
