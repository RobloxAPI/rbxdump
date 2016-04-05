// The rbxapi package is used to parse and handle information about the Roblox
// Lua API.
package rbxapi

import (
	"sort"
)

// TagGroup is a list of known tags that appear only in certain contexts.
type TagGroup struct {
	Name string
	Tags []string
}

var (
	// Each tag may appear on any item.
	MetadataItem = TagGroup{
		Name: "MetadataItem",
		Tags: []string{
			"notbrowsable",
			"deprecated",
			"backend",
		},
	}
	// Each tag may appear on any Class.
	MetadataClass = TagGroup{
		Name: "MetadataClass",
		Tags: []string{
			"notCreatable",
		},
	}
	// Each tag may appear on any Property.
	MetadataProperty = TagGroup{
		Name: "MetadataProperty",
		Tags: []string{
			"hidden",
			"readonly",
			"writeonly",
		},
	}
	// Each tag may appear on any Callback.
	MetadataCallback = TagGroup{
		Name: "MetadataCallback",
		Tags: []string{
			"noyield",
		},
	}
	// One tag may appear on any Member.
	MemberSecurity = TagGroup{
		Name: "MemberSecurity",
		Tags: []string{
			"LocalUserSecurity",
			"PluginSecurity",
			"RobloxPlaceSecurity",
			"RobloxScriptSecurity",
			"RobloxSecurity",
			"WritePlayerSecurity",
		},
	}
)

// GroupOrder is a list of known tag groups. If an item has multiple tags,
// this indicates the order in which those tags should appear.
var GroupOrder = []TagGroup{
	MetadataItem,
	MetadataClass,
	MetadataProperty,
	MetadataCallback,
	MemberSecurity,
}

// Taggable is an API item that may have a number of tags attached to it.
type Taggable interface {
	// Tag returns whether a given tag is enabled.
	Tag(tag string) (enabled bool)
	// SetTag enables all of the given tags.
	SetTag(tags ...string)
	// UnsetTag disables all of the given tags.
	UnsetTag(tags ...string)
	// Tags returns a list of enabled tags.
	Tags() (tags []string)
	// TagCount returns the number of enabled tags.
	TagCount() (n int)
}

type taggable map[string]bool

func (t taggable) Tag(tag string) bool {
	return t[tag]
}

func (t taggable) SetTag(tags ...string) {
	for _, tag := range tags {
		t[tag] = true
	}
}

func (t taggable) UnsetTag(tags ...string) {
	for _, tag := range tags {
		delete(t, tag)
	}
}

func (t taggable) Tags() []string {
	a := make([]string, len(t))
	i := 0
	for tag := range t {
		a[i] = tag
		i++
	}
	sort.Strings(a)
	return a
}

func (t taggable) TagCount() int {
	return len(t)
}

// API represents the root of an API dump.
type API struct {
	Classes map[string]*Class
	Enums   map[string]*Enum
}

func NewAPI() *API {
	return &API{
		Classes: make(map[string]*Class),
		Enums:   make(map[string]*Enum),
	}
}

// ClassTree is a node in a tree representing the class hierarchy of an API.
// ClassTrees can be sorted by name with the sort package.
type ClassTree struct {
	// Name is the name of the class the node represents.
	Name string

	// Class is the Class that the node represents.
	Class *Class

	// Subclasses is a list containing each class that inherits from this
	// class.
	Subclasses []*ClassTree
}

func (t *ClassTree) Len() int {
	return len(t.Subclasses)
}
func (t *ClassTree) Less(i, j int) bool {
	return t.Subclasses[i].Name < t.Subclasses[j].Name
}
func (t *ClassTree) Swap(i, j int) {
	t.Subclasses[i], t.Subclasses[j] = t.Subclasses[j], t.Subclasses[i]
}

// list populates a list of classes that will be sorted by hierarchy.
func (t *ClassTree) list(l []*Class) []*Class {
	l = append(l, t.Class)
	for _, subt := range t.Subclasses {
		l = subt.list(l)
	}
	return l
}

