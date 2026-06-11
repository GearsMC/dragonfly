package permission

const (
	GroupUser     = "dfmc.group.user"
	GroupOperator = "dfmc.group.operator"
	GroupConsole  = "dfmc.group.console"

	// Rol şablonları - oyunculara bu roller atanabilir
	RoleBuilder   = "dfmc.role.builder"   // Harita yapımcısı
	RoleModerator = "dfmc.role.moderator" // Moderatör
	RoleHelper    = "dfmc.role.helper"    // Yardımcı

	CommandHelp             = "dfmc.command.help"
	CommandList             = "dfmc.command.list"
	CommandMe               = "dfmc.command.me"
	CommandVersion          = "dfmc.command.version"
	CommandStatus           = "dfmc.command.status"
	CommandTPS              = "dfmc.command.tps"
	CommandStop             = "dfmc.command.stop"
	CommandOP               = "dfmc.command.op"
	CommandDeOP             = "dfmc.command.deop"
	CommandBan              = "dfmc.command.ban"
	CommandBanIP            = "dfmc.command.ban_ip"
	CommandClear            = "dfmc.command.clear"
	CommandDefaultGameMode  = "dfmc.command.default_gamemode"
	CommandDifficulty       = "dfmc.command.difficulty"
	CommandEffect           = "dfmc.command.effect"
	CommandEnchant          = "dfmc.command.enchant"
	CommandExecute          = "dfmc.command.execute"
	CommandGameMode         = "dfmc.command.gamemode"
	CommandGameRule         = "dfmc.command.gamerule"
	CommandGive             = "dfmc.command.give"
	CommandKick             = "dfmc.command.kick"
	CommandPardon           = "dfmc.command.pardon"
	CommandPardonIP         = "dfmc.command.pardon_ip"
	CommandSay              = "dfmc.command.say"
	CommandSeed             = "dfmc.command.seed"
	CommandSetWorldSpawn    = "dfmc.command.setworldspawn"
	CommandSpawnPoint       = "dfmc.command.spawnpoint"
	CommandSummon           = "dfmc.command.summon"
	CommandTeleport         = "dfmc.command.teleport"
	CommandTime             = "dfmc.command.time"
	CommandTitle            = "dfmc.command.title"
	CommandWeather          = "dfmc.command.weather"
	CommandWhitelist        = "dfmc.command.whitelist"
	CommandXP               = "dfmc.command.xp"
	CommandViewOtherOutputs = "dfmc.command.view_other_outputs"

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
		CommandBan:                     true,
		CommandBanIP:                   true,
		CommandClear:                   true,
		CommandDefaultGameMode:         true,
		CommandDifficulty:              true,
		CommandEffect:                  true,
		CommandEnchant:                 true,
		CommandExecute:                 true,
		CommandGameMode:                true,
		CommandGameRule:                true,
		CommandGive:                    true,
		CommandKick:                    true,
		CommandPardon:                  true,
		CommandPardonIP:                true,
		CommandSay:                     true,
		CommandSeed:                    true,
		CommandSetWorldSpawn:           true,
		CommandSpawnPoint:              true,
		CommandSummon:                  true,
		CommandTeleport:                true,
		CommandTime:                    true,
		CommandTitle:                   true,
		CommandWeather:                 true,
		CommandWhitelist:               true,
		CommandXP:                      true,
		CommandViewOtherOutputs:        true,
		AbilityOperatorCommandQuickBar: true,
		BlockUseCommandBlock:           true,
	}))
	registry.Register(New(GroupConsole, "Konsol kaynaklarının tüm temel izinleri.").WithChildren(map[string]bool{
		GroupOperator: true,
	}))

	// Rol şablonları - Bu roller oyunculara opsiyonel olarak atanabilir
	// 1. BUILDER ROLE - Harita yapımcıları için
	registry.Register(New(RoleBuilder, "Harita yapımcısı rolü").WithChildren(map[string]bool{
		CommandSetWorldSpawn: true,
		CommandSpawnPoint:    true,
		CommandGameRule:      true,
		CommandExecute:       true,
		CommandSummon:        true,
		AbilityFlySurvival:   true,
		BlockUseCommandBlock: true,
	}))

	// 2. MODERATOR ROLE - Moderatörler için
	registry.Register(New(RoleModerator, "Moderatör rolü").WithChildren(map[string]bool{
		CommandKick:     true,
		CommandBan:      true,
		CommandBanIP:    true,
		CommandPardon:   true,
		CommandPardonIP: true,
		AbilityChat:     true,
	}))

	// 3. HELPER ROLE - Yardımcılar için
	registry.Register(New(RoleHelper, "Yardımcı rolü").WithChildren(map[string]bool{
		CommandHelp:       true,
		CommandList:       true,
		CommandSay:        true,
		CommandSpawnPoint: true,
		CommandTeleport:   true,
		AbilityChat:       true,
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
		New(CommandBan, "Oyuncu banlama izni."),
		New(CommandBanIP, "IP banlama izni."),
		New(CommandClear, "Envanter temizleme izni."),
		New(CommandDefaultGameMode, "Varsayılan oyun modunu değiştirme izni."),
		New(CommandDifficulty, "Zorluk ayarını değiştirme izni."),
		New(CommandEffect, "Efekt verme veya kaldırma izni."),
		New(CommandEnchant, "Eşya büyüleme izni."),
		New(CommandExecute, "Başka kaynak adına komut çalıştırma izni."),
		New(CommandGameMode, "Oyun modu değiştirme izni."),
		New(CommandGameRule, "Oyun kuralı değiştirme izni."),
		New(CommandGive, "Eşya verme izni."),
		New(CommandKick, "Oyuncu atma izni."),
		New(CommandPardon, "Oyuncu banını kaldırma izni."),
		New(CommandPardonIP, "IP banını kaldırma izni."),
		New(CommandSay, "Sunucu duyurusu gönderme izni."),
		New(CommandSeed, "Dünya seed bilgisini görme izni."),
		New(CommandSetWorldSpawn, "Dünya doğma noktasını değiştirme izni."),
		New(CommandSpawnPoint, "Oyuncu doğma noktasını değiştirme izni."),
		New(CommandSummon, "Varlık çağırma izni."),
		New(CommandTeleport, "Işınlanma komutu kullanma izni."),
		New(CommandTime, "Dünya zamanını değiştirme izni."),
		New(CommandTitle, "Title mesajı gönderme izni."),
		New(CommandWeather, "Hava durumunu değiştirme izni."),
		New(CommandWhitelist, "Whitelist yönetme izni."),
		New(CommandXP, "Deneyim verme veya alma izni."),
		New(CommandViewOtherOutputs, "Diğer command output mesajlarını görme izni."),
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
