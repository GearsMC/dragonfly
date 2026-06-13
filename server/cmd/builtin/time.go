package builtin

import (
	"strconv"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// parseTime, "day", "night" gibi ön ayarları veya düz sayıyı int'e çevirir.
// Hata durumunda output'a yazıp -1 döner.
func parseTime(src cmd.Source, output *cmd.Output, value string) int {
	switch value {
	case "day":
		return 1000
	case "night":
		return 13000
	case "midnight":
		return 18000
	case "noon":
		return 6000
	case "sunrise":
		return 23000
	case "sunset":
		return 12000
	default:
		n, err := strconv.Atoi(value)
		if err != nil {
			output.Errorm(src, "%df.cmd.time.error.value", value)
			return -1
		}
		return n
	}
}

// TimeSetCommand, /time set <değer> komutu.
type TimeSetCommand struct {
	Value string
}

func (c TimeSetCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	t := parseTime(src, output, c.Value)
	if t < 0 {
		return
	}
	tx.World().SetTime(t)
	output.Printt(i18n.T("%commands.time.set", 1), t)
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
	t := parseTime(src, output, c.Value)
	if t < 0 {
		return
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