// TreeList receives a list of ClassTree objects, and returns a flat list of
// the classes in the tree, ordered by the hierarchy of the tree.
func TreeList(tree []*ClassTree) (list []*Class) {
	for _, t := range tree {
		list = t.list(list)
	}
	return list
}

// ClassTree returns a list of ClassTree nodes. The list contains root nodes
// of the tree (classes that have no superclass). It also contains classes
// whose superclass does not exist.
func (api *API) ClassTree() []*ClassTree {
	nodes := make(map[string]*ClassTree)
	for name, class := range api.Classes {
		nodes[name] = &ClassTree{Name: name, Class: class}
	}

	root := &ClassTree{}
	for name, class := range api.Classes {
		tree := nodes[name]
		if stree, ok := nodes[class.Superclass]; ok {
			stree.Subclasses = append(stree.Subclasses, tree)
		} else {
			root.Subclasses = append(root.Subclasses, tree)
		}
	}

	for _, node := range nodes {
		sort.Sort(node)
	}
	sort.Sort(root)

	return root.Subclasses
}

// SortClassesByName sorts classes by name.
type SortClassesByName []*Class

func (l SortClassesByName) Len() int           { return len(l) }
func (l SortClassesByName) Less(i, j int) bool { return l[i].Name < l[j].Name }
func (l SortClassesByName) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// ClassList returns the classes of the API in a list sorted by name.
func (api *API) ClassList() []*Class {
	list := make([]*Class, len(api.Classes))
	i := 0
	for _, class := range api.Classes {
		list[i] = class
		i++
	}
	sort.Sort(SortClassesByName(list))
	return list
}

// SortEnumsByName sorts enums by name.
type SortEnumsByName []*Enum

func (l SortEnumsByName) Len() int           { return len(l) }
func (l SortEnumsByName) Less(i, j int) bool { return l[i].Name < l[j].Name }
func (l SortEnumsByName) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// EnumList returns the enums of the API in a list sorted by name.
func (api *API) EnumList() []*Enum {
	list := make([]*Enum, len(api.Enums))
	i := 0
	for _, enum := range api.Enums {
		list[i] = enum
		i++
	}
	sort.Sort(SortEnumsByName(list))
	return list
}

type Class struct {
	Name       string
	Superclass string
	Members    map[string]Member
	Taggable
}

func NewClass(name string) *Class {
	return &Class{
		Name:     name,
		Members:  make(map[string]Member),
		Taggable: make(taggable),
	}
}

func (class *Class) String() string {
	return "Class " + class.Name
}

// SortMembersByName sorts a list of members by name.
type SortMembersByName []Member

func (l SortMembersByName) Len() int           { return len(l) }
func (l SortMembersByName) Less(i, j int) bool { return l[i].Name() < l[j].Name() }
func (l SortMembersByName) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// SortMembersByType sorts a list of members by member type, then by name.
type SortMembersByType []Member

// Sort in reverse so that anything else (0) is sorted to the end of the list.
var memberTypeOrder = map[string]int{
	"Property":      5,
	"Function":      4,
	"YieldFunction": 3,
	"Event":         2,
	"Callback":      1,
}

func (l SortMembersByType) Len() int {
	return len(l)
}
func (l SortMembersByType) Less(i, j int) bool {
	mi, mj := l[i], l[j]
	oi := memberTypeOrder[mi.Type()]
	oj := memberTypeOrder[mj.Type()]
	if oi == oj {
		return mi.Name() < mj.Name()
	}
	return oi > oj
}
func (l SortMembersByType) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// MemberList returns the members of the class as a list, sorted by name.
func (class *Class) MemberList() []Member {
	a := make([]Member, len(class.Members))
	i := 0
	for _, member := range class.Members {
		a[i] = member
		i++
	}
	sort.Sort(SortMembersByName(a))
	return a
}

// Argument is a single argument of a Function, YieldFunction, Callback, or
// Event.
type Argument struct {
	// Type is the value type of the argument.
	Type string
	// Name is a name given to the argument.
	Name string
	// Default is an optional string that indicates the default value of the
	// argument. Only Functions and YieldFunctions may have a default value.
	Default *string
}

