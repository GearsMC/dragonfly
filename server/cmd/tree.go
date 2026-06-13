package cmd

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/df-mc/dragonfly/server/player/chat"
)

type nodeKind uint8

const (
	nodeRoot nodeKind = iota
	nodeLiteral
	nodeArgument
)

// Tree, bir komutun root -> node -> leaf şeklindeki çalıştırılabilir ağacıdır.
// Her leaf bir Runnable çalıştırır; node seviyesindeki permissionlar hem client command tree'de hem de runtime'da uygulanır.
type Tree struct {
	root *Node
}

// Node, command tree içindeki literal veya argument düğümüdür.
type Node struct {
	kind        nodeKind
	name        string
	description chat.Translation
	value       any
	optional    bool
	suffix      string
	enumType    string
	suggestions SuggestionProvider
	permissions []string
	children    []*Node
	runnable    Runnable
}

// SuggestionProvider, argument node'u için source'a göre autocomplete seçenekleri üretir.
type SuggestionProvider func(source Source) []string

// Root, yeni bir command tree root düğümü oluşturur.
func Root(children ...*Node) *Node {
	return (&Node{kind: nodeRoot}).Then(children...)
}

// Literal, oyuncunun komutta aynen yazması gereken statik bir node oluşturur.
func Literal(name string) *Node {
	if name == "" {
		panic("literal node adı boş olamaz")
	}
	return &Node{kind: nodeLiteral, name: name, value: SubCommand{}}
}

