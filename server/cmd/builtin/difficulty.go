package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// DifficultyCommand, /difficulty komutu.
type DifficultyCommand struct {
	Difficulty string
}

// Run, difficulty komutunu çalıştırır.
func (d DifficultyCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Zorluk parse et
	diff, err := cmd.ParseDifficulty(src, d.Difficulty)
	if err != nil {
		output.Error(err)
		return
	}

	if tx == nil {
		output.Error("World transaction kullanılamıyor")
		return
	}

	output.Printf("Zorluk seviyesi %s olarak ayarlandı.", diff)

	// Output ayarları
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandDifficulty)
}

// init, difficulty komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("zorluk", "").
			Executes(&DifficultyCommand{}),
	)

	cmd.Register(cmd.NewWithTree(
		"difficulty",
		"Sunucu zorluk seviyesini değiştirir.",
		[]string{"diff"},
		tree,
	).WithPermissions(permission.CommandDifficulty))
}
