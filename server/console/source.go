package console

import (
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Source is the command source used by the server terminal.
type Source struct {
	writer   *Writer
	position mgl64.Vec3
}

// NewSource creates a new console command source.
func NewSource(writer *Writer, position mgl64.Vec3) *Source {
	return &Source{writer: writer, position: position}
}

// Console marks this Source as the server console.
func (*Source) Console() bool {
	return true
}

// Position returns the command execution position of the console.
func (s *Source) Position() mgl64.Vec3 {
	return s.position
}

// PermissionXUID, console için kalıcı oyuncu XUID'si olmadığını belirtir.
func (*Source) PermissionXUID() string {
	return ""
}

// PermissionName, permission loglarında console kaynağını okunabilir şekilde tanımlar.
func (*Source) PermissionName() string {
	return i18n.M(nil, "%df.console.permission_name")
}

// PermissionState, console için her permission'ı Allow kabul eder.
func (*Source) PermissionState(string) permission.State {
	return permission.Allow
}

// HasPermission, console için her permission'a izin verir.
func (*Source) HasPermission(string) bool {
	return true
}

// HasCommandPermission grants every command permission to the server console.
func (s *Source) HasCommandPermission(permission string) bool {
	return s.HasPermission(permission)
}

// PermissionCalculator, console'un sabit izin calculator'ını döndürür.
func (*Source) PermissionCalculator() permission.Calculator {
	return permission.ConstantCalculator{State: permission.Allow}
}

// SetPermissionCalculator, console için işlem yapmaz; console her zaman tam yetkilidir.
func (*Source) SetPermissionCalculator(permission.Calculator) {}

// RecalculatePermissions, console için işlem yapmaz.
func (*Source) RecalculatePermissions() {}

// OnPermissionChange, console için işlem yapmaz.
func (*Source) OnPermissionChange() {}

// RefreshPermissions, console için işlem yapmaz.
func (*Source) RefreshPermissions() {}

// RefreshCommands, console için işlem yapmaz.
func (*Source) RefreshCommands() {}

// Compile-time contract checks.
var (
	_ permission.Permissible = (*Source)(nil)
	_ cmd.PermissionSource   = (*Source)(nil)
	_ cmd.ConsoleSource      = (*Source)(nil)
)

// SendCommandOutput writes command output to the terminal.
func (s *Source) SendCommandOutput(output *cmd.Output) {
	if output == nil || s.writer == nil {
		return
	}
	for _, message := range output.Messages() {
		s.writeLines(text.Green+i18n.M(nil, "%df.console.command_output_prefix")+text.White+message.String()+text.Reset, false)
	}
	for _, err := range output.Errors() {
		s.writeLines(text.Red+i18n.M(nil, "%df.console.command_error_prefix")+text.White+err.Error()+text.Reset, true)
	}
}

func (s *Source) writeLines(message string, trim bool) {
	if trim {
		message = strings.TrimSpace(message)
	}
	for _, line := range strings.Split(message, "\n") {
		if line == "" {
			continue
		}
		s.writer.Line(line)
	}
}
