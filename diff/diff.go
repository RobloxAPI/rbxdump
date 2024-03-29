package diff

import (
	"slices"
	"sort"

	"github.com/robloxapi/rbxdump"
)

// compareParams compares two parameter lists.
func compareParams(prev, next []rbxdump.Parameter) bool {
	if len(prev) != len(next) {
		return false
	}
	for i := range prev {
		pparam := prev[i]
		nparam := next[i]
		if nparam.Type != pparam.Type {
			return false
		}
		if nparam.Name != pparam.Name {
			return false
		}
		if nparam.Optional != pparam.Optional {
			return false
		}
		if nparam.Optional && nparam.Default != pparam.Default {
			return false
		}
	}
	return true
}

// compareMemberTypes returns whether two rbxdump.Members are the same type.
func compareMemberTypes(prev, next rbxdump.Member) (ok bool) {
	switch next.(type) {
	case *rbxdump.Property:
		_, ok = prev.(*rbxdump.Property)
	case *rbxdump.Function:
		_, ok = prev.(*rbxdump.Function)
	case *rbxdump.Event:
		_, ok = prev.(*rbxdump.Event)
	case *rbxdump.Callback:
		_, ok = prev.(*rbxdump.Callback)
	case nil:
		// Return true so that prev can be compared with nil.
		ok = true
	}
	return ok
}

// compareFields compares two sets of fields, returning the difference. Only
// handles field types returned by rbxdump elements.
func compareFields(prev, next rbxdump.Fields) rbxdump.Fields {
	fields := rbxdump.Fields{}
	for name, n := range next {
		p := prev[name]
		switch n := n.(type) {
		case bool:
			if p, ok := p.(bool); !ok || p != n {
				fields[name] = n
			}
		case string:
			if p, ok := p.(string); !ok || p != n {
				fields[name] = n
			}
		case []string:
			if v, ok := p.([]string); !ok || !slices.Equal(v, n) {
				fields[name] = n
			}
		case int:
			if v, ok := p.(int); !ok || v != n {
				fields[name] = n
			}
		case rbxdump.Type:
			pa := []rbxdump.Type{}
			na := []rbxdump.Type{n}
			switch p := p.(type) {
			case rbxdump.Type:
				pa = []rbxdump.Type{p}
			case []rbxdump.Type:
				pa = p
			default:
				continue
			}
			if !slices.Equal(pa, na) {
				fields[name] = n
			}
		case []rbxdump.Type:
			pa := []rbxdump.Type{}
			na := n
			switch p := p.(type) {
			case rbxdump.Type:
				pa = []rbxdump.Type{p}
			case []rbxdump.Type:
				pa = p
			default:
				continue
			}
			if !slices.Equal(pa, na) {
				fields[name] = n
			}
		case rbxdump.PreferredDescriptor:
			if p, ok := p.(rbxdump.PreferredDescriptor); !ok || p != n {
				fields[name] = n
			}
		case rbxdump.Tags:
			if p, ok := p.(rbxdump.Tags); !ok || !slices.Equal(p, n) {
				fields[name] = n
			}
		case []rbxdump.Parameter:
			if p, ok := p.([]rbxdump.Parameter); !ok || !compareParams(p, n) {
				fields[name] = n
			}
		}
	}
	for name := range prev {
		if _, ok := next[name]; ok {
			continue
		}
		fields[name] = nil
	}
	return fields
}

// appendFields appends a template action containing fields. If separate is
// true, then each field will produce a separate action. Otherwise, all fields
// will be grouped into one action.
func appendFields(actions []Action, separate bool, fields rbxdump.Fields, template Action) []Action {
	if len(fields) == 0 {
		return actions
	}

	if !separate {
		template.Fields = fields
		return append(actions, template)
	}

	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value := fields[name]
		template.Fields = rbxdump.Fields{name: value}
		actions = append(actions, template)
	}
	return actions
}

// Diff is a Differ that finds differences between two rbxdump.Root values.
type Diff struct {
	Prev, Next *rbxdump.Root
	// If true, then each change action will have exactly one field. Otherwise,
	// an action will have all changed fields grouped together.
	SeparateFields bool
}

