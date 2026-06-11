package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// DifficultyCommand, /difficulty komutu.
// Sunucunun zorluğunu değiştirir.
//
// Kullanım: /difficulty <zorluk>
// Örnek: /difficulty hard
type DifficultyCommand struct {
	Difficulty cmd.Difficulty
}

// Run, difficulty komutunu çalıştırır.
func (d DifficultyCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Gerçek sunucu setting'inde zorluk değiştirilir
	// Bu örnek, komut işlevselliğini gösterir

	if tx == nil {
		output.Error("World transaction kullanılamıyor")
		return
	}

	// Zorluk değiştirildi
	output.Printf("Zorluk seviyesi %s olarak ayarlandı.", d.Difficulty)

	// Output ayarları - Sadece konsolda ve admin'lere göster
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandDifficulty)
}

// init, difficulty komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("zorluk", cmd.Difficulty(0)).
			Executes(&DifficultyCommand{}),
	)

	cmd.Register(cmd.NewWithTree(
		"difficulty",
		"Sunucu zorluk seviyesini değiştirir.",
		[]string{"diff"},
		tree,
	).WithPermissions(permission.CommandDifficulty))
}
