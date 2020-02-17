// The rbxdump package is used to serialize Roblox Lua API dumps.
package rbxdump

import "sort"

// Fields describes a set of names mapped to values.
type Fields map[string]interface{}

// Fielder is implemented by any value that can get and set its fields from a
// Fields value.
type Fielder interface {
	// Fields returns the set of fields present in the value.
	Fields() Fields
	// SetFields sets the fields of the value. Values must not be retained; they
	// should be copied if necessary. Invalid fields are ignored.
	SetFields(Fields)
}

// Root represents the top-level structure of an API.
type Root struct {
	Classes map[string]*Class
	Enums   map[string]*Enum
}

// GetClasses returns the Classes in the root as a slice, ordered by name.
func (root *Root) GetClasses() []*Class {
	list := make([]*Class, 0, len(root.Classes))
	for _, class := range root.Classes {
		list = append(list, class)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// GetEnums returns the Enums in the root as a slice, ordered by name.
func (root *Root) GetEnums() []*Enum {
	list := make([]*Enum, 0, len(root.Enums))
	for _, enum := range root.Enums {
		list = append(list, enum)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
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

func (root *Root) Fields() Fields {
	return Fields{
		"Classes": root.Classes,
		"Enums":   root.Enums,
	}
}

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

// Class represents a class descriptor.
type Class struct {
	Name           string
	Superclass     string
	MemoryCategory string
	Members        map[string]Member
	Tags
}

type Member interface {
	Fielder
	member()
	MemberType() string
	MemberName() string
	MemberCopy() Member
}

func MemberTypeOrder(memberType string) int {
	switch memberType {
	case "Property":
		return 0
	case "Function":
		return 1
	case "Event":
		return 2
	case "Callback":
		return 3
	}
	return -1
}

// GetMembers returns a list of members belonging to the class.
func (class *Class) GetMembers() []Member {
	list := make([]Member, 0, len(class.Members))
	for _, member := range class.Members {
		list = append(list, member)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].MemberName() < list[j].MemberName() })
	return list
}

// Copy returns a deep copy of the class descriptor.
//
// Copy implements the Class interface.
func (class *Class) Copy() *Class {
	cclass := *class
	cclass.Members = make(map[string]Member, len(class.Members))
	for name, member := range class.Members {
		cclass.Members[name] = member.MemberCopy()
	}
	cclass.Tags = class.GetTags()
	return &cclass
}

func (class *Class) Fields() Fields {
	return Fields{
		"Name":           class.Name,
		"Supere":         class.Superclass,
		"MemoryCategory": class.MemoryCategory,
		"Members":        class.Members,
		"Tags":           class.Tags,
	}
}

func (class *Class) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			class.Name = v
		}
	}
	if v, ok := fields["Superclass"]; ok {
		if v, ok := v.(string); ok {
			class.Superclass = v
		}
	}
	if v, ok := fields["MemoryCategory"]; ok {
		if v, ok := v.(string); ok {
			class.MemoryCategory = v
		}
	}
	if v, ok := fields["Members"]; ok {
		if v, ok := v.(map[string]Member); ok {
			class.Members = make(map[string]Member, len(v))
			for k, v := range v {
				class.Members[k] = v.MemberCopy()
			}
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			class.Tags = v.GetTags()
		}
	}
}

// Property represents a class member of the Property member type.
type Property struct {
	Name          string
	ValueType     Type
	Category      string
	ReadSecurity  string
	WriteSecurity string
	CanLoad       bool
	CanSave       bool
	Tags
}

// member implements the Member interface.
func (Property) member() {}

// MemberType returns a string indicating the the type of member.
func (member *Property) MemberType() string { return "Property" }

// MemberName returns the name of the member.
func (member *Property) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
func (member *Property) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the property.
func (member *Property) Copy() *Property {
	cmember := *member
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

func (member *Property) Fields() Fields {
	return Fields{
		"Name":          member.Name,
		"ValueType":     member.ValueType,
		"Category":      member.Category,
		"ReadSecurity":  member.ReadSecurity,
		"WriteSecurity": member.WriteSecurity,
		"CanLoad":       member.CanLoad,
		"CanSave":       member.CanSave,
		"Tags":          member.Tags,
	}
}

func (member *Property) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			member.Name = v
		}
	}
	if v, ok := fields["ValueType"]; ok {
		if v, ok := v.(Type); ok {
			member.ValueType = v
		}
	}
	if v, ok := fields["Category"]; ok {
		if v, ok := v.(string); ok {
			member.Category = v
		}
	}
	if v, ok := fields["ReadSecurity"]; ok {
		if v, ok := v.(string); ok {
			member.ReadSecurity = v
		}
	}
	if v, ok := fields["WriteSecurity"]; ok {
		if v, ok := v.(string); ok {
			member.WriteSecurity = v
		}
	}
	if v, ok := fields["CanLoad"]; ok {
		if v, ok := v.(bool); ok {
			member.CanLoad = v
		}
	}
	if v, ok := fields["CanSave"]; ok {
		if v, ok := v.(bool); ok {
			member.CanSave = v
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			member.Tags = v.GetTags()
		}
	}
}

// Function represents a class member of the Function member type.
type Function struct {
	Name       string
	Parameters []Parameter
	ReturnType Type
	Security   string
	Tags
}

// member implements the Member interface.
func (Function) member() {}

// MemberType returns a string indicating the the type of member.
func (member *Function) MemberType() string { return "Function" }

