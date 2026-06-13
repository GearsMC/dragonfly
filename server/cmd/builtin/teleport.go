package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TeleportToPosCommand, /tp <x y z> — kendini koordinata ışınlar.
type TeleportToPosCommand struct {
	Pos mgl64.Vec3
}

func (c TeleportToPosCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if s, ok := src.(cmd.Target); ok {
		players := resolvePlayers([]cmd.Target{s})
		if len(players) == 0 {
			output.Errorm(src, "%df.cmd.teleport.error.source")
			return
		}
		players[0].Teleport(c.Pos)
		output.Printm(src, "%df.cmd.teleport.self.pos", players[0].Name(), c.Pos[0], c.Pos[1], c.Pos[2])
	} else {
		output.Errorm(src, "%df.generic.console.only")
		return
	}
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTeleport)
}

// TeleportToPlayerCommand, /tp <hedef> — kendini hedef oyuncuya ışınlar.
type TeleportToPlayerCommand struct {
	Dest []cmd.Target
}

func (c TeleportToPlayerCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if s, ok := src.(cmd.Target); ok {
		selfPlayers := resolvePlayers([]cmd.Target{s})
		if len(selfPlayers) == 0 {
			output.Errorm(src, "%df.cmd.teleport.error.source")
			return
		}
		destPlayers := resolvePlayers(c.Dest)
		if len(destPlayers) == 0 {
			output.Errorm(src, "%df.cmd.teleport.error.destination")
			return
		}
		destPos := destPlayers[0].Position()
		selfPlayers[0].Teleport(destPos)
		output.Printm(src, "%df.cmd.teleport.self.player", selfPlayers[0].Name(), destPlayers[0].Name())
	} else {
		output.Errorm(src, "%df.generic.console.only")
		return
	}
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTeleport)
}

// TeleportPlayerToPosCommand, /tp <kurban> <x y z> — oyuncuyu koordinata ışınlar.
type TeleportPlayerToPosCommand struct {
	Victims []cmd.Target
	Pos     mgl64.Vec3
}

func (c TeleportPlayerToPosCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	victims := resolvePlayers(c.Victims)
	if len(victims) == 0 {
		output.Errorm(src, "%df.cmd.teleport.error.victim")
		return
	}
	for _, p := range victims {
		p.Teleport(c.Pos)
	}
	if len(victims) == 1 {
		output.Printm(src, "%df.cmd.teleport.other.pos.single", victims[0].Name(), c.Pos[0], c.Pos[1], c.Pos[2])
	} else {
		output.Printm(src, "%df.cmd.teleport.other.pos", len(victims), c.Pos[0], c.Pos[1], c.Pos[2])
	}
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTeleport)
}

// TeleportPlayerToPlayerCommand, /tp <kurban> <hedef> — oyuncuyu oyuncuya ışınlar.
type TeleportPlayerToPlayerCommand struct {
	Victims []cmd.Target
	Dest    []cmd.Target
}

func (c TeleportPlayerToPlayerCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	victims := resolvePlayers(c.Victims)
	if len(victims) == 0 {
		output.Errorm(src, "%df.cmd.teleport.error.victim")
		return
	}
	destPlayers := resolvePlayers(c.Dest)
	if len(destPlayers) == 0 {
		output.Errorm(src, "%df.cmd.teleport.error.destination")
		return
	}
	destPos := destPlayers[0].Position()
	for _, p := range victims {
		p.Teleport(destPos)
	}
	if len(victims) == 1 {
		output.Printm(src, "%df.cmd.teleport.other.player.single", victims[0].Name(), destPlayers[0].Name())
	} else {
		output.Printm(src, "%df.cmd.teleport.other.player", len(victims), destPlayers[0].Name())
	}
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTeleport)
}

// init, teleport komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("teleport", i18n.D("%df.cmd.teleport.description"),
		[]string{"tp"},
		cmd.NewCommandTree(
			// /tp <x y z>
			cmd.Argument("konum", mgl64.Vec3{}).Executes(&TeleportToPosCommand{}),
			// /tp <hedef>
			cmd.Argument("hedef", []cmd.Target{}).Executes(&TeleportToPlayerCommand{}),
			// /tp <kurban> <x y z>
			cmd.Argument("kurban", []cmd.Target{}).Then(
				cmd.Argument("konum", mgl64.Vec3{}).Executes(&TeleportPlayerToPosCommand{}),
			),
			// /tp <kurban> <hedef>
			cmd.Argument("kurban2", []cmd.Target{}).Then(
				cmd.Argument("hedef2", []cmd.Target{}).Executes(&TeleportPlayerToPlayerCommand{}),
			),
		),
	).WithPermissions(permission.CommandTeleport))
}
