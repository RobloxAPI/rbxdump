// The patch package is used to represent information about differences
// between Roblox Lua API structures.
//
// Similar to the rbxapi package, this package provides a common interface,
// which is to be implemented elsewhere.
//
// The subpackages rbxapi/diff and rbxapijson/diff provide implementations of
// this interface.
package patch

import (
	"github.com/robloxapi/rbxapi"
)

// Differ is implemented by any value that has a Diff method, which returns
// the differences between two structures as a list of Actions.
//
// Returned actions may point to values within either structure. As such,
// these actions should be considered invalid when either structure changes.
type Differ interface {
	Diff() []Action
}

// Patcher is implemented by any value that has a Patch method, which applies
// a given list of Actions to a structure. Actions with information that is
// irrelevant, incomplete, or invalid can be ignored.
//
// Ideally, when the APIs "origin" and "target" are compared with a Differ,
// and the returned list of Actions are passed to a Patcher, the end result is
// that origin is transformed to match target exactly. This should be the case
// when origin and target come from the same implementation. In practice,
// origin and target may have different underlying implementations, in which
// case it may not be possible for all information to transferred from one to
// the other.
//
// Implementers must ensure that referred information is properly copied, so
// that values are not shared between structures.
type Patcher interface {
	Patch([]Action)
}

// Action represents a single unit of difference between one API structure and
// another.
type Action interface {
	// GetType returns the type of action.
	GetType() Type
	// GetField returns the name of the field being changed, when the action
	// is a Change type. Must return an empty string otherwise.
	GetField() string
	// GetPrev returns the old value of the field being changed, when the
	// action is a Change type. Must return nil otherwise.
	GetPrev() interface{}
	// GetNext returns the new value of the field being changed, when the
	// action is a Change type. Must return nil otherwise.
	GetNext() interface{}
	// String returns a string representation of the action, which is
	// implementation-dependent.
	String() string
}

// Type indicates the kind of transformation performed by an Action.
type Type int

const (
	Remove Type = -1 // The action removes data.
	Change Type = 0  // The action changes data.
	Add    Type = 1  // The action adds data.
)

// String returns a string representation of the action type.
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

// Class represents an Action that applies to a rbxapi.Class.
type Class interface {
	// GetClass returns the class associated with the action. When adding or
	// removing, this is the class being added or removed. When changing, the
	// class is meant to be used only for context.
	GetClass() rbxapi.Class
	Action
}

// Member represents an Action that applies to a rbxapi.Member.
type Member interface {
	// GetClass returns the class associated with the action. The class is
	// meant to be used only for context.
	GetClass() rbxapi.Class
	// GetMember returns the member associated with the action. When adding or
	// removing, this is the member being added or removed. When changing, the
	// member is meant to be used only for context.
	GetMember() rbxapi.Member
	Action
}

// Enum represents an Action that applies to a rbxapi.Enum.
type Enum interface {
	// GetEnum returns the enum associated with the action. When adding or
	// removing, this is the enum being added or removed. When changing, the
	// enum is meant to be used only for context.
	GetEnum() rbxapi.Enum
	Action
}

// EnumItem represents an Action that applies to a rbxapi.EnumItem.
type EnumItem interface {
	// GetEnum returns the enum associated with the action. The enum is meant
	// to be used only for context.
	GetEnum() rbxapi.Enum
	// GetEnumItem returns the enum item associated with the action. When
	// adding or removing, this is the item being added or removed. When
	// changing, the item is meant to be used only for context.
	GetEnumItem() rbxapi.EnumItem
	Action
}
