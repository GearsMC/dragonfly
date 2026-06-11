package builtin

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GiveCommand, /give komutu.
// Oyunculara belirtilen eşyayı istenen miktarda verir.
//
// Kullanım: /give <oyuncu> <eşya> [miktar]
// Örnekler:
//   /give @p diamond 64        - en yakın oyuncuya 64 elmas
//   /give Steve apple 10       - Steve'e 10 elma
//   /give @a iron 1            - tüm oyunculara 1 demir
//   /give @s stick             - kendine 1 çubuk
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

	// Eşya ismini world.Item tipine dönüştür
	itemType := itemByName(g.Item)
	if itemType == nil {
		output.Errorf("Bilinmeyen eşya: %s", g.Item)
		return
	}

	// Miktarı belirle (varsayılan 1)
	amount := int32(1)
	if amt, ok := g.Amount.Load(); ok {
		if amt < 1 || amt > 64 {
			output.Errorf("Miktar 1 ile 64 arasında olmalıdır, alındı: %d", amt)
			return
		}
		amount = amt
	}

	// Eşya yığınını oluştur
	stack := item.NewStack(itemType, int(amount))

	// Her hedef oyuncuya eşyayı ver
	givenCount := 0
	for _, p := range players {
		n, err := p.Inventory().AddItem(stack)
		if err != nil {
			output.Errorf("%s oyuncusuna eşya verilirken hata: %v", p.Name(), err)
			continue
		}
		givenCount++
		_ = n
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
