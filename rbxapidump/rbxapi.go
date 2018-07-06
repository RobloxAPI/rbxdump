package rbxapidump

import (
	"github.com/robloxapi/rbxapi"
	"sort"
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

type Class struct {
	Name         string
	Superclass   string
	Members      []rbxapi.Member
	NotCreatable bool
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

func (class *Class) GetMember(name string) rbxapi.Member {
	for _, member := range class.Members {
		if member.GetName() == name {
			return member
		}
	}
	return nil
}

type Property struct {
	Name          string
	Class         string
	ValueType     string
	Hidden        bool
	ReadOnly      bool
	WriteOnly     bool
	ReadSecurity  string
	WriteSecurity string
	Tags
}

func (member *Property) GetMemberType() string     { return "Property" }
func (member *Property) GetName() string           { return member.Name }
func (member *Property) GetValueType() rbxapi.Type { return Type(member.ValueType) }
func (member *Property) GetSecurity() (read, write string) {
	return member.ReadSecurity, member.WriteSecurity
}

type Function struct {
	Name       string
	Class      string
	ReturnType string
	Parameters []Parameter
	Security   string
	Tags
}

func (member *Function) GetMemberType() string      { return "Function" }
func (member *Function) GetName() string            { return member.Name }
func (member *Function) GetReturnType() rbxapi.Type { return Type(member.ReturnType) }
func (member *Function) GetSecurity() string        { return member.Security }
func (member *Function) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

type YieldFunction Function

func (member *YieldFunction) GetMemberType() string      { return "Function" }
func (member *YieldFunction) GetName() string            { return member.Name }
func (member *YieldFunction) GetReturnType() rbxapi.Type { return Type(member.ReturnType) }
func (member *YieldFunction) GetSecurity() string        { return member.Security }
func (member *YieldFunction) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

type Event struct {
	Name       string
	Class      string
	Parameters []Parameter
	Security   string
	Tags
}

func (member *Event) GetMemberType() string { return "Event" }
func (member *Event) GetName() string       { return member.Name }
func (member *Event) GetSecurity() string   { return member.Security }
func (member *Event) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list
}

type Callback struct {
	Name       string
	Class      string
	ReturnType string
	Parameters []Parameter
	NoYield    bool
	Security   string
	Tags
}

func (member *Callback) GetMemberType() string      { return "Callback" }
func (member *Callback) GetName() string            { return member.Name }
func (member *Callback) GetReturnType() rbxapi.Type { return Type(member.ReturnType) }
func (member *Callback) GetSecurity() string        { return member.Security }
func (member *Callback) GetParameters() []rbxapi.Parameter {
	list := make([]rbxapi.Parameter, len(member.Parameters))
	for i, param := range member.Parameters {
		list[i] = param
	}
	return list

}

type Parameter struct {
	Type    string
	Name    string
	Default *string
}

func (param Parameter) GetType() rbxapi.Type { return Type(param.Type) }
func (param Parameter) GetName() string      { return param.Name }
func (param Parameter) GetDefault() (value string, ok bool) {
	if param.Default != nil {
		return *param.Default, true
	}
	return "", false
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

type EnumItem struct {
	Enum  string
	Name  string
	Value int
	Tags
}

func (item *EnumItem) GetName() string { return item.Name }
func (item *EnumItem) GetValue() int   { return item.Value }

type Tags map[string]struct{}

func (tags Tags) GetTag(tag string) bool {
	_, ok := tags[tag]
	return ok
}
func (tags Tags) LenTags() int { return len(tags) }
func (tags *Tags) SetTag(tag ...string) {
	if *tags == nil {
		*tags = Tags{}
	}
	for _, t := range tag {
		(*tags)[t] = struct{}{}
	}
}
func (tags Tags) UnsetTag(tag ...string) {
	if tags == nil {
		return
	}
	for _, t := range tag {
		delete(tags, t)
	}
}
func (tags Tags) GetTags() []string {
	list := make([]string, 0, len(tags))
	for tag := range tags {
		list = append(list, tag)
	}
	sort.Strings(list)
	return list
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
