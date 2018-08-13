// The rbxapijson package implements the rbxapi package as a codec for the
// Roblox API dump in JSON format.
//
// Note that this package is an implementation of a non-standardized format.
// Therefore, this package's API is subject to change as the format changes.
package rbxapijson

import (
	"github.com/robloxapi/rbxapi"
)

// Root represents the top-level structure of an API.
type Root struct {
	Classes []*Class
	Enums   []*Enum
}

// GetClasses returns a list of class descriptors present in the API.
//
// GetClasses implements the rbxapi.Root interface.
func (root *Root) GetClasses() []rbxapi.Class {
	list := make([]rbxapi.Class, len(root.Classes))
	for i, class := range root.Classes {
		list[i] = class
	}
	return list
}

// GetClass returns the first class descriptor of the given name, or nil if no
// class of the given name is present.
//
// GetClass implements the rbxapi.Root interface.
func (root *Root) GetClass(name string) rbxapi.Class {
	for _, class := range root.Classes {
		if class.Name == name {
			return class
		}
	}
	return nil
}

// GetEnums returns a list of enum descriptors present in the API.
//
// GetEnums implements the rbxapi.Root interface.
func (root *Root) GetEnums() []rbxapi.Enum {
	list := make([]rbxapi.Enum, len(root.Enums))
	for i, enum := range root.Enums {
		list[i] = enum
	}
	return list
}

// GetEnum returns the first enum descriptor of the given name, or nil if no
// enum of the given name is present.
//
// GetEnum implements the rbxapi.Root interface.
func (root *Root) GetEnum(name string) rbxapi.Enum {
	for _, enum := range root.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
}

// Copy returns a deep copy of the API structure.
//
// Copy implements the rbxapi.Root interface.
func (root *Root) Copy() rbxapi.Root {
	croot := &Root{
		Classes: make([]*Class, len(root.Classes)),
		Enums:   make([]*Enum, len(root.Enums)),
	}
	for i, class := range root.Classes {
		croot.Classes[i] = class.Copy().(*Class)
	}
	for i, enum := range root.Enums {
		croot.Enums[i] = enum.Copy().(*Enum)
	}
	return croot
}

// Class represents a class descriptor.
type Class struct {
	Name           string
	Superclass     string
	MemoryCategory string
	Members        []rbxapi.Member
	Tags           `json:",omitempty"`
}

// GetName returns the class name.
//
// GetName implements the rbxapi.Class interface.
func (class *Class) GetName() string {
	return class.Name
}

// GetSuperclass returns the name of the class that this class inherits from.
//
// GetSuperclass implements the rbxapi.Class interface.
func (class *Class) GetSuperclass() string {
	return class.Superclass
}

// GetMembers returns a list of member descriptors belonging to the class.
//
// GetMembers implements the rbxapi.Class interface.
func (class *Class) GetMembers() []rbxapi.Member {
	list := make([]rbxapi.Member, len(class.Members))
	copy(list, class.Members)
	return list
}

// GetMember returns the first member descriptor of the given name, or nil if
// no member of the given name is present.
//
// GetMember implements the rbxapi.Class interface.
func (class *Class) GetMember(name string) rbxapi.Member {
	for _, member := range class.Members {
		if member.GetName() == name {
			return member
		}
	}
	return nil
}

// Copy returns a deep copy of the class descriptor.
//
// Copy implements the rbxapi.Class interface.
func (class *Class) Copy() rbxapi.Class {
	cclass := *class
	cclass.Members = make([]rbxapi.Member, len(class.Members))
	for i, member := range class.Members {
		cclass.Members[i] = member.Copy()
	}
	cclass.Tags = class.CopyTags()
	return &cclass
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
	Tags          `json:",omitempty"`
}

// GetMemberType returns a string indicating the the type of member.
//
// GetMemberType implements the rbxapi.Member interface.
func (member *Property) GetMemberType() string {
	return "Property"
}

// GetName returns the name of the member.
//
// GetName implements the rbxapi.Member interface.
func (member *Property) GetName() string {
	return member.Name
}

// Copy returns a deep copy of the member descriptor.
//
// Copy implements the rbxapi.Member interface.
func (member *Property) Copy() rbxapi.Member {
	cmember := *member
	cmember.Tags = member.CopyTags()
	return &cmember
}

