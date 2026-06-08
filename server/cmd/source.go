package cmd

// Source represents a source of a command execution. Commands may limit the sources that can run them by
// implementing the Allower interface.
// Source implements Target. A Source must always be able to target itself.
type Source interface {
	Target
	// SendCommandOutput sends a command output to the source. The way the output is applied, depends on what
	// kind of source it is.
	// SendCommandOutput is called by a Command automatically after being run.
	SendCommandOutput(o *Output)
}

// ConsoleSource may be implemented by a Source that represents the server
// console. Console sources may execute commands without a leading slash.
type ConsoleSource interface {
	Source
	Console() bool
}

// PermissionSource may be implemented by a Source that can evaluate command
// permissions. Sources that do not implement this interface can only execute
// commands without explicit permissions.
type PermissionSource interface {
	Source
	HasCommandPermission(permission string) bool
}
