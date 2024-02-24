// The rbxdump package is used to represent Roblox Lua API dumps.
package rbxdump

import (
	"slices"
	"sort"
)

// Fields describes a set of names mapped to values, for the purpose of diffing
// and patching fields of an element. Entries may be set and deleted, but values
// must not be modified.
type Fields map[string]any

// Fielder is implemented by any value that can get and set its fields from a
// Fields map.
type Fielder interface {
	// Fields populates f with the fields of the value that are present in f.
	// Fields that are not present in the value are deleted from f. If f is nil,
	// then a new Fields is created and populated with all fields. Returns the
	// populated Fields.
	//
	// Must include only fields that are expected to be diffed and patched.
	//
	// Implementations may retain returned values.
	Fields(f Fields) Fields
	// SetFields receives fields from f and sets them on the value. Irrelevant
	// fields are ignored.
	//
	// If possible, SetFields may convert any type to the field's canonical
	// type. At a minimu, the implemntation must convert the field's type, and a
	// general interface representation, according to the json package.
	//
	// Implementations must not retain received values; they should be copied if
	// necessary.
	SetFields(f Fields)
}

// Attempts to assign fields[name] to v. If the field is not present, then no
// assignment is made. If the field is nil, then zero is assigned. Otherwise,
// the field is assigned if it is a T or pointer to T. Returns whether the
// assignment was successful.
func normalize[T any](v *T, fields Fields, name string) bool {
	u, ok := fields[name]
	if !ok {
		return false
	}
	return convert(v, u)
}

func convert[T any](v *T, u any) bool {
	switch u := u.(type) {
	case nil:
		var z T
		*v = z
		return true
	case T:
		*v = u
		return true
	case *T:
		*v = *u
		return true
	}
	return false
}

// Any numeric value.
type number interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Attempts to assign fields[name] to v. If the field is not present, then no
// assignment is made. If the field is nil, then zero is assigned. Otherwise,
// the field is assigned if it is a number. Returns whether the assignment was
// successful.
func normalizeNumber[T number](v *T, fields Fields, name string) bool {
	u, ok := fields[name]
	if !ok {
		return false
	}
	return convertNumber(v, u)
}

func convertNumber[T number](v *T, u any) bool {
	switch u := u.(type) {
	case nil:
		*v = T(0)
	case float32:
		*v = T(u)
	case float64:
		*v = T(u)
	case int16:
		*v = T(u)
	case int32:
		*v = T(u)
	case int64:
		*v = T(u)
	case int8:
		*v = T(u)
	case int:
		*v = T(u)
	case uint16:
		*v = T(u)
	case uint32:
		*v = T(u)
	case uint64:
		*v = T(u)
	case uint8:
		*v = T(u)
	case uint:
		*v = T(u)
	case uintptr:
		*v = T(u)
	default:
		return false
	}
	return true
}

// Receives any value and attempts to convert it to the value. Returns success.
type normalizer interface {
	normalize(any) bool
}

// Attempts to assign fields[name] to v. If the field is not present, then no
// assignment is made. Otherwise, the field is assigned according it its
// normalize method. Returns whether the assignment was successful.
func normalizeType[T normalizer](v T, fields Fields, name string) bool {
	f, ok := fields[name]
	if !ok {
		return false
	}
	return convertType(v, f)
}

func convertType[T normalizer](v T, u any) bool {
	return v.normalize(u)
}

func normalizeSlice[T any](v *[]T, fields Fields, name string, get func(v *T, u any) bool) bool {
	u, ok := fields[name]
	if !ok {
		return false
	}
	return convertSlice(v, u, get)
}

func convertSlice[T any](v *[]T, u any, get func(v *T, u any) bool) bool {
	switch u := u.(type) {
	case nil:
		*v = nil
		return true
	case T:
		*v = []T{u}
		return true
	case *T:
		*v = []T{*u}
		return true
	case map[string]any:
		var ts [1]T
		if !get(&ts[0], u) {
			return false
		}
		*v = ts[:]
	case []T:
		*v = slices.Clone(u)
		return true
	case []any:
		return convertAnySlice(v, u, get)
	case []map[string]any:
		return convertAnySlice(v, u, get)
	}
	return false
}

func convertAnySlice[U any, T any](v *[]T, u []U, get func(v *T, u any) bool) bool {
	ts := make([]T, len(u))
	for i, u := range u {
		if !get(&ts[i], u) {
			return false
		}
	}
	*v = ts
	return true
}

func normalizeParameters(v *[]Parameter, fields Fields, name string) bool {
	return normalizeSlice(v, fields, name, func(v *Parameter, u any) bool {
		return (*v).normalize(u)
	})
}

