package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
	"strings"
)

type ClassAction struct {
	Type  patch.Type
	Class rbxapi.Class
	Field string
	Prev  patch.Value
	Next  patch.Value
}

func (a *ClassAction) GetClass() rbxapi.Class { return a.Class }
func (a *ClassAction) GetType() patch.Type    { return a.Type }
func (a *ClassAction) GetField() string       { return a.Field }
func (a *ClassAction) GetPrev() patch.Value   { return a.Prev }
func (a *ClassAction) GetNext() patch.Value   { return a.Next }
func (a *ClassAction) String() string {
	switch a.Type {
	case patch.Add, patch.Remove:
		members := a.Class.GetMembers()
		ms := make([]string, len(members)*2)
		for _, member := range members {
			ms = append(ms, "\n\t")
			ms = append(ms, (&MemberAction{Type: a.Type, Class: a.Class, Member: member}).String())
		}
		return a.Type.String() +
			" Class " + a.Class.GetName() +
			strings.Join(ms, "")
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of class " + a.Class.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

type MemberAction struct {
	Type   patch.Type
	Class  rbxapi.Class
	Member rbxapi.Member
	Field  string
	Prev   patch.Value
	Next   patch.Value
}

func (a *MemberAction) GetClass() rbxapi.Class   { return a.Class }
func (a *MemberAction) GetMember() rbxapi.Member { return a.Member }
func (a *MemberAction) GetType() patch.Type      { return a.Type }
func (a *MemberAction) GetField() string         { return a.Field }
func (a *MemberAction) GetPrev() patch.Value     { return a.Prev }
func (a *MemberAction) GetNext() patch.Value     { return a.Next }
func (a *MemberAction) String() string {
	var class string
	if a.Class != nil {
		class = a.Class.GetName() + "."
	}
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" " + a.Member.GetMemberType() +
			class + "." + a.Member.GetName()
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of " + a.Member.GetMemberType() +
			" " + class + a.Member.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

type EnumAction struct {
	Type  patch.Type
	Enum  rbxapi.Enum
	Field string
	Prev  patch.Value
	Next  patch.Value
}

func (a *EnumAction) GetEnum() rbxapi.Enum { return a.Enum }
func (a *EnumAction) GetType() patch.Type  { return a.Type }
func (a *EnumAction) GetField() string     { return a.Field }
func (a *EnumAction) GetPrev() patch.Value { return a.Prev }
func (a *EnumAction) GetNext() patch.Value { return a.Next }
func (a *EnumAction) String() string {
	switch a.Type {
	case patch.Add, patch.Remove:
		items := a.Enum.GetItems()
		is := make([]string, len(items)*2)
		for _, item := range items {
			is = append(is, "\n\t")
			is = append(is, (&EnumItemAction{Type: a.Type, Enum: a.Enum, Item: item}).String())
		}
		return a.Type.String() +
			" Enum " + a.Enum.GetName() +
			strings.Join(is, "")
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of Enum " + a.Enum.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

type EnumItemAction struct {
	Type  patch.Type
	Enum  rbxapi.Enum
	Item  rbxapi.EnumItem
	Field string
	Prev  patch.Value
	Next  patch.Value
}

func (a *EnumItemAction) GetEnum() rbxapi.Enum         { return a.Enum }
func (a *EnumItemAction) GetEnumItem() rbxapi.EnumItem { return a.Item }
func (a *EnumItemAction) GetType() patch.Type          { return a.Type }
func (a *EnumItemAction) GetField() string             { return a.Field }
func (a *EnumItemAction) GetPrev() patch.Value         { return a.Prev }
func (a *EnumItemAction) GetNext() patch.Value         { return a.Next }
func (a *EnumItemAction) String() string {
	var enum string
	if a.Enum != nil {
		enum = a.Enum.GetName() + "."
	}
	switch a.Type {
	case patch.Add, patch.Remove:
		return a.Type.String() +
			" EnumItem " + enum + a.Item.GetName()
	case patch.Change:
		return a.Type.String() +
			" field " + a.Field +
			" of EnumItem " + enum + a.Item.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}
