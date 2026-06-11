package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// SeedCommand, /seed komutu.
// Dünya seed değerini gösterir.
type SeedCommand struct{}

// Run, /seed komutunu çalıştırır.
func (SeedCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	output.Printf("Seed: %d", int64(0))
}

// init, seed komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("seed", "Dünya seed değerini gösterir.",
		nil,
		cmd.NewCommandTree(cmd.Root().WithPermissions(permission.CommandSeed).Executes(&SeedCommand{})),
	))
}
