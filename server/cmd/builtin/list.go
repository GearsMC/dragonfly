package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// ListCommand, /list komutu.
// Sunucudaki çevrimiçi oyuncuları listeler.
type ListCommand struct{}

// Run, /list komutunu çalıştırır.
func (ListCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	var names []string
	for e := range tx.Players() {
		if p, ok := e.(*player.Player); ok {
			names = append(names, p.Name())
		}
	}
	if len(names) == 0 {
		output.Printm(src, "%df.cmd.list.empty")
		return
	}
	output.Printm(src, "%df.cmd.list.format", len(names), strings.Join(names, ", "))
}

// init, list komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("list", i18n.D("%df.cmd.list.description"),
		[]string{"players", "online"},
		cmd.NewCommandTree(cmd.Root().WithPermissions(permission.CommandList).Executes(&ListCommand{})),
	))
}
