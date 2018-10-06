package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
	"strconv"
	"strings"
)

// toString converts common API value types to strings.
func toString(v interface{}) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case rbxapi.Type:
		return v.String()
	case []string:
		return "[" + strings.Join(v, ", ") + "]"
	case rbxapi.Parameters:
		n := v.GetLength()
		ss := make([]string, n)
		for i := 0; i < n; i++ {
			param := v.GetParameter(i)
			ss[i] = param.GetType().String() + " " + param.GetName()
			if def, ok := param.GetDefault(); ok {
				ss[i] += " = " + def
			}
		}
		return "(" + strings.Join(ss, ", ") + ")"
	}
	return "<unknown value>"
}

func tagsToString(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return " " + toString(tags)
}

// ClassAction represents a patch.Action that applies to a rbxapi.Class.
type ClassAction struct {
	Type  patch.Type
	Class rbxapi.Class
	Field string
	Prev  interface{}
	Next  interface{}
}

func (a *ClassAction) GetClass() rbxapi.Class { return a.Class }
func (a *ClassAction) GetType() patch.Type    { return a.Type }
func (a *ClassAction) GetField() string       { return a.Field }
func (a *ClassAction) GetPrev() interface{}   { return a.Prev }
func (a *ClassAction) GetNext() interface{}   { return a.Next }
func (a *ClassAction) String() string {
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" Class " + a.Class.GetName() +
			tagsToString(a.Class.GetTags())
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of class " + a.Class.GetName() +
			" from " + toString(a.Prev) +
			" to " + toString(a.Next)
	}
	return ""
}

// MemberAction represents a patch.Action that applies to a rbxapi.Member.
type MemberAction struct {
	Type   patch.Type
	Class  rbxapi.Class
	Member rbxapi.Member
	Field  string
	Prev   interface{}
	Next   interface{}
}

func (a *MemberAction) GetClass() rbxapi.Class   { return a.Class }
func (a *MemberAction) GetMember() rbxapi.Member { return a.Member }
func (a *MemberAction) GetType() patch.Type      { return a.Type }
func (a *MemberAction) GetField() string         { return a.Field }
func (a *MemberAction) GetPrev() interface{}     { return a.Prev }
func (a *MemberAction) GetNext() interface{}     { return a.Next }
func (a *MemberAction) String() string {
	var class string
	if a.Class != nil {
		class = a.Class.GetName() + "."
	}
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" " + a.Member.GetMemberType() +
			" " + class + a.Member.GetName() +
			tagsToString(a.Member.GetTags())
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of " + a.Member.GetMemberType() +
			" " + class + a.Member.GetName() +
			" from " + toString(a.Prev) +
			" to " + toString(a.Next)
	}
	return ""
}

// EnumAction represents a patch.Action that applies to a rbxapi.Enum.
type EnumAction struct {
	Type  patch.Type
	Enum  rbxapi.Enum
	Field string
	Prev  interface{}
	Next  interface{}
}

func (a *EnumAction) GetEnum() rbxapi.Enum { return a.Enum }
func (a *EnumAction) GetType() patch.Type  { return a.Type }
func (a *EnumAction) GetField() string     { return a.Field }
func (a *EnumAction) GetPrev() interface{} { return a.Prev }
func (a *EnumAction) GetNext() interface{} { return a.Next }
func (a *EnumAction) String() string {
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" Enum " + a.Enum.GetName() +
			tagsToString(a.Enum.GetTags())
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of Enum " + a.Enum.GetName() +
			" from " + toString(a.Prev) +
			" to " + toString(a.Next)
	}
	return ""
}

// EnumItemAction represents a patch.Action that applies to a rbxapi.EnumItem.
type EnumItemAction struct {
	Type     patch.Type
	Enum     rbxapi.Enum
	EnumItem rbxapi.EnumItem
	Field    string
	Prev     interface{}
	Next     interface{}
}

func (a *EnumItemAction) GetEnum() rbxapi.Enum         { return a.Enum }
func (a *EnumItemAction) GetEnumItem() rbxapi.EnumItem { return a.EnumItem }
func (a *EnumItemAction) GetType() patch.Type          { return a.Type }
func (a *EnumItemAction) GetField() string             { return a.Field }
func (a *EnumItemAction) GetPrev() interface{}         { return a.Prev }
func (a *EnumItemAction) GetNext() interface{}         { return a.Next }
func (a *EnumItemAction) String() string {
	var enum string
	if a.Enum != nil {
		enum = a.Enum.GetName() + "."
	}
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" EnumItem " + enum + a.EnumItem.GetName() +
			tagsToString(a.EnumItem.GetTags())
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of EnumItem " + enum + a.EnumItem.GetName() +
			" from " + toString(a.Prev) +
			" to " + toString(a.Next)
	}
	return ""
}
