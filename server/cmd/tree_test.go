package cmd

import (
	"testing"

	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type treeSource struct {
	permissions map[string]bool
	output      *Output
}

func (s *treeSource) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (s *treeSource) SendCommandOutput(output *Output) {
	s.output = output
}

func (s *treeSource) HasCommandPermission(permission string) bool {
	return s.permissions[permission]
}

type treeSetCommand struct {
	Name string
}

func (c treeSetCommand) Run(_ Source, o *Output, _ *world.Tx) {
	o.Print("set:" + c.Name)
}

type treeQueryCommand struct {
	Name string
}

func (c treeQueryCommand) Run(_ Source, o *Output, _ *world.Tx) {
	o.Print("query:" + c.Name)
}

func TestCommandTreeFiltersNodePermissions(t *testing.T) {
	command := NewWithTree("permission", chat.Untranslated("Permission test."), nil, NewCommandTree(
		Literal("set").WithPermissions("test.permission.set").Then(
			Argument("name", "").Executes(treeSetCommand{}),
		),
		Literal("query").WithPermissions("test.permission.query").Then(
			Argument("name", "").Executes(treeQueryCommand{}),
		),
	))
	source := &treeSource{permissions: map[string]bool{"test.permission.query": true}}

	params := command.Params(source)
	if len(params) != 1 {
		t.Fatalf("yalnızca izinli leaf görünmeli, got %d", len(params))
	}
	if params[0][0].Name != "query" {
		t.Fatalf("query leaf görünmeli, got %q", params[0][0].Name)
	}

	command.Execute("set lexa", source, nil)
	if source.output == nil || source.output.ErrorCount() == 0 {
		t.Fatal("izinsiz leaf çalışmamalı")
	}

	command.Execute("query lexa", source, nil)
	if source.output == nil || source.output.MessageCount() != 1 {
		t.Fatal("izinli leaf çalışmalı")
	}
	if got := source.output.Messages()[0].String(); got != "query:lexa" {
		t.Fatalf("beklenmeyen çıktı: %q", got)
	}
}

func TestCommandTreeGeneratedFromRunnable(t *testing.T) {
	command := New("example", chat.Untranslated("Example."), nil, structCommand{})
	source := &treeSource{permissions: map[string]bool{}}

	params := command.Params(source)
	if len(params) != 1 || len(params[0]) != 2 {
		t.Fatalf("reflection runnable tree parametreleri üretmeli, got %#v", params)
	}
	if params[0][0].Name != "sub" || params[0][1].Name != "Name" {
		t.Fatalf("beklenmeyen params: %#v", params[0])
	}

	command.Execute("sub lexa", source, nil)
	if source.output == nil || source.output.MessageCount() != 1 {
		t.Fatal("reflection tree çalışmalı")
	}
	if got := source.output.Messages()[0].String(); got != "sub:lexa" {
		t.Fatalf("beklenmeyen çıktı: %q", got)
	}
}

func TestDispatchUsesLeafPermissions(t *testing.T) {
	command := NewWithTree("treepermtest", chat.Untranslated("Permission dispatch test."), nil, NewCommandTree(
		Literal("admin").WithPermissions("test.dispatch.admin").Executes(treeNoopCommand{}),
	))
	Register(command)

	source := &treeSource{permissions: map[string]bool{}}
	called := false
	if Dispatch("/treepermtest admin", source, nil, func(Command, []string) bool {
		called = true
		return true
	}) {
		t.Fatal("yetkisiz leaf dispatch başarılı görünmemeli")
	}
	if called {
		t.Fatal("yetkisiz leaf before hook'a düşmemeli")
	}
	if source.output == nil || source.output.ErrorCount() == 0 {
		t.Fatal("yetkisiz leaf unknown error üretmeli")
	}
}

func TestCommandTreeRootPermission(t *testing.T) {
	command := NewWithTree("rootperm", chat.Untranslated("Root permission test."), nil, NewCommandTree(
		Root().WithPermissions("test.root").Executes(treeNoopCommand{}),
	))

	denied := &treeSource{permissions: map[string]bool{}}
	if len(command.Runnables(denied)) != 0 {
		t.Fatal("root permission olmayan kaynak runnable görmemeli")
	}

	allowed := &treeSource{permissions: map[string]bool{"test.root": true}}
	if len(command.Runnables(allowed)) != 1 {
		t.Fatal("root permission olan kaynak runnable görmeli")
	}
}

func TestCommandTreeContextRunnableParsesValues(t *testing.T) {
	var captured any
	command := NewWithTree("context", chat.Untranslated("Context command test."), nil, NewCommandTree(
		Argument("name", "", ArgumentSuggestions("PlayerName", func(Source) []string {
			return []string{"lexa5936"}
		})).ExecutesFunc(func(ctx *Context) {
			var ok bool
			captured, ok = ctx.Value("name")
			if !ok {
				ctx.Error("name argument yok")
				return
			}
			ctx.Print("context:" + captured.(string))
		}),
	))
	source := &treeSource{permissions: map[string]bool{}}

	params := command.Params(source)
	if len(params) != 1 || len(params[0]) != 1 {
		t.Fatalf("context command parametre üretmeli, got %#v", params)
	}
	if params[0][0].EnumType != "PlayerName" {
		t.Fatalf("suggestion enum tipi korunmalı, got %q", params[0][0].EnumType)
	}
	if suggestions := params[0][0].Suggestions(source); len(suggestions) != 1 || suggestions[0] != "lexa5936" {
		t.Fatalf("suggestion provider korunmalı, got %#v", suggestions)
	}

	command.Execute("lexa5936", source, nil)
	if captured != "lexa5936" {
		t.Fatalf("context argument değeri parse edilmeli, got %#v", captured)
	}
	if source.output == nil || source.output.MessageCount() != 1 {
		t.Fatal("context runnable output üretmeli")
	}
	if got := source.output.Messages()[0].String(); got != "context:lexa5936" {
		t.Fatalf("beklenmeyen çıktı: %q", got)
	}
}

func TestListenRegistryChanges(t *testing.T) {
	called := 0
	unregister := ListenRegistryChanges(func() {
		called++
	})
	Register(NewWithTree("registrylistenertest", chat.Untranslated("Registry listener test."), nil, NewCommandTree(
		Root().Executes(treeNoopCommand{}),
	)))
	if called != 1 {
		t.Fatalf("listener bir kez çağrılmalı, got %d", called)
	}

	unregister()
	Register(NewWithTree("registrylistenertesttwo", chat.Untranslated("Registry listener test two."), nil, NewCommandTree(
		Root().Executes(treeNoopCommand{}),
	)))
	if called != 1 {
		t.Fatalf("unregister sonrası listener çağrılmamalı, got %d", called)
	}
}

type treeNoopCommand struct{}

func (treeNoopCommand) Run(Source, *Output, *world.Tx) {}

type structCommand struct {
	Sub  SubCommand `cmd:"sub"`
	Name string
}

func (c structCommand) Run(_ Source, o *Output, _ *world.Tx) {
	o.Print("sub:" + c.Name)
}
