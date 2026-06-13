package builtin

import (
	"strconv"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// TimeSetCommand, /time set <değer> komutu.
type TimeSetCommand struct {
	Value int32
}

func (c TimeSetCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	tx.World().SetTime(int(c.Value))
	output.Printt(i18n.T("%commands.time.set", 1), c.Value)
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// TimeAddCommand, /time add <değer> komutu.
type TimeAddCommand struct {
	Value int32
}

func (c TimeAddCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	current := tx.World().Time()
	tx.World().SetTime(current + int(c.Value))
	output.Printt(i18n.T("%commands.time.added", 1), c.Value)
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// TimeQueryCommand, /time query komutu.
type TimeQueryCommand struct{}

func (TimeQueryCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	t := tx.World().Time()
	day := t / 24000
	output.Printm(src, "%df.cmd.time.query", t, day, t%24000)
}

// TimeStartCommand, /time start komutu — zaman döngüsünü başlatır.
type TimeStartCommand struct{}

func (TimeStartCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	tx.World().StartTime()
	output.Printm(src, "%df.cmd.time.started")
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// TimeStopCommand, /time stop komutu — zaman döngüsünü durdurur.
type TimeStopCommand struct{}

func (TimeStopCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	tx.World().StopTime()
	output.Printm(src, "%df.cmd.time.stopped")
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// TimeDirectCommand, /time <değer> komutu — doğrudan preset veya sayı.
type TimeDirectCommand struct {
	Value string
}

func (c TimeDirectCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var t int
	switch c.Value {
	case "day":
		t = 1000
	case "night":
		t = 13000
	case "midnight":
		t = 18000
	case "noon":
		t = 6000
	case "sunrise":
		t = 23000
	case "sunset":
		t = 12000
	default:
		n, err := strconv.Atoi(c.Value)
		if err != nil {
			output.Errorm(src, "%df.cmd.time.error.value", c.Value)
			return
		}
		t = n
	}
	tx.World().SetTime(t)
	output.Printt(i18n.T("%commands.time.set", 1), t)
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// init, time komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("time", i18n.D("%df.cmd.time.description"),
		nil,
		cmd.NewCommandTree(
			cmd.Literal("set").Then(
				cmd.Argument("değer", int32(0)).Executes(&TimeSetCommand{}),
			),
			cmd.Literal("add").Then(
				cmd.Argument("değer", int32(0)).Executes(&TimeAddCommand{}),
			),
			cmd.Literal("query").Executes(&TimeQueryCommand{}),
			cmd.Literal("start").Executes(&TimeStartCommand{}),
			cmd.Literal("stop").Executes(&TimeStopCommand{}),
			cmd.Argument("değer", "").Executes(&TimeDirectCommand{}),
		),
	).WithPermissions(permission.CommandTime))
}
