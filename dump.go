// The rbxdump package is used to represent Roblox Lua API dumps.
package rbxdump

import (
	"slices"
	"sort"
)

// Fields describes a set of names mapped to values.
type Fields map[string]interface{}

// Fielder is implemented by any value that can get and set its fields from a
// Fields map.
type Fielder interface {
	// Fields returns the set of fields present in the value. Values may be
	// retained by the implementation.
	Fields() Fields
	// SetFields sets the fields of the value. Values must not be retained; they
	// should be copied if necessary. Invalid fields are ignored.
	SetFields(Fields)
}

// Attempts to get a slice of T from v. get gets an individual element of type
// T.
//
// When v is a value of type T, it is converted to a single-element slice of
// that value.
func convertSlice[T any](u *[]T, v any, get func(u *T, v any) bool) bool {
	switch v := v.(type) {
	case nil:
		*u = nil
		return true
	case T:
		*u = []T{v}
		return true
	case *T:
		*u = []T{*v}
		return true
	case map[string]any:
		ts := [1]T{}
		if !get(&ts[0], v) {
			return false
		}
		*u = ts[:]
		return true
	case []T:
		*u = slices.Clone(v)
		return true
	case []any:
		return convertAnySlice(u, v, get)
	case []map[string]any:
		return convertAnySlice(u, v, get)
	}
	return false
}

// Attempts to get a slice of type T from a slice of type U. get gets an
// individual element of type T.
func convertAnySlice[U any, T any](u *[]T, v []U, get func(u *T, v any) bool) bool {
	ts := make([]T, len(v))
	for i, v := range v {
		var t T
		if !get(&t, v) {
			return false
		}
		ts[i] = t
	}
	*u = ts
	return true
}

// Attempts to get a Parameter from v.
func convertParameter(u *Parameter, v any) bool {
	switch v := v.(type) {
	case nil:
		*u = Parameter{}
		return true
	case Parameter:
		*u = v
		return true
	case *Parameter:
		*u = *v
		return true
	case map[string]any:
		var p Parameter
		var ok bool
		if !convertType(&p.Type, v["Type"]) {
			return false
		}
		if p.Name, ok = v["Name"].(string); !ok {
			return false
		}
		if p.Optional, ok = v["Optional"].(bool); !ok {
			return false
		}
		if p.Optional {
			if p.Default, ok = v["Default"].(string); !ok {
				return false
			}
		}
		*u = p
		return true
	}
	return false
}

func convertType(u *Type, v any) bool {
	switch v := v.(type) {
	case nil:
		*u = Type{}
		return true
	case Type:
		*u = v
		return true
	case *Type:
		*u = *v
		return true
	case map[string]any:
		var t Type
		var ok bool
		if t.Category, ok = v["Category"].(string); !ok {
			return false
		}
		if t.Name, ok = v["Name"].(string); !ok {
			return false
		}
		*u = t
		return true
	}
	return false
}

func getPreferredDescriptor(u *PreferredDescriptor, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	switch v := v.(type) {
	case nil:
		*u = PreferredDescriptor{}
		return true
	case PreferredDescriptor:
		*u = v
	case *PreferredDescriptor:
		*u = *v
	case map[string]any:
		var d PreferredDescriptor
		var ok bool
		if d.Name, ok = v["Name"].(string); !ok {
			return false
		}
		if d.ThreadSafety, ok = v["ThreadSafety"].(string); !ok {
			return false
		}
		*u = d
		return true
	}
	return false
}

// Attempts to get a slice of strings from v.
func convertStrings(u *[]string, v any) bool {
	switch v := v.(type) {
	case nil:
		*u = nil
	case []string:
		*u = slices.Clone(v)
		return true
	case []any:
		return convertAnySlice(u, v, func(u *string, v any) bool {
			if v, ok := v.(string); ok {
				*u = v
				return true
			}
			return false
		})
	}
	return false
}

// Attempts to get a value of type T from fields[key].
func getPrimitive[T any](u *T, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	switch v := v.(type) {
	case nil:
		var t T
		*u = t
		return true
	case T:
		*u = v
		return true
	}
	return false
}

type number interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Attempts to get a number of type T from number value fields[key].
func getNumber[T number](u *T, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	switch v := v.(type) {
	case nil:
		*u = T(0)
	case float32:
		*u = T(v)
	case float64:
		*u = T(v)
	case int16:
		*u = T(v)
	case int32:
		*u = T(v)
	case int64:
		*u = T(v)
	case int8:
		*u = T(v)
	case int:
		*u = T(v)
	case uint16:
		*u = T(v)
	case uint32:
		*u = T(v)
	case uint64:
		*u = T(v)
	case uint8:
		*u = T(v)
	case uint:
		*u = T(v)
	case uintptr:
		*u = T(v)
	default:
		return false
	}
	return true
}

