package builtin

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// RegisterServer registers built-in server control commands.
func RegisterServer(srv *server.Server) {
	cmd.Register(cmd.New("stop", "Stops the server.", nil, stopCommand{srv: srv}))
}

type stopCommand struct {
	srv *server.Server `cmd:"-"`
}

func (c stopCommand) Allow(src cmd.Source) bool {
	_, ok := src.(cmd.ConsoleSource)
	return ok
}

func (c stopCommand) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	o.Print(text.Yellow + "Stopping the server..." + text.Reset)
	go func() {
		_ = c.srv.Close()
	}()
}
