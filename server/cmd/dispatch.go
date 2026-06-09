package cmd

import (
	"encoding/csv"
	"strings"

	"github.com/df-mc/dragonfly/server/world"
)

// BeforeExecute is called after command lookup and permission checks, but
// before the Command executes. Returning false cancels execution.
type BeforeExecute func(command Command, args []string) bool

// Dispatch parses a command line, looks up the command and executes it for the
// Source passed. Player-like sources must use a leading slash, while
// ConsoleSource implementations may omit it.
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
		source.SendCommandOutput(output)
		return false
	}

	if before != nil && !before(command, ArgumentPreview(args)) {
		return true
	}

	command.Execute(args, source, tx)
	return true
}

// ParseCommandLine extracts the command name and raw argument string from a
// command line. The returned command name is normalised to lowercase.
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

// ArgumentPreview returns a best-effort split of the raw argument string. It is
// intended for command execution hooks and mirrors the parser used by Command.
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
