package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
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
		output.Printf("* %s %s", named.Name(), m.Action)
	} else {
		output.Printf("* %s", m.Action)
	}
}

// init, me komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("me", "Üçüncü şahıs eylem mesajı gönderir.",
		nil,
		cmd.NewCommandTree(cmd.GreedyText("eylem").Executes(&MeCommand{})),
	).WithPermissions(permission.CommandMe))
}
