// The diff package provides an implementation of the patch package for the
// generic rbxapi types.
package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
)

// compareAndCopyTags compares two string slices, and return copies if they
// are not equal.
func compareAndCopyTags(prev, next []string) (eq bool, p, n []string) {
	if len(prev) == len(next) {
		for i, s := range prev {
			if next[i] != s {
				goto neq
			}
		}
		return true, nil, nil
	}
neq:
	p = make([]string, len(prev))
	n = make([]string, len(next))
	copy(p, prev)
	copy(n, next)
	return false, p, n
}

// compareAndCopyParameters compares two parameter slices, and return copies
// if they are not equal.
func compareAndCopyParameters(prev, next []rbxapi.Parameter) (eq bool, p, n []rbxapi.Parameter) {
	if len(prev) != len(next) {
		for i, s := range prev {
			if next[i] != s {
				goto neq
			}
		}
		return true, nil, nil
	}
neq:
	p = make([]rbxapi.Parameter, len(prev))
	n = make([]rbxapi.Parameter, len(next))
	copy(p, prev)
	copy(n, next)
	return true, p, n
}

// Diff is a patch.Differ that finds differences between two rbxapi.Root
// values.
type Diff struct {
	Prev, Next rbxapi.Root
}

// Diff implements the patch.Differ interface.
func (d *Diff) Diff() (actions []patch.Action) {
	{
		var names map[string]struct{}
		if d.Prev != nil {
			classes := d.Prev.GetClasses()
			names = make(map[string]struct{}, len(classes))
			if d.Next == nil {
				for _, p := range classes {
					names[p.GetName()] = struct{}{}
					actions = append(actions, &ClassAction{Type: patch.Remove, Class: p})
				}
			} else {
				for _, p := range classes {
					names[p.GetName()] = struct{}{}
					n := d.Next.GetClass(p.GetName())
					if n == nil {
						actions = append(actions, &ClassAction{Type: patch.Remove, Class: p})
						continue
					}
					actions = append(actions, (&DiffClass{p, n, false}).Diff()...)
				}
			}
		}
		if d.Next != nil {
			for _, n := range d.Next.GetClasses() {
				if _, ok := names[n.GetName()]; !ok {
					actions = append(actions, &ClassAction{Type: patch.Add, Class: n})
				}
			}
		}
	}
	{
		var names map[string]struct{}
		if d.Prev != nil {
			enums := d.Prev.GetEnums()
			names = make(map[string]struct{}, len(enums))
			if d.Next == nil {
				for _, p := range enums {
					names[p.GetName()] = struct{}{}
					actions = append(actions, &EnumAction{Type: patch.Remove, Enum: p})
				}
			} else {
				for _, p := range enums {
					names[p.GetName()] = struct{}{}
					n := d.Next.GetEnum(p.GetName())
					if n == nil {
						actions = append(actions, &EnumAction{Type: patch.Remove, Enum: p})
						continue
					}
					actions = append(actions, (&DiffEnum{p, n, false}).Diff()...)
				}
			}
		}
		if d.Next != nil {
			for _, n := range d.Next.GetEnums() {
				if _, ok := names[n.GetName()]; !ok {
					actions = append(actions, &EnumAction{Type: patch.Add, Enum: n})
				}
			}
		}
	}
	return
}

// DiffClass is a patch.Differ that finds differences between two rbxapi.Class
// values.
type DiffClass struct {
	Prev, Next rbxapi.Class
	// ExcludeMembers indicates whether members should be diffed.
	ExcludeMembers bool
}

// Diff implements the patch.Differ interface.
func (d *DiffClass) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &ClassAction{Type: patch.Add, Class: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &ClassAction{Type: patch.Remove, Class: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &ClassAction{patch.Change, d.Prev, "Name", p, n})
	}
	if p, n := d.Prev.GetSuperclass(), d.Next.GetSuperclass(); p != n {
		actions = append(actions, &ClassAction{patch.Change, d.Prev, "Superclass", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &ClassAction{patch.Change, d.Prev, "Tags", p, n})
	}
	if !d.ExcludeMembers {
		members := d.Prev.GetMembers()
		names := make(map[string]struct{}, len(members))
		for _, p := range members {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetMember(p.GetName())
			if n == nil {
				actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Prev, Member: p})
				continue
			}
			switch p.GetMemberType() {
			case "Property":
				if p, ok := n.(rbxapi.Property); ok {
					if n, ok := n.(rbxapi.Property); ok {
						actions = append(actions, (&DiffProperty{d.Prev, p, n}).Diff()...)
						continue
					}
				}
			case "Function":
				if p, ok := n.(rbxapi.Function); ok {
					if n, ok := n.(rbxapi.Function); ok {
						actions = append(actions, (&DiffFunction{d.Prev, p, n}).Diff()...)
						continue
					}
				}
			case "Event":
				if p, ok := n.(rbxapi.Event); ok {
					if n, ok := n.(rbxapi.Event); ok {
						actions = append(actions, (&DiffEvent{d.Prev, p, n}).Diff()...)
						continue
					}
				}
			case "Callback":
				if p, ok := n.(rbxapi.Callback); ok {
					if n, ok := n.(rbxapi.Callback); ok {
						actions = append(actions, (&DiffCallback{d.Prev, p, n}).Diff()...)
						continue
					}
				}
			}
			actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Prev, Member: p})
			actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Prev, Member: p})
		}
		for _, n := range d.Next.GetMembers() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Prev, Member: n})
			}
		}
	}
	return
}

