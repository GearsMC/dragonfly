package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GiveCommand, /give komutu.
// Oyunculara belirtilen eşyayı istenen miktarda verir.
// world.ItemByName() ile merkezi item registry'sinden eşya araması yapar.
//
// Kullanım: /give <oyuncu> <eşya> [miktar]
// Örnekler:
//
//	/give @p diamond 64        - en yakın oyuncuya 64 elmas
//	/give Steve apple 10       - Steve'e 10 elma
//	/give @a iron_ingot 1      - tüm oyunculara 1 demir
//	/give @s stick             - kendine 1 çubuk
type GiveCommand struct {
	Target []cmd.Target
	Item   string
	Amount cmd.Optional[int32]
}

// Run, give komutunu çalıştırır.
func (g GiveCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Hedef oyuncuları çözümle
	players := resolvePlayers(tx, g.Target)
	if len(players) == 0 {
		output.Error("Hedef oyuncu bulunamadı.")
		return
	}

	// Eşya adını normalize et: "minecraft:" ön eki yoksa ekle
	itemName := g.Item
	if !strings.Contains(itemName, ":") {
		itemName = "minecraft:" + itemName
	}

	// Merkezi item registry'sinden eşyayı bul
	itemType, ok := world.ItemByName(itemName, 0)
	if !ok {
		output.Errorf("Bilinmeyen eşya: %s", g.Item)
		return
	}

	// Miktarı belirle (varsayılan 1)
	amount := int32(1)
	if amt, ok := g.Amount.Load(); ok {
		if amt < 1 || amt > 1024 {
			output.Errorf("Miktar 1 ile 1024 arasında olmalıdır, alındı: %d", amt)
			return
		}
		amount = amt
	}

	// Eşya yığınını oluştur ve dağıt
	givenCount := 0
	for _, p := range players {
		remaining := int(amount)
		for remaining > 0 {
			// En fazla stack boyutu kadar ver (genelde 64)
			give := remaining
			if give > 64 {
				give = 64
			}
			remaining -= give

			stack := item.NewStack(itemType, give)
			if _, err := p.Inventory().AddItem(stack); err != nil {
				output.Errorf("%s oyuncusuna eşya verilirken hata: %v", p.Name(), err)
				continue
			}
		}
		givenCount++
	}

	// Başarı çıktısı
	if givenCount == 1 {
		output.Printf("%s oyuncusuna %d x %s verildi.", players[0].Name(), amount, g.Item)
	} else {
		output.Printf("%d oyuncuya %d x %s verildi.", givenCount, amount, g.Item)
	}

	// Çıktı kapsamını ayarla
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGive)
}

// init, give komutunu kaydeder.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("oyuncu", []cmd.Target{}).
			Then(
				cmd.Argument("eşya", "").
					Then(
						cmd.Argument("miktar", int32(1)).Optional().
							Executes(&GiveCommand{}),
					),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"give",
		"Oyuncuya eşya verir.",
		nil,
		tree,
	).WithPermissions(permission.CommandGive))
}