// GetSecurity returns the security context associated with the property's
// read and write access.
//
// GetSecurity implements the rbxapi.Property interface.
func (member *Property) GetSecurity() (read, write string) {
	return member.ReadSecurity, member.WriteSecurity
}

// GetValueType returns the type of value stored in the property.
//
// GetValueType implements the rbxapi.Property interface.
func (member *Property) GetValueType() rbxapi.Type {
	return member.ValueType
}

// Function represents a class member of the Function member type.
type Function struct {
	Name       string
	Parameters []Parameter
	ReturnType Type
	Security   string
	Tags       `json:",omitempty"`
}

// GetMemberType returns a string indicating the the type of member.
//
// GetMemberType implements the rbxapi.Member interface.
func (member *Function) GetMemberType() string {
	return "Function"
}

// GetName returns the name of the member.
//
// GetName implements the rbxapi.Member interface.
func (member *Function) GetName() string {
	return member.Name
}

// Copy returns a deep copy of the member descriptor.
//
// Copy implements the rbxapi.Member interface.
func (member *Function) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

// GetSecurity returns the security context of the member's access.
//
// GetSecurity implements the rbxapi.Function interface.
func (member *Function) GetSecurity() string {
	return member.Security
}

// GetParameters returns the list of parameters describing the arguments
// passed to the function. These parameters may have default values.
//
// GetParameters implements the rbxapi.Function interface.
func (member *Function) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

// GetReturnType returns the type of value returned by the function.
//
// GetReturnType implements the rbxapi.Function interface.
func (member *Function) GetReturnType() rbxapi.Type {
	return member.ReturnType
}

// Event represents a class member of the Event member type.
type Event struct {
	Name       string
	Parameters []Parameter
	Security   string
	Tags       `json:",omitempty"`
}

// GetMemberType returns a string indicating the the type of member.
//
// GetMemberType implements the rbxapi.Member interface.
func (member *Event) GetMemberType() string {
	return "Event"
}

// GetName returns the name of the member.
//
// GetName implements the rbxapi.Member interface.
func (member *Event) GetName() string {
	return member.Name
}

// Copy returns a deep copy of the member descriptor.
//
// Copy implements the rbxapi.Member interface.
func (member *Event) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

// GetSecurity returns the security context of the member's access.
//
// GetSecurity implements the rbxapi.Event interface.
func (member *Event) GetSecurity() string {
	return member.Security
}

// GetParameters returns the list of parameters describing the arguments
// received from the event. These parameters cannot have default values.
//
// GetParameters implements the rbxapi.Event interface.
func (member *Event) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

// Callback represents a class member of the Callback member type.
type Callback struct {
	Name       string
	Parameters []Parameter
	ReturnType Type
	Security   string
	Tags       `json:",omitempty"`
}

// GetMemberType returns a string indicating the the type of member.
//
// GetMemberType implements the rbxapi.Member interface.
func (member *Callback) GetMemberType() string {
	return "Callback"
}

// GetName returns the name of the member.
//
// GetName implements the rbxapi.Member interface.
func (member *Callback) GetName() string {
	return member.Name
}

// Copy returns a deep copy of the member descriptor.
//
// Copy implements the rbxapi.Member interface.
func (member *Callback) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

// GetSecurity returns the security context of the member's access.
//
// GetSecurity implements the rbxapi.Callback interface.
func (member *Callback) GetSecurity() string {
	return member.Security
}

// GetParameters returns the list of parameters describing the arguments
// passed to the callback. These parameters cannot have default values.
//
// GetParameters implements the rbxapi.Callback interface.
func (member *Callback) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

// GetReturnType returns the type of value that is returned by the callback.
//
// GetReturnType implements the rbxapi.Callback interface.
func (member *Callback) GetReturnType() rbxapi.Type {
	return member.ReturnType
}

// Parameter represents a parameter of a function, event, or callback member.
type Parameter struct {
	Type    Type
	Name    string
	Default *string `json:",omitempty"`
}

// GetType returns the type of the parameter value.
//
// GetType implements the rbxapi.Parameter interface.
func (param Parameter) GetType() rbxapi.Type {
	return param.Type
}

// GetName returns the name describing the parameter.
//
// GetName implements the rbxapi.Parameter interface.
func (param Parameter) GetName() string {
	return param.Name
}