func normalizeReturnType(v *[]Type, fields Fields, name string) bool {
	return normalizeSlice(v, fields, name, func(v *Type, u any) bool {
		return (*v).normalize(u)
	})
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

// Fields implements the Fielder interface.
func (class *Class) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["Superclass"] = class.Superclass
		fields["MemoryCategory"] = class.MemoryCategory
		fields["PreferredDescriptor"] = class.PreferredDescriptor
		fields["Tags"] = class.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "Superclass":
			fields[name] = class.Superclass
		case "MemoryCategory":
			fields[name] = class.MemoryCategory
		case "PreferredDescriptor":
			fields["PreferredDescriptor"] = class.PreferredDescriptor
		case "Tags":
			fields["Tags"] = class.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (class *Class) SetFields(fields Fields) {
	normalize[string](&class.Superclass, fields, "Superclass")
	normalize[string](&class.MemoryCategory, fields, "MemoryCategory")
	normalizeType(&class.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&class.Tags, fields, "Tags")
}

// Property is a Member that represents a class property.
type Property struct {
	Name                string
	ValueType           Type
	Default             string
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

func (member *Property) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["ValueType"] = member.ValueType
		fields["Default"] = member.Default
		fields["Category"] = member.Category
		fields["ReadSecurity"] = member.ReadSecurity
		fields["WriteSecurity"] = member.WriteSecurity
		fields["CanLoad"] = member.CanLoad
		fields["CanSave"] = member.CanSave
		fields["ThreadSafety"] = member.ThreadSafety
		fields["PreferredDescriptor"] = member.PreferredDescriptor
		fields["Tags"] = member.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "ValueType":
			fields[name] = member.ValueType
		case "Default":
			fields[name] = member.Default
		case "Category":
			fields[name] = member.Category
		case "ReadSecurity":
			fields[name] = member.ReadSecurity
		case "WriteSecurity":
			fields[name] = member.WriteSecurity
		case "CanLoad":
			fields[name] = member.CanLoad
		case "CanSave":
			fields[name] = member.CanSave
		case "ThreadSafety":
			fields[name] = member.ThreadSafety
		case "PreferredDescriptor":
			fields[name] = member.PreferredDescriptor
		case "Tags":
			fields[name] = member.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

func (member *Property) SetFields(fields Fields) {
	normalizeType(&member.ValueType, fields, "ValueType")
	normalize[string](&member.Default, fields, "Default")
	normalize[string](&member.Category, fields, "Category")
	normalize[string](&member.ReadSecurity, fields, "ReadSecurity")
	normalize[string](&member.WriteSecurity, fields, "WriteSecurity")
	normalize[bool](&member.CanLoad, fields, "CanLoad")
	normalize[bool](&member.CanSave, fields, "CanSave")
	normalize[string](&member.ThreadSafety, fields, "ThreadSafety")
	normalizeType(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&member.Tags, fields, "Tags")
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
func (member *Function) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["Parameters"] = member.Parameters
		fields["ReturnType"] = member.ReturnType
		fields["Security"] = member.Security
		fields["ThreadSafety"] = member.ThreadSafety
		fields["PreferredDescriptor"] = member.PreferredDescriptor
		fields["Tags"] = member.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "Parameters":
			fields["Parameters"] = member.Parameters
		case "ReturnType":
			fields["ReturnType"] = member.ReturnType
		case "Security":
			fields["Security"] = member.Security
		case "ThreadSafety":
			fields["ThreadSafety"] = member.ThreadSafety
		case "PreferredDescriptor":
			fields["PreferredDescriptor"] = member.PreferredDescriptor
		case "Tags":
			fields["Tags"] = member.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (member *Function) SetFields(fields Fields) {
	normalizeParameters(&member.Parameters, fields, "Parameters")
	normalizeReturnType(&member.ReturnType, fields, "ReturnType")
	normalize[string](&member.Security, fields, "Security")
	normalize[string](&member.ThreadSafety, fields, "ThreadSafety")
	normalizeType(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&member.Tags, fields, "Tags")
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
func (member *Event) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["Parameters"] = member.Parameters
		fields["Security"] = member.Security
		fields["ThreadSafety"] = member.ThreadSafety
		fields["PreferredDescriptor"] = member.PreferredDescriptor
		fields["Tags"] = member.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "Parameters":
			fields["Parameters"] = member.Parameters
		case "Security":
			fields["Security"] = member.Security
		case "ThreadSafety":
			fields["ThreadSafety"] = member.ThreadSafety
		case "PreferredDescriptor":
			fields["PreferredDescriptor"] = member.PreferredDescriptor
		case "Tags":
			fields["Tags"] = member.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (member *Event) SetFields(fields Fields) {
	normalizeParameters(&member.Parameters, fields, "Parameters")
	normalize[string](&member.Security, fields, "Security")
	normalize[string](&member.ThreadSafety, fields, "ThreadSafety")
	normalizeType(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&member.Tags, fields, "Tags")
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
func (member *Callback) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["Parameters"] = member.Parameters
		fields["ReturnType"] = member.ReturnType
		fields["Security"] = member.Security
		fields["ThreadSafety"] = member.ThreadSafety
		fields["PreferredDescriptor"] = member.PreferredDescriptor
		fields["Tags"] = member.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "Parameters":
			fields["Parameters"] = member.Parameters
		case "ReturnType":
			fields["ReturnType"] = member.ReturnType
		case "Security":
			fields["Security"] = member.Security
		case "ThreadSafety":
			fields["ThreadSafety"] = member.ThreadSafety
		case "PreferredDescriptor":
			fields["PreferredDescriptor"] = member.PreferredDescriptor
		case "Tags":
			fields["Tags"] = member.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (member *Callback) SetFields(fields Fields) {
	normalizeParameters(&member.Parameters, fields, "Parameters")
	normalizeReturnType(&member.ReturnType, fields, "ReturnType")
	normalize[string](&member.Security, fields, "Security")
	normalize[string](&member.ThreadSafety, fields, "ThreadSafety")
	normalizeType(&member.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&member.Tags, fields, "Tags")
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
func (enum *Enum) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["PreferredDescriptor"] = enum.PreferredDescriptor
		fields["Tags"] = enum.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "PreferredDescriptor":
			fields["PreferredDescriptor"] = enum.PreferredDescriptor
		case "Tags":
			fields["Tags"] = enum.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (enum *Enum) SetFields(fields Fields) {
	normalizeType(&enum.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&enum.Tags, fields, "Tags")
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
func (item *EnumItem) Fields(fields Fields) Fields {
	if fields == nil {
		fields = Fields{}
		fields["Value"] = item.Value
		fields["Index"] = item.Index
		fields["LegacyNames"] = item.LegacyNames
		fields["PreferredDescriptor"] = item.PreferredDescriptor
		fields["Tags"] = item.Tags
		return fields
	}
	for name := range fields {
		switch name {
		case "Value":
			fields["Value"] = item.Value
		case "Index":
			fields["Index"] = item.Index
		case "LegacyNames":
			fields["LegacyNames"] = item.LegacyNames
		case "PreferredDescriptor":
			fields[name] = item.PreferredDescriptor
		case "Tags":
			fields[name] = item.Tags
		default:
			delete(fields, name)
		}
	}
	return fields
}

// SetFields implements the Fielder interface.
func (item *EnumItem) SetFields(fields Fields) {
	normalizeNumber(&item.Value, fields, "Value")
	normalize(&item.Index, fields, "Index")
	normalizeSlice(&item.LegacyNames, fields, "LegacyNames", convert)
	normalizeType(&item.PreferredDescriptor, fields, "PreferredDescriptor")
	normalizeType(&item.Tags, fields, "Tags")
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

func (p *Parameter) normalize(u any) bool {
	switch u := u.(type) {
	case nil:
		*p = Parameter{}
		return true
	case Parameter:
		*p = u
		return true
	case *Parameter:
		*p = *u
		return true
	case map[string]any:
		var v Parameter
		if !convertType(&v.Type, u["Type"]) {
			return false
		}
		if !convert(&v.Name, u["Name"]) {
			return false
		}
		if !convert(&v.Optional, u["Optional"]) {
			return false
		}
		if v.Optional {
			if !convert(&v.Default, u["Default"]) {
				return false
			}
		}
		*p = v
		return true
	}
	return false
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

func (t *Type) normalize(u any) bool {
	switch u := u.(type) {
	case nil:
		*t = Type{}
		return true
	case Type:
		*t = u
		return true
	case *Type:
		*t = *u
		return true
	case map[string]any:
		var v Type
		if !convert(&v.Category, u["Category"]) {
			return false
		}
		if !convert(&v.Name, u["Name"]) {
			return false
		}
		*t = v
		return true
	}
	return false
}

// PreferredDescriptor refers to an alternative of a deprecated descriptor.
type PreferredDescriptor struct {
	// The name of the alternative descriptor.
	Name         string
	ThreadSafety string
}

func (p *PreferredDescriptor) normalize(u any) bool {
	switch u := u.(type) {
	case nil:
		*p = PreferredDescriptor{}
		return true
	case PreferredDescriptor:
		*p = u
		return true
	case *PreferredDescriptor:
		*p = *u
		return true
	case map[string]any:
		var v PreferredDescriptor
		if !convert(&v.Name, u["Name"]) {
			return false
		}
		if !convert(&v.ThreadSafety, u["ThreadSafety"]) {
			return false
		}
		*p = v
		return true
	}
	return false
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

func (tags *Tags) normalize(u any) bool {
	switch u := u.(type) {
	case nil:
		*tags = nil
		return true
	case Tags:
		*tags = u.GetTags()
		return true
	case []string:
		*tags = slices.Clone(u)
		return true
	case []any:
		return convertAnySlice((*[]string)(tags), u, convert)
	}
	return false
}
