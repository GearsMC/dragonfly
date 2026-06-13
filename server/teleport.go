package server

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TeleportToLevel, bir oyuncuyu belirtilen Level ve dimension'a ışınlar.
// levelName boş bırakılırsa varsayılan Level kullanılır. Hedef dimension World
// bulunamazsa hata döner.
func (srv *Server) TeleportToLevel(p *player.Player, levelName string, dim world.Dimension, pos mgl64.Vec3) error {
	target := srv.levelManager.LookupWorld(levelName, dim)
	if target == nil {
		if levelName == "" {
			levelName = srv.levelManager.DefaultLevel().Name
		}
		return fmt.Errorf("target world not found: level=%q dimension=%v", levelName, dim)
	}

	currentTx := p.Tx()
	if currentTx.World() == target {
		// Aynı dimension World içinde sadece pozisyon değiştir.
		p.Teleport(pos)
		return nil
	}

	// Farklı dimension World: entity'yi çıkar ve hedef dünyaya ekle.
	handle := currentTx.RemoveEntity(p)
	target.Exec(func(tx *world.Tx) {
		np := tx.AddEntity(handle).(*player.Player)
		np.Teleport(pos)
	})
	return nil
}
