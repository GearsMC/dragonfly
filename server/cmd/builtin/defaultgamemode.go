package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// DefaultGameModeCommand, /defaultgamemode komutu.
// Sunucuya yeni katılan oyuncuların varsayılan oyun modunu değiştirir.
type DefaultGameModeCommand struct {
	Mode string
}

// Run, /defaultgamemode komutunu çalıştırır.
func (d DefaultGameModeCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	mode, err := cmd.ParseGameMode(src, d.Mode)
	if err != nil {
		output.Error(err)
		return
	}
	tx.World().SetDefaultGameMode(mode.ToWorldGameMode())
	output.Printt(i18n.T("%commands.defaultgamemode.success", 1), mode)
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandDefaultGameMode)
}

// init, defaultgamemode komutunu kaydeder.
func init() {
	cmd.Register(cmd.NewWithTree("defaultgamemode",
		i18n.D("%df.cmd.defaultgamemode.description"),
		[]string{"dgm"},
		cmd.NewCommandTree(
			cmd.Argument("mod", "", cmd.ArgumentSuggestions("GameMode", gamemodeSuggestions)).
				Executes(&DefaultGameModeCommand{}),
		),
	).WithPermissions(permission.CommandDefaultGameMode))
}
