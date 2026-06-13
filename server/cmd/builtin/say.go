package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// SayCommand, /say komutu.
// Tüm oyunculara sunucu duyurusu mesajı gönderir.
type SayCommand struct {
	Message cmd.Varargs
}

// Run, /say komutunu çalıştırır.
func (s SayCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if named, ok := src.(cmd.NamedTarget); ok {
		output.Printm(src, "%df.cmd.say.format", named.Name(), s.Message)
	} else {
		output.Printm(src, "%df.cmd.say.format.anon", s.Message)
	}
}

// init, say komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("say", i18n.D("%df.cmd.say.description"),
		nil,
		cmd.NewCommandTree(cmd.GreedyText("mesaj").Executes(&SayCommand{})),
	).WithPermissions(permission.CommandSay))
}
