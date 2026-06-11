package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// VersionCommand, /version komutu.
// Sunucu ve MC sürüm bilgilerini gösterir.
type VersionCommand struct{}

// Run, /version komutunu çalıştırır.
func (VersionCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	output.Printf("Dragonfly - Minecraft Bedrock %s", "1.21.60")
	output.Print("github.com/df-mc/dragonfly")
}

// init, version komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("version", "Sunucu sürüm bilgisini gösterir.",
		[]string{"ver"},
		cmd.NewCommandTree(cmd.Root().WithPermissions(permission.CommandVersion).Executes(&VersionCommand{})),
	))
}
