package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// VersionCommand, /version komutu.
// Sunucu ve MC sürüm bilgilerini gösterir.
type VersionCommand struct{}

// Run, /version komutunu çalıştırır.
func (VersionCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	output.Printm(src, "%df.cmd.version.format", "1.21.60")
	output.Printm(src, "%df.cmd.version.source")
}

// init, version komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("version", i18n.D("%df.cmd.version.description"),
		[]string{"ver"},
		cmd.NewCommandTree(cmd.Root().WithPermissions(permission.CommandVersion).Executes(&VersionCommand{})),
	))
}
