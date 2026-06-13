package builtin

import (
	"errors"
	"strings"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RegisterWorld registers the built-in /world multi-world management command.
func RegisterWorld(srv *server.Server) {
	cmd.Register(cmd.NewWithTree("world", i18n.D("%df.cmd.world.description"), nil, cmd.NewCommandTree(
		cmd.Literal("list").Executes(worldListCommand{srv: srv}),
		cmd.Literal("tp").Then(
			cmd.Argument("level", "").Executes(worldTPNoDimCommand{srv: srv}).Then(
				cmd.Argument("dimension", "").Executes(worldTPWithDimCommand{srv: srv}),
			),
		),
		cmd.Literal("create").Then(
			cmd.Argument("name", "").Executes(worldCreateCommand{srv: srv}),
		),
		cmd.Literal("unload").Then(
			cmd.Argument("name", "").Executes(worldUnloadCommand{srv: srv}),
		),
		cmd.Literal("setdefault").Then(
			cmd.Argument("name", "").Executes(worldSetDefaultCommand{srv: srv}),
		),
	)).WithPermissions(permission.CommandWorld))
}

// worldListCommand lists all loaded levels.
type worldListCommand struct {
	srv *server.Server `cmd:"-"`
}

func (c worldListCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	levels := c.srv.LevelManager().Levels()
	if len(levels) == 0 {
		output.Printm(src, "%df.cmd.world.list.empty")
		return
	}
	names := make([]string, len(levels))
	for i, l := range levels {
		names[i] = l.Name
	}
	output.Printm(src, "%df.cmd.world.list", len(names), strings.Join(names, ", "))
}

// worldTPNoDimCommand teleports the source player to a level spawn, defaulting
// to the Overworld dimension. Bound to the leaf "/world tp <level>".
type worldTPNoDimCommand struct {
	srv   *server.Server `cmd:"-"`
	Level string
}

func (c worldTPNoDimCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	teleportToLevel(c.srv, src, output, c.Level, world.Overworld)
}

// worldTPWithDimCommand teleports the source player to a specific (level,
// dimension) pair. Bound to the leaf "/world tp <level> <dimension>".
type worldTPWithDimCommand struct {
	srv       *server.Server `cmd:"-"`
	Level     string
	Dimension string
}

func (c worldTPWithDimCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	dim, err := parseDimension(c.Dimension)
	if err != nil {
		output.Errorm(src, "%df.cmd.world.error.invalid_dimension", c.Dimension)
		return
	}
	teleportToLevel(c.srv, src, output, c.Level, dim)
}

// teleportToLevel is the shared teleport logic used by both /world tp leaves.
// It validates that the source is a player, resolves the target world for the
// given level name and dimension, and finally teleports the player to that
// world's spawn point.
func teleportToLevel(srv *server.Server, src cmd.Source, output *cmd.Output, levelName string, dim world.Dimension) {
	p, ok := src.(*player.Player)
	if !ok {
		output.Errorm(src, "%df.generic.console.only")
		return
	}

	target := srv.LevelManager().LookupWorld(levelName, dim)
	if target == nil {
		output.Errorm(src, "%df.cmd.world.error.not_found", levelName)
		return
	}

	if err := srv.TeleportToLevel(p, levelName, dim, target.Spawn().Vec3Centre().Add(mgl64.Vec3{0, 1.62})); err != nil {
		output.Errorm(src, "%df.cmd.world.error.teleport", err)
		return
	}
	output.Printm(src, "%df.cmd.world.tp.success", levelName, dim)
}

// worldCreateCommand creates a new empty level.
type worldCreateCommand struct {
	srv  *server.Server `cmd:"-"`
	Name string
}

func (c worldCreateCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if _, err := c.srv.LevelManager().LoadLevel(c.Name); err != nil {
		output.Errorm(src, "%df.cmd.world.error.create", c.Name, err)
		return
	}
	output.Printm(src, "%df.cmd.world.create.success", c.Name)
}

// worldUnloadCommand unloads a level after moving players out.
type worldUnloadCommand struct {
	srv  *server.Server `cmd:"-"`
	Name string
}

func (c worldUnloadCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	lvl := c.srv.LevelManager().Level(c.Name)
	if lvl == nil {
		output.Errorm(src, "%df.cmd.world.error.not_found", c.Name)
		return
	}

	defaultLevel := c.srv.LevelManager().DefaultLevel()
	if defaultLevel == nil {
		output.Errorm(src, "%df.cmd.world.error.no_default")
		return
	}
	spawn := defaultLevel.Overworld.Spawn().Vec3Centre().Add(mgl64.Vec3{0, 1.62})

	for p := range c.srv.Players(nil) {
		if p.Tx().World() == lvl.Overworld || p.Tx().World() == lvl.Nether || p.Tx().World() == lvl.End {
			_ = c.srv.TeleportToLevel(p, defaultLevel.Name, world.Overworld, spawn)
		}
	}

	if err := c.srv.LevelManager().UnloadLevel(c.Name); err != nil {
		output.Errorm(src, "%df.cmd.world.error.unload", c.Name, err)
		return
	}
	output.Printm(src, "%df.cmd.world.unload.success", c.Name)
}

// worldSetDefaultCommand sets the default level.
type worldSetDefaultCommand struct {
	srv  *server.Server `cmd:"-"`
	Name string
}

func (c worldSetDefaultCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if err := c.srv.LevelManager().SetDefaultLevel(c.Name); err != nil {
		output.Errorm(src, "%df.cmd.world.error.setdefault", c.Name, err)
		return
	}
	output.Printm(src, "%df.cmd.world.setdefault.success", c.Name)
}

// parseDimension converts a string to a world.Dimension.
func parseDimension(s string) (world.Dimension, error) {
	switch strings.ToLower(s) {
	case "overworld", "o", "0":
		return world.Overworld, nil
	case "nether", "n", "1":
		return world.Nether, nil
	case "end", "e", "2":
		return world.End, nil
	default:
		return nil, errors.New(i18n.R("%df.cmd.world.error.invalid_dimension", s))
	}
}
