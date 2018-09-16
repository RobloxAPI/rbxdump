package rbxapidump

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
)

// copyClass returns a deep copy of a generic rbxapi.Class.
func copyClass(class rbxapi.Class) *Class {
	members := class.GetMembers()
	c := Class{
		Name:       class.GetName(),
		Superclass: class.GetSuperclass(),
		Members:    make([]rbxapi.Member, 0, len(members)),
		Tags:       Tags(class.GetTags()),
	}
	for _, member := range members {
		if member := copyMember(member); member != nil {
			switch member := member.(type) {
			case *Property:
				member.Class = class.GetName()
			case *Function:
				member.Class = class.GetName()
			case *Event:
				member.Class = class.GetName()
			case *Callback:
				member.Class = class.GetName()
			}
			c.Members = append(c.Members, member)
		}
	}
	return &c
}

// copyMember returns a deep copy of a generic rbxapi.Member.
func copyMember(member rbxapi.Member) rbxapi.Member {
	switch member := member.(type) {
	case rbxapi.Property:
		return &Property{
			Name:      member.GetName(),
			ValueType: copyType(member.GetValueType()),
			Tags:      Tags(member.GetTags()),
		}
	case rbxapi.Function:
		// Function and Callback have the same methods.
		switch member.GetMemberType() {
		case "Function":
			return &Function{
				Name:       member.GetName(),
				ReturnType: copyType(member.GetReturnType()),
				Parameters: copyParameters(member.GetParameters()),
				Tags:       Tags(member.GetTags()),
			}
		case "Callback":
			return &Callback{
				Name:       member.GetName(),
				ReturnType: copyType(member.GetReturnType()),
				Parameters: copyParameters(member.GetParameters()),
				Tags:       Tags(member.GetTags()),
			}
		}
	case rbxapi.Event:
		return &Event{
			Name:       member.GetName(),
			Parameters: copyParameters(member.GetParameters()),
			Tags:       Tags(member.GetTags()),
		}
	}
	return nil
}

// copyEnum returns a deep copy of a generic rbxapi.Enum.
func copyEnum(enum rbxapi.Enum) *Enum {
	items := enum.GetEnumItems()
	e := Enum{
		Name:      enum.GetName(),
		EnumItems: make([]*EnumItem, 0, len(items)),
		Tags:      Tags(enum.GetTags()),
	}
	for _, item := range items {
		if item := copyEnumItem(item); item != nil {
			item.Enum = enum.GetName()
			e.EnumItems = append(e.EnumItems, item)
		}
	}
	return &e
}

// copyEnumItem returns a deep copy of a generic rbxapi.EnumItem.
func copyEnumItem(item rbxapi.EnumItem) *EnumItem {
	return &EnumItem{
		Name:  item.GetName(),
		Value: item.GetValue(),
		Tags:  item.GetTags(),
	}
}

// copyParameters returns a deep copy of a list of generic rbxapi.Parameter
// values.
func copyParameters(params []rbxapi.Parameter) []Parameter {
	p := make([]Parameter, len(params))
	for i, param := range params {
		p[i].Type = copyType(param.GetType())
		p[i].Name = param.GetName()
		if d, ok := param.GetDefault(); ok {
			p[i].Default = &d
		}
	}
	return p
}

// copyType returns a deep copy of a generic rbxapi.Type.
func copyType(typ rbxapi.Type) Type {
	var t Type
	t.SetFromType(typ)
	return t
}