// Attempts to get a slice of Parameters from fields[key].
func getParameters(u *[]Parameter, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	return convertSlice(u, v, convertParameter)
}

// Attempts to get a slice of Types from fields[key].
func getReturnType(u *[]Type, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	return convertSlice(u, v, convertType)
}

// Attempts to get a Type from fields[key].
func getType(u *Type, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	return convertType(u, v)
}

// Attempts to get a slice of strings from fields[key].
func getStrings(u *[]string, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	return convertStrings(u, v)
}

// Attempts to get Tags from fields[key].
func getTags(u *Tags, fields Fields, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	switch v := v.(type) {
	case nil:
		*u = nil
		return true
	case Tags:
		*u = v.GetTags()
		return true
	case []string:
		*u = slices.Clone(v)
		return true
	case []any:
		var s []string
		if !convertStrings(&s, v) {
			return false
		}
		*u = s
		return true
	}
	return false
}

// If pd is non-zero, then assign it in fields to "PreferredDescriptor".
func includePD(fields Fields, pd PreferredDescriptor) {
	if pd != (PreferredDescriptor{}) {
		fields["PreferredDescriptor"] = pd
	}
}

// Root represents the top-level structure of an API dump.
type Root struct {
	Classes map[string]*Class
	Enums   map[string]*Enum
}

// sortClasses sorts Class values by name.
type sortClasses []*Class

func (a sortClasses) Len() int           { return len(a) }
func (a sortClasses) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortClasses) Less(i, j int) bool { return a[i].Name < a[j].Name }

// GetClasses returns the Classes in the root as a slice, ordered by name.
func (root *Root) GetClasses() []*Class {
	list := make([]*Class, 0, len(root.Classes))
	for _, class := range root.Classes {
		list = append(list, class)
	}
	sort.Sort(sortClasses(list))
	return list
}

// sortEnums sorts Enum values by name.
type sortEnums []*Enum

func (a sortEnums) Len() int           { return len(a) }
func (a sortEnums) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortEnums) Less(i, j int) bool { return a[i].Name < a[j].Name }

// GetEnums returns the Enums in the root as a slice, ordered by name.
func (root *Root) GetEnums() []*Enum {
	list := make([]*Enum, 0, len(root.Enums))
	for _, enum := range root.Enums {
		list = append(list, enum)
	}
	sort.Sort(sortEnums(list))
	return list
}

// Copy returns a deep copy of the root.
func (root *Root) Copy() *Root {
	croot := &Root{
		Classes: make(map[string]*Class, len(root.Classes)),
		Enums:   make(map[string]*Enum, len(root.Enums)),
	}
	for name, class := range root.Classes {
		croot.Classes[name] = class.Copy()
	}
	for name, enum := range root.Enums {
		croot.Enums[name] = enum.Copy()
	}
	return croot
}

// Fields implements the Fielder interface.
func (root *Root) Fields() Fields {
	return Fields{
		"Classes": root.Classes,
		"Enums":   root.Enums,
	}
}

// SetFields implements the Fielder interface.
func (root *Root) SetFields(fields Fields) {
	if v, ok := fields["Classes"]; ok {
		if v, ok := v.(map[string]*Class); ok {
			root.Classes = make(map[string]*Class, len(v))
			for k, v := range v {
				root.Classes[k] = v.Copy()
			}
		}
	}
	if v, ok := fields["Enums"]; ok {
		if v, ok := v.(map[string]*Enum); ok {
			root.Enums = make(map[string]*Enum, len(v))
			for k, v := range v {
				root.Enums[k] = v.Copy()
			}
		}
	}
}

// Class represents a class defined in an API dump.
type Class struct {
	Name                string
	Superclass          string
	MemoryCategory      string
	Members             map[string]Member
	PreferredDescriptor PreferredDescriptor
	Tags
}

// Member represents a member of a Class.
type Member interface {
	Fielder
	Tagger
	// member prevents external types from implementing the interface.
	member()
	// MemberType returns a string indicating the type of member.
	MemberType() string
	// MemberName returns the name of the member.
	MemberName() string
	// MemberCopy returns a deep copy of the member.
	MemberCopy() Member
}

// sortMembers sorts Member values by name.
type sortMembers []Member

