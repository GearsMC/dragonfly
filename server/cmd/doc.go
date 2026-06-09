// Package cmd, Minecraft komutlarını root/literal/argument node'larından oluşan bir command tree ile yönetir.
// Node seviyesinde permission verilebilir; client'a gönderilen AvailableCommands paketi ve runtime execution aynı
// tree leaf'lerinden beslendiği için görünen komut ile çalıştırılabilen komut aynı permission kararını kullanır.
//
// Yeni komutlar cmd.NewWithTree(), cmd.Root(), cmd.Literal() ve cmd.Argument() ile açık tree olarak tanımlanabilir.
// cmd.New() ise struct tabanlı eski tanımı otomatik olarak aynı command tree modeline çevirir. Runnable struct içindeki
// exported alanlar sırayla argument olarak parse edilir; SubCommand alanları literal node'a çevrilir. Export edilmeyen
// veya `cmd:"-"` tag'iyle yok sayılan alanlar parametre olarak parse edilmez.
//
// A Runnable may have exported fields only of the following types:
// int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint,
// float32, float64, string, bool, mgl64.Vec3, Varargs, []Target, cmd.SubCommand, Optional[T] (to make a parameter
// optional), or a type that implements the cmd.Parameter or cmd.Enum interface. cmd.Enum implementations must be of the
// type string.
// Fields in the Runnable struct may have `cmd:` struct tag to specify the name and suffix of a parameter as such:
//
//	type T struct {
//	    Param int `cmd:"name,suffix"`
//	}
//
// If no name is set, the field name is used. Additionally, the name as specified in the struct tag may be '-' to make
// the parser ignore the field. In this case, the field does not have to be of one of the types above.
//
// Commands may be registered using the cmd.Register() method. By itself, this method will not ensure that the
// client will be able to use the command: The user of the cmd package must handle commands itself and run the
// appropriate one using the cmd.ByAlias function.
package cmd
