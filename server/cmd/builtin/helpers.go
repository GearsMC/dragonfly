package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
)

// resolvePlayers, []cmd.Target hedef listesini *player.Player listesine dönüştürür.
// cmd.Target'lar zaten world transaction'dan gelen entity referanslarıdır,
// bu yüzden direkt type-assert yeterlidir, isim eşleşmesine gerek yoktur.
func resolvePlayers(targets []cmd.Target) []*player.Player {
	var resolved []*player.Player
	for _, t := range targets {
		if p, ok := t.(*player.Player); ok {
			resolved = append(resolved, p)
		}
	}
	return resolved
}
