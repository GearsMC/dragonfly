package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// MeCommand, /me komutu.
// Oyuncunun kendi adına üçüncü şahıs eylem mesajı gönderir.
type MeCommand struct {
	Action cmd.Varargs
}

// Run, /me komutunu çalıştırır.
func (m MeCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if named, ok := src.(cmd.NamedTarget); ok {
		output.Printm(src, "%df.cmd.me.format", named.Name(), m.Action)
	} else {
		output.Printm(src, "%df.cmd.me.format.anon", m.Action)
	}
}

// init, me komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("me", i18n.D("%df.cmd.me.description"),
		nil,
		cmd.NewCommandTree(cmd.GreedyText("eylem").Executes(&MeCommand{})),
	).WithPermissions(permission.CommandMe))
}
