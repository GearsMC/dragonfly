package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GameModeCommand, /gamemode komutu.
type GameModeCommand struct {
	Mode   string
	Target cmd.Optional[[]cmd.Target]
}

// Run, gamemode komutunu çalıştırır.
func (g GameModeCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Mode'u parse et
	mode, err := cmd.ParseGameMode(src, g.Mode)
	if err != nil {
		output.Error(err)
		return
	}

	targets, ok := g.Target.Load()
	if !ok {
		// Kaynağın kendisi
		if targeter, ok := src.(cmd.Target); ok {
			targets = []cmd.Target{targeter}
		}
	}

	if len(targets) == 0 {
		output.Errort(cmd.MessageNoTargets)
		return
	}

	// Başarı mesajı
	output.Printf("%d oyuncunun oyun modu %s olarak değiştirildi.", len(targets), mode)

	// Output ayarları
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGameMode)
}

// init, gamemode komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("mod", "").
			Then(
				cmd.Argument("oyuncular", []cmd.Target{}).Optional().
					Executes(&GameModeCommand{}),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"gamemode",
		"Oyuncu oyun modunu değiştirir.",
		[]string{"gm"},
		tree,
	).WithPermissions(permission.CommandGameMode))
}
