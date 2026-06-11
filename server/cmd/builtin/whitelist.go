package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/whitelist"
	"github.com/df-mc/dragonfly/server/world"
)

var whitelistInstance *whitelist.Whitelist

// RegisterWhitelist, whitelist komutunu ve whitelist yönetici referansını sunucuya bağlar.
func RegisterWhitelist(srv *server.Server, wl *whitelist.Whitelist) {
	whitelistInstance = wl

	cmd.Register(cmd.NewWithTree("whitelist",
		"Whitelist yönetimi.",
		[]string{"wl"},
		cmd.NewCommandTree(
			cmd.Argument("eylem", "", cmd.ArgumentSuggestions("WhitelistAction", func(_ cmd.Source) []string {
				return []string{"on", "off", "add", "remove", "list", "reload"}
			})).
				Then(
					cmd.Argument("oyuncu", "").Optional().
						Executes(&WhitelistCommand{srv: srv}),
				),
		),
	).WithPermissions(permission.CommandWhitelist))
}

// WhitelistCommand, /whitelist komutu.
// Whitelist'i yönetir: aç/kapat, oyuncu ekle/çıkar, listele.
type WhitelistCommand struct {
	srv    *server.Server `cmd:"-"`
	Action string
	Player cmd.Optional[string]
}

// Run, /whitelist komutunu çalıştırır.
func (w WhitelistCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	if whitelistInstance == nil {
		output.Error("Whitelist yöneticisi başlatılmamış.")
		return
	}

	switch strings.ToLower(w.Action) {
	case "on", "enable":
		if err := whitelistInstance.SetEnabled(true); err != nil {
			output.Errorf("Whitelist açılamadı: %v", err)
			return
		}
		output.Print("Whitelist açıldı.")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "off", "disable":
		if err := whitelistInstance.SetEnabled(false); err != nil {
			output.Errorf("Whitelist kapatılamadı: %v", err)
			return
		}
		output.Print("Whitelist kapatıldı.")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "add", "a":
		name, ok := w.Player.Load()
		if !ok || name == "" {
			output.Error("Kullanım: /whitelist add <oyuncu|xuid>")
			return
		}
		if isXUIDLike(name) {
			if err := whitelistInstance.Add(name, ""); err != nil {
				output.Errorf("Eklenemedi: %v", err)
				return
			}
			output.Printf("%s whitelist'e eklendi.", name)
		} else {
			identity, ok, err := resolveOperatorTarget(w.srv, name)
			if err != nil {
				output.Errorf("Kimlik çözülemedi: %v", err)
				return
			}
			if !ok {
				output.Errorf("%s daha önce sunucuya bağlanmamış, XUID'sini manuel girin.", name)
				return
			}
			if err := whitelistInstance.Add(identity.xuid, identity.name); err != nil {
				output.Errorf("Eklenemedi: %v", err)
				return
			}
			output.Printf("%s whitelist'e eklendi.", identity.display())
		}
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "remove", "r", "delete":
		name, ok := w.Player.Load()
		if !ok || name == "" {
			output.Error("Kullanım: /whitelist remove <oyuncu|xuid>")
			return
		}
		if isXUIDLike(name) {
			if err := whitelistInstance.Remove(name); err != nil {
				output.Errorf("Çıkarılamadı: %v", err)
				return
			}
			output.Printf("%s whitelist'ten çıkarıldı.", name)
		} else {
			identity, ok, err := resolveOperatorTarget(w.srv, name)
			if err != nil {
				output.Errorf("Kimlik çözülemedi: %v", err)
				return
			}
			if !ok {
				if xuid, found := whitelistInstance.ResolveByName(name); found {
					if err := whitelistInstance.Remove(xuid); err != nil {
						output.Errorf("Çıkarılamadı: %v", err)
						return
					}
					output.Printf("%s whitelist'ten çıkarıldı.", name)
				}
				output.Errorf("%s bulunamadı.", name)
				return
			}
			if err := whitelistInstance.Remove(identity.xuid); err != nil {
				output.Errorf("Çıkarılamadı: %v", err)
				return
			}
			output.Printf("%s whitelist'ten çıkarıldı.", identity.display())
		}
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "list", "l":
		entries := whitelistInstance.Entries()
		enabled := whitelistInstance.Enabled()
		if len(entries) == 0 {
			if enabled {
				output.Print("Whitelist açık ama hiç oyuncu yok.")
			} else {
				output.Print("Whitelist kapalı, hiç oyuncu yok.")
			}
		} else {
			output.Printf("Whitelist (%s) — %d oyuncu:", map[bool]string{true: "açık", false: "kapalı"}[enabled], len(entries))
			for _, e := range entries {
				if e.LastKnownName != "" {
					output.Printf("  %s (%s)", e.LastKnownName, e.XUID)
				} else {
					output.Printf("  %s", e.XUID)
				}
			}
		}

	case "reload":
		if err := whitelistInstance.Reload(); err != nil {
			output.Errorf("Yeniden yüklenemedi: %v", err)
			return
		}
		output.Print("Whitelist yeniden yüklendi.")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	default:
		output.Error("Kullanım: /whitelist <on|off|add|remove|list|reload>")
	}
}

func isXUIDLike(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 10
}
