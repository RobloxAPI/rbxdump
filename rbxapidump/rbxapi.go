// The rbxapidump package implements the rbxapi interface as a codec for the
// Roblox API dump format.
package rbxapidump

import (
	"github.com/robloxapi/rbxapi"
	"strings"
)

type Root struct {
	Classes []*Class
	Enums   []*Enum
}

func (root *Root) GetClasses() []rbxapi.Class {
	list := make([]rbxapi.Class, len(root.Classes))
	for i, class := range root.Classes {
		list[i] = class
	}
	return list
}

func (root *Root) GetClass(name string) rbxapi.Class {
	for _, class := range root.Classes {
		if class.Name == name {
			return class
		}
	}
	return nil
}

func (root *Root) GetEnums() []rbxapi.Enum {
	list := make([]rbxapi.Enum, len(root.Enums))
	for i, enum := range root.Enums {
		list[i] = enum
	}
	return list
}

func (root *Root) GetEnum(name string) rbxapi.Enum {
	for _, enum := range root.Enums {
		if enum.Name == name {
			return enum
		}
	}
	return nil
}

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

type Class struct {
	Name       string
	Superclass string
	Members    []rbxapi.Member
	Tags
}

func (class *Class) GetName() string {
	return class.Name
}

func (class *Class) GetSuperclass() string {
	return class.Superclass
}

func (class *Class) GetMembers() []rbxapi.Member {
	list := make([]rbxapi.Member, len(class.Members))
	copy(list, class.Members)
	return list
}

func (class *Class) Copy() rbxapi.Class {
	cclass := *class
	cclass.Members = make([]rbxapi.Member, len(class.Members))
	for i, member := range class.Members {
		cclass.Members[i] = member.Copy()
	}
	cclass.Tags = class.CopyTags()
	return &cclass
}

func (class *Class) GetMember(name string) rbxapi.Member {
	for _, member := range class.Members {
		if member.GetName() == name {
			return member
		}
	}
	return nil
}

func getSecurity(tags Tags) string {
	for _, tag := range tags {
		if strings.Contains(tag, "Security") || strings.Contains(tag, "security") {
			return tag
		}
	}
	return ""
}

type Property struct {
	Name      string
	Class     string
	ValueType Type
	Tags
}

func (member *Property) GetMemberType() string     { return "Property" }
func (member *Property) GetName() string           { return member.Name }
func (member *Property) GetValueType() rbxapi.Type { return member.ValueType }
func (member *Property) GetSecurity() (read, write string) {
	const prefix = "ScriptWriteRestricted: ["
	const suffix = "]"
	for _, tag := range member.Tags {
		if write == "" && strings.HasPrefix(tag, prefix) {
			write = tag[len(prefix) : len(tag)-len(suffix)]
			if read != "" {
				break
			}
		} else if read == "" && (strings.Contains(tag, "Security") || strings.Contains(tag, "security")) {
			read = tag
			if write != "" {
				break
			}
		}
	}
	return read, write
}
func (member *Property) Copy() rbxapi.Member {
	cmember := *member
	cmember.Tags = member.CopyTags()
	return &cmember
}

type Function struct {
	Name       string
	Class      string
	ReturnType Type
	Parameters []Parameter
	Tags
}

func (member *Function) GetMemberType() string      { return "Function" }
func (member *Function) GetName() string            { return member.Name }
func (member *Function) GetReturnType() rbxapi.Type { return member.ReturnType }
func (member *Function) GetSecurity() string        { return getSecurity(member.Tags) }
func (member *Function) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}
func (member *Function) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

type YieldFunction Function

func (member *YieldFunction) GetMemberType() string      { return "Function" }
func (member *YieldFunction) GetName() string            { return member.Name }
func (member *YieldFunction) GetReturnType() rbxapi.Type { return member.ReturnType }
func (member *YieldFunction) GetSecurity() string        { return getSecurity(member.Tags) }
func (member *YieldFunction) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}
func (member *YieldFunction) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