func (arg Argument) String() string {
	if arg.Default {
		return arg.Type + " " + arg.Name + " = " + *arg.Default
	}
	return arg.Type + " " + arg.Name
}

// Member is a single member of a Class.
type Member interface {
	// Type returns a string representation of the member type.
	Type() string
	// Name returns the name of the member.
	Name() string
	// Class returns the name of the class the member belongs to.
	Class() string
	// String returns a string representation of the member.
	String() string
	Taggable
}

type Property struct {
	MemberName  string
	MemberClass string
	ValueType   string
	Taggable
}

func NewProperty(name, class string) *Property {
	return &Property{
		MemberName:  name,
		MemberClass: class,
		Taggable:    make(taggable),
	}
}

func (m *Property) Type() string {
	return "Property"
}

func (m *Property) Name() string {
	return m.MemberName
}

func (m *Property) Class() string {
	return m.MemberClass
}

func (m *Property) String() string {
	return "Property " + m.MemberClass + "." + m.MemberName
}

type Function struct {
	MemberName  string
	MemberClass string
	ReturnType  string
	Arguments   []Argument
	Taggable
}

func NewFunction(name, class string) *Function {
	return &Function{
		MemberName:  name,
		MemberClass: class,
		Taggable:    make(taggable),
	}
}

func (m *Function) Type() string {
	return "Function"
}

func (m *Function) Name() string {
	return m.MemberName
}

func (m *Function) Class() string {
	return m.MemberClass
}

func (m *Function) String() string {
	return "Function " + m.MemberClass + "." + m.MemberName
}

type YieldFunction struct {
	MemberName  string
	MemberClass string
	ReturnType  string
	Arguments   []Argument
	Taggable
}

func NewYieldFunction(name, class string) *YieldFunction {
	return &YieldFunction{
		MemberName:  name,
		MemberClass: class,
		Taggable:    make(taggable),
	}
}

func (m *YieldFunction) Type() string {
	return "YieldFunction"
}

func (m *YieldFunction) Name() string {
	return m.MemberName
}

func (m *YieldFunction) Class() string {
	return m.MemberClass
}

func (m *YieldFunction) String() string {
	return "YieldFunction " + m.MemberClass + "." + m.MemberName
}

type Event struct {
	MemberName  string
	MemberClass string
	Arguments   []Argument
	Taggable
}

func NewEvent(name, class string) *Event {
	return &Event{
		MemberName:  name,
		MemberClass: class,
		Taggable:    make(taggable),
	}
}

func (m *Event) Type() string {
	return "Event"
}

func (m *Event) Name() string {
	return m.MemberName
}

func (m *Event) Class() string {
	return m.MemberClass
}

func (m *Event) String() string {
	return "Event " + m.MemberClass + "." + m.MemberName
}

type Callback struct {
	MemberName  string
	MemberClass string
	ReturnType  string
	Arguments   []Argument
	Taggable
}

func NewCallback(name, class string) *Callback {
	return &Callback{
		MemberName:  name,
		MemberClass: class,
		Taggable:    make(taggable),
	}
}

func (m *Callback) Type() string {
	return "Callback"
}

func (m *Callback) Name() string {
	return m.MemberName
}

func (m *Callback) Class() string {
	return m.MemberClass
}

func (m *Callback) String() string {
	return "Callback " + m.MemberClass + "." + m.MemberName
}

type Enum struct {
	Name  string
	Items []*EnumItem
	Taggable
}

func NewEnum(name string) *Enum {
	return &Enum{
		Name:     name,
		Taggable: make(taggable),
	}
}

func (enum *Enum) Item(name string) *EnumItem {
	for _, item := range enum.Items {
		if item.Name == name {
			return item
		}
	}
	return nil
}

func (enum *Enum) String() string {
	return "Enum " + enum.Name
}

type EnumItem struct {
	Enum  string
	Name  string
	Value int
	Taggable
}

func NewEnumItem(name, enum string) *EnumItem {
	return &EnumItem{
		Name:     name,
		Enum:     enum,
		Taggable: make(taggable),
	}
}

func (item *EnumItem) String() string {
	return "EnumItem " + item.Enum + "." + item.Name
}
