// The patch package represents information about differences between API
// structures.
package patch

import (
	"github.com/robloxapi/rbxapi"
	"strings"
)

// Differ is implemented by any value that has a Diff method, which returns
// the differences between two structures as a list of Actions.
type Differ interface {
	Diff() []Action
}

// Action represents a single unit of difference between one API structure and
// another.
type Action interface {
	// GetType returns the type of action.
	GetType() Type
	// GetField returns the name of the field being changed, when the action
	// is a Change type. Returns an empty string otherwise.
	GetField() string
	// GetPrev returns the old value of the field being changed, when the
	// action is a Change type. Returns nil otherwise.
	GetPrev() Value
	// GetNext returns the new value of the field being changed, when the
	// action is a Change type. Returns nil otherwise.
	GetNext() Value
	// String returns a string representation of the action.
	String() string
}

// Type indicates the kind of transformation performed by an Action.
type Type int

const (
	Remove Type = -1 // The action removes a piece of data.
	Change Type = 0  // The action changes a piece of data.
	Add    Type = 1  // The action adds a piece of data.
)

func (t Type) String() string {
	switch t {
	case Remove:
		return "Remove"
	case Change:
		return "Change"
	case Add:
		return "Add"
	}
	return ""
}

// Value represents a comparable value.
type Value interface {
	// Equal returns whether the value is equal to another value.
	Equal(Value) bool
	// String returns a string representation of the value.
	String() string
}

// Class represents an Action that applies to a rbxapi.Class.
type Class struct {
	Type  Type
	Class rbxapi.Class
	Field string
	Prev  Value
	Next  Value
}

func (a *Class) GetType() Type    { return a.Type }
func (a *Class) GetField() string { return a.Field }
func (a *Class) GetPrev() Value   { return a.Prev }
func (a *Class) GetNext() Value   { return a.Next }
func (a *Class) String() string {
	switch a.Type {
	case Add, Remove:
		members := a.Class.GetMembers()
		ms := make([]string, len(members)*2)
		for _, member := range members {
			ms = append(ms, "\n\t")
			ms = append(ms, (&Member{Type: a.Type, Class: a.Class, Member: member}).String())
		}
		return a.Type.String() +
			" Class " + a.Class.GetName() +
			strings.Join(ms, "")
	case Change:
		return a.Type.String() +
			" field " + a.Field +
			" of class " + a.Class.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

// Member represents an Action that applies to a rbxapi.Member.
//
// The Class field is the outer structure of the member. It is used primarily
// for context, so it may be omitted.
type Member struct {
	Type   Type
	Class  rbxapi.Class
	Member rbxapi.Member
	Field  string
	Prev   Value
	Next   Value
}

func (a *Member) GetType() Type    { return a.Type }
func (a *Member) GetField() string { return a.Field }
func (a *Member) GetPrev() Value   { return a.Prev }
func (a *Member) GetNext() Value   { return a.Next }
func (a *Member) String() string {
	var class string
	if a.Class != nil {
		class = a.Class.GetName() + "."
	}
	switch a.Type {
	case Add, Remove:
		return a.Type.String() +
			" " + a.Member.GetMemberType() +
			class + "." + a.Member.GetName()
	case Change:
		return a.Type.String() +
			" field " + a.Field +
			" of " + a.Member.GetMemberType() +
			" " + class + a.Member.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

// Enum represents an Action that applies to a rbxapi.Enum.
type Enum struct {
	Type  Type
	Enum  rbxapi.Enum
	Field string
	Prev  Value
	Next  Value
}

func (a *Enum) GetType() Type    { return a.Type }
func (a *Enum) GetField() string { return a.Field }
func (a *Enum) GetPrev() Value   { return a.Prev }
func (a *Enum) GetNext() Value   { return a.Next }
func (a *Enum) String() string {
	switch a.Type {
	case Add, Remove:
		items := a.Enum.GetItems()
		is := make([]string, len(items)*2)
		for _, item := range items {
			is = append(is, "\n\t")
			is = append(is, (&EnumItem{Type: a.Type, Enum: a.Enum, Item: item}).String())
		}
		return a.Type.String() +
			" Enum " + a.Enum.GetName() +
			strings.Join(is, "")
	case Change:
		return a.Type.String() +
			" field " + a.Field +
			" of Enum " + a.Enum.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}

// EnumItem represents an Action that applies to a rbxapi.EnumItem.
//
// The Enum field is the outer structure of the enum item. It is used
// primarily for context, so it may be omitted.
type EnumItem struct {
	Type  Type
	Enum  rbxapi.Enum
	Item  rbxapi.EnumItem
	Field string
	Prev  Value
	Next  Value
}

func (a *EnumItem) GetType() Type    { return a.Type }
func (a *EnumItem) GetField() string { return a.Field }
func (a *EnumItem) GetPrev() Value   { return a.Prev }
func (a *EnumItem) GetNext() Value   { return a.Next }
func (a *EnumItem) String() string {
	var enum string
	if a.Enum != nil {
		enum = a.Enum.GetName() + "."
	}
	switch a.Type {
	case Add, Remove:
		return a.Type.String() +
			" EnumItem " + enum + a.Item.GetName()
	case Change:
		return a.Type.String() +
			" field " + a.Field +
			" of EnumItem " + enum + a.Item.GetName() +
			" from " + a.Prev.String() +
			" to " + a.Next.String()
	}
	return ""
}
