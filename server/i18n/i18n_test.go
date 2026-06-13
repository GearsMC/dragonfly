package i18n

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func TestRegistry(t *testing.T) {
	r := New()
	r.Register(language.Turkish, map[string]string{
		"%df.test.hello": "Merhaba %v",
	})
	r.Register(language.English, map[string]string{
		"%df.test.hello": "Hello %v",
	})

	if got := r.Resolvef(language.Turkish, "%df.test.hello", "Dunya"); got != "§rMerhaba Dunya" {
		t.Errorf("Turkish resolve = %q, want %q", got, "§rMerhaba Dunya")
	}
	if got := r.Resolvef(language.English, "%df.test.hello", "World"); got != "§rHello World" {
		t.Errorf("English resolve = %q, want %q", got, "§rHello World")
	}
}

func TestVanillaKey(t *testing.T) {
	tr := T("%commands.gamemode.success.self", 1)
	resolved := tr.F("creative").Resolve(language.Turkish)
	if resolved != "%commands.gamemode.success.self" {
		t.Errorf("vanilla resolve = %q, want raw key", resolved)
	}
}

func TestIsCustom(t *testing.T) {
	if !IsCustom("%df.cmd.test") {
		t.Error("df.cmd.test should be custom")
	}
	if IsCustom("%commands.gamemode.success.self") {
		t.Error("vanilla key should not be custom")
	}
}

func TestFormatArgsUnordered(t *testing.T) {
	r := New()
	r.Register(language.Turkish, map[string]string{
		"%df.test.unordered": "Merhaba %s, %d",
	})
	if got := r.Resolvef(language.Turkish, "%df.test.unordered", "Dunya", 42); got != "§rMerhaba Dunya, 42" {
		t.Errorf("unordered = %q, want %q", got, "§rMerhaba Dunya, 42")
	}
}

func TestFormatArgsOrdered(t *testing.T) {
	r := New()
	r.Register(language.Turkish, map[string]string{
		"%df.test.ordered": "%2 %1",
	})
	if got := r.Resolvef(language.Turkish, "%df.test.ordered", "Dunya", "Merhaba"); got != "§rMerhaba Dunya" {
		t.Errorf("ordered = %q, want %q", got, "§rMerhaba Dunya")
	}
}

func TestFormatArgsMixed(t *testing.T) {
	r := New()
	r.Register(language.Turkish, map[string]string{
		"%df.test.mixed": "%v %2",
	})
	if got := r.Resolvef(language.Turkish, "%df.test.mixed", "A", "B", "C"); got != "§rA C" {
		t.Errorf("mixed = %q, want %q", got, "§rA C")
	}
}

func TestFormatArgsEscape(t *testing.T) {
	r := New()
	r.Register(language.Turkish, map[string]string{
		"%df.test.escape": "%%v",
	})
	if got := r.Resolvef(language.Turkish, "%df.test.escape", "x"); got != "§r%v" {
		t.Errorf("escape = %q, want %q", got, "§r%v")
	}
}

// customKeyRe, Go kaynak dosyalarinda kullanilan %df.* key string literal'larini bulur.
var customKeyRe = regexp.MustCompile(`"%df\.[a-zA-Z0-9_.]+"`)

// TestAllCustomKeysRegistered, üretim kodunda kullanilan tüm %df.* custom key'lerin
// gömülü Türkçe TOML'da tanımlı olduğunu doğrular.
func TestAllCustomKeysRegistered(t *testing.T) {
	used := findUsedCustomKeys(t)
	var missing []string
	for key := range used {
		if defaultRegistry.Resolvef(language.Turkish, key) == "§r"+key {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("gömülü tr.toml'da eksik çeviri key'leri: %v", missing)
	}
}

func findUsedCustomKeys(t *testing.T) map[string]bool {
	keys := make(map[string]bool)
	root := ".."
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range customKeyRe.FindAllString(string(data), -1) {
			keys[strings.Trim(m, `"`)] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("kaynak dosyalari taranirken hata: %v", err)
	}
	return keys
}
