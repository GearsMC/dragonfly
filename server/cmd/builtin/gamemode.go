package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// Player, oyuncu entity interface'i
type Player interface {
	cmd.Target
	Name() string
	// SetGameMode oyuncunun oyun modunu değiştirir
	SetGameMode(mode cmd.GameMode)
}

// GameModeCommand, /gamemode komutu.
// Oyuncu veya oyuncu gruplarının oyun modunu değiştirir.
//
// Kullanım: /gamemode <mode> [oyuncular]
// Örnek: /gamemode creative @a
type GameModeCommand struct {
	Mode   cmd.GameMode
	Target cmd.Optional[[]cmd.Target]
}

// Run, gamemode komutunu çalıştırır.
func (g GameModeCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	targets, ok := g.Target.Load()
	if !ok {
		// Hedef belirtilmezse, kaynağa sadece kendisi hedef olabilir
		if targeter, ok := src.(cmd.Target); ok {
			targets = []cmd.Target{targeter}
		} else {
			output.Errort(cmd.MessageUnknown, "gamemode")
			return
		}
	}

	if len(targets) == 0 {
		output.Errort(cmd.MessageNoTargets)
		return
	}

	// Tüm hedeflerin oyun modunu değiştir
	for _, target := range targets {
		// Gerekse oyuncu türü kontrol et (geçerli bir entity hedefi olmalı)
		if player, ok := target.(Player); ok {
			player.SetGameMode(g.Mode)

			// Başarı mesajı
			output.Printf("Oyun modu %s için %s olarak değiştirildi.", player.Name(), g.Mode)
		} else {
			output.Printf("Geçerli olmayan hedef: %v", target)
		}
	}

	// Output ayarları
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGameMode)
}

// init, gamemode komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("mode", cmd.GameMode(0)).
			Then(
				cmd.Argument("oyuncular", []cmd.Target{}).
					Executes(&GameModeCommand{}),
			).
			Executes(&GameModeCommand{}),
	)

	cmd.Register(cmd.NewWithTree(
		"gamemode",
		"Oyuncu oyun modunu değiştirir.",
		[]string{"gm"},
		tree,
	).WithPermissions(permission.CommandGameMode))
}
