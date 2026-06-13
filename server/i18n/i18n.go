package i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"golang.org/x/text/language"
)

// orderedPlaceholderRe, format string içindeki %1, %2 gibi sıralı placeholder'ları bulur.
var orderedPlaceholderRe = regexp.MustCompile(`%([0-9]+)`)

//go:embed embedded/*.toml
var embeddedFiles embed.FS

var (
	defaultRegistry = New()
	defaultLang     = language.Turkish
	defaultLangMu   sync.RWMutex
)

func init() {
	entries, err := embeddedFiles.ReadDir("embedded")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".toml") {
			continue
		}
		data, err := embeddedFiles.ReadFile("embedded/" + name)
		if err != nil {
			continue
		}
		var m map[string]string
		if err := toml.Unmarshal(data, &m); err != nil {
			continue
		}
		// *_<grup>.toml dosyaları: ilk alt çizgiden sonrasını atıp kalanını
		// dil kodu olarak yorumluyoruz (örn. tr_internal, tr_world -> tr).
		base := strings.TrimSuffix(name, ".toml")
		if i := strings.Index(base, "_"); i != -1 {
			base = base[:i]
		}
		lang, err := language.Parse(base)
		if err != nil {
			continue
		}
		defaultRegistry.Register(lang, m)
	}
}

// Registry, dil çevirilerini saklar ve çözümler.
type Registry struct {
	mu        sync.RWMutex
	data      map[language.Tag]map[string]string
	supported []language.Tag
	matcher   language.Matcher
}

// New boş bir Registry oluşturur.
func New() *Registry {
	r := &Registry{data: make(map[language.Tag]map[string]string)}
	r.rebuildMatcher()
	return r
}

// SetDefault, sunucu varsayılan dilini ayarlar.
// Console gibi locale bilgisi olmayan kaynaklar bu dili kullanır.
func SetDefault(lang language.Tag) {
	defaultLangMu.Lock()
	defaultLang = lang
	defaultLangMu.Unlock()
}

// Default, şu anki varsayılan dili döndürür.
func Default() language.Tag {
	defaultLangMu.RLock()
	defer defaultLangMu.RUnlock()
	return defaultLang
}

// Register, bir dil için çeviri girişleri ekler veya günceller.
func (r *Registry) Register(lang language.Tag, entries map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.data[lang]
	if !ok {
		m = make(map[string]string)
		r.data[lang] = m
	}
	for k, v := range entries {
		m[k] = v
	}
	r.rebuildMatcher()
}

// Register, varsayılan registry'ye çeviri ekler.
func Register(lang language.Tag, entries map[string]string) {
	defaultRegistry.Register(lang, entries)
}

// Load, bir TOML dosyasını yükler. Dosya adı locale'i belirler (örn. tr.toml).
func (r *Registry) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var entries map[string]string
	if err := toml.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("%s: %w", defaultRegistry.Resolvef(Default(), "%df.internal.i18n.parse_toml", path), err)
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lang, err := language.Parse(base)
	if err != nil {
		return fmt.Errorf("%s: %w", defaultRegistry.Resolvef(Default(), "%df.internal.i18n.invalid_locale_filename", base), err)
	}
	r.Register(lang, entries)
	return nil
}

// Load, varsayılan registry'ye bir TOML dosyası yükler.
func Load(path string) error {
	return defaultRegistry.Load(path)
}

