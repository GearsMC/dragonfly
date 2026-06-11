package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// resolvePlayers, []cmd.Target hedef listesini *player.Player listesine dönüştürür.
// isim eşleşmesi case-insensitive yapılır.
// tx.Players() iterator'ü üzerinden geçerli oyuncular taranır.
func resolvePlayers(tx *world.Tx, targets []cmd.Target) []*player.Player {
	if tx == nil || len(targets) == 0 {
		return nil
	}

	// İsim → oyuncu eşleşme tablosu
	playerMap := make(map[string]*player.Player)
	for e := range tx.Players() {
		if p, ok := e.(*player.Player); ok {
			playerMap[strings.ToLower(p.Name())] = p
		}
	}

	// Hedefleri çözümle
	var resolved []*player.Player
	for _, t := range targets {
		if named, ok := t.(cmd.NamedTarget); ok {
			if p, ok := playerMap[strings.ToLower(named.Name())]; ok {
				resolved = append(resolved, p)
			}
		}
	}
	return resolved
}