// DiffProperty is a patch.Differ that finds differences between two
// rbxapi.Property values.
type DiffProperty struct {
	// Class is the outer structure of the Prev value.
	Class      rbxapi.Class
	Prev, Next rbxapi.Property
}

// Diff implements the patch.Differ interface.
func (d *DiffProperty) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Class, Member: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Class, Member: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Name", p, n})
	}
	if p, n := (d.Prev.GetValueType()), d.Next.GetValueType(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "ValueType", p, n})
	}
	pr, pw := d.Prev.GetSecurity()
	nr, nw := d.Next.GetSecurity()
	if pr != nr {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "ReadSecurity", pr, nr})
	}
	if pw != nw {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "WriteSecurity", pw, nw})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

// DiffFunction is a patch.Differ that finds differences between two
// rbxapi.Function values.
type DiffFunction struct {
	// Class is the outer structure of the Prev value.
	Class      rbxapi.Class
	Prev, Next rbxapi.Function
}

// Diff implements the patch.Differ interface.
func (d *DiffFunction) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Class, Member: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Class, Member: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Name", p, n})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.GetParameters(), d.Next.GetParameters()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if p, n := (d.Prev.GetReturnType()), d.Next.GetReturnType(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "ReturnType", p, n})
	}
	if p, n := (d.Prev.GetSecurity()), d.Next.GetSecurity(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Security", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

// DiffEvent is a patch.Differ that finds differences between two rbxapi.Event
// values.
type DiffEvent struct {
	// Class is the outer structure of the Prev value.
	Class      rbxapi.Class
	Prev, Next rbxapi.Event
}

// Diff implements the patch.Differ interface.
func (d *DiffEvent) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Class, Member: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Class, Member: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Name", p, n})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.GetParameters(), d.Next.GetParameters()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if p, n := (d.Prev.GetSecurity()), d.Next.GetSecurity(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Security", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

// DiffCallback is a patch.Differ that finds differences between two
// rbxapi.Callback values.
type DiffCallback struct {
	// Class is the outer structure of the Prev value.
	Class      rbxapi.Class
	Prev, Next rbxapi.Callback
}

// Diff implements the patch.Differ interface.
func (d *DiffCallback) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &MemberAction{Type: patch.Add, Class: d.Class, Member: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &MemberAction{Type: patch.Remove, Class: d.Class, Member: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Name", p, n})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.GetParameters(), d.Next.GetParameters()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if p, n := (d.Prev.GetReturnType()), d.Next.GetReturnType(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "ReturnType", p, n})
	}
	if p, n := (d.Prev.GetSecurity()), d.Next.GetSecurity(); p != n {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Security", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

// DiffEnum is a patch.Differ that finds differences between two rbxapi.Enum
// values.
type DiffEnum struct {
	Prev, Next rbxapi.Enum
	// ExcludeItems indicates whether enum items should be diffed.
	ExcludeItems bool
}

// Diff implements the patch.Differ interface.
func (d *DiffEnum) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &EnumAction{Type: patch.Add, Enum: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &EnumAction{Type: patch.Remove, Enum: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &EnumAction{patch.Change, d.Prev, "Name", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &EnumAction{patch.Change, d.Prev, "Tags", p, n})
	}
	if !d.ExcludeItems {
		items := d.Prev.GetItems()
		names := make(map[string]struct{}, len(items))
		for _, p := range items {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetItem(p.GetName())
			if n == nil {
				actions = append(actions, &EnumItemAction{Type: patch.Remove, Enum: d.Prev, Item: p})
				continue
			}
			actions = append(actions, (&DiffEnumItem{d.Prev, p, n}).Diff()...)
		}
		for _, n := range d.Next.GetItems() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &EnumItemAction{Type: patch.Add, Enum: d.Prev, Item: n})
			}
		}
	}
	return
}

// DiffEnumItem is a patch.Differ that finds differences between two
// rbxapi.EnumItem values.
type DiffEnumItem struct {
	// Enum is the outer structure of the Prev value.
	Enum       rbxapi.Enum
	Prev, Next rbxapi.EnumItem
}

// Diff implements the patch.Differ interface.
func (d *DiffEnumItem) Diff() (actions []patch.Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, &EnumItemAction{Type: patch.Add, Enum: d.Enum, Item: d.Next})
		return
	} else if d.Next == nil {
		actions = append(actions, &EnumItemAction{Type: patch.Remove, Enum: d.Enum, Item: d.Prev})
		return
	}
	if p, n := (d.Prev.GetName()), d.Next.GetName(); p != n {
		actions = append(actions, &EnumItemAction{patch.Change, d.Enum, d.Prev, "Name", p, n})
	}
	if p, n := (d.Prev.GetValue()), d.Next.GetValue(); p != n {
		actions = append(actions, &EnumItemAction{patch.Change, d.Enum, d.Prev, "Value", p, n})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &EnumItemAction{patch.Change, d.Enum, d.Prev, "Tags", p, n})
	}
	return
}
