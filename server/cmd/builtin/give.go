package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GiveCommand, /give komutu.
type GiveCommand struct {
	Target []cmd.Target
	Item   string
	Amount cmd.Optional[int32]
}

// Run, give komutunu çalıştırır.
func (g GiveCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	if len(g.Target) == 0 {
		output.Error("Geçersiz hedef")
		return
	}

	amount := int32(1)
	if amt, ok := g.Amount.Load(); ok {
		if amt <= 0 || amt > 64 {
			output.Errorf("Miktar 1-64 arasında olmalıdır")
			return
		}
		amount = amt
	}

	output.Printf("%v oyuncuya %d x %s verildi.", g.Target[0], amount, g.Item)

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGive)
}

// init, give komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("oyuncu", []cmd.Target{}).
			Then(
				cmd.Argument("eşya", "").
					Then(
						cmd.Argument("miktar", int32(1)).Optional().
							Executes(&GiveCommand{}),
					),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"give",
		"Oyuncuya eşya verir.",
		nil,
		tree,
	).WithPermissions(permission.CommandGive))
}
