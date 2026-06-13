package cmd

import (
	"errors"
	"fmt"

	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/player/chat"
)

// BroadcastScope, komut çıktısının yayınlanma kapsamını belirler.
// Hangiplere hangi çıktıların gösterilmesi gerektiğini kontrol eder.
type BroadcastScope string

const (
	// BroadcastConsole, çıktının sadece admin konsoluna gösterilmesini belirtir.
	// Komut çalıştıran oyunculara gösterilmez.
	BroadcastConsole BroadcastScope = "console"

	// BroadcastRequester, çıktının sadece komutu çalıştıran kaynağa gösterilmesini belirtir.
	// Hiç kimseye broadcast edilmez.
	BroadcastRequester BroadcastScope = "requester"

	// BroadcastPermitted, çıktının belirtilen izinlere sahip kişilere gösterilmesini belirtir.
	// RequiredPermissions alanı dikkate alınır.
	BroadcastPermitted BroadcastScope = "permitted"

	// BroadcastAll, çıktının herkese (tüm izin seviyeleri) gösterilmesini belirtir.
	// RequiredPermissions dikkate alınmaz.
	BroadcastAll BroadcastScope = "all"
)

// Output, komut yürütülmesinin çıktısını tutar. Başarı mesajları ve hata mesajları
// içerir ve bunlar komutu çalıştıran kaynağa gönderilir.
// YENİ: Broadcast kapsamı ve gerekli izinler desteklenir.
type Output struct {
	errors              []error
	messages            []fmt.Stringer
	broadcastScope      BroadcastScope // Çıktının nereye yayınlanması gerektiği
	requiredPermissions []string       // BroadcastPermitted olduğunda gerekli izinler
}

// Errorf, hata mesajını formatlar ve komut çıktısına ekler.
func (o *Output) Errorf(format string, a ...any) {
	o.errors = append(o.errors, fmt.Errorf(format, a...))
}

// Error, hata mesajını formatlar ve komut çıktısına ekler.
func (o *Output) Error(a ...any) {
	if len(a) == 1 {
		if err, ok := a[0].(error); ok {
			o.errors = append(o.errors, err)
			return
		}
	}
	o.errors = append(o.errors, errors.New(fmt.Sprint(a...)))
}

// Errort, çevirili bir hata mesajı ekler ve fonksiyon argümanları ile parametrelendirir.
// Argüman sayısı yanlışsa Errort panik yapar.
func (o *Output) Errort(t chat.Translation, a ...any) {
	o.errors = append(o.errors, t.F(a...))
}

// Printf, (başarı) mesajını formatlar ve komut çıktısına ekler.
func (o *Output) Printf(format string, a ...any) {
	o.messages = append(o.messages, stringer(fmt.Sprintf(format, a...)))
}

// Print, (başarı) mesajını formatlar ve komut çıktısına ekler.
func (o *Output) Print(a ...any) {
	o.messages = append(o.messages, stringer(fmt.Sprint(a...)))
}

// Printt, çevirili bir (başarı) mesajı ekler ve fonksiyon argümanları ile parametrelendirir.
// Argüman sayısı yanlışsa Printt panik yapar.
func (o *Output) Printt(t chat.Translation, a ...any) {
	o.messages = append(o.messages, t.F(a...))
}

// Printm, custom i18n key'i kaynağın locale'ine göre çözümler ve çıktıya ekler.
func (o *Output) Printm(src any, key string, args ...any) {
	o.messages = append(o.messages, stringer(i18n.M(src, key, args...)))
}

// Errorm, custom i18n key'ini hata olarak çıktıya ekler.
func (o *Output) Errorm(src any, key string, args ...any) {
	o.errors = append(o.errors, errors.New(i18n.M(src, key, args...)))
}

// Errors, komut çıktısına eklenen tüm hataları döndürür. Genellikle
// sadece bir hata mesajı ayarlanır: Bir hata mesajından sonra,
// bir komutun yürütülmesi tipik olarak sonlanır.
func (o *Output) Errors() []error {
	return o.errors
}

// ErrorCount, komut çıktısının sahip olduğu hata sayısını döndürür.
func (o *Output) ErrorCount() int {
	return len(o.errors)
}

// Messages, komut çıktısına eklenen tüm mesajları döndürür. Mevcut
// mesaj miktarı çağrılan komuta bağlıdır.
func (o *Output) Messages() []fmt.Stringer {
	return o.messages
}

// MessageCount, komut çıktısının sahip olduğu (başarı) mesaj sayısını döndürür.
func (o *Output) MessageCount() int {
	return len(o.messages)
}

// SetBroadcastScope, bu çıktının yayınlanacağı kapsamı ayarlar.
// Varsayılan değer BroadcastAll'dur.
func (o *Output) SetBroadcastScope(scope BroadcastScope) *Output {
	o.broadcastScope = scope
	return o
}

// BroadcastScope, bu çıktının yayınlanacağı kapsamı döndürür.
func (o *Output) BroadcastScope() BroadcastScope {
	if o.broadcastScope == "" {
		return BroadcastAll
	}
	return o.broadcastScope
}

// SetRequiredPermissions, BroadcastPermitted kapsamı kullanıldığında
// bu çıktıyı görmek için gerekli izinleri ayarlar.
func (o *Output) SetRequiredPermissions(permissions ...string) *Output {
	o.requiredPermissions = make([]string, len(permissions))
	copy(o.requiredPermissions, permissions)
	return o
}

// RequiredPermissions, bu çıktıyı görmek için gerekli izinleri döndürür.
func (o *Output) RequiredPermissions() []string {
	return o.requiredPermissions
}

type stringer string

func (s stringer) String() string { return string(s) }

// Genel komut mesajlari. Vanilla key'ler client tarafindan çevrilir.
var (
	MessageSyntax           = i18n.T("%commands.generic.syntax", 3)
	MessageUsage            = i18n.T("%commands.generic.usage", 1)
	MessageUnknown          = i18n.T("%commands.generic.unknown", 1)
	MessageNoTargets        = i18n.T("%commands.generic.noTargetMatch", 0)
	MessageNumberInvalid    = i18n.T("%commands.generic.num.invalid", 1)
	MessageBooleanInvalid   = i18n.T("%commands.generic.boolean.invalid", 1)
	MessagePlayerNotFound   = i18n.T("%commands.generic.player.notFound", 0)
	MessageParameterInvalid = i18n.T("%commands.generic.parameter.invalid", 1)
)
