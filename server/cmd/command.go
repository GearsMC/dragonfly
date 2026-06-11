package cmd

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"reflect"
	"slices"
	"strings"

	"github.com/df-mc/dragonfly/server/world"
)

// Runnable represents a Command that may be run by a Command source. The Command must be a struct type and
// its fields represent the parameters of the Command. When the Run method is called, these fields are set
// and may be used for behaviour in the Command. Fields unexported or ignored using the `cmd:"-"` struct tag (see
// below) have their values copied but retained.
// A Runnable may have exported fields only of the following types:
// int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint,
// float32, float64, string, bool, mgl64.Vec3, Varargs, []Target, cmd.SubCommand, Optional[T] (to make a parameter
// optional), or a type that implements the cmd.Parameter or cmd.Enum interface. cmd.Enum implementations must be of the
// type string.
// Fields in the Runnable struct may have `cmd:` struct tag to specify the name and suffix of a parameter as such:
//
//	type T struct {
//	    Param int `cmd:"name,suffix"`
//	}
//
// If no name is set, the field name is used. Additionally, the name as specified in the struct tag may be '-' to make
// the parser ignore the field. In this case, the field does not have to be of one of the types above.
type Runnable interface {
	// Run runs the Command, using the arguments passed to the Command. The source is passed to the method,
	// which is the source of the Command execution, and the output is passed, to which messages may be
	// added which get sent to the source.
	Run(src Source, o *Output, tx *world.Tx)
}

// Allower, Runnable de uygulanacak bir tür tarafından uygulanabilir ve
// komutu çalıştırbilen kaynakları sınırlandırmak için kullanılabilir.
type Allower interface {
	// Allow, geçilen Source'un komutu çalıştırmasına izin verilip verilmediğini kontrol eder.
	// Kaynağa komutu çalıştırması izni verilirse true değer döndürülür.
	Allow(src Source) bool
}

// SenderTypeAllower, Runnable de uygulanacak bir tür tarafından uygulanabilir ve
// komutu çalıştırbilen kaynakların tipini sınırlandırmak için kullanılabilir.
type SenderTypeAllower interface {
	// AllowedSenderType, bu komutta çalıştırabilen gönderici türlerini döndürür.
	// Komutu çalıştırabilen hiçbir gönderici varsa SenderTypeAny döndürülebilir.
	AllowedSenderType() SenderType
}

// Command is a wrapper around a Runnable. It provides additional identity and utility methods for the actual
// runnable command so that it may be identified more easily.
type Command struct {
	v           []reflect.Value
	name        string
	description string
	usage       string
	aliases     []string
	permissions []string
	tree        *Tree
	leaves      []commandLeaf
}

// New returns a new Command using the name and description passed. Command
// names and aliases are all converted to lowercase. The Runnable passed must
// be a (pointer to a) struct, with its fields representing the parameters of
// the command. When the command is run, the Run method of the Runnable will be
// called after all fields have their values from the parsed command set. If r
// is not a struct or a pointer to a struct, New panics.
func New(name, description string, aliases []string, r ...Runnable) Command {
	runnableValues := make([]reflect.Value, len(r))
	for i, runnable := range r {
		runnableValues[i] = normaliseRunnable(runnable)
	}
	return newCommand(name, description, aliases, treeFromRunnables(runnableValues), runnableValues)
}

// NewWithTree, açık command tree tanımıyla yeni bir Command oluşturur.
func NewWithTree(name, description string, aliases []string, tree *Tree) Command {
	return newCommand(name, description, aliases, tree, nil)
}

func newCommand(name, description string, aliases []string, tree *Tree, runnableValues []reflect.Value) Command {
	name = strings.ToLower(name)
	aliases = slices.Clone(aliases)
	for i, alias := range aliases {
		aliases[i] = strings.ToLower(alias)
	}
	if len(aliases) > 0 && slices.Index(aliases, name) == -1 {
		aliases = append(aliases, name)
	}

	if tree == nil {
		tree = NewCommandTree()
	}
	leaves := tree.leaves()

	return Command{
		name:        name,
		description: description,
		aliases:     aliases,
		v:           runnableValues,
		tree:        tree,
		leaves:      leaves,
		usage:       usageFromLeaves(name, leaves),
	}
}

