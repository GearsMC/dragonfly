package builtin

import (
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// WeatherClearCommand, /weather clear komutu.
type WeatherClearCommand struct{}

func (WeatherClearCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	tx.World().StopRaining()
	tx.World().StopThundering()
	output.Printt(i18n.S("%commands.weather.clear"))
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandWeather)
}

// WeatherRainCommand, /weather rain [süre] komutu.
type WeatherRainCommand struct {
	Duration cmd.Optional[int32]
}

func (c WeatherRainCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	dur := time.Hour
	if d, ok := c.Duration.Load(); ok && d > 0 {
		dur = time.Duration(d) * time.Second
	}
	tx.World().StopThundering()
	tx.World().StartRaining(dur)
	output.Printt(i18n.S("%commands.weather.rain"))
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandWeather)
}

// WeatherThunderCommand, /weather thunder [süre] komutu.
type WeatherThunderCommand struct {
	Duration cmd.Optional[int32]
}

func (c WeatherThunderCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	dur := time.Hour
	if d, ok := c.Duration.Load(); ok && d > 0 {
		dur = time.Duration(d) * time.Second
	}
	tx.World().StartThundering(dur)
	output.Printt(i18n.S("%commands.weather.thunder"))
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandWeather)
}

// WeatherQueryCommand, /weather query komutu.
type WeatherQueryCommand struct{}

func (WeatherQueryCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	raining := tx.Raining()
	thundering := tx.Thundering()
	switch {
	case thundering && raining:
		output.Printm(src, "%df.cmd.weather.query.thunder")
	case raining:
		output.Printm(src, "%df.cmd.weather.query.rain")
	default:
		output.Printm(src, "%df.cmd.weather.query.clear")
	}
}

// WeatherDirectCommand, /weather <tür> komutu (eski direkt kullanım).
type WeatherDirectCommand struct {
	Type string
}

func (c WeatherDirectCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	switch c.Type {
	case "clear", "sunny":
		tx.World().StopRaining()
		tx.World().StopThundering()
		output.Printt(i18n.S("%commands.weather.clear"))
	case "rain":
		tx.World().StopThundering()
		tx.World().StartRaining(time.Hour)
		output.Printt(i18n.S("%commands.weather.rain"))
	case "thunder":
		tx.World().StartThundering(time.Hour)
		output.Printt(i18n.S("%commands.weather.thunder"))
	default:
		output.Errorm(src, "%df.cmd.weather.error.type")
		return
	}
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandWeather)
}

// init, weather komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("weather", i18n.D("%df.cmd.weather.description"),
		nil,
		cmd.NewCommandTree(
			cmd.Literal("clear").Executes(&WeatherClearCommand{}),
			cmd.Literal("rain").Then(
				cmd.Argument("süre", int32(0)).Optional().Executes(&WeatherRainCommand{}),
			),
			cmd.Literal("thunder").Then(
				cmd.Argument("süre", int32(0)).Optional().Executes(&WeatherThunderCommand{}),
			),
			cmd.Literal("query").Executes(&WeatherQueryCommand{}),
			cmd.Argument("tür", "").Executes(&WeatherDirectCommand{}),
		),
	).WithPermissions(permission.CommandWeather))
}
