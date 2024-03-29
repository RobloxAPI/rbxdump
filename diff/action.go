// The diff package provides operations for diffing and patching rbxdump
// structures.
package diff

import (
	"encoding/json"
	"fmt"

	"github.com/robloxapi/rbxdump"
)

// Differ is implemented by any value that has a Diff method, which returns
// the differences between two structures as a list of Actions.
type Differ interface {
	Diff() []Action
}

// Patcher is implemented by any value that has a Patch method, which applies
// a given list of Actions to a structure. Actions with information that is
// irrelevant, incomplete, or invalid can be ignored.
//
// Ideally, when the API's "origin" and "target" are compared with a Differ,
// and the returned list of Actions are passed to a Patcher, the end result is
// that origin is transformed to match target exactly.
//
// Implementations must ensure that values contained within action Fields are
// copied, so that they are not shared between structures.
type Patcher interface {
	Patch([]Action)
}

// Inverter is implemented by any value that has an Inverse method, which
// receives a list of Actions and produces an inverse list of Actions. That is,
// for a list of actions P that patches structure A into B, the inverse of P
// will patch structure B back into A.
type Inverter interface {
	Inverse([]Action) []Action
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
	return "<invalid>"
}

// Element indicates the type of element to which an Action applies.
type Element int

const (
	_ Element = iota - 1
	Invalid
	Class
	Property
	Function
	Event
	Callback
	Enum
	EnumItem
)

// FromElement returns the Element corresponding to the given rbxdump element
// type or a pointer to such. Returns Invalid for any other type.
func FromElement(e any) Element {
	switch e.(type) {
	case rbxdump.Class, *rbxdump.Class:
		return Class
	case rbxdump.Property, *rbxdump.Property:
		return Property
	case rbxdump.Function, *rbxdump.Function:
		return Function
	case rbxdump.Event, *rbxdump.Event:
		return Event
	case rbxdump.Callback, *rbxdump.Callback:
		return Callback
	case rbxdump.Enum, *rbxdump.Enum:
		return Enum
	case rbxdump.EnumItem, *rbxdump.EnumItem:
		return EnumItem
	}
	return Invalid
}

// String returns a string representation of the element type.
func (e Element) String() string {
	switch e {
	case Class:
		return "Class"
	case Property:
		return "Property"
	case Function:
		return "Function"
	case Event:
		return "Event"
	case Callback:
		return "Callback"
	case Enum:
		return "Enum"
	case EnumItem:
		return "EnumItem"
	}
	return "<invalid>"
}

// Primary returns the primary element of this element.
func (e Element) Primary() Element {
	switch e {
	case Class:
		return Class
	case Property:
		return Class
	case Function:
		return Class
	case Event:
		return Class
	case Callback:
		return Class
	case Enum:
		return Enum
	case EnumItem:
		return Enum
	}
	return Invalid
}

// IsValid returns whether the value is a valid element.
func (e Element) IsValid() bool {
	return Class <= e && e <= EnumItem
}

// IsMember returns whether the element is a class member.
func (e Element) IsMember() bool {
	return Property <= e && e <= Callback
}

func (e Element) MarshalJSON() (b []byte, err error) {
	s := e.String()
	b = make([]byte, 0, len(s)+2)
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, '"')
	return b, nil
}

func (e *Element) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch v {
	case "Class":
		*e = Class
	case "Property":
		*e = Property
	case "Function":
		*e = Function
	case "Event":
		*e = Event
	case "Callback":
		*e = Callback
	case "Enum":
		*e = Enum
	case "EnumItem":
		*e = EnumItem
	default:
		*e = Invalid
	}
	return nil
}

// Action describes a single unit of difference between two dump structures.
type Action struct {
	// Type is the kind of transformation performed by the action.
	Type Type
	// Element is the type of element to which the action applies.
	Element Element
	// Primary is the name of the primary element.
	Primary string
	// Secondary is the name of the secondary element. Applies only to Property,
	// Function, Event, Callback, and EnumItem elements.
	Secondary string `json:",omitempty"`
	// Fields describes fields of the element. If Type is Add, this describes
	// the initial values. If Type is Change, this describes the new values.
	Fields rbxdump.Fields `json:",omitempty"`
}

// ToFielder returns a new element corresponding to the action's element type,
// as a Fielder. The Name field is set according to the action's identifiers.
// Returns nil if the type is invalid.
func (a Action) ToFielder() rbxdump.Fielder {
	switch a.Element {
	case Class:
		return &rbxdump.Class{Name: a.Primary}
	case Property:
		return &rbxdump.Property{Name: a.Secondary}
	case Function:
		return &rbxdump.Function{Name: a.Secondary}
	case Event:
		return &rbxdump.Event{Name: a.Secondary}
	case Callback:
		return &rbxdump.Callback{Name: a.Secondary}
	case Enum:
		return &rbxdump.Enum{Name: a.Primary}
	case EnumItem:
		return &rbxdump.EnumItem{Name: a.Secondary}
	default:
		return nil
	}
}

// ToPrimaryFielder is similar to ToFielder, but returns a new element
// corresponding to the action's primary element type.
func (a Action) ToPrimaryFielder() rbxdump.Fielder {
	switch a.Element {
	case Class:
		return &rbxdump.Class{Name: a.Primary}
	case Property:
		return &rbxdump.Class{Name: a.Primary}
	case Function:
		return &rbxdump.Class{Name: a.Primary}
	case Event:
		return &rbxdump.Class{Name: a.Primary}
	case Callback:
		return &rbxdump.Class{Name: a.Primary}
	case Enum:
		return &rbxdump.Enum{Name: a.Primary}
	case EnumItem:
		return &rbxdump.Enum{Name: a.Primary}
	default:
		return nil
	}
}

// ToMember returns a new member element corresponding to the action's element
// type, as a Member. The Name field is set according to the action's
// identifiers. Returns nil if the type is not a member.
func (a Action) ToMember() rbxdump.Member {
	switch a.Element {
	case Property:
		return &rbxdump.Property{Name: a.Secondary}
	case Function:
		return &rbxdump.Function{Name: a.Secondary}
	case Event:
		return &rbxdump.Event{Name: a.Secondary}
	case Callback:
		return &rbxdump.Callback{Name: a.Secondary}
	default:
		return nil
	}
}

func (a *Action) UnmarshalJSON(b []byte) error {
	var action struct {
		Type      Type
		Element   Element
		Primary   string
		Secondary string
		Fields    rbxdump.Fields
	}
	if err := json.Unmarshal(b, &action); err != nil {
		return err
	}
	if action.Fields == nil {
		action.Fields = rbxdump.Fields{}
	} else if len(action.Fields) > 0 {
		// Convert generic JSON structure to rbxdump values.
		if f := Action(action).ToFielder(); f != nil {
			f.SetFields(action.Fields)
			action.Fields = f.Fields(action.Fields)
		} else {
			action.Fields = rbxdump.Fields{}
		}
	}
	*a = Action(action)
	return nil
}

func (a Action) String() string {
	s := a.Type.String() + " " + a.Element.String() + " " + a.Primary
	switch a.Element {
	case Property, Function, Event, Callback, EnumItem:
		s += "." + a.Secondary
	}
	if len(a.Fields) > 0 {
		s += ": " + fmt.Sprintf("%v", a.Fields)
	}
	return s
}