// Name returns the name of the command. The name is guaranteed to be lowercase and will never have spaces in
// it. This name is used to call the command, and is shown in the /help list.
func (cmd Command) Name() string {
	return cmd.name
}

// Description returns the description of the command. The description is shown in the /help list, and
// provides information on the functionality of a command.
func (cmd Command) Description() string {
	return cmd.description
}

// Usage returns the usage of the command. The usage will be roughly equal to the one showed by the client
// in-game.
func (cmd Command) Usage() string {
	return cmd.usage
}

// Aliases returns a list of aliases for the command. In addition to the name of the command, the command may
// be called using one of these aliases.
func (cmd Command) Aliases() []string {
	return cmd.aliases
}

// WithPermissions returns a copy of the Command requiring the permissions
// passed. All permissions must be granted by the Source executing the command.
func (cmd Command) WithPermissions(permissions ...string) Command {
	cmd.permissions = slices.Clone(permissions)
	return cmd
}

// Permissions returns the permissions required to execute this Command.
func (cmd Command) Permissions() []string {
	return slices.Clone(cmd.permissions)
}

// CanRun checks if the Source passed is allowed to run this Command before
// runnable-specific checks are applied.
func (cmd Command) CanRun(src Source) bool {
	if len(cmd.permissions) == 0 {
		return true
	}
	permissionSource, ok := src.(PermissionSource)
	if !ok {
		return false
	}
	for _, permission := range cmd.permissions {
		if !permissionSource.HasCommandPermission(permission) {
			return false
		}
	}
	return true
}

// Execute executes the Command as a source with the args passed. The args are parsed assuming they do not
// start with the command name. Execute will attempt to parse and execute one Runnable at a time. If one of
// the Runnable was able to parse args correctly, it will be executed and no more Runnables will be attempted
// to be run.
// If parsing of all Runnables was unsuccessful, a command output with an error message is sent to the Source
// passed, and the Run method of the Runnables are not called.
// The Source passed must not be nil. The method will panic if a nil Source is passed.
func (cmd Command) Execute(args string, source Source, tx *world.Tx) {
	if source == nil {
		panic("execute: invalid command source: source must not be nil")
	}
	output := &Output{}
	defer source.SendCommandOutput(output)
	if !cmd.CanRun(source) {
		output.Errort(MessageUnknown, cmd.name)
		return
	}

	var leastErroneous error
	var leastArgsLeft *Line

	for _, leaf := range cmd.leaves {
		line, err := cmd.executeLeaf(leaf, args, source, output, tx)
		if err == nil {
			// Command was executed successfully: We won't execute any of the other Runnable values passed, as
			// we've already found an overload that works.
			return
		}
		if line == nil {
			// This Runnable was not runnable by the source passed. Only if no error was yet set, we set an
			// error for the wrong source.
			if leastErroneous == nil {
				leastErroneous = err
			}
			continue
		}
		if leastArgsLeft == nil || line.Len() <= leastArgsLeft.Len() {
			// If the line had less (or equal) arguments left than the previous lowest, we update the error,
			// so that we can return an error that applies for the most successful Runnable.
			leastErroneous = err
			leastArgsLeft = line
		}
	}
	// No working Runnable found for the arguments passed. We add the most
	// applicable error to the output and stop there.
	if leastArgsLeft != nil {
		output.Error(leastArgsLeft.SyntaxError())
	}
	output.Error(leastErroneous)
}

// ParamInfo holds the information of a parameter in a Runnable. Information of a parameter may be obtained
// by calling Command.Params().
type ParamInfo struct {
	Name        string
	Description string
	Value       any
	Optional    bool
	Suffix      string
	EnumType    string
	Suggestions SuggestionProvider
}

