package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// DifficultyCommand, /difficulty komutu.
// Sunucunun zorluk seviyesini peaceful, easy, normal veya hard olarak değiştirir.
//
// Kullanım: /difficulty <zorluk>
// Örnekler:
//
//	/difficulty hard      - zorluğu hard yap
//	/difficulty peaceful  - zorluğu peaceful yap
//	/diff e               - alias ile easy
type DifficultyCommand struct {
	Difficulty string
}

// Run, difficulty komutunu çalıştırır.
func (d DifficultyCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// String zorluk değerini cmd.Difficulty'ye dönüştür
	diff, err := cmd.ParseDifficulty(src, d.Difficulty)
	if err != nil {
		output.Error(err)
		return
	}

	// cmd.Difficulty'yi world.Difficulty interface'ine dönüştür
	worldDiff := diff.ToWorldDifficulty()

	// World transaction üzerinden zorluğu ayarla
	if tx != nil {
		tx.World().SetDifficulty(worldDiff)
	} else {
		output.Error("Dünya işlemi kullanılamıyor.")
		return
	}

	// Başarı çıktısı
	output.Printf("Zorluk seviyesi %s olarak ayarlandı.", diff)

	// Çıktı kapsamını ayarla - sadece admin izni olanlar görsün
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandDifficulty)
}

// difficultySuggestions, /difficulty komutunda client autocomplete'si için
// zorluk seviyesi seçeneklerini döndürür.
func difficultySuggestions(_ cmd.Source) []string {
	return []string{
		"0", "peaceful", "p",
		"1", "easy", "e",
		"2", "normal", "n",
		"3", "hard", "h",
	}
}

// init, difficulty komutunu kaydeder.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("zorluk", "", cmd.ArgumentSuggestions("Difficulty", difficultySuggestions)).
			Executes(&DifficultyCommand{}),
	)

	cmd.Register(cmd.NewWithTree(
		"difficulty",
		"Sunucu zorluk seviyesini değiştirir.",
		[]string{"diff"},
		tree,
	).WithPermissions(permission.CommandDifficulty))
}
