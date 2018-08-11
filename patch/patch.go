// The patch package represents information about differences between API
// structures.
package patch

import (
	"github.com/robloxapi/rbxapi"
)

// Differ is implemented by any value that has a Diff method, which returns
// the differences between two structures as a list of Actions.
type Differ interface {
	Diff() []Action
}

// Patcher is implemented by any value that has a Patch method, which applies
// a given list of Actions to a structure.
//
// Actions with irrelevant or incomplete information are ignored.
type Patcher interface {
	Patch([]Action)
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
type Class interface {
	GetClass() rbxapi.Class
	Action
}

// Member represents an Action that applies to a rbxapi.Member.
type Member interface {
	GetClass() rbxapi.Class
	GetMember() rbxapi.Member
	Action
}

// Enum represents an Action that applies to a rbxapi.Enum.
type Enum interface {
	GetEnum() rbxapi.Enum
	Action
}

// EnumItem represents an Action that applies to a rbxapi.EnumItem.
type EnumItem interface {
	GetEnum() rbxapi.Enum
	GetItem() rbxapi.EnumItem
	Action
}
