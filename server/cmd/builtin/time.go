package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// TimeCommand, /time komutu.
// Dünya zamanını ayarlar veya sorgular.
type TimeCommand struct {
	Action cmd.Varargs
}

// Run, /time komutunu çalıştırır.
func (t TimeCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	action := string(t.Action)

	// /time query → zamanı göster
	if action == "query" || action == "" {
		output.Printf("Sunucu zamanı: %d (gün: %d)", tx.World().Time(), tx.World().Time()%24000)
		return
	}

	// /time set <değer> → zamanı ayarla
	if action == "set" {
		output.Error("Kullanım: /time set <değer>")
		return
	}
	if action == "add" {
		output.Error("Kullanım: /time add <değer>")
		return
	}

	// /time <değer> → doğrudan ayarla
	// Gündüz/değer kontrolleri
	switch action {
	case "day", "1000":
		tx.World().SetTime(1000)
		output.Print("Zaman gündüz olarak ayarlandı.")
	case "night", "13000":
		tx.World().SetTime(13000)
		output.Print("Zaman gece olarak ayarlandı.")
	case "midnight", "18000":
		tx.World().SetTime(18000)
		output.Print("Zaman gece yarısı olarak ayarlandı.")
	case "noon", "6000":
		tx.World().SetTime(6000)
		output.Print("Zaman öğlen olarak ayarlandı.")
	case "sunrise", "23000":
		tx.World().SetTime(23000)
		output.Print("Zaman gün doğumu olarak ayarlandı.")
	case "sunset", "12000":
		tx.World().SetTime(12000)
		output.Print("Zaman gün batımı olarak ayarlandı.")
	default:
		output.Error("Geçersiz zaman değeri: day, night, noon, midnight, sunrise, sunset")
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTime)
}

// init, time komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("time", "Dünya zamanını değiştirir.",
		nil,
		cmd.NewCommandTree(
			cmd.GreedyText("eylem").Executes(&TimeCommand{}),
		),
	).WithPermissions(permission.CommandTime))
}
