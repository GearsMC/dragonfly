package cmd

import (
	"encoding/csv"
	"strings"

	"github.com/df-mc/dragonfly/server/world"
)

// BeforeExecute, komut araması ve izin kontrolleri yapıldıktan sonra,
// ancak Command yürütülmesinden önce çağrılır. false döndürmek yürütülmeyi iptal eder.
type BeforeExecute func(command Command, args []string) bool

// Dispatch, komut satırını ayrıştırır, komutu arar ve geçilen Source için
// yürütür. Oyuncu benzeri kaynaklar başında slash kullanmalıdır, ancak
// ConsoleSource uygulamaları bunu atlayabilir.
func Dispatch(commandLine string, source Source, tx *world.Tx, before BeforeExecute) bool {
	if source == nil {
		panic("dispatch: invalid command source: source must not be nil")
	}

	name, args, ok := ParseCommandLine(commandLine, source)
	if !ok {
		return false
	}

	command, ok := ByAlias(name)
	if !ok || len(command.Runnables(source)) == 0 {
		output := &Output{}
		output.Errort(MessageUnknown, name)
		// Bilinmeyen komut hatası her zaman gönderici tarafından görülür
		source.SendCommandOutput(output)
		return false
	}

	if before != nil && !before(command, ArgumentPreview(args)) {
		return true
	}

	command.Execute(args, source, tx)
	return true
}

// ParseCommandLine, komut satırından komut adını ve ham argüman dizesini ayıklayan fonksiyondur.
// Döndürülen komut adı küçük harflere normalize edilir.
func ParseCommandLine(commandLine string, source Source) (name string, args string, ok bool) {
	commandLine = strings.TrimSpace(commandLine)
	if commandLine == "" {
		return "", "", false
	}

	if stripped, slash := strings.CutPrefix(commandLine, "/"); slash {
		commandLine = stripped
	} else {
		console, ok := source.(ConsoleSource)
		if !ok || !console.Console() {
			return "", "", false
		}
	}

	commandLine = strings.TrimSpace(commandLine)
	if commandLine == "" {
		return "", "", false
	}

	name, args, _ = strings.Cut(commandLine, " ")
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "", "", false
	}
	return name, strings.TrimSpace(args), true
}

// ArgumentPreview, ham argüman dizesinin en iyi çabasıyla bölünmesini döndürür.
// Komut yürütme kancaları için tasarlanmış ve Command tarafından kullanılan ayrıştırıcıyı yansıtır.
func ArgumentPreview(args string) []string {
	if args == "" {
		return nil
	}
	reader := csv.NewReader(strings.NewReader(args))
	reader.Comma, reader.LazyQuotes = ' ', true
	record, err := reader.Read()
	if err != nil {
		return strings.Fields(args)
	}
	return record
}

// FilterOutput, kaynak tarafından erişilebilecek çıktı mesajlarını filtreler.
// BroadcastScope'a göre mesajları gösterir veya gizler.
func FilterOutput(output *Output, source Source) *Output {
	scope := output.BroadcastScope()

	switch scope {
	case BroadcastConsole:
		// Sadece konsoldan görülür
		if _, isConsole := source.(ConsoleSource); !isConsole {
			output.messages = nil
		}

	case BroadcastRequester:
		// Sadece isteyen kaynağa gönderilir (zaten yapılmış)
		// Bu durumda filtreleme gerekmez

	case BroadcastPermitted:
		// Belirtilen izinlere sahip kaynaklara gönderilir
		perms := output.RequiredPermissions()
		if len(perms) > 0 {
			if permSource, ok := source.(PermissionSource); ok {
				// İznesi olmayan oyuncu ise mesajları gizle
				hasAllPermissions := true
				for _, perm := range perms {
					if !permSource.HasCommandPermission(perm) {
						hasAllPermissions = false
						break
					}
				}
				if !hasAllPermissions {
					output.messages = nil
				}
			} else {
				// Izin kontrol edilemiyor, mesajları gizle
				output.messages = nil
			}
		}

	case BroadcastAll:
		// Herkese göster, filtreleme yok
	}

	return output
}
