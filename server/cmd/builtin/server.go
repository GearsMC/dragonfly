package builtin

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// RegisterServer registers built-in server control commands.
func RegisterServer(srv *server.Server) {
	cmd.Register(cmd.NewWithTree("stop", "Sunucuyu durdurur.", nil, cmd.NewCommandTree(
		cmd.Root().WithPermissions(permission.CommandStop).Executes(stopCommand{srv: srv}),
	)))
}

type stopCommand struct {
	srv *server.Server `cmd:"-"`
}

func (c stopCommand) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	o.Print(text.Yellow + "Sunucu durduruluyor..." + text.Reset)
	go func() {
		_ = c.srv.Close()
	}()
}