func (a sortMembers) Len() int           { return len(a) }
func (a sortMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortMembers) Less(i, j int) bool { return a[i].MemberName() < a[j].MemberName() }

// GetMembers returns a list of members belonging to the class, sorted by name.
func (class *Class) GetMembers() []Member {
	list := make([]Member, 0, len(class.Members))
	for _, member := range class.Members {
		list = append(list, member)
	}
	sort.Sort(sortMembers(list))
	return list
}

// Copy returns a deep copy of the class.
func (class *Class) Copy() *Class {
	cclass := *class
	cclass.Members = make(map[string]Member, len(class.Members))
	for name, member := range class.Members {
		cclass.Members[name] = member.MemberCopy()
	}
	cclass.Tags = class.GetTags()
	return &cclass
}

// Fields implements the Fielder interface. Does not return the Members field.
func (class *Class) Fields() Fields {
	fields := Fields{
		"Name":           class.Name,
		"Superclass":     class.Superclass,
		"MemoryCategory": class.MemoryCategory,
		"Tags":           class.Tags,
	}
	includePD(fields, class.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (class *Class) SetFields(fields Fields) {
	getPrimitive(&class.Name, fields, "Name")
	getPrimitive(&class.Superclass, fields, "Superclass")
	getPrimitive(&class.MemoryCategory, fields, "MemoryCategory")
	getPreferredDescriptor(&class.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&class.Tags, fields, "Tags")
	if v, ok := fields["Members"]; ok {
		if v, ok := v.(map[string]Member); ok {
			class.Members = make(map[string]Member, len(v))
			for k, v := range v {
				class.Members[k] = v.MemberCopy()
			}
		}
	}
}

// Property is a Member that represents a class property.
type Property struct {
	Name                string
	ValueType           Type
	Category            string
	ReadSecurity        string
	WriteSecurity       string
	CanLoad             bool
	CanSave             bool
	ThreadSafety        string
	PreferredDescriptor PreferredDescriptor
	Tags
}

// member implements the Member interface.
func (Property) member() {}

// MemberType returns a string indicating the the type of member.
//
// MemberType implements the Member interface.
func (member *Property) MemberType() string { return "Property" }

// MemberName returns the name of the member.
//
// MemberType implements the Member interface.
func (member *Property) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
//
// MemberType implements the Member interface.
func (member *Property) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the property.
func (member *Property) Copy() *Property {
	cmember := *member
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

// Fields implements the Fielder interface.
func (member *Property) Fields() Fields {
	fields := Fields{
		"Name":          member.Name,
		"ValueType":     member.ValueType,
		"Category":      member.Category,
		"ReadSecurity":  member.ReadSecurity,
		"WriteSecurity": member.WriteSecurity,
		"CanLoad":       member.CanLoad,
		"CanSave":       member.CanSave,
		"ThreadSafety":  member.ThreadSafety,
		"Tags":          member.Tags,
	}
	includePD(fields, member.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (member *Property) SetFields(fields Fields) {
	getPrimitive(&member.Name, fields, "Name")
	getType(&member.ValueType, fields, "ValueType")
	getPrimitive(&member.Category, fields, "Category")
	getPrimitive(&member.ReadSecurity, fields, "ReadSecurity")
	getPrimitive(&member.WriteSecurity, fields, "WriteSecurity")
	getPrimitive(&member.CanLoad, fields, "CanLoad")
	getPrimitive(&member.CanSave, fields, "CanSave")
	getPrimitive(&member.ThreadSafety, fields, "ThreadSafety")
	getPreferredDescriptor(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&member.Tags, fields, "Tags")
}

// Function is a Member that represents a class function.
type Function struct {
	Name                string
	Parameters          []Parameter
	ReturnType          []Type
	Security            string
	ThreadSafety        string
	PreferredDescriptor PreferredDescriptor
	Tags
}

// member implements the Member interface.
func (Function) member() {}

// MemberType returns a string indicating the the type of member.
//
// MemberType implements the Member interface.
func (member *Function) MemberType() string { return "Function" }

// MemberName returns the name of the member.
//
// MemberType implements the Member interface.
func (member *Function) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
//
// MemberType implements the Member interface.
func (member *Function) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the function.
func (member *Function) Copy() *Function {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

// Fields implements the Fielder interface.
func (member *Function) Fields() Fields {
	fields := Fields{
		"Name":         member.Name,
		"Parameters":   member.Parameters,
		"ReturnType":   member.ReturnType,
		"Security":     member.Security,
		"ThreadSafety": member.ThreadSafety,
		"Tags":         member.Tags,
	}
	includePD(fields, member.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (member *Function) SetFields(fields Fields) {
	getPrimitive(&member.Name, fields, "Name")
	getParameters(&member.Parameters, fields, "Parameters")
	getReturnType(&member.ReturnType, fields, "ReturnType")
	getPrimitive(&member.Security, fields, "Security")
	getPrimitive(&member.ThreadSafety, fields, "ThreadSafety")
	getPreferredDescriptor(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&member.Tags, fields, "Tags")
}

// Event is a Member that represents a class event.
type Event struct {
	Name                string
	Parameters          []Parameter
	Security            string
	ThreadSafety        string
	PreferredDescriptor PreferredDescriptor
	Tags
}

// member implements the Member interface.
func (Event) member() {}

// MemberType returns a string indicating the the type of member.
//
// MemberType implements the Member interface.
func (member *Event) MemberType() string { return "Event" }

// MemberName returns the name of the member.
//
// MemberType implements the Member interface.
func (member *Event) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
//
// MemberType implements the Member interface.
func (member *Event) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the event.
func (member *Event) Copy() *Event {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

// Fields implements the Fielder interface.
func (member *Event) Fields() Fields {
	fields := Fields{
		"Name":         member.Name,
		"Parameters":   member.Parameters,
		"Security":     member.Security,
		"ThreadSafety": member.ThreadSafety,
		"Tags":         member.Tags,
	}
	includePD(fields, member.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (member *Event) SetFields(fields Fields) {
	getPrimitive(&member.Name, fields, "Name")
	getParameters(&member.Parameters, fields, "Parameters")
	getPrimitive(&member.Security, fields, "Security")
	getPrimitive(&member.ThreadSafety, fields, "ThreadSafety")
	getPreferredDescriptor(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&member.Tags, fields, "Tags")
}

// Callback is a Member that represents a class callback.
type Callback struct {
	Name                string
	Parameters          []Parameter
	ReturnType          []Type
	Security            string
	ThreadSafety        string
	PreferredDescriptor PreferredDescriptor
	Tags
}

// member implements the Member interface.
func (Callback) member() {}

// MemberType returns a string indicating the the type of member.
//
// MemberType implements the Member interface.
func (member *Callback) MemberType() string { return "Callback" }

// MemberName returns the name of the member.
//
// MemberType implements the Member interface.
func (member *Callback) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
//
// MemberType implements the Member interface.
func (member *Callback) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the callback.
func (member *Callback) Copy() *Callback {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

// Fields implements the Fielder interface.
func (member *Callback) Fields() Fields {
	fields := Fields{
		"Name":         member.Name,
		"Parameters":   member.Parameters,
		"ReturnType":   member.ReturnType,
		"Security":     member.Security,
		"ThreadSafety": member.ThreadSafety,
		"Tags":         member.Tags,
	}
	includePD(fields, member.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (member *Callback) SetFields(fields Fields) {
	getPrimitive(&member.Name, fields, "Name")
	getParameters(&member.Parameters, fields, "Parameters")
	getReturnType(&member.ReturnType, fields, "ReturnType")
	getPrimitive(&member.Security, fields, "Security")
	getPrimitive(&member.ThreadSafety, fields, "ThreadSafety")
	getPreferredDescriptor(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&member.Tags, fields, "Tags")
}

// Enum represents an enum defined in an API dump.
type Enum struct {
	Name                string
	Items               map[string]*EnumItem
	PreferredDescriptor PreferredDescriptor
	Tags
}

// sortEnumItems sorts Member values by Index, then Value, then Name.
type sortEnumItems []*EnumItem

func (a sortEnumItems) Len() int      { return len(a) }
func (a sortEnumItems) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortEnumItems) Less(i, j int) bool {
	if a[i].Index == a[j].Index {
		if a[i].Name == a[j].Name {
			return a[i].Value < a[j].Value
		}
		return a[i].Name < a[j].Name
	}
	return a[i].Index < a[j].Index
}

// GetEnumItems returns a list of items of the enum.
func (enum *Enum) GetEnumItems() []*EnumItem {
	list := make([]*EnumItem, 0, len(enum.Items))
	for _, item := range enum.Items {
		list = append(list, item)
	}
	sort.Sort(sortEnumItems(list))
	return list
}

// Copy returns a deep copy of the enum.
func (enum *Enum) Copy() *Enum {
	cenum := *enum
	cenum.Items = make(map[string]*EnumItem, len(enum.Items))
	for name, item := range enum.Items {
		cenum.Items[name] = item.Copy()
	}
	cenum.Tags = Tags(enum.GetTags())
	return &cenum
}

// Fields implements the Fielder interface. Does not return the Items field.
func (enum *Enum) Fields() Fields {
	fields := Fields{
		"Name": enum.Name,
		"Tags": enum.Tags,
	}
	includePD(fields, enum.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (enum *Enum) SetFields(fields Fields) {
	getPrimitive(&enum.Name, fields, "Name")
	getPreferredDescriptor(&enum.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&enum.Tags, fields, "Tags")
	if v, ok := fields["Items"]; ok {
		if v, ok := v.(map[string]*EnumItem); ok {
			enum.Items = make(map[string]*EnumItem, len(v))
			for k, v := range v {
				enum.Items[k] = v.Copy()
			}
		}
	}
}

// EnumItem represents an enum item.
type EnumItem struct {
	Name                string
	Value               int
	Index               int // Index determines the item's order among its sibling items.
	LegacyNames         []string
	PreferredDescriptor PreferredDescriptor
	Tags
}

// Copy returns a deep copy of the enum item.
func (item *EnumItem) Copy() *EnumItem {
	citem := *item
	citem.Tags = Tags(item.GetTags())
	citem.LegacyNames = slices.Clone(item.LegacyNames)
	return &citem
}

// Fields implements the Fielder interface.
func (item *EnumItem) Fields() Fields {
	fields := Fields{
		"Name":        item.Name,
		"Value":       item.Value,
		"Index":       item.Index,
		"LegacyNames": item.LegacyNames,
		"Tags":        item.Tags,
	}
	includePD(fields, item.PreferredDescriptor)
	return fields
}

// SetFields implements the Fielder interface.
func (item *EnumItem) SetFields(fields Fields) {
	getPrimitive(&item.Name, fields, "Name")
	getNumber(&item.Value, fields, "Value")
	getNumber(&item.Index, fields, "Index")
	getStrings(&item.LegacyNames, fields, "LegacyNames")
	getPreferredDescriptor(&item.PreferredDescriptor, fields, "PreferredDescriptor")
	getTags(&item.Tags, fields, "Tags")
}

// Parameter represents a parameter of a function, event, or callback member.
type Parameter struct {
	Type     Type
	Name     string
	Optional bool
	Default  string
}

// CopyParams returns a copy of the given parameters.
func CopyParams(p []Parameter) []Parameter {
	c := make([]Parameter, len(p))
	copy(c, p)
	return c
}

// Type represents a value type.
type Type struct {
	Category string
	Name     string
}

// String returns a string representation of the type.
func (typ Type) String() string {
	if typ.Category == "" {
		return typ.Name
	}
	return typ.Category + ":" + typ.Name
}

// PreferredDescriptor refers to an alternative of a deprecated descriptor.
type PreferredDescriptor struct {
	// The name of the alternative descriptor.
	Name         string
	ThreadSafety string
}

// Tagger is implemented by any value that contains a set of tags.
type Tagger interface {
	// GetTag returns whether the given tag is present.
	GetTag(tag string) bool
	// GetTags returns a list of all present tags. Implementations must not
	// retain the result.
	GetTags() []string
	// SetTag adds one or more tags. Duplicate tags are removed.
	SetTag(tag ...string)
	// UnsetTag removes one or more tags. Duplicate tags are removed.
	UnsetTag(tag ...string)
}

// Tags implements a Tagger, to be embedded by taggable elements.
type Tags []string

// GetTag returns whether the given tag is present.
func (tags Tags) GetTag(tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetTags returns a list of all present tags.
func (tags Tags) GetTags() []string {
	list := make([]string, len(tags))
	copy(list, tags)
	return list
}

// SetTag adds one or more tags. Duplicate tags are removed.
func (tags *Tags) SetTag(tag ...string) {
	*tags = append(*tags, tag...)
loop:
	for i, n := 1, len(*tags); i < n; {
		for j := 0; j < i; j++ {
			if (*tags)[i] == (*tags)[j] {
				*tags = append((*tags)[:i], (*tags)[i+1:]...)
				n--
				continue loop
			}
		}
		i++
	}
}

// UnsetTag removes one or more tags. Duplicate tags are removed.
func (tags *Tags) UnsetTag(tag ...string) {
loop:
	for i, n := 0, len(*tags); i < n; {
		for j := 0; j < len(tag); j++ {
			if (*tags)[i] == tag[j] {
				*tags = append((*tags)[:i], (*tags)[i+1:]...)
				n--
				continue loop
			}
		}
		i++
	}
}
