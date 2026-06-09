package cmd

import "github.com/df-mc/dragonfly/server/world"

// Context, command execution sırasında komuta verilen tüm ortak bilgileri taşır.
// Yeni komutlar mümkün olduğunda RunContext kullanarak bu tek giriş noktasından çalışmalıdır.
type Context struct {
	Command     Command
	Source      Source
	Output      *Output
	Tx          *world.Tx
	Args        []string
	Values      map[string]any
	Permissions []string
}

// HasPermission, source command permission desteği taşıyorsa verilen permission'ı kontrol eder.
func (ctx *Context) HasPermission(permission string) bool {
	permissionSource, ok := ctx.Source.(PermissionSource)
	return ok && permissionSource.HasCommandPermission(permission)
}

// Value, command tree tarafından parse edilen argument değerini döndürür.
func (ctx *Context) Value(name string) (any, bool) {
	value, ok := ctx.Values[name]
	return value, ok
}

// Print, command output'a başarı mesajı ekler.
func (ctx *Context) Print(a ...any) {
	ctx.Output.Print(a...)
}

// Printf, command output'a formatlı başarı mesajı ekler.
func (ctx *Context) Printf(format string, a ...any) {
	ctx.Output.Printf(format, a...)
}

// Error, command output'a hata mesajı ekler.
func (ctx *Context) Error(a ...any) {
	ctx.Output.Error(a...)
}

// Errorf, command output'a formatlı hata mesajı ekler.
func (ctx *Context) Errorf(format string, a ...any) {
	ctx.Output.Errorf(format, a...)
}

// ContextRunnable, klasik Run imzası yerine Context ile çalışabilen command runnable'dır.
type ContextRunnable interface {
	RunContext(ctx *Context)
}

// HandlerFunc, command tree leaf'inde doğrudan fonksiyon çalıştırmak için kullanılan runnable'dır.
type HandlerFunc func(ctx *Context)

// Run, HandlerFunc için klasik Runnable kontratını sağlar. Gerçek çalışma RunContext üzerinden yapılır.
func (HandlerFunc) Run(Source, *Output, *world.Tx) {}

// RunContext, HandlerFunc'i çalıştırır.
func (f HandlerFunc) RunContext(ctx *Context) {
	f(ctx)
}
