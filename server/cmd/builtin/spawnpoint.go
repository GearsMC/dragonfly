package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SpawnPointCommand, /spawnpoint komutu.
// Oyuncunun doğma noktasını ayarlar.
type SpawnPointCommand struct {
	Player   cmd.Optional[[]cmd.Target]
	Position cmd.Optional[mgl64.Vec3]
}

// Run, /spawnpoint komutunu çalıştırır.
func (s SpawnPointCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var targets []cmd.Target
	if p, ok := s.Player.Load(); ok {
		targets = p
	} else if t, ok := src.(cmd.Target); ok {
		targets = []cmd.Target{t}
	}
	if len(targets) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	for _, p := range players {
		if pos, ok := s.Position.Load(); ok {
			p.SetSpawnPosition(cubePosFromVec3(pos), tx.World())
			output.Printf("%s oyuncusunun doğma noktası ayarlandı.", p.Name())
		} else {
			p.SetSpawnPosition(cubePosFromVec3(p.Position()), tx.World())
			output.Printf("%s oyuncusunun doğma noktası bulunduğu konuma ayarlandı.", p.Name())
		}
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandSpawnPoint)
}

// init, spawnpoint komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("spawnpoint", "Doğma noktasını ayarlar.",
		nil,
		cmd.NewCommandTree(
			cmd.Argument("oyuncu", []cmd.Target{}).Optional().
				Then(
					cmd.Argument("konum", mgl64.Vec3{}).Optional().
						Executes(&SpawnPointCommand{}),
				),
		),
	).WithPermissions(permission.CommandSpawnPoint))
}
