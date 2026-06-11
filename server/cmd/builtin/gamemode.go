package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GameModeCommand, /gamemode komutu.
// Oyuncuların oyun modunu survival, creative, adventure veya spectator
// olarak değiştirir.
//
// Kullanım: /gamemode <mod> [oyuncular]
// Örnekler:
//   /gamemode creative          - kendi oyun modunu değiştir
//   /gamemode survival @a       - tüm oyuncuları survival yap
//   /gamemode adventure Steve   - Steve'i adventure yap
//   /gm c                       - alias ile creative
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
	players := resolvePlayers(tx, targets)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	// Her oyuncunun oyun modunu değiştir
	for _, p := range players {
		p.SetGameMode(worldMode)
	}

	// Başarı çıktısı
	if len(players) == 1 {
		output.Printf("%s oyuncusunun oyun modu %s olarak değiştirildi.", players[0].Name(), mode)
	} else {
		output.Printf("%d oyuncunun oyun modu %s olarak değiştirildi.", len(players), mode)
	}

	// Çıktı kapsamını ayarla - sadece gamemode izni olanlar görsün
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGameMode)
}

// init, gamemode komutunu kaydeder.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("mod", "").
			Then(
				cmd.Argument("oyuncular", []cmd.Target{}).Optional().
					Executes(&GameModeCommand{}),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"gamemode",
		"Oyuncu oyun modunu değiştirir.",
		[]string{"gm"},
		tree,
	).WithPermissions(permission.CommandGameMode))
}
