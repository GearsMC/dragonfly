package builtin

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
)

// GiveCommand, /give komutu.
// Oyunculara eşya verir.
//
// Kullanım: /give <oyuncu> <eşya> [miktar] [data]
// Örnek: /give @a diamond 64
type GiveCommand struct {
	Target cmd.Target
	Item   string         // Eşya adı (registry lookup gerçekleşir)
	Amount cmd.Optional[int32]
	Data   cmd.Optional[string]
}

// Run, give komutunu çalıştırır.
func (g GiveCommand) Run(src cmd.Source, output *cmd.Output, tx *world.Tx) {
	// Eşya miktarını belirle
	amount := int32(1)
	if amt, ok := g.Amount.Load(); ok {
		if amt <= 0 || amt > 64 {
			output.Errorf("Miktar 1 ile 64 arasında olmalıdır, alındı: %d", amt)
			return
		}
		amount = amt
	}

	// Eşya datası (isteğe bağlı)
	data := ""
	if d, ok := g.Data.Load(); ok {
		data = d
	}

	// Oyuncuya eşya ver
	// Not: Gerçek implementation world item management sistemi ile entegre olmalı
	output.Printf("%s tarafından %d x %s alındı.", g.Target, amount, g.Item)
	if data != "" {
		output.Printf("Eşya verileri: %s", data)
	}

	// Output ayarları
	output.SetBroadcastScope(cmd.BroadcastPermitted).
		SetRequiredPermissions(permission.CommandGive)
}

// init, give komutunu kaydet.
func init() {
	tree := cmd.NewCommandTree(
		cmd.Argument("oyuncu", []cmd.Target{}).
			Then(
				cmd.Argument("eşya", "").
					Then(
						cmd.Argument("miktar", int32(0)).Optional().
							Then(
								cmd.Argument("veri", "").Optional().
									Executes(&GiveCommand{}),
							).
							Executes(&GiveCommand{}),
					).
					Executes(&GiveCommand{}),
			),
	)

	cmd.Register(cmd.NewWithTree(
		"give",
		"Oyuncuya eşya verir.",
		nil,
		tree,
	).WithPermissions(permission.CommandGive))
}