// LoadFolder, bir klasördeki tüm .toml dosyalarını yükler.
func (r *Registry) LoadFolder(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".toml") {
			continue
		}
		if err := r.Load(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// LoadFolder, varsayılan registry'ye bir klasör yükler.
func LoadFolder(dir string) error {
	return defaultRegistry.LoadFolder(dir)
}

// T, vanilla translation key için chat.Translation döndürür.
// Client bu key'i kendi dilinde çevirir.
func T(key string, params int) chat.Translation {
	return chat.Translate(translatorString{key}, params, resolveFallback(key)).Raw()
}

// S, parametresiz vanilla key için kısayoldur.
func S(key string) chat.Translation {
	return T(key, 0)
}

// LocaleSource, locale bilgisi sağlayan herhangi bir kaynaktır.
type LocaleSource interface {
	Locale() language.Tag
}

// M, custom key'i kaynağin locale'ine göre sunucu tarafinda çözümler ve
// parametreleri yerlestirir. Sonuç raw text olarak kullanilir.
func M(src any, key string, args ...any) string {
	locale := Default()
	if ls, ok := src.(LocaleSource); ok {
		locale = ls.Locale()
	}
	return defaultRegistry.Resolvef(locale, key, args...)
}

// R, sunucu varsayilan dilinde custom key'i çözümler ve parametreleri yerlestirir.
// Console loglari ve internal hata mesajlari için kullanilir.
func R(key string, args ...any) string {
	return defaultRegistry.Resolvef(Default(), key, args...)
}

// D, sunucu tarafinda çözümlenen bir chat.Translation döndürür.
// Command description gibi client'in kendi çeviremeyeceği metinler için kullanilir.
// Raw döndürür; böylece client command tree'de §r öneki olmadan gönderilir.
func D(key string) chat.Translation {
	return chat.Translate(translatorString{key}, 0, resolveFallback(key)).Raw()
}

// Resolvef, registry'den format string'i çeker ve argümanlarla formatlar.
// %v, %s, %d sıralı olmayan; %1, %2 ... sıralı placeholder'lari destekler.
func (r *Registry) Resolvef(locale language.Tag, key string, args ...any) string {
	format := r.resolveWithFallback(locale, key)
	if format == "" {
		format = key
	}
	// text.Colourf'ye sabit "%s" formati ile veriyoruz; böylece go vet
	// "non-constant format string" uyarmaz ve renk tag'leri çözümlenir.
	return text.Colourf("%s", formatArgs(format, args...))
}

// formatArgs, format string içindeki placeholder'lari argümanlarla değiştirir.
// Sırali olmayan: %v, %s, %d (argüman sirasina göre tüketilir).
// Sırali: %1, %2 ... (kalan argümanlardan 1-bazli seçer).
// %% ile literal '%' yazilabilir.
func formatArgs(format string, args ...any) string {
	strs := make([]string, len(args))
	for i, a := range args {
		strs[i] = fmt.Sprint(a)
	}

	var sb strings.Builder
	argIdx := 0
	for i := 0; i < len(format); {
		if format[i] == '%' && i+1 < len(format) {
			ch := format[i+1]
			if ch == '%' {
				sb.WriteByte('%')
				i += 2
				continue
			}
			if ch == 'v' || ch == 's' || ch == 'd' {
				if argIdx < len(strs) {
					sb.WriteString(strs[argIdx])
					argIdx++
				} else {
					sb.WriteByte('%')
					sb.WriteByte(ch)
				}
				i += 2
				continue
			}
		}
		sb.WriteByte(format[i])
		i++
	}
	result := sb.String()

	// Kalan argümanlari %1, %2 ... ile eşleştir.
	remaining := strs[argIdx:]
	if len(remaining) > 0 {
		result = orderedPlaceholderRe.ReplaceAllStringFunc(result, func(match string) string {
			n, _ := strconv.Atoi(match[1:])
			if n >= 1 && n <= len(remaining) {
				return remaining[n-1]
			}
			return match
		})
	}

	return result
}

// IsCustom, bir key'in custom (%df.) olup olmadığını kontrol eder.
// Custom olmayan key'ler vanilla olarak client'a gönderilir.
func IsCustom(key string) bool {
	return strings.HasPrefix(key, "%df.")
}

func resolveFallback(key string) string {
	if v := defaultRegistry.resolveWithFallback(Default(), key); v != "" {
		return v
	}
	return key
}

func (r *Registry) resolveWithFallback(locale language.Tag, key string) string {
	if v := r.resolve(locale, key); v != "" {
		return v
	}
	if v := r.resolve(Default(), key); v != "" {
		return v
	}
	return ""
}

func (r *Registry) resolve(locale language.Tag, key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.matcher == nil {
		return ""
	}
	tag, _, _ := r.matcher.Match(locale)
	m, ok := r.data[tag]
	if !ok {
		return ""
	}
	return m[key]
}

func (r *Registry) rebuildMatcher() {
	r.supported = make([]language.Tag, 0, len(r.data))
	for k := range r.data {
		r.supported = append(r.supported, k)
	}
	r.matcher = language.NewMatcher(r.supported)
}

type translatorString struct {
	key string
}

func (t translatorString) Resolve(l language.Tag) string {
	if !IsCustom(t.key) {
		return t.key
	}
	return defaultRegistry.resolveWithFallback(l, t.key)
}
