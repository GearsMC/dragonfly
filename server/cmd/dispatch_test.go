package cmd

import (
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestParseCommandLineRequiresSlashForRegularSources(t *testing.T) {
	name, args, ok := ParseCommandLine("status now", regularSource{})
	if ok {
		t.Fatalf("expected command without slash to be ignored, got name=%q args=%q", name, args)
	}
}

func TestParseCommandLineAllowsConsoleWithoutSlash(t *testing.T) {
	name, args, ok := ParseCommandLine("status now", consoleSource{})
	if !ok {
		t.Fatal("expected console command without slash to parse")
	}
	if name != "status" || args != "now" {
		t.Fatalf("unexpected parsed command: name=%q args=%q", name, args)
	}
}

func TestParseCommandLineRejectsDisabledConsoleWithoutSlash(t *testing.T) {
	name, args, ok := ParseCommandLine("status now", disabledConsoleSource{})
	if ok {
		t.Fatalf("expected disabled console command without slash to be ignored, got name=%q args=%q", name, args)
	}
}

func TestArgumentPreviewKeepsQuotedArguments(t *testing.T) {
	args := ArgumentPreview(`hello "two words" tail`)
	if len(args) != 3 || args[0] != "hello" || args[1] != "two words" || args[2] != "tail" {
		t.Fatalf("unexpected arguments: %#v", args)
	}
}

type regularSource struct{}

func (regularSource) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (regularSource) SendCommandOutput(*Output) {}

type consoleSource struct {
	regularSource
}

func (consoleSource) Console() bool {
	return true
}

type disabledConsoleSource struct {
	regularSource
}

func (disabledConsoleSource) Console() bool {
	return false
}
