package permission

const (
	GroupUser     = "dfmc.group.user"
	GroupOperator = "dfmc.group.operator"
	GroupConsole  = "dfmc.group.console"

	CommandHelp    = "dfmc.command.help"
	CommandList    = "dfmc.command.list"
	CommandMe      = "dfmc.command.me"
	CommandVersion = "dfmc.command.version"
	CommandStatus  = "dfmc.command.status"
	CommandTPS     = "dfmc.command.tps"
	CommandStop    = "dfmc.command.stop"
	CommandOP      = "dfmc.command.op"
	CommandDeOP    = "dfmc.command.deop"

	AbilityChat                    = "dfmc.ability.chat"
	AbilityOperatorCommandQuickBar = "dfmc.ability.operator_command_quick_bar"
)

// RegisterDefaults, DFMC'nin temel kullanıcı ve operatör permission ağacını registry'ye ekler.
func RegisterDefaults(registry *Registry) {
	registry.Register(New(GroupUser, "Varsayılan oyuncu izinleri.").WithChildren(map[string]bool{
		CommandHelp:    true,
		CommandList:    true,
		CommandMe:      true,
		CommandVersion: true,
		AbilityChat:    true,
	}))
	registry.Register(New(GroupOperator, "Sunucu operatörü izinleri.").WithChildren(map[string]bool{
		GroupUser:                      true,
		CommandStatus:                  true,
		CommandTPS:                     true,
		CommandStop:                    true,
		CommandOP:                      true,
		CommandDeOP:                    true,
		AbilityOperatorCommandQuickBar: true,
	}))
	registry.Register(New(GroupConsole, "Konsol kaynaklarının tüm temel izinleri.").WithChildren(map[string]bool{
		GroupOperator: true,
	}))
	for _, permission := range []Permission{
		New(CommandHelp, "Yardım komutunu kullanma izni."),
		New(CommandList, "Oyuncu listesini görme izni."),
		New(CommandMe, "Me komutunu kullanma izni."),
		New(CommandVersion, "Sürüm bilgisini görme izni."),
		New(CommandStatus, "Sunucu durum bilgisini görme izni."),
		New(CommandTPS, "TPS bilgisini görme izni."),
		New(CommandStop, "Sunucuyu durdurma izni."),
		New(CommandOP, "Oyuncuya operatör yetkisi verme izni."),
		New(CommandDeOP, "Oyuncudan operatör yetkisi alma izni."),
		New(AbilityChat, "Sohbet gönderme izni."),
		New(AbilityOperatorCommandQuickBar, "Operatör komut hızlı erişim çubuğunu kullanma izni."),
	} {
		registry.Register(permission)
	}
}
