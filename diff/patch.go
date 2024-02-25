package diff

import (
	"maps"

	"github.com/robloxapi/rbxdump"
)

// Patch is used to transform the embedded rbxdump.Root by applying a list of
// Actions.
type Patch struct {
	*rbxdump.Root
}

// Patch implements the Patcher interface.
func (root *Patch) Patch(actions []Action) {
	if root.Root == nil {
		root.Root = &rbxdump.Root{}
	}
	for i, action := range actions {
		switch action.Element {
		case Class:
			switch action.Type {
			case Add:
				if class := root.Classes[action.Primary]; class != nil {
					class.SetFields(action.Fields)
				} else {
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
			if class := root.Classes[action.Primary]; class != nil {
				(&PatchClass{class}).Patch(actions[i : i+1])
			}
		case Enum:
			switch action.Type {
			case Add:
				if enum := root.Enums[action.Primary]; enum != nil {
					enum.SetFields(action.Fields)
				} else {
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
			if enum := root.Enums[action.Primary]; enum != nil {
				(&PatchEnum{enum}).Patch(actions[i : i+1])
			}
		}
	}
}

// Inverse implements the Inverter interface by producing the inverse of actions
// according to the root.
func (root Patch) Inverse(actions []Action) []Action {
	reversed := make([]Action, len(actions))
	for i, action := range actions {
		rev := action
		rev.Type = -rev.Type
		rev.Fields = maps.Clone(rev.Fields)
		switch rev.Type {
		case Remove:
			rev.Fields = nil
			goto finish
		case Change:
			if root.Root != nil {
				switch rev.Element {
				case Class:
					if class, ok := root.Classes[rev.Primary]; ok {
						rev.Fields = class.Fields(rev.Fields)
						goto finish
					}
				case Property, Function, Event, Callback:
					if class, ok := root.Classes[rev.Primary]; ok {
						if member, ok := class.Members[rev.Secondary]; ok {
							rev.Fields = member.Fields(rev.Fields)
							goto finish
						}
					}
				case Enum:
					if enum, ok := root.Enums[rev.Primary]; ok {
						rev.Fields = enum.Fields(rev.Fields)
						goto finish
					}
				case EnumItem:
					if enum, ok := root.Enums[rev.Primary]; ok {
						if item, ok := enum.Items[rev.Secondary]; ok {
							rev.Fields = item.Fields(rev.Fields)
							goto finish
						}
					}
				}
			}
		case Add:
			if root.Root != nil {
				switch rev.Element {
				case Class:
					if class, ok := root.Classes[rev.Primary]; ok {
						rev.Fields = class.Fields(rev.Fields)
						goto finish
					}
				case Property, Function, Event, Callback:
					if class, ok := root.Classes[rev.Primary]; ok {
						if member, ok := class.Members[rev.Secondary]; ok {
							rev.Fields = member.Fields(rev.Fields)
							goto finish
						}
					}
				case Enum:
					if enum, ok := root.Enums[rev.Primary]; ok {
						rev.Fields = enum.Fields(rev.Fields)
						goto finish
					}
				case EnumItem:
					if enum, ok := root.Enums[rev.Primary]; ok {
						if item, ok := enum.Items[rev.Secondary]; ok {
							rev.Fields = item.Fields(rev.Fields)
							goto finish
						}
					}
				}
			}
		}
		if fielder := action.ToFielder(); fielder != nil {
			rev.Fields = fielder.Fields(rev.Fields)
		} else {
			rev.Fields = rbxdump.Fields{}
		}
	finish:
		reversed[i] = rev
	}
	return reversed
}

// PatchClass is used to transform the embedded rbxdump.Class by applying a list
// of Actions.
type PatchClass struct {
	*rbxdump.Class
}

// Patch implements the Patcher interface.
func (class *PatchClass) Patch(actions []Action) {
	if class.Class == nil {
		class.Class = &rbxdump.Class{}
	}
	for _, action := range actions {
		switch action.Element {
		case Class:
			if action.Type == Change {
				class.SetFields(action.Fields)
			}
		case Property:
			patchMember[*rbxdump.Property](class, action)
		case Function:
			patchMember[*rbxdump.Function](class, action)
		case Event:
			patchMember[*rbxdump.Event](class, action)
		case Callback:
			patchMember[*rbxdump.Callback](class, action)
		}
	}
}

func patchMember[T rbxdump.Member](class *PatchClass, action Action) {
	switch action.Type {
	case Add:
		if member, ok := class.Members[action.Secondary].(T); ok {
			// Change matching type.
			member.SetFields(action.Fields)
			return
		}
		// Add new member or overwrite member of non-matching type.
		if member := action.ToMember(); member != nil {
			member.SetFields(action.Fields)
			if class.Members == nil {
				class.Members = map[string]rbxdump.Member{}
			}
			class.Members[action.Secondary] = member
		}
	case Remove:
		if _, ok := class.Members[action.Secondary].(T); ok {
			// Remove only if type matches.
			delete(class.Members, action.Secondary)
		}
	case Change:
		if member, ok := class.Members[action.Secondary].(T); ok {
			// Change only if type matches.
			member.SetFields(action.Fields)
		}
	}
}

// PatchProperty is used to transform the embedded rbxdump.Property by applying
// a list of Actions.
type PatchProperty struct {
	*rbxdump.Property
}

// Patch implements the Patcher interface.
func (member *PatchProperty) Patch(actions []Action) {
	if member.Property == nil {
		member.Property = &rbxdump.Property{}
	}
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
func (member *PatchFunction) Patch(actions []Action) {
	if member.Function == nil {
		member.Function = &rbxdump.Function{}
	}
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
func (member *PatchEvent) Patch(actions []Action) {
	if member.Event == nil {
		member.Event = &rbxdump.Event{}
	}
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
func (member *PatchCallback) Patch(actions []Action) {
	if member.Callback == nil {
		member.Callback = &rbxdump.Callback{}
	}
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
func (enum *PatchEnum) Patch(actions []Action) {
	if enum.Enum == nil {
		enum.Enum = &rbxdump.Enum{}
	}
	for _, action := range actions {
		switch action.Element {
		case Enum:
			if action.Type == Change {
				enum.SetFields(action.Fields)
			}
		case EnumItem:
			switch action.Type {
			case Add:
				if item := enum.Items[action.Secondary]; item != nil {
					item.SetFields(action.Fields)
				} else {
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
func (item *PatchEnumItem) Patch(actions []Action) {
	if item.EnumItem == nil {
		item.EnumItem = &rbxdump.EnumItem{}
	}
	for _, action := range actions {
		switch action.Element {
		case EnumItem:
			if action.Type == Change {
				item.SetFields(action.Fields)
			}
		}
	}
}
