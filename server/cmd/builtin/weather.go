package builtin

import (
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// WeatherCommand, /weather komutu.
// Hava durumunu değiştirir.
type WeatherCommand struct {
	Type string
}

// Run, /weather komutunu çalıştırır.
func (w WeatherCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	switch w.Type {
	case "clear", "sunny":
		tx.World().StopRaining()
		tx.World().StopThundering()
		output.Print("Hava açık olarak ayarlandı.")
	case "rain":
		tx.World().StopThundering()
		tx.World().StartRaining(time.Hour)
		output.Print("Hava yağmurlu olarak ayarlandı.")
	case "thunder":
		tx.World().StartThundering(time.Hour)
		output.Print("Hava fırtınalı olarak ayarlandı.")
	default:
		output.Error("Geçersiz hava türü: clear, rain, thunder")
		return
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandWeather)
}

// init, weather komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("weather", "Hava durumunu değiştirir.",
		nil,
		cmd.NewCommandTree(
			cmd.Argument("tür", "").
				Executes(&WeatherCommand{}),
		),
	).WithPermissions(permission.CommandWeather))
}