// Params returns a list of all parameters of the runnables. No assumptions should be done on the values that
// they hold: Only the types are guaranteed to be consistent.
func (cmd Command) Params(src Source) [][]ParamInfo {
	if !cmd.CanRun(src) {
		return nil
	}
	params := make([][]ParamInfo, 0, len(cmd.leaves))
	for _, leaf := range cmd.leaves {
		if !cmd.leafCanRun(leaf, src) {
			continue
		}
		if d, ok := leaf.runnable.Interface().(ParamDescriber); ok {
			params = append(params, d.DescribeParams(src))
			continue
		}
		params = append(params, slices.Clone(leaf.usageParams))
	}
	return params
}

// Runnables returns a map of all Runnable implementations of the Command that a Source can execute.
func (cmd Command) Runnables(src Source) map[int]Runnable {
	if !cmd.CanRun(src) {
		return nil
	}
	m := make(map[int]Runnable, len(cmd.leaves))
	for _, leaf := range cmd.leaves {
		if cmd.leafCanRun(leaf, src) {
			m[leaf.id] = leaf.runnable.Interface().(Runnable)
		}
	}
	return m
}

// String returns the usage of the command. The usage will be roughly equal to the one showed by the client
// in-game.
func (cmd Command) String() string {
	return cmd.usage
}

func (cmd Command) executeLeaf(leaf commandLeaf, args string, source Source, output *Output, tx *world.Tx) (*Line, error) {
	if !cmd.leafCanRun(leaf, source) {
		return nil, MessageUnknown.F(cmd.name)
	}

	var argFrags []string
	if args != "" {
		r := csv.NewReader(strings.NewReader(args))
		r.Comma, r.LazyQuotes = ' ', true
		record, err := r.Read()
		if err != nil {
			return nil, MessageUsage.F(cmd.Usage())
		}
		argFrags = record
	}

	signature := leaf.runnable
	if leaf.runnable.Kind() == reflect.Struct {
		cp := reflect.New(leaf.runnable.Type())
		cp.Elem().Set(leaf.runnable)
		signature = cp.Elem()
	}
	parser := parser{}
	arguments := &Line{args: argFrags, src: source, seen: []string{"/" + cmd.name}, cmd: cmd}
	contextValues := map[string]any{}

	for _, param := range leaf.params {
		if param.literal {
			parser.currentField = param.Name
			val := reflect.New(reflect.TypeOf(SubCommand{})).Elem()
			if err, _ := parser.parseArgument(arguments, val, false, param.Name, source, tx); err != nil {
				return arguments, err
			}
			continue
		}

		if signature.Kind() != reflect.Struct {
			parser.currentField = param.Name
			val := reflect.New(reflect.TypeOf(param.Value)).Elem()
			err, success := parser.parseArgument(arguments, val, param.Optional, param.Name, source, tx)
			if err != nil {
				return arguments, err
			}
			if success {
				contextValues[param.Name] = val.Interface()
			}
			continue
		}

		field := signature.FieldByName(param.fieldName)
		parser.currentField = param.fieldName
		opt := optional(field)
		val := field
		if opt {
			val = reflect.New(field.Field(0).Type()).Elem()
		}

		err, success := parser.parseArgument(arguments, val, opt, param.Name, source, tx)
		if err != nil {
			return arguments, err
		}
		if success && opt {
			field.Set(reflect.ValueOf(field.Interface().(optionalT).with(val.Interface())))
		}
	}
	if arguments.Len() != 0 {
		return arguments, arguments.UsageError()
	}

	runnable := signature.Interface().(Runnable)
	if contextRunnable, ok := runnable.(ContextRunnable); ok {
		contextRunnable.RunContext(&Context{
			Command:     cmd,
			Source:      source,
			Output:      output,
			Tx:          tx,
			Args:        ArgumentPreview(args),
			Values:      contextValues,
			Permissions: slices.Clone(leaf.permissions),
		})
		return arguments, nil
	}
	runnable.Run(source, output, tx)
	return arguments, nil
}

