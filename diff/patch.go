package diff

import (
	"github.com/robloxapi/rbxdump"
)

// Patch is used to transform the embedded rbxdump.Root by applying a list of
// Actions.
type Patch struct {
	*rbxdump.Root
}

// Patch implements the Patcher interface.
func (root Patch) Patch(actions []Action) {
	for i, action := range actions {
		switch action.Element {
		case Class:
			switch action.Type {
			case Add:
				if _, ok := root.Classes[action.Primary]; !ok {
					class := rbxdump.Class{Name: action.Primary}
					class.SetFields(action.Fields)
					if root.Classes == nil {
						root.Classes = map[string]*rbxdump.Class{}
					}
					root.Classes[action.Primary] = &class
				}
			case Remove:
				delete(root.Classes, action.Primary)
			case Change:
				if class := root.Classes[action.Primary]; class != nil {
					class.SetFields(action.Fields)
				}
			}
		case Property, Function, Event, Callback:
			if class, ok := root.Classes[action.Primary]; ok {
				PatchClass{class}.Patch(actions[i : i+1])
			}
		case Enum:
			switch action.Type {
			case Add:
				if _, ok := root.Enums[action.Primary]; !ok {
					enum := rbxdump.Enum{Name: action.Primary}
					enum.SetFields(action.Fields)
					if root.Enums == nil {
						root.Enums = map[string]*rbxdump.Enum{}
					}
					root.Enums[action.Primary] = &enum
				}
			case Remove:
				delete(root.Enums, action.Primary)
			case Change:
				if enum := root.Enums[action.Primary]; enum != nil {
					enum.SetFields(action.Fields)
				}
			}
		case EnumItem:
			if enum, ok := root.Enums[action.Primary]; ok {
				PatchEnum{enum}.Patch(actions[i : i+1])
			}
		}
	}
}

// PatchClass is used to transform the embedded rbxdump.Class by applying a list
// of Actions.
type PatchClass struct {
	*rbxdump.Class
}

// Patch implements the Patcher interface.
func (class PatchClass) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Class:
			if action.Type == Change {
				class.SetFields(action.Fields)
			}
		case Property:
			switch action.Type {
			case Add:
				if _, ok := class.Members[action.Secondary]; !ok {
					member := rbxdump.Property{Name: action.Secondary}
					member.SetFields(action.Fields)
					if class.Members == nil {
						class.Members = map[string]rbxdump.Member{}
					}
					class.Members[action.Secondary] = &member
				}
			case Remove:
				if member, ok := class.Members[action.Secondary]; ok {
					if _, ok := member.(*rbxdump.Property); ok {
						delete(class.Members, action.Secondary)
					}
				}
			case Change:
				if member, ok := class.Members[action.Secondary]; ok {
					if member, ok := member.(*rbxdump.Property); ok {
						member.SetFields(action.Fields)
					}
				}
			}
		case Function:
			switch action.Type {
			case Add:
				if _, ok := class.Members[action.Secondary]; !ok {
					member := rbxdump.Function{Name: action.Secondary}
					member.SetFields(action.Fields)
					if class.Members == nil {
						class.Members = map[string]rbxdump.Member{}
					}
					class.Members[action.Secondary] = &member
				}
			case Remove:
				if member, ok := class.Members[action.Secondary]; ok {
					if _, ok := member.(*rbxdump.Function); ok {
						delete(class.Members, action.Secondary)
					}
				}
			case Change:
				if member, ok := class.Members[action.Secondary]; ok {
					if member, ok := member.(*rbxdump.Function); ok {
						member.SetFields(action.Fields)
					}
				}
			}
		case Event:
			switch action.Type {
			case Add:
				if _, ok := class.Members[action.Secondary]; !ok {
					member := rbxdump.Event{Name: action.Secondary}
					member.SetFields(action.Fields)
					if class.Members == nil {
						class.Members = map[string]rbxdump.Member{}
					}
					class.Members[action.Secondary] = &member
				}
			case Remove:
				if member, ok := class.Members[action.Secondary]; ok {
					if _, ok := member.(*rbxdump.Event); ok {
						delete(class.Members, action.Secondary)
					}
				}
			case Change:
				if member, ok := class.Members[action.Secondary]; ok {
					if member, ok := member.(*rbxdump.Event); ok {
						member.SetFields(action.Fields)
					}
				}
			}
		case Callback:
			switch action.Type {
			case Add:
				if _, ok := class.Members[action.Secondary]; !ok {
					member := rbxdump.Callback{Name: action.Secondary}
					member.SetFields(action.Fields)
					if class.Members == nil {
						class.Members = map[string]rbxdump.Member{}
					}
					class.Members[action.Secondary] = &member
				}
			case Remove:
				if member, ok := class.Members[action.Secondary]; ok {
					if _, ok := member.(*rbxdump.Callback); ok {
						delete(class.Members, action.Secondary)
					}
				}
			case Change:
				if member, ok := class.Members[action.Secondary]; ok {
					if member, ok := member.(*rbxdump.Callback); ok {
						member.SetFields(action.Fields)
					}
				}
			}
		}
	}
}