type Event struct {
	Name       string
	Class      string
	Parameters []Parameter
	Tags
}

func (member *Event) GetMemberType() string { return "Event" }
func (member *Event) GetName() string       { return member.Name }
func (member *Event) GetSecurity() string   { return getSecurity(member.Tags) }
func (member *Event) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}
func (member *Event) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

type Callback struct {
	Name       string
	Class      string
	ReturnType Type
	Parameters []Parameter
	Tags
}

func (member *Callback) GetMemberType() string      { return "Callback" }
func (member *Callback) GetName() string            { return member.Name }
func (member *Callback) GetReturnType() rbxapi.Type { return member.ReturnType }
func (member *Callback) GetSecurity() string        { return getSecurity(member.Tags) }
func (member *Callback) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list

}
func (member *Callback) Copy() rbxapi.Member {
	cmember := *member
	cmember.Parameters = make([]Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		cmember.Parameters[i] = param.Copy().(Parameter)
	}
	cmember.Tags = member.CopyTags()
	return &cmember
}

type Parameter struct {
	Type    Type
	Name    string
	Default *string
}

func (param Parameter) GetType() rbxapi.Type { return param.Type }
func (param Parameter) GetName() string      { return param.Name }
func (param Parameter) GetDefault() (value string, ok bool) {
	if param.Default != nil {
		return *param.Default, true
	}
	return "", false
}
func (param Parameter) Copy() rbxapi.Parameter {
	cparam := param
	d := *param.Default
	cparam.Default = &d
	return cparam
}

type Enum struct {
	Name  string
	Items []*EnumItem
	Tags
}

func (enum *Enum) GetName() string { return enum.Name }
func (enum *Enum) GetItems() []rbxapi.EnumItem {
	list := make([]rbxapi.EnumItem, len(enum.Items))
	for i, item := range enum.Items {
		list[i] = item
	}
	return list
}
func (enum *Enum) GetItem(name string) rbxapi.EnumItem {
	for _, item := range enum.Items {
		if item.GetName() == name {
			return item
		}
	}
	return nil
}
func (enum *Enum) Copy() rbxapi.Enum {
	cenum := *enum
	cenum.Items = make([]*EnumItem, len(enum.Items))
	for i, item := range enum.Items {
		cenum.Items[i] = item.Copy().(*EnumItem)
	}
	cenum.Tags = enum.CopyTags()
	return &cenum
}

type EnumItem struct {
	Enum  string
	Name  string
	Value int
	Tags
}

func (item *EnumItem) GetName() string { return item.Name }
func (item *EnumItem) GetValue() int   { return item.Value }
func (item *EnumItem) Copy() rbxapi.EnumItem {
	citem := *item
	citem.Tags = item.CopyTags()
	return &citem
}

type Tags []string

func (tags Tags) GetTag(tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
func (tags Tags) LenTags() int {
	return len(tags)
}
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
func (tags Tags) GetTags() []string {
	list := make([]string, 0, len(tags))
	copy(list, tags)
	return list
}
func (tags Tags) CopyTags() Tags {
	ctags := make(Tags, len(tags))
	copy(ctags, tags)
	return ctags
}

type Type string

func (typ Type) GetName() string {
	if i := strings.Index(string(typ), ":"); i >= 0 {
		return string(typ[i+1:])
	}
	return string(typ)
}
func (typ Type) GetCategory() string {
	if i := strings.Index(string(typ), ":"); i >= 0 {
		return string(typ[:i])
	}
	return ""
}
func (typ Type) String() string {
	return string(typ)
}
func (typ Type) Copy() rbxapi.Type {
	return typ
}
func (typ *Type) SetFromType(t rbxapi.Type) {
	if cat := t.GetCategory(); cat == "" {
		*typ = Type(t.GetName())
	} else {
		*typ = Type(cat + ":" + t.GetName())
	}
}
