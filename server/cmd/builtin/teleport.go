package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TeleportCommand, /teleport komutu.
// Oyuncuları bir konuma veya başka bir oyuncuya ışınlar.
type TeleportCommand struct {
	Target      cmd.Optional[[]cmd.Target]
	Destination cmd.Optional[mgl64.Vec3]
}

// Run, /teleport komutunu çalıştırır.
func (t TeleportCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Hedef ve konum belirle
	var targets []cmd.Target
	if p, ok := t.Target.Load(); ok && len(p) > 0 {
		targets = p
	} else if s, ok := src.(cmd.Target); ok {
		targets = []cmd.Target{s}
	}

	players := resolvePlayers(targets)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	// Konum belirle
	if pos, ok := t.Destination.Load(); ok {
		// Belirtilen pozisyona ışınla
		for _, p := range players {
			p.Teleport(pos)
		}
		output.Printf("%d oyuncu ışınlandı.", len(players))
	} else if s, ok := src.(cmd.Target); ok {
		// Konum verilmediyse kaynağın konumuna ışınla (kendi kendine = zaten orada)
		if len(players) == 1 && players[0].Position() == s.Position() {
			output.Error("Zaten bu konumdasın.")
			return
		}
		for _, p := range players {
			p.Teleport(s.Position())
		}
		output.Printf("%d oyuncu ışınlandı.", len(players))
	}

	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandTeleport)
}

// init, teleport komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("teleport", "Oyuncuları ışınlar.",
		[]string{"tp"},
		cmd.NewCommandTree(
			cmd.Argument("hedef", []cmd.Target{}).Optional().
				Then(
					cmd.Argument("konum", mgl64.Vec3{}).Optional().
						Executes(&TeleportCommand{}),
				),
		),
	).WithPermissions(permission.CommandTeleport))
}