// Diff implements the Differ interface.
func (d Diff) Diff() (actions []Action) {
	if d.Prev != nil && d.Next != nil {
		for _, p := range d.Prev.GetClasses() {
			n := d.Next.Classes[p.Name]
			actions = append(actions, DiffClass{Prev: p, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
		for _, n := range d.Next.GetClasses() {
			if p := d.Prev.Classes[n.Name]; p == nil {
				actions = append(actions, DiffClass{Next: n, SeparateFields: d.SeparateFields}.Diff()...)
			}
		}
		for _, p := range d.Prev.GetEnums() {
			n := d.Next.Enums[p.Name]
			actions = append(actions, DiffEnum{Prev: p, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
		for _, n := range d.Next.GetEnums() {
			if p := d.Prev.Enums[n.Name]; p == nil {
				actions = append(actions, DiffEnum{Next: n, SeparateFields: d.SeparateFields}.Diff()...)
			}
		}
	} else if d.Prev != nil {
		for _, p := range d.Prev.GetClasses() {
			actions = append(actions, DiffClass{Prev: p, SeparateFields: d.SeparateFields}.Diff()...)
		}
		for _, p := range d.Prev.GetEnums() {
			actions = append(actions, DiffEnum{Prev: p, SeparateFields: d.SeparateFields}.Diff()...)
		}
	} else if d.Next != nil {
		for _, n := range d.Next.GetClasses() {
			actions = append(actions, DiffClass{Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
		for _, n := range d.Next.GetEnums() {
			actions = append(actions, DiffEnum{Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
	}
	return actions
}

// DiffClass is a Differ that finds differences between two rbxdump.Class
// values.
type DiffClass struct {
	Prev, Next *rbxdump.Class
	// ExcludeMembers indicates whether members should be diffed.
	ExcludeMembers bool
	// If true, then each change action will have exactly one field. Otherwise,
	// an action will have all changed fields grouped together.
	SeparateFields bool
}

// Diff implements the Differ interface.
func (d DiffClass) Diff() (actions []Action) {
	// Handle both-nil case.
	if d.Prev == nil && d.Next == nil {
		return actions
	}

	// Handle either-nil case.
	if d.Prev == nil {
		actions = append(actions, Action{
			Type:    Add,
			Element: Class,
			Primary: d.Next.Name,
			Fields:  d.Next.Fields(nil),
		})
		if !d.ExcludeMembers {
			for _, member := range d.Next.GetMembers() {
				actions = append(actions, DiffMember{Class: d.Next.Name, Next: member, SeparateFields: d.SeparateFields}.Diff()...)
			}
		}
		return actions
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:    Remove,
			Element: Class,
			Primary: d.Prev.Name,
		})
		return actions
	}

	// Compare and append fields.
	fields := compareFields(d.Prev.Fields(nil), d.Next.Fields(nil))
	actions = appendFields(actions, d.SeparateFields, fields, Action{
		Type:    Change,
		Element: Class,
		Primary: d.Prev.Name,
	})

	// Compare members.
	if d.ExcludeMembers {
		return actions
	}
	for _, p := range d.Prev.GetMembers() {
		n := d.Next.Members[p.MemberName()]
		if compareMemberTypes(p, n) {
			actions = append(actions, DiffMember{Class: d.Prev.Name, Prev: p, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
			continue
		}
		// Member names match, but have different element types. Resolve by
		// removing the previous and adding the next.
		actions = append(actions, DiffMember{Class: d.Prev.Name, Prev: p, SeparateFields: d.SeparateFields}.Diff()...)
		actions = append(actions, DiffMember{Class: d.Prev.Name, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
	}
	for _, n := range d.Next.GetMembers() {
		if _, ok := d.Prev.Members[n.MemberName()]; !ok {
			actions = append(actions, DiffMember{Class: d.Prev.Name, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
	}

	return actions
}

// DiffMember is a Differ that finds differences between two
// rbxdump.Member values.
type DiffMember struct {
	// Class is the name of the outer structure of the Prev value.
	Class      string
	Prev, Next rbxdump.Member
	// If true, then each change action will have exactly one field. Otherwise,
	// an action will have all changed fields grouped together.
	SeparateFields bool
}

// Diff implements the Differ interface.
func (d DiffMember) Diff() (actions []Action) {
	// Handle both-nil case.
	if d.Prev == nil && d.Next == nil {
		return
	}

	// Handle either-nil case.
	if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Element:   FromElement(d.Next),
			Primary:   d.Class,
			Secondary: d.Next.MemberName(),
			Fields:    d.Next.Fields(nil),
		})
		return actions
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Element:   FromElement(d.Prev),
			Primary:   d.Class,
			Secondary: d.Prev.MemberName(),
		})
		return actions
	}

	// Compare and append fields.
	fields := compareFields(d.Prev.Fields(nil), d.Next.Fields(nil))
	actions = appendFields(actions, d.SeparateFields, fields, Action{
		Type:      Change,
		Element:   FromElement(d.Prev),
		Primary:   d.Class,
		Secondary: d.Prev.MemberName(),
	})
	return actions
}

// DiffEnum is a Differ that finds differences between two rbxdump.Enum
// values.
type DiffEnum struct {
	Prev, Next *rbxdump.Enum
	// ExcludeEnumItems indicates whether enum items should be diffed.
	ExcludeEnumItems bool
	// If true, then each change action will have exactly one field. Otherwise,
	// an action will have all changed fields grouped together.
	SeparateFields bool
}

// Diff implements the Differ interface.
func (d DiffEnum) Diff() (actions []Action) {
	// Handle both-nil case.
	if d.Prev == nil && d.Next == nil {
		return actions
	}

	// Handle either-nil case.
	if d.Prev == nil {
		actions = append(actions, Action{
			Type:    Add,
			Element: Enum,
			Primary: d.Next.Name,
			Fields:  d.Next.Fields(nil),
		})
		if !d.ExcludeEnumItems {
			for _, item := range d.Next.GetEnumItems() {
				actions = append(actions, DiffEnumItem{Enum: d.Next.Name, Next: item, SeparateFields: d.SeparateFields}.Diff()...)
			}
		}
		return actions
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:    Remove,
			Element: Enum,
			Primary: d.Prev.Name,
		})
		return actions
	}

	// Compare and append fields.
	fields := compareFields(d.Prev.Fields(nil), d.Next.Fields(nil))
	actions = appendFields(actions, d.SeparateFields, fields, Action{
		Type:    Change,
		Element: Enum,
		Primary: d.Prev.Name,
	})

	// Compare items.
	if d.ExcludeEnumItems {
		return actions
	}
	for _, p := range d.Prev.GetEnumItems() {
		n := d.Next.Items[p.Name]
		actions = append(actions, DiffEnumItem{Enum: d.Prev.Name, Prev: p, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
	}
	for _, n := range d.Next.GetEnumItems() {
		if _, ok := d.Prev.Items[n.Name]; !ok {
			actions = append(actions, DiffEnumItem{Enum: d.Prev.Name, Next: n, SeparateFields: d.SeparateFields}.Diff()...)
		}
	}

	return actions
}

// DiffEnumItem is a Differ that finds differences between two
// rbxdump.EnumItem values.
type DiffEnumItem struct {
	// Enum is the name of the outer structure of the Prev value.
	Enum       string
	Prev, Next *rbxdump.EnumItem
	// If true, then each change action will have exactly one field. Otherwise,
	// an action will have all changed fields grouped together.
	SeparateFields bool
}

// Diff implements the Differ interface.
func (d DiffEnumItem) Diff() (actions []Action) {
	// Handle both-nil case.
	if d.Prev == nil && d.Next == nil {
		return
	}

	// Handle either-nil case.
	if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Next.Name,
			Fields:    d.Next.Fields(nil),
		})
		return actions
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Prev.Name,
		})
		return actions
	}

	// Compare and append fields.
	fields := compareFields(d.Prev.Fields(nil), d.Next.Fields(nil))
	actions = appendFields(actions, d.SeparateFields, fields, Action{
		Type:      Change,
		Element:   EnumItem,
		Primary:   d.Enum,
		Secondary: d.Prev.Name,
	})
	return actions
}
