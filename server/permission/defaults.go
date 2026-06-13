package permission

const (
	GroupUser     = "dfmc.group.user"
	GroupOperator = "dfmc.group.operator"
	GroupConsole  = "dfmc.group.console"

	CommandList            = "dfmc.command.list"
	CommandMe              = "dfmc.command.me"
	CommandVersion         = "dfmc.command.version"
	CommandWorld           = "dfmc.command.world"
	CommandStatus          = "dfmc.command.status"
	CommandTPS             = "dfmc.command.tps"
	CommandStop            = "dfmc.command.stop"
	CommandOP              = "dfmc.command.op"
	CommandDeOP            = "dfmc.command.deop"
	CommandClear           = "dfmc.command.clear"
	CommandDefaultGameMode = "dfmc.command.default_gamemode"
	CommandDifficulty      = "dfmc.command.difficulty"
	CommandEffect          = "dfmc.command.effect"
	CommandEnchant         = "dfmc.command.enchant"
	CommandGameMode        = "dfmc.command.gamemode"
	CommandGive            = "dfmc.command.give"
	CommandKick            = "dfmc.command.kick"
	CommandSay             = "dfmc.command.say"
	CommandSeed            = "dfmc.command.seed"
	CommandSetWorldSpawn   = "dfmc.command.setworldspawn"
	CommandSpawnPoint      = "dfmc.command.spawnpoint"
	CommandTeleport        = "dfmc.command.teleport"
	CommandTime            = "dfmc.command.time"
	CommandWeather         = "dfmc.command.weather"
	CommandWhitelist       = "dfmc.command.whitelist"
	CommandXP              = "dfmc.command.xp"

	AbilityChat                    = "dfmc.ability.chat"
	AbilityOperatorCommandQuickBar = "dfmc.ability.operator_command_quick_bar"
	AbilityFlySurvival             = "dfmc.ability.fly.survival"
	AbilityFlyCreative             = "dfmc.ability.fly.creative"
	AbilityFlyAdventure            = "dfmc.ability.fly.adventure"

	BlockUseCommandBlock = "dfmc.block.command_block.use"
)

// RegisterDefaults, DFMC'nin temel kullanıcı ve operatör permission ağacını registry'ye ekler.
func RegisterDefaults(registry *Registry) {
	registry.Register(New(GroupUser, "Varsayılan oyuncu izinleri.").WithChildren(map[string]bool{
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
		CommandClear:                   true,
		CommandDefaultGameMode:         true,
		CommandDifficulty:              true,
		CommandEffect:                  true,
		CommandEnchant:                 true,
		CommandGameMode:                true,
		CommandGive:                    true,
		CommandKick:                    true,
		CommandSay:                     true,
		CommandSeed:                    true,
		CommandSetWorldSpawn:           true,
		CommandSpawnPoint:              true,
		CommandTeleport:                true,
		CommandTime:                    true,
		CommandWeather:                 true,
		CommandWhitelist:               true,
		CommandWorld:                   true,
		CommandXP:                      true,
		AbilityOperatorCommandQuickBar: true,
		AbilityFlySurvival:             true,
		AbilityFlyCreative:             true,
		AbilityFlyAdventure:            true,
		BlockUseCommandBlock:           true,
	}))
	registry.Register(New(GroupConsole, "Konsol kaynaklarının tüm temel izinleri.").WithChildren(map[string]bool{
		GroupOperator: true,
	}))

	for _, permission := range []Permission{
		New(CommandList, "Oyuncu listesini görme izni."),
		New(CommandMe, "Me komutunu kullanma izni."),
		New(CommandVersion, "Sürüm bilgisini görme izni."),
		New(CommandStatus, "Sunucu durum bilgisini görme izni."),
		New(CommandTPS, "TPS bilgisini görme izni."),
		New(CommandStop, "Sunucuyu durdurma izni."),
		New(CommandOP, "Oyuncuya operatör yetkisi verme izni."),
		New(CommandDeOP, "Oyuncudan operatör yetkisi alma izni."),
		New(CommandClear, "Envanter temizleme izni."),
		New(CommandDefaultGameMode, "Varsayılan oyun modunu değiştirme izni."),
		New(CommandDifficulty, "Zorluk ayarını değiştirme izni."),
		New(CommandEffect, "Efekt verme veya kaldırma izni."),
		New(CommandEnchant, "Eşya büyüleme izni."),
		New(CommandGameMode, "Oyun modu değiştirme izni."),
		New(CommandGive, "Eşya verme izni."),
		New(CommandKick, "Oyuncu atma izni."),
		New(CommandSay, "Sunucu duyurusu gönderme izni."),
		New(CommandSeed, "Dünya seed bilgisini görme izni."),
		New(CommandSetWorldSpawn, "Dünya doğma noktasını değiştirme izni."),
		New(CommandSpawnPoint, "Oyuncu doğma noktasını değiştirme izni."),
		New(CommandTeleport, "Işınlanma komutu kullanma izni."),
		New(CommandTime, "Dünya zamanını değiştirme izni."),
		New(CommandWeather, "Hava durumunu değiştirme izni."),
		New(CommandWhitelist, "Whitelist yönetme izni."),
		New(CommandWorld, "Çoklu dünya komutunu kullanma izni."),
		New(CommandXP, "Deneyim verme veya alma izni."),
		New(AbilityChat, "Sohbet gönderme izni."),
		New(AbilityOperatorCommandQuickBar, "Operatör komut hızlı erişim çubuğunu kullanma izni."),
		New(AbilityFlySurvival, "Survival modda uçma izni."),
		New(AbilityFlyCreative, "Creative modda uçma izni."),
		New(AbilityFlyAdventure, "Adventure modda uçma izni."),
		New(BlockUseCommandBlock, "Komut bloğu kullanma izni."),
	} {
		registry.Register(permission)
	}
}
