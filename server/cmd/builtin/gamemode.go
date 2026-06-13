package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GameModeCommand, /gamemode komutu.
// Oyuncuların oyun modunu survival, creative, adventure veya spectator
// olarak değiştirir.
//
// Kullanım: /gamemode <mod> [oyuncular]
// Örnekler:
//
//	/gamemode creative          - kendi oyun modunu değiştir
//	/gamemode survival @a       - tüm oyuncuları survival yap
//	/gamemode adventure Steve   - Steve'i adventure yap
//	/gm c                       - alias ile creative
type GameModeCommand struct {
	Mode   string
	Target cmd.Optional[[]cmd.Target]
}

// Run, gamemode komutunu çalıştırır.
func (g GameModeCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// String oyun modunu cmd.GameMode'a dönüştür
	mode, err := cmd.ParseGameMode(src, g.Mode)
	if err != nil {
		output.Error(err)
		return
	}
	// cmd.GameMode'u world.GameMode interface'ine dönüştür
	worldMode := mode.ToWorldGameMode()

	// Hedef oyuncuları belirle
	targets, ok := g.Target.Load()
	if !ok || len(targets) == 0 {
		// Hedef belirtilmediyse komutu çalıştıran kaynağı hedef al
		if targeter, ok := src.(cmd.Target); ok {
			targets = []cmd.Target{targeter}
		}
	}
	if len(targets) == 0 {
		output.Errort(cmd.MessageNoTargets)
		return
	}

	// Hedefleri gerçek oyunculara çözümle
	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Errorm(src, "%df.generic.target.notfound")
		return
	}

	// Her oyuncunun oyun modunu değiştir
	for _, p := range players {
		p.SetGameMode(worldMode)
	}

	// Başarı çıktısı
	if len(players) == 1 {
		if players[0] == src {
			output.Printt(i18n.T("%commands.gamemode.success.self", 1), mode)
		} else {
			output.Printt(i18n.T("%commands.gamemode.success.other", 2), mode, players[0].Name())
		}
	} else {
		output.Printm(src, "%df.cmd.gamemode.success.multi", len(players), mode)
	}

	// Çıktı kapsamını ayarla - sadece gamemode izni olanlar görsün
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGameMode)
}

// gamemodeSuggestions, /gamemode komutunda client autocomplete'si için
// oyun modu seçeneklerini döndürür.
func gamemodeSuggestions(_ cmd.Source) []string {
	return []string{
		"0", "survival", "s",
		"1", "creative", "c",
		"2", "adventure", "a",
		"3", "spectator", "spc",
	}
}

// init, gamemode komutunu kaydeder.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("mod", "", cmd.ArgumentSuggestions("GameMode", gamemodeSuggestions)).
			Then(
				cmd.Argument("oyuncular", []cmd.Target{}).Optional().
					Executes(&GameModeCommand{}),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"gamemode",
		i18n.D("%df.cmd.gamemode.description"),
		[]string{"gm"},
		tree,
	).WithPermissions(permission.CommandGameMode))
}