// MemberName returns the name of the member.
func (member *Function) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
func (member *Function) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the member descriptor.
func (member *Function) Copy() *Function {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

func (member *Function) Fields() Fields {
	return Fields{
		"Name":       member.Name,
		"Parameters": member.Parameters,
		"ReturnType": member.ReturnType,
		"Security":   member.Security,
		"Tags":       member.Tags,
	}
}

func (member *Function) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			member.Name = v
		}
	}
	if v, ok := fields["Parameters"]; ok {
		if v, ok := v.([]Parameter); ok {
			member.Parameters = CopyParams(v)
		}
	}
	if v, ok := fields["ReturnType"]; ok {
		if v, ok := v.(Type); ok {
			member.ReturnType = v
		}
	}
	if v, ok := fields["Security"]; ok {
		if v, ok := v.(string); ok {
			member.Security = v
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			member.Tags = v.GetTags()
		}
	}
}

// Event represents a class member of the Event member type.
type Event struct {
	Name       string
	Parameters []Parameter
	Security   string
	Tags
}

// member implements the Member interface.
func (Event) member() {}

// MemberType returns a string indicating the the type of member.
func (member *Event) MemberType() string { return "Event" }

// MemberName returns the name of the member.
func (member *Event) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
func (member *Event) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the member descriptor.
func (member *Event) Copy() *Event {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

func (member *Event) Fields() Fields {
	return Fields{
		"Name":       member.Name,
		"Parameters": member.Parameters,
		"Security":   member.Security,
		"Tags":       member.Tags,
	}
}

func (member *Event) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			member.Name = v
		}
	}
	if v, ok := fields["Parameters"]; ok {
		if v, ok := v.([]Parameter); ok {
			member.Parameters = CopyParams(v)
		}
	}
	if v, ok := fields["Security"]; ok {
		if v, ok := v.(string); ok {
			member.Security = v
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			member.Tags = v.GetTags()
		}
	}
}

// Callback represents a class member of the Callback member type.
type Callback struct {
	Name       string
	Parameters []Parameter
	ReturnType Type
	Security   string
	Tags
}

// member implements the Member interface.
func (Callback) member() {}

// MemberType returns a string indicating the the type of member.
func (member *Callback) MemberType() string { return "Callback" }

// MemberName returns the name of the member.
func (member *Callback) MemberName() string { return member.Name }

// MemberCopy returns a deep copy of the member.
func (member *Callback) MemberCopy() Member { return member.Copy() }

// Copy returns a deep copy of the member descriptor.
func (member *Callback) Copy() *Callback {
	cmember := *member
	cmember.Parameters = CopyParams(member.Parameters)
	cmember.Tags = Tags(member.GetTags())
	return &cmember
}

func (member *Callback) Fields() Fields {
	return Fields{
		"Name":       member.Name,
		"Parameters": member.Parameters,
		"ReturnType": member.ReturnType,
		"Security":   member.Security,
		"Tags":       member.Tags,
	}
}

func (member *Callback) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			member.Name = v
		}
	}
	if v, ok := fields["Parameters"]; ok {
		if v, ok := v.([]Parameter); ok {
			member.Parameters = CopyParams(v)
		}
	}
	if v, ok := fields["ReturnType"]; ok {
		if v, ok := v.(Type); ok {
			member.ReturnType = v
		}
	}
	if v, ok := fields["Security"]; ok {
		if v, ok := v.(string); ok {
			member.Security = v
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			member.Tags = v.GetTags()
		}
	}
}

// Enum represents an enum descriptor.
type Enum struct {
	Name  string
	Items map[string]*EnumItem
	Tags
}

// GetEnumItems returns a list of items of the enum.
func (enum *Enum) GetEnumItems() []*EnumItem {
	list := make([]*EnumItem, 0, len(enum.Items))
	for _, item := range enum.Items {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

// Copy returns a deep copy of the enum descriptor.
func (enum *Enum) Copy() *Enum {
	cenum := *enum
	cenum.Items = make(map[string]*EnumItem, len(enum.Items))
	for name, item := range enum.Items {
		cenum.Items[name] = item.Copy()
	}
	cenum.Tags = Tags(enum.GetTags())
	return &cenum
}

func (enum *Enum) Fields() Fields {
	return Fields{
		"Name":      enum.Name,
		"EnumItems": enum.Items,
		"Tags":      enum.Tags,
	}
}

func (enum *Enum) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			enum.Name = v
		}
	}
	if v, ok := fields["EnumItems"]; ok {
		if v, ok := v.(map[string]*EnumItem); ok {
			enum.Items = make(map[string]*EnumItem, len(v))
			for k, v := range v {
				enum.Items[k] = v.Copy()
			}
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			enum.Tags = v.GetTags()
		}
	}
}

// EnumItem represents an enum item descriptor.
type EnumItem struct {
	Name  string
	Value int
	Index int // Index determines the item's order among its sibling items.
	Tags
}

// Copy returns a deep copy of the enum item descriptor.
func (item *EnumItem) Copy() *EnumItem {
	citem := *item
	citem.Tags = Tags(item.GetTags())
	return &citem
}

func (item *EnumItem) Fields() Fields {
	return Fields{
		"Name":  item.Name,
		"Value": item.Value,
		"Tags":  item.Tags,
	}
}

func (item *EnumItem) SetFields(fields Fields) {
	if v, ok := fields["Name"]; ok {
		if v, ok := v.(string); ok {
			item.Name = v
		}
	}
	if v, ok := fields["Value"]; ok {
		if v, ok := v.(int); ok {
			item.Value = v
		}
	}
	if v, ok := fields["Tags"]; ok {
		if v, ok := v.(Tags); ok {
			item.Tags = v.GetTags()
		}
	}
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

// Tagger is implemented by any value that contains a set of tags.
type Tagger interface {
	// GetTag returns whether the given tag is present.
	GetTag() bool
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