// Patch transforms the API structure by applying a list of patch actions.
//
// Patch implements the patch.Patcher interface.
func (root *Root) Patch(actions []patch.Action) {
	for _, action := range actions {
		switch action := action.(type) {
		case patch.Member:
			amember := action.GetMember()
			if amember == nil {
				continue
			}
			var class *Class
			{
				aclass := action.GetClass()
				if aclass == nil {
					continue
				}
				name := aclass.GetName()
				for _, c := range root.Classes {
					if c.Name == name {
						class = c
						break
					}
				}
			}
			if class == nil {
				continue
			}
			switch action.GetType() {
			case patch.Remove:
				name := amember.GetName()
				for i, member := range class.Members {
					if member.GetName() == name {
						copy(class.Members[i:], class.Members[i+1:])
						class.Members[len(class.Members)-1] = nil
						class.Members = class.Members[:len(class.Members)-1]
						break
					}
				}
			case patch.Add:
				class.Members = append(class.Members, copyMember(amember))
			case patch.Change:
				var member rbxapi.Member
				{
					name := amember.GetName()
					mtype := amember.GetMemberType()
					for _, m := range class.Members {
						if m.GetName() == name && m.GetMemberType() == mtype {
							member = m
							break
						}
					}
				}
				if member == nil {
					continue
				}
				switch member := member.(type) {
				case *Property:
					switch action.GetField() {
					case "Name":
						if v, ok := action.GetNext().(string); ok {
							member.Name = v
						}
					case "ValueType":
						switch v := action.GetNext().(type) {
						case rbxapi.Type:
							member.ValueType.SetFromType(v)
						case string:
							member.ValueType = Type(v)
						}
					case "Tags":
						if v, ok := action.GetNext().([]string); ok {
							member.Tags = Tags(Tags(v).GetTags())
						}
					}
				case *Function:
					switch action.GetField() {
					case "Name":
						if v, ok := action.GetNext().(string); ok {
							member.Name = v
						}
					case "ReturnType":
						switch v := action.GetNext().(type) {
						case rbxapi.Type:
							member.ReturnType.SetFromType(v)
						case string:
							member.ReturnType = Type(v)
						}
					case "Parameters":
						if v, ok := action.GetNext().([]rbxapi.Parameter); ok {
							member.Parameters = make([]Parameter, len(v))
							for i, param := range v {
								member.Parameters[i].Type.SetFromType(param.GetType())
								member.Parameters[i].Name = param.GetName()
								if d, ok := param.GetDefault(); ok {
									member.Parameters[i].Default = &d
								}
							}
						}
					case "Tags":
						if v, ok := action.GetNext().([]string); ok {
							member.Tags = Tags(Tags(v).GetTags())
						}
					}
				case *Event:
					switch action.GetField() {
					case "Name":
						if v, ok := action.GetNext().(string); ok {
							member.Name = v
						}
					case "Parameters":
						if v, ok := action.GetNext().([]rbxapi.Parameter); ok {
							member.Parameters = make([]Parameter, len(v))
							for i, param := range v {
								member.Parameters[i].Type.SetFromType(param.GetType())
								member.Parameters[i].Name = param.GetName()
								if d, ok := param.GetDefault(); ok {
									member.Parameters[i].Default = &d
								}
							}
						}
					case "Tags":
						if v, ok := action.GetNext().([]string); ok {
							member.Tags = Tags(Tags(v).GetTags())
						}
					}
				case *Callback:
					switch action.GetField() {
					case "Name":
						if v, ok := action.GetNext().(string); ok {
							member.Name = v
						}
					case "ReturnType":
						switch v := action.GetNext().(type) {
						case rbxapi.Type:
							member.ReturnType.SetFromType(v)
						case string:
							member.ReturnType = Type(v)
						}
					case "Parameters":
						if v, ok := action.GetNext().([]rbxapi.Parameter); ok {
							member.Parameters = make([]Parameter, len(v))
							for i, param := range v {
								member.Parameters[i].Type.SetFromType(param.GetType())
								member.Parameters[i].Name = param.GetName()
								if d, ok := param.GetDefault(); ok {
									member.Parameters[i].Default = &d
								}
							}
						}
					case "Tags":
						if v, ok := action.GetNext().([]string); ok {
							member.Tags = Tags(Tags(v).GetTags())
						}
					}
				}
			}
		case patch.Class:
			aclass := action.GetClass()
			if aclass == nil {
				continue
			}
			switch action.GetType() {
			case patch.Remove:
				name := aclass.GetName()
				for i, class := range root.Classes {
					if class.Name == name {
						copy(root.Classes[i:], root.Classes[i+1:])
						root.Classes[len(root.Classes)-1] = nil
						root.Classes = root.Classes[:len(root.Classes)-1]
						break
					}
				}
			case patch.Add:
				root.Classes = append(root.Classes, copyClass(aclass))
			case patch.Change:
				var class *Class
				{
					name := aclass.GetName()
					for _, c := range root.Classes {
						if c.Name == name {
							class = c
							break
						}
					}
				}
				if class == nil {
					continue
				}
				switch action.GetField() {
				case "Name":
					if v, ok := action.GetNext().(string); ok {
						class.Name = v
					}
				case "Superclass":
					if v, ok := action.GetNext().(string); ok {
						class.Superclass = v
					}
				case "Tags":
					if v, ok := action.GetNext().([]string); ok {
						class.Tags = Tags(Tags(v).GetTags())
					}
				}
			}
		case patch.EnumItem:
			aitem := action.GetEnumItem()
			if aitem == nil {
				continue
			}
			var enum *Enum
			{
				aenum := action.GetEnum()
				if aenum == nil {
					continue
				}
				name := aenum.GetName()
				for _, e := range root.Enums {
					if e.Name == name {
						enum = e
						break
					}
				}
			}
			if enum == nil {
				continue
			}
			switch action.GetType() {
			case patch.Remove:
				name := aitem.GetName()
				for i, item := range enum.EnumItems {
					if item.GetName() == name {
						copy(enum.EnumItems[i:], enum.EnumItems[i+1:])
						enum.EnumItems[len(enum.EnumItems)-1] = nil
						enum.EnumItems = enum.EnumItems[:len(enum.EnumItems)-1]
						break
					}
				}
			case patch.Add:
				enum.EnumItems = append(enum.EnumItems, copyEnumItem(aitem))
			case patch.Change:
				var item *EnumItem
				{
					name := aitem.GetName()
					for _, i := range enum.EnumItems {
						if i.GetName() == name {
							item = i
							break
						}
					}
				}
				if item == nil {
					continue
				}
				switch action.GetField() {
				case "Name":
					if v, ok := action.GetNext().(string); ok {
						item.Name = v
					}
				case "Value":
					if v, ok := action.GetNext().(int); ok {
						item.Value = v
					}
				case "Tags":
					if v, ok := action.GetNext().([]string); ok {
						item.Tags = Tags(Tags(v).GetTags())
					}
				}
			}
		case patch.Enum:
			aenum := action.GetEnum()
			if aenum == nil {
				continue
			}
			switch action.GetType() {
			case patch.Remove:
				name := aenum.GetName()
				for i, enum := range root.Enums {
					if enum.Name == name {
						copy(root.Enums[i:], root.Enums[i+1:])
						root.Enums[len(root.Enums)-1] = nil
						root.Enums = root.Enums[:len(root.Enums)-1]
						break
					}
				}
			case patch.Add:
				root.Enums = append(root.Enums, copyEnum(aenum))
			case patch.Change:
				var enum *Enum
				{
					name := aenum.GetName()
					for _, e := range root.Enums {
						if e.Name == name {
							enum = e
							break
						}
					}
				}
				if enum == nil {
					continue
				}
				switch action.GetField() {
				case "Name":
					if v, ok := action.GetNext().(string); ok {
						enum.Name = v
					}
				case "Tags":
					if v, ok := action.GetNext().([]string); ok {
						enum.Tags = Tags(Tags(v).GetTags())
					}
				}
			}
		}
	}
}