// PatchProperty is used to transform the embedded rbxdump.Property by applying
// a list of Actions.
type PatchProperty struct {
	*rbxdump.Property
}

// Patch implements the Patcher interface.
func (member PatchProperty) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Property:
			if action.Type == Change {
				member.SetFields(action.Fields)
			}
		}
	}
}

// PatchFunction is used to transform the embedded rbxdump.Function by applying
// a list of Actions.
type PatchFunction struct {
	*rbxdump.Function
}

// Patch implements the Patcher interface.
func (member PatchFunction) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Function:
			if action.Type == Change {
				member.SetFields(action.Fields)
			}
		}
	}
}

// PatchEvent is used to transform the embedded rbxdump.Event by applying a list
// of Actions.
type PatchEvent struct {
	*rbxdump.Event
}

// Patch implements the Patcher interface.
func (member PatchEvent) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Event:
			if action.Type == Change {
				member.SetFields(action.Fields)
			}
		}
	}
}

// PatchCallback is used to transform the embedded rbxdump.Callback by applying
// a list of Actions.
type PatchCallback struct {
	*rbxdump.Callback
}

// Patch implements the Patcher interface.
func (member PatchCallback) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Callback:
			if action.Type == Change {
				member.SetFields(action.Fields)
			}
		}
	}
}

// PatchEnum is used to transform the embedded rbxdump.Enum by applying a list
// of Actions.
type PatchEnum struct {
	*rbxdump.Enum
}

// Patch implements the Patcher interface.
func (enum PatchEnum) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case Enum:
			if action.Type == Change {
				enum.SetFields(action.Fields)
			}
		case EnumItem:
			switch action.Type {
			case Add:
				if _, ok := enum.Items[action.Secondary]; !ok {
					item := rbxdump.EnumItem{Name: action.Secondary}
					item.SetFields(action.Fields)
					if enum.Items == nil {
						enum.Items = map[string]*rbxdump.EnumItem{}
					}
					enum.Items[action.Secondary] = &item
				}
			case Remove:
				delete(enum.Items, action.Secondary)
			case Change:
				if item, ok := enum.Items[action.Secondary]; ok {
					item.SetFields(action.Fields)
				}
			}
		}
	}
}

// PatchEnumItem is used to transform the embedded rbxdump.EnumItem by applying
// a list of Actions.
type PatchEnumItem struct {
	*rbxdump.EnumItem
}

// Patch implements the Patcher interface.
func (item PatchEnumItem) Patch(actions []Action) {
	for _, action := range actions {
		switch action.Element {
		case EnumItem:
			if action.Type == Change {
				item.SetFields(action.Fields)
			}
		}
	}
}