// Argument, parser tarafından okunacak dinamik bir argument node'u oluşturur.
// value, argument tipini temsil eden sıfır değer olmalıdır.
func Argument(name string, value any, opts ...ArgumentOption) *Node {
	if name == "" {
		panic("argument node adı boş olamaz")
	}
	if value == nil {
		panic("argument node tipi nil olamaz")
	}
	n := &Node{kind: nodeArgument, name: name, value: value}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// ArgumentOption, Argument node davranışını değiştirir.
type ArgumentOption func(*Node)

// OptionalArgument, argument node'unu opsiyonel yapar.
func OptionalArgument() ArgumentOption {
	return func(n *Node) {
		n.optional = true
	}
}

// Optional, argument node'unu opsiyonel yapar (convenience method).
// OptionalArgument() ile aynı yerini tutar.
func (n *Node) Optional() *Node {
	n.optional = true
	return n
}

// ArgumentSuffix, argument için client tarafında gösterilecek suffix bilgisini ayarlar.
func ArgumentSuffix(suffix string) ArgumentOption {
	return func(n *Node) {
		n.suffix = suffix
	}
}

// ArgumentSuggestions, argument node'u için client autocomplete seçenekleri üretir.
func ArgumentSuggestions(enumType string, provider SuggestionProvider) ArgumentOption {
	return func(n *Node) {
		n.enumType = enumType
		n.suggestions = provider
	}
}

// GreedyText, kalan tüm argümanları tek text değeri olarak tüketen argument node'u oluşturur.
func GreedyText(name string, opts ...ArgumentOption) *Node {
	return Argument(name, Varargs(""), opts...)
}

// NewCommandTree, verilen root çocuklarıyla yeni bir command tree oluşturur.
func NewCommandTree(children ...*Node) *Tree {
	return &Tree{root: Root(children...)}
}

// Then, node altına çocuk node'lar ekler.
func (n *Node) Then(children ...*Node) *Node {
	for _, child := range children {
		if child == nil {
			panic("command tree child nil olamaz")
		}
		n.children = append(n.children, child)
	}
	return n
}

// Executes, node'u çalıştırılabilir leaf yapar.
func (n *Node) Executes(runnable Runnable) *Node {
	if runnable == nil {
		panic("command tree runnable nil olamaz")
	}
	n.runnable = runnable
	return n
}

// ExecutesFunc, node'u Context kullanan fonksiyon leaf'i yapar.
func (n *Node) ExecutesFunc(fn func(ctx *Context)) *Node {
	return n.Executes(HandlerFunc(fn))
}

// WithPermissions, node ve altındaki leafler için gerekli permissionları ayarlar.
func (n *Node) WithPermissions(permissions ...string) *Node {
	n.permissions = slices.Clone(permissions)
	return n
}

// WithDescription, node için okunabilir açıklama metadata'sı ayarlar.
func (n *Node) WithDescription(description chat.Translation) *Node {
	n.description = description
	return n
}

func (n *Node) paramInfo() ParamInfo {
	return ParamInfo{
		Name:        n.name,
		Description: n.description,
		Value:       n.value,
		Optional:    n.optional,
		Suffix:      n.suffix,
		EnumType:    n.enumType,
		Suggestions: n.suggestions,
	}
}

type commandLeaf struct {
	id          int
	runnable    reflect.Value
	params      []treeParam
	permissions []string
	usageParams []ParamInfo
}

type treeParam struct {
	ParamInfo
	literal   bool
	fieldName string
}

func treeFromRunnables(runnables []reflect.Value) *Tree {
	root := Root()
	for _, runnable := range runnables {
		root.addRunnablePath(runnable)
	}
	return &Tree{root: root}
}

func (n *Node) addRunnablePath(runnable reflect.Value) {
	current := n
	for _, t := range exportedFields(runnable) {
		field := runnable.FieldByName(t.Name)
		if _, ok := field.Interface().(SubCommand); ok {
			current = current.Then(Literal(name(t))).children[len(current.children)-1]
			continue
		}

		value := unwrap(field).Interface()
		opts := []ArgumentOption{}
		if optional(field) {
			opts = append(opts, OptionalArgument())
		}
		if suffix := suffix(t); suffix != "" {
			opts = append(opts, ArgumentSuffix(suffix))
		}
		current = current.Then(Argument(name(t), value, opts...)).children[len(current.children)-1]
	}
	current.runnable = runnable.Interface().(Runnable)
}

func (t *Tree) leaves() []commandLeaf {
	if t == nil || t.root == nil {
		return nil
	}
	var leaves []commandLeaf
	t.root.walk(nil, nil, &leaves)
	for i := range leaves {
		leaves[i].id = i
		leaves[i].bindFields()
	}
	return leaves
}

func (n *Node) walk(params []treeParam, permissions []string, leaves *[]commandLeaf) {
	permissions = appendPermissions(permissions, n.permissions)
	switch n.kind {
	case nodeLiteral:
		params = append(params, treeParam{ParamInfo: n.paramInfo(), literal: true})
	case nodeArgument:
		params = append(params, treeParam{ParamInfo: n.paramInfo()})
	}
	if n.runnable != nil {
		value := normaliseTreeRunnable(n.runnable)
		*leaves = append(*leaves, commandLeaf{
			runnable:    value,
			params:      cloneTreeParams(params),
			permissions: slices.Clone(permissions),
			usageParams: usageParams(params),
		})
	}
	for _, child := range n.children {
		child.walk(params, permissions, leaves)
	}
}

func (l *commandLeaf) bindFields() {
	if l.runnable.Kind() != reflect.Struct {
		return
	}
	fields := exportedFields(l.runnable)
	next := 0
	for i := range l.params {
		if l.params[i].literal {
			continue
		}
		for next < len(fields) {
			field := l.runnable.FieldByName(fields[next].Name)
			if _, ok := field.Interface().(SubCommand); ok {
				next++
				continue
			}
			l.params[i].fieldName = fields[next].Name
			next++
			break
		}
		if l.params[i].fieldName == "" {
			panic(fmt.Sprintf("command tree leaf %T için yeterli runnable alanı yok", l.runnable.Interface()))
		}
	}
	for next < len(fields) {
		field := l.runnable.FieldByName(fields[next].Name)
		if _, ok := field.Interface().(SubCommand); !ok {
			panic(fmt.Sprintf("command tree leaf %T için kullanılmayan runnable alanı var: %s", l.runnable.Interface(), fields[next].Name))
		}
		next++
	}
}

func normaliseTreeRunnable(runnable Runnable) reflect.Value {
	t := reflect.TypeOf(runnable)
	if t.Kind() != reflect.Struct && (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) {
		if _, ok := runnable.(ContextRunnable); ok {
			return reflect.ValueOf(runnable)
		}
	}
	return normaliseRunnable(runnable)
}

func normaliseRunnable(runnable Runnable) reflect.Value {
	t := reflect.TypeOf(runnable)
	if t.Kind() != reflect.Struct && (t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct) {
		panic(fmt.Sprintf("Runnable r struct veya struct pointer olmalı, got %v", t.Kind()))
	}
	original := reflect.ValueOf(runnable)
	if t.Kind() == reflect.Ptr {
		original = original.Elem()
	}
	cp := reflect.New(original.Type()).Elem()
	cp.Set(original)
	if err := verifySignature(cp); err != nil {
		panic(err.Error())
	}
	return cp
}

func appendPermissions(base []string, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	out := slices.Clone(base)
	return append(out, extra...)
}

func cloneTreeParams(params []treeParam) []treeParam {
	clone := make([]treeParam, len(params))
	copy(clone, params)
	return clone
}

func usageParams(params []treeParam) []ParamInfo {
	infos := make([]ParamInfo, len(params))
	for i, param := range params {
		infos[i] = param.ParamInfo
	}
	return infos
}
