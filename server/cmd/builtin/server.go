package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// RegisterServer registers built-in server control commands.
func RegisterServer(srv *server.Server) {
	cmd.Register(cmd.NewWithTree("stop", i18n.D("%df.cmd.stop.description"), nil, cmd.NewCommandTree(
		cmd.Root().WithPermissions(permission.CommandStop).Executes(stopCommand{srv: srv}),
	)))
	cmd.Register(cmd.NewWithTree("op", i18n.D("%df.cmd.op.description"), nil, cmd.NewCommandTree(
		cmd.Root().WithPermissions(permission.CommandOP).Then(
			cmd.Argument("player", "").Executes(opCommand{srv: srv}),
		),
	)))
	cmd.Register(cmd.NewWithTree("deop", i18n.D("%df.cmd.deop.description"), nil, cmd.NewCommandTree(
		cmd.Root().WithPermissions(permission.CommandDeOP).Then(
			cmd.Argument("player", "").Executes(deopCommand{srv: srv}),
		),
	)))
}

type stopCommand struct {
	srv *server.Server `cmd:"-"`
}

func (c stopCommand) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	o.Printm(src, "%df.server.stop")
	go func() {
		_ = c.srv.Close()
	}()
}

type opCommand struct {
	srv    *server.Server `cmd:"-"`
	Player string
}

func (c opCommand) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	identity, ok, err := resolveOperatorTarget(c.srv, c.Player)
	if err != nil {
		o.Errorm(src, "%df.server.op.error.identity", err)
		return
	}
	if !ok {
		o.Printm(src, "%df.server.op.notfound", c.Player)
		return
	}
	if c.srv.IsOperatorXUID(identity.xuid) {
		o.Printm(src, "%df.server.op.already", identity.display())
		return
	}
	if err := c.srv.SetOperatorXUID(identity.xuid, identity.name, true); err != nil {
		o.Errorm(src, "%df.server.op.error.identity", err)
		return
	}
	o.Printm(src, "%df.server.op.success", identity.display())
}

type deopCommand struct {
	srv    *server.Server `cmd:"-"`
	Player string
}

func (c deopCommand) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	identity, ok, err := resolveOperatorTarget(c.srv, c.Player)
	if err != nil {
		o.Errorm(src, "%df.server.op.error.identity", err)
		return
	}
	if !ok {
		o.Printm(src, "%df.server.op.notfound", c.Player)
		return
	}
	if !c.srv.IsOperatorXUID(identity.xuid) {
		o.Printm(src, "%df.server.deop.not", identity.display())
		return
	}
	if err := c.srv.SetOperatorXUID(identity.xuid, identity.name, false); err != nil {
		o.Errorm(src, "%df.server.op.error.identity", err)
		return
	}
	o.Printm(src, "%df.server.deop.success", identity.display())
}

type operatorTarget struct {
	xuid string
	name string
}

func (t operatorTarget) display() string {
	if t.name == "" {
		return t.xuid
	}
	return t.name + " (" + t.xuid + ")"
}

func resolveOperatorTarget(srv *server.Server, input string) (operatorTarget, bool, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return operatorTarget{}, false, nil
	}
	if isXUID(input) {
		identity, ok, err := srv.ResolveIdentityByXUID(input)
		if err != nil {
			return operatorTarget{}, false, err
		}
		if ok {
			return operatorTargetFromIdentity(identity.XUID, identity.LastKnownName, input), true, nil
		}
		identity, ok, err = srv.ResolveIdentityByName(input)
		if err != nil {
			return operatorTarget{}, false, err
		}
		if ok {
			return operatorTargetFromIdentity(identity.XUID, identity.LastKnownName, input), true, nil
		}
		return operatorTarget{xuid: input}, true, nil
	}

	identity, ok, err := srv.ResolveIdentityByName(input)
	if err != nil {
		return operatorTarget{}, false, err
	}
	if !ok {
		return operatorTarget{}, false, nil
	}
	return operatorTargetFromIdentity(identity.XUID, identity.LastKnownName, input), true, nil
}

func operatorTargetFromIdentity(xuid, name, fallbackName string) operatorTarget {
	if name == "" {
		name = fallbackName
	}
	return operatorTarget{xuid: xuid, name: name}
}

func isXUID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