// GetDefault returns a string representing the default value of the
// parameter, and whether a default value is present.
//
// GetDefault implements the rbxapi.Parameter interface.
func (param Parameter) GetDefault() (value string, ok bool) {
	if param.Default != nil {
		return *param.Default, true
	}
	return "", false
}

// Copy returns a deep copy of the parameter.
//
// Copy implements the rbxapi.Parameter interface.
func (param Parameter) Copy() rbxapi.Parameter {
	cparam := param
	d := *param.Default
	cparam.Default = &d
	return cparam
}

// Enum represents an enum descriptor.
type Enum struct {
	Name  string
	Items []*EnumItem
	Tags  `json:",omitempty"`
}

// GetName returns the name of the enum.
//
// GetName implements the rbxapi.Enum interface.
func (enum *Enum) GetName() string {
	return enum.Name
}

// GetItems returns a list of items of the enum.
//
// GetItems implements the rbxapi.Enum interface.
func (enum *Enum) GetItems() []rbxapi.EnumItem {
	list := make([]rbxapi.EnumItem, len(enum.Items))
	for i, item := range enum.Items {
		list[i] = item
	}
	return list
}

// GetItem returns the first item of the given name, or nil if no item of the
// given name is present.
//
// GetItem implements the rbxapi.Enum interface.
func (enum *Enum) GetItem(name string) rbxapi.EnumItem {
	for _, item := range enum.Items {
		if item.GetName() == name {
			return item
		}
	}
	return nil
}

// Copy returns a deep copy of the enum descriptor.
//
// Copy implements the rbxapi.Enum interface.
func (enum *Enum) Copy() rbxapi.Enum {
	cenum := *enum
	cenum.Items = make([]*EnumItem, len(enum.Items))
	for i, item := range enum.Items {
		cenum.Items[i] = item.Copy().(*EnumItem)
	}
	cenum.Tags = enum.CopyTags()
	return &cenum
}

// EnumItem represents an enum item descriptor.
type EnumItem struct {
	Name  string
	Value int
	Tags  `json:",omitempty"`
}

// GetName returns the name of the enum item.
//
// GetName implements the rbxapi.EnumItem interface.
func (item *EnumItem) GetName() string {
	return item.Name
}

// GetValue returns the value of the enum item.
//
// GetValue implements the rbxapi.EnumItem interface.
func (item *EnumItem) GetValue() int {
	return item.Value
}

// Copy returns a deep copy of the enum item descriptor.
//
// Copy implements the rbxapi.EnumItem interface.
func (item *EnumItem) Copy() rbxapi.EnumItem {
	citem := *item
	citem.Tags = item.CopyTags()
	return &citem
}

// Tags contains the list of tags of a descriptor.
type Tags []string

// GetTag returns whether the given tag is present in the descriptor.
//
// GetTag implements the rbxapi.Taggable interface.
func (tags Tags) GetTag(tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetTags returns a list of all tags present in the descriptor.
//
// GetTags implements the rbxapi.Taggable interface.
func (tags Tags) GetTags() []string {
	list := make([]string, 0, len(tags))
	copy(list, tags)
	return list
}

// LenTags returns the number of tags in the list.
func (tags Tags) LenTags() int {
	return len(tags)
}

// SetTag adds one or more tags to the list. Duplicate tags are removed.
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

// UnsetTag removes one or more tags from the list. Duplicate tags are
// removed.
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

// CopyTags returns a copy of the tag list.
func (tags Tags) CopyTags() Tags {
	ctags := make(Tags, len(tags))
	copy(ctags, tags)
	return ctags
}

// Type represents a value type.
type Type struct {
	Category string
	Name     string
}

// GetName returns the name of the type.
//
// GetName implements the rbxapi.Type interface.
func (typ Type) GetName() string {
	return typ.Name
}

// GetCategory returns the category of the type.
//
// GetCategory implements the rbxapi.Type interface.
func (typ Type) GetCategory() string {
	return typ.Category
}

// String returns a string representation of the entire type.
//
// String implements the rbxapi.Type interface.
func (typ Type) String() string {
	if typ.Category == "" {
		return typ.Name
	}
	return typ.Category + ":" + typ.Name
}

// Copy returns a deep copy of the type.
//
// Copy implements the rbxapi.Type interface.
func (typ Type) Copy() rbxapi.Type {
	return typ
}
