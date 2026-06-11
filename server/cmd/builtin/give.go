package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GiveCommand, /give komutu.
// Oyunculara belirtilen eşyayı istenen miktar ve veri değeri ile verir.
// world.ItemByName() ile merkezi item registry'sinden eşya araması yapar.
//
// Kullanım: /give <oyuncu> <eşya> [miktar] [veri]
// Örnekler:
//
//	/give @p diamond 64          - en yakın oyuncuya 64 elmas
//	/give Steve apple 10         - Steve'e 10 elma
//	/give @a wool 32 5           - tüm oyunculara 32 mavi yün (data=5)
//	/give @s stick               - kendine 1 çubuk
type GiveCommand struct {
	Target []cmd.Target
	Item   string
	Amount cmd.Optional[int32]
	Data   cmd.Optional[int32]
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

	// Veri değerini al (varsayılan 0)
	data := int32(0)
	if d, ok := g.Data.Load(); ok {
		data = d
	}

	// Merkezi item registry'sinden eşyayı bul (meta verisi ile)
	itemType, ok := world.ItemByName(itemName, int16(data))
	if !ok {
		output.Errorf("Bilinmeyen eşya: %s", g.Item)
		return
	}

	// Miktarı belirle (varsayılan 1)
	amount := int32(1)
	if amt, ok := g.Amount.Load(); ok {
		if amt < 1 || amt > 32767 {
			output.Errorf("Miktar 1 ile 32767 arasında olmalıdır, alındı: %d", amt)
			return
		}
		amount = amt
	}

	// Eşyayı dağıt
	givenCount := 0
	for _, p := range players {
		remaining := int(amount)
		for remaining > 0 {
			// Her stack için 64'er olarak böl (eşya max stack boyutu)
			give := remaining
			if give > 64 {
				give = 64
			}
			remaining -= give

			stack := item.NewStack(itemType, give)
			if data > 0 {
				if durable, ok := itemType.(item.Durable); ok {
					maxDur := durable.DurabilityInfo().MaxDurability
					remaining := maxDur - int(data)
					if remaining < 1 {
						remaining = 1
					}
					stack = stack.WithDurability(remaining)
				}
			}
			if _, err := p.Inventory().AddItem(stack); err != nil {
				// Envanter doluysa hatayı bildir
				if remaining+give > 0 {
					output.Errorf("%s oyuncusuna eşya verilirken envanter dolu: %v", p.Name(), err)
				}
				break
			}
		}
		givenCount++
	}

	// Başarı çıktısı
	if data > 0 {
		if givenCount == 1 {
			output.Printf("%s oyuncusuna %d x %s (veri: %d) verildi.", players[0].Name(), amount, g.Item, data)
		} else {
			output.Printf("%d oyuncuya %d x %s (veri: %d) verildi.", givenCount, amount, g.Item, data)
		}
	} else {
		if givenCount == 1 {
			output.Printf("%s oyuncusuna %d x %s verildi.", players[0].Name(), amount, g.Item)
		} else {
			output.Printf("%d oyuncuya %d x %s verildi.", givenCount, amount, g.Item)
		}
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
				cmd.Argument("itemName", "", cmd.ArgumentSuggestions("Item", func(_ cmd.Source) []string {
					return nil
				})).
					Then(
						cmd.Argument("miktar", int32(1)).Optional().
							Then(
								cmd.Argument("veri", int32(0)).Optional().
									Executes(&GiveCommand{}),
							),
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