func (cmd Command) leafCanRun(leaf commandLeaf, src Source) bool {
	// İzinleri kontrol et
	if len(leaf.permissions) > 0 {
		permissionSource, ok := src.(PermissionSource)
		if !ok {
			return false
		}
		for _, permission := range leaf.permissions {
			if !permissionSource.HasCommandPermission(permission) {
				return false
			}
		}
	}

	v := leaf.runnable.Interface().(Runnable)

	// Gönderici tipini kontrol et
	if senderTypeAllower, ok := v.(SenderTypeAllower); ok {
		allowedTypes := senderTypeAllower.AllowedSenderType()
		if allowedTypes != SenderTypeAny {
			// Kaynağın gönderici tipini belirle
			var sourceType SenderType
			if senderTypeSource, ok := src.(SenderTypeSource); ok {
				sourceType = senderTypeSource.SenderTypeOf()
			} else if _, isConsole := src.(ConsoleSource); isConsole {
				sourceType = SenderTypeServer
			} else {
				sourceType = SenderTypePlayer
			}

			// İzin verilen tiplerle eşleştir - eşitlik kontrolü
			isAllowed := false
			if allowedTypes == SenderTypeAny {
				isAllowed = true
			} else if sourceType == allowedTypes {
				isAllowed = true
			}

			if !isAllowed {
				return false
			}
		}
	}

	// Allower interface'i kontrol et
	if allower, ok := v.(Allower); ok {
		return allower.Allow(src)
	}
	return true
}

func usageFromLeaves(commandName string, leaves []commandLeaf) string {
	usages := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		parts := make([]string, 0, len(leaf.usageParams)+1)
		parts = append(parts, "/"+commandName)
		for _, param := range leaf.usageParams {
			typeName := typeNameOf(param.Value, param.Name)
			if _, ok := param.Value.(SubCommand); ok {
				parts = append(parts, typeName)
				continue
			}
			if param.Optional {
				parts = append(parts, "["+param.Name+": "+typeName+"]"+param.Suffix)
				continue
			}
			parts = append(parts, "<"+param.Name+": "+typeName+">"+param.Suffix)
		}
		usages = append(usages, strings.Join(parts, " "))
	}
	return strings.Join(usages, "\n")
}

// verifySignature verifies the passed struct pointer value signature to ensure it is a valid command,
// checking things such as the validity of the optional struct tags.
// If not valid, an error is returned.
func verifySignature(command reflect.Value) error {
	optionalField := false
	for _, t := range exportedFields(command) {
		field := command.FieldByName(t.Name)

		// If the field is not optional, while the last field WAS optional, we return an error, as this is
		// not parsable in an expected way.
		opt := optional(field)
		if !opt && optionalField {
			return fmt.Errorf("command must only have optional parameters at the end")
		}
		val := field
		if opt {
			val = reflect.New(field.Field(0).Type()).Elem()
		}
		if _, ok := val.Interface().(Enum); ok && val.Kind() != reflect.String {
			return fmt.Errorf("parameters implementing Enum must be of the type string")
		}
		optionalField = opt
	}
	return nil
}

// exportedFields returns all exported struct fields of the reflect.Value passed. It returns the fields as returned by
// reflect.VisibleFields, but filters out unexported fields, anonymous fields and fields that have a name value in the
// 'cmd' tag of '-'.
func exportedFields(command reflect.Value) []reflect.StructField {
	visible := reflect.VisibleFields(command.Type())
	fields := make([]reflect.StructField, 0, len(visible))

	for _, t := range visible {
		if !ast.IsExported(t.Name) || name(t) == "-" || t.Anonymous {
			continue
		}
		field := command.FieldByName(t.Name)
		if !field.CanSet() {
			continue
		}
		fields = append(fields, t)
	}
	return fields
}
