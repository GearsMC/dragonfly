package builtin

import (
	"strings"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/whitelist"
	"github.com/df-mc/dragonfly/server/world"
)

var whitelistInstance *whitelist.Whitelist

// RegisterWhitelist, whitelist komutunu ve whitelist yönetici referansını sunucuya bağlar.
func RegisterWhitelist(srv *server.Server, wl *whitelist.Whitelist) {
	whitelistInstance = wl

	cmd.Register(cmd.NewWithTree("whitelist",
		i18n.D("%df.cmd.whitelist.description"),
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
		output.Errorm(src, "%df.whitelist.notinitialized")
		return
	}

	switch strings.ToLower(w.Action) {
	case "on", "enable":
		if err := whitelistInstance.SetEnabled(true); err != nil {
			output.Errorm(src, "%df.whitelist.add.error", err)
			return
		}
		output.Printm(src, "%df.whitelist.enabled")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "off", "disable":
		if err := whitelistInstance.SetEnabled(false); err != nil {
			output.Errorm(src, "%df.whitelist.remove.error", err)
			return
		}
		output.Printm(src, "%df.whitelist.disabled")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "add", "a":
		name, ok := w.Player.Load()
		if !ok || name == "" {
			output.Errorm(src, "%df.whitelist.add.usage")
			return
		}
		if isXUIDLike(name) {
			if err := whitelistInstance.Add(name, ""); err != nil {
				output.Errorm(src, "%df.whitelist.add.error", err)
				return
			}
			output.Printm(src, "%df.whitelist.add.success", name)
		} else {
			identity, ok, err := resolveOperatorTarget(w.srv, name)
			if err != nil {
				output.Errorm(src, "%df.whitelist.resolve.error", err)
				return
			}
			if !ok {
				output.Errorm(src, "%df.whitelist.offline", name)
				return
			}
			if err := whitelistInstance.Add(identity.xuid, identity.name); err != nil {
				output.Errorm(src, "%df.whitelist.add.error", err)
				return
			}
			output.Printm(src, "%df.whitelist.add.success", identity.display())
		}
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "remove", "r", "delete":
		name, ok := w.Player.Load()
		if !ok || name == "" {
			output.Errorm(src, "%df.whitelist.remove.usage")
			return
		}
		if isXUIDLike(name) {
			if err := whitelistInstance.Remove(name); err != nil {
				output.Errorm(src, "%df.whitelist.remove.error", err)
				return
			}
			output.Printm(src, "%df.whitelist.remove.success", name)
		} else {
			identity, ok, err := resolveOperatorTarget(w.srv, name)
			if err != nil {
				output.Errorm(src, "%df.whitelist.resolve.error", err)
				return
			}
			if !ok {
				if xuid, found := whitelistInstance.ResolveByName(name); found {
					if err := whitelistInstance.Remove(xuid); err != nil {
						output.Errorm(src, "%df.whitelist.remove.error", err)
						return
					}
					output.Printm(src, "%df.whitelist.remove.success", name)
					output.SetBroadcastScope(cmd.BroadcastPermitted).
						SetRequiredPermissions(permission.CommandWhitelist)
					return
				}
				output.Errorm(src, "%df.whitelist.notfound", name)
				return
			}
			if err := whitelistInstance.Remove(identity.xuid); err != nil {
				output.Errorm(src, "%df.whitelist.remove.error", err)
				return
			}
			output.Printm(src, "%df.whitelist.remove.success", identity.display())
		}
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	case "list", "l":
		entries := whitelistInstance.Entries()
		enabled := whitelistInstance.Enabled()
		if len(entries) == 0 {
			if enabled {
				output.Printm(src, "%df.whitelist.list.empty.enabled")
			} else {
				output.Printm(src, "%df.whitelist.list.empty.disabled")
			}
		} else {
			status := i18n.M(src, "%df.whitelist.status.disabled")
			if enabled {
				status = i18n.M(src, "%df.whitelist.status.enabled")
			}
			output.Printm(src, "%df.whitelist.list.header", status, len(entries))
			for _, e := range entries {
				if e.LastKnownName != "" {
					output.Printm(src, "%df.whitelist.list.entry", e.LastKnownName, e.XUID)
				} else {
					output.Printm(src, "%df.whitelist.list.entry.xuid", e.XUID)
				}
			}
		}

	case "reload":
		if err := whitelistInstance.Reload(); err != nil {
			output.Errorm(src, "%df.whitelist.reload.error", err)
			return
		}
		output.Printm(src, "%df.whitelist.reload.success")
		output.SetBroadcastScope(cmd.BroadcastPermitted).
			SetRequiredPermissions(permission.CommandWhitelist)

	default:
		output.Errorm(src, "%df.whitelist.usage")
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
