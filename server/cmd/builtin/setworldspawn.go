package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SetWorldSpawnCommand, /setworldspawn komutu.
// Dünyanın varsayılan doğma noktasını ayarlar.
type SetWorldSpawnCommand struct {
	Position cmd.Optional[mgl64.Vec3]
}

// Run, /setworldspawn komutunu çalıştırır.
func (s SetWorldSpawnCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	if pos, ok := s.Position.Load(); ok {
		tx.World().SetSpawn(cubePosFromVec3(pos))
		output.Printf("Dünya doğma noktası ayarlandı.")
	} else if t, ok := src.(cmd.Target); ok {
		tx.World().SetSpawn(cubePosFromVec3(t.Position()))
		output.Printf("Dünya doğma noktası bulunduğun konuma ayarlandı.")
	} else {
		output.Error("Konum belirlenemedi.")
		return
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandSetWorldSpawn)
}

// init, setworldspawn komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("setworldspawn", "Dünya doğma noktasını ayarlar.",
		nil,
		cmd.NewCommandTree(
			cmd.Argument("konum", mgl64.Vec3{}).Optional().
				Executes(&SetWorldSpawnCommand{}),
		),
	).WithPermissions(permission.CommandSetWorldSpawn))
}
