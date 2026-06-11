package cmd

// SenderType, komut kaynağının tipini temsil eder.
// Komut yürütücüsünün hangi türden kaynaklar tarafından çalıştırılabileceğini
// belirtmek için kullanılır.
type SenderType uint8

const (
	// SenderTypeAny, herhangi bir kaynaktan komutu yürütülebilir anlamına gelir.
	SenderTypeAny SenderType = 0

	// SenderTypeServer, komutu sadece sunucu konsolundan yürütülebilir anlamına gelir.
	// Admin-only işlemler için kullanılır.
	SenderTypeServer SenderType = 1 << iota

	// SenderTypePlayer, komutu oyunculardan yürütülebilir anlamına gelir.
	// Hem gerçek oyuncu hem de sahte oyuncu (NPC) dahil.
	SenderTypePlayer

	// SenderTypeActualPlayer, komutu sadece gerçek oyunculardan yürütülebilir anlamına gelir.
	// Komut blokları veya NPC'ler tarafından çalıştırılamaz.
	SenderTypeActualPlayer

	// SenderTypeEntity, komutu herhangi bir entity'den yürütülebilir anlamına gelir.
	SenderTypeEntity

	// SenderTypeConsole, komutu konsoldan yürütülebilir anlamına gelir.
	// Server + Console konsolu içerir.
	SenderTypeConsole = SenderTypeServer
)

// Source, komut yürütülmesinin kaynağını temsil eder. Komutlar,
// Allower interface'ini uygulayarak komutu çalıştırabilen kaynakları sınırlandırabilirler.
// Source, Target interface'ini uygular. Bir Source her zaman kendisini hedefleyebilmelidir.
type Source interface {
	Target
	// SendCommandOutput, komut çıktısını kaynağa gönderir. Çıktının nasıl uygulanacağı
	// kaynağın ne türden olduğuna bağlıdır.
	// SendCommandOutput, bir Command tarafından çalıştıktan otomatik olarak çağrılır.
	SendCommandOutput(o *Output)
}

// ConsoleSource, sunucu konsolunu temsil eden bir Source tarafından uygulanabilir.
// Konsol kaynakları başında slash olmadan komutları yürütebilir.
type ConsoleSource interface {
	Source
	Console() bool
}

// PermissionSource, komut izinlerini değerlendirebilen bir Source tarafından uygulanabilir.
// Bu interface'i uygulamayan kaynaklar sadece açık izni olmayan komutları yürütebilir.
type PermissionSource interface {
	Source
	HasCommandPermission(permission string) bool
}

// SenderTypeSource, komut kaynağının tipini bildirebilen bir Source tarafından uygulanabilir.
// Bu interface'i uygulamayan kaynaklar SenderTypeAny olarak kabul edilir.
type SenderTypeSource interface {
	Source
	// SenderTypeOf, bu kaynağın tipini döndürür.
	SenderTypeOf() SenderType
}
