// The diff package provides an implementation of the patch package for the
// rbxapijson types.
package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/diff"
	"github.com/robloxapi/rbxapi/patch"
	"github.com/robloxapi/rbxapi/rbxapijson"
)

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

func compareAndCopyParameters(prev, next []rbxapijson.Parameter) (eq bool, p, n []rbxapi.Parameter) {
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
	for i := range prev {
		p[i] = prev[i]
	}
	for i := range next {
		n[i] = next[i]
	}
	return true, p, n
}

type Diff struct {
	Prev, Next *rbxapijson.Root
}

func (d *Diff) Diff() (actions []patch.Action) {
	{
		names := make(map[string]struct{}, len(d.Prev.Classes))
		for _, p := range d.Prev.Classes {
			names[p.Name] = struct{}{}
			n, _ := d.Next.GetClass(p.Name).(*rbxapijson.Class)
			if n == nil {
				actions = append(actions, &diff.ClassAction{Type: patch.Remove, Class: p})
				continue
			}
			actions = append(actions, (&DiffClass{p, n}).Diff()...)
		}
		for _, n := range d.Next.Classes {
			if _, ok := names[n.Name]; !ok {
				actions = append(actions, &diff.ClassAction{Type: patch.Add, Class: n})
			}
		}
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Enums))
		for _, p := range d.Prev.Enums {
			names[p.Name] = struct{}{}
			n, _ := d.Next.GetEnum(p.Name).(*rbxapijson.Enum)
			if n == nil {
				actions = append(actions, &diff.EnumAction{Type: patch.Remove, Enum: p})
				continue
			}
			actions = append(actions, (&DiffEnum{p, n}).Diff()...)
		}
		for _, n := range d.Next.Enums {
			if _, ok := names[n.Name]; !ok {
				actions = append(actions, &diff.EnumAction{Type: patch.Add, Enum: n})
			}
		}
	}
	return
}

type DiffClass struct {
	Prev, Next *rbxapijson.Class
}

func (d *DiffClass) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.ClassAction{patch.Change, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if d.Prev.Superclass != d.Next.Superclass {
		actions = append(actions, &diff.ClassAction{patch.Change, d.Prev, "Superclass", d.Prev.Superclass, d.Next.Superclass})
	}
	if d.Prev.MemoryCategory != d.Next.MemoryCategory {
		actions = append(actions, &diff.ClassAction{patch.Change, d.Prev, "MemoryCategory", d.Prev.MemoryCategory, d.Next.MemoryCategory})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.ClassAction{patch.Change, d.Prev, "Tags", p, n})
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Members))
		for _, p := range d.Prev.Members {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetMember(p.GetName())
			if n == nil {
				actions = append(actions, &diff.MemberAction{Type: patch.Remove, Class: d.Prev, Member: p})
				continue
			}
			switch p := p.(type) {
			case *rbxapijson.Property:
				if n, ok := n.(*rbxapijson.Property); ok {
					actions = append(actions, (&DiffProperty{d.Prev, p, n}).Diff()...)
					continue
				}
			case *rbxapijson.Function:
				if n, ok := n.(*rbxapijson.Function); ok {
					actions = append(actions, (&DiffFunction{d.Prev, p, n}).Diff()...)
					continue
				}
			case *rbxapijson.Event:
				if n, ok := n.(*rbxapijson.Event); ok {
					actions = append(actions, (&DiffEvent{d.Prev, p, n}).Diff()...)
					continue
				}
			case *rbxapijson.Callback:
				if n, ok := n.(*rbxapijson.Callback); ok {
					actions = append(actions, (&DiffCallback{d.Prev, p, n}).Diff()...)
					continue
				}
			}
			actions = append(actions, &diff.MemberAction{Type: patch.Remove, Class: d.Prev, Member: p})
			actions = append(actions, &diff.MemberAction{Type: patch.Add, Class: d.Prev, Member: p})
		}
		for _, n := range d.Next.Members {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &diff.MemberAction{Type: patch.Add, Class: d.Prev, Member: n})
			}
		}
	}
	return
}

type DiffProperty struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Property
}

func (d *DiffProperty) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if d.Prev.ValueType != d.Next.ValueType {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "ValueType", d.Prev.ValueType, d.Next.ValueType})
	}
	if d.Prev.Category != d.Next.Category {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Category", d.Prev.Category, d.Next.Category})
	}
	if d.Prev.ReadSecurity != d.Next.ReadSecurity {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "ReadSecurity", d.Prev.ReadSecurity, d.Next.ReadSecurity})
	}
	if d.Prev.WriteSecurity != d.Next.WriteSecurity {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "WriteSecurity", d.Prev.WriteSecurity, d.Next.WriteSecurity})
	}
	if d.Prev.CanLoad != d.Next.CanLoad {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "CanLoad", d.Prev.CanLoad, d.Next.CanLoad})
	}
	if d.Prev.CanSave != d.Next.CanSave {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "CanSave", d.Prev.CanSave, d.Next.CanSave})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

type DiffFunction struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Function
}

func (d *DiffFunction) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.Parameters, d.Next.Parameters); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if d.Prev.ReturnType != d.Next.ReturnType {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "ReturnType", d.Prev.ReturnType, d.Next.ReturnType})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Security", d.Prev.Security, d.Next.Security})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

type DiffEvent struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Event
}

func (d *DiffEvent) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.Parameters, d.Next.Parameters); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Security", d.Prev.Security, d.Next.Security})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

type DiffCallback struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Callback
}

func (d *DiffCallback) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if eq, p, n := compareAndCopyParameters(d.Prev.Parameters, d.Next.Parameters); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Parameters", p, n})
	}
	if d.Prev.ReturnType != d.Next.ReturnType {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "ReturnType", d.Prev.ReturnType, d.Next.ReturnType})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Security", d.Prev.Security, d.Next.Security})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.MemberAction{patch.Change, d.Class, d.Prev, "Tags", p, n})
	}
	return
}

type DiffEnum struct {
	Prev, Next *rbxapijson.Enum
}

func (d *DiffEnum) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.EnumAction{patch.Change, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.EnumAction{patch.Change, d.Prev, "Tags", p, n})
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Items))
		for _, p := range d.Prev.Items {
			names[p.GetName()] = struct{}{}
			n, _ := d.Next.GetItem(p.GetName()).(*rbxapijson.EnumItem)
			if n == nil {
				actions = append(actions, &diff.EnumItemAction{Type: patch.Remove, Enum: d.Prev, Item: p})
				continue
			}
			actions = append(actions, (&DiffEnumItem{d.Prev, p, n}).Diff()...)
		}
		for _, n := range d.Next.Items {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &diff.EnumItemAction{Type: patch.Add, Enum: d.Prev, Item: n})
			}
		}
	}
	return
}

type DiffEnumItem struct {
	Enum       *rbxapijson.Enum
	Prev, Next *rbxapijson.EnumItem
}

func (d *DiffEnumItem) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &diff.EnumItemAction{patch.Change, d.Enum, d.Prev, "Name", d.Prev.Name, d.Next.Name})
	}
	if d.Prev.Value != d.Next.Value {
		actions = append(actions, &diff.EnumItemAction{patch.Change, d.Enum, d.Prev, "Value", d.Prev.Value, d.Next.Value})
	}
	if eq, p, n := compareAndCopyTags(d.Prev.GetTags(), d.Next.GetTags()); !eq {
		actions = append(actions, &diff.EnumItemAction{patch.Change, d.Enum, d.Prev, "Tags", p, n})
	}
	return
}
