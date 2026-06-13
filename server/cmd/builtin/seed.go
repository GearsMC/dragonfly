package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// SeedCommand, /seed komutu.
// Dünya seed değerini gösterir.
type SeedCommand struct{}

// Run, /seed komutunu çalıştırır.
func (SeedCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	output.Printm(src, "%df.cmd.seed.success", int64(0))
}

// init, seed komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("seed", i18n.D("%df.cmd.seed.description"),
		nil,
		cmd.NewCommandTree(cmd.Root().WithPermissions(permission.CommandSeed).Executes(&SeedCommand{})),
	))
}
