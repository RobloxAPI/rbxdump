// The diff package provides an implementation of the patch package for the
// rbxapijson types.
package diff

import (
	"github.com/robloxapi/rbxapi/diff"
	"github.com/robloxapi/rbxapi/patch"
	"github.com/robloxapi/rbxapi/rbxapijson"
)

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
				actions = append(actions, &patch.Class{Type: patch.Remove, Class: p})
				continue
			}
			actions = append(actions, (&DiffClass{p, n}).Diff()...)
		}
		for _, n := range d.Next.Classes {
			if _, ok := names[n.Name]; !ok {
				actions = append(actions, &patch.Class{Type: patch.Add, Class: n})
			}
		}
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Enums))
		for _, p := range d.Prev.Enums {
			names[p.Name] = struct{}{}
			n, _ := d.Next.GetEnum(p.Name).(*rbxapijson.Enum)
			if n == nil {
				actions = append(actions, &patch.Enum{Type: patch.Remove, Enum: p})
				continue
			}
			actions = append(actions, (&DiffEnum{p, n}).Diff()...)
		}
		for _, n := range d.Next.Enums {
			if _, ok := names[n.Name]; !ok {
				actions = append(actions, &patch.Enum{Type: patch.Add, Enum: n})
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
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if d.Prev.Superclass != d.Next.Superclass {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Superclass", diff.Value{d.Prev.Superclass}, diff.Value{d.Next.Superclass}})
	}
	if d.Prev.MemoryCategory != d.Next.MemoryCategory {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "MemoryCategory", diff.Value{d.Prev.MemoryCategory}, diff.Value{d.Next.MemoryCategory}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Tags", pv, nv})
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Members))
		for _, p := range d.Prev.Members {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetMember(p.GetName())
			if n == nil {
				actions = append(actions, &patch.Member{Type: patch.Remove, Class: d.Prev, Member: p})
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
			actions = append(actions, &patch.Member{Type: patch.Remove, Class: d.Prev, Member: p})
			actions = append(actions, &patch.Member{Type: patch.Add, Class: d.Prev, Member: p})
		}
		for _, n := range d.Next.Members {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.Member{Type: patch.Add, Class: d.Prev, Member: n})
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
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if d.Prev.ValueType != d.Next.ValueType {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ValueType", diff.Value{d.Prev.ValueType}, diff.Value{d.Next.ValueType}})
	}
	if d.Prev.Category != d.Next.Category {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Category", diff.Value{d.Prev.Category}, diff.Value{d.Next.Category}})
	}
	if d.Prev.ReadSecurity != d.Next.ReadSecurity {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReadSecurity", diff.Value{d.Prev.ReadSecurity}, diff.Value{d.Next.ReadSecurity}})
	}
	if d.Prev.WriteSecurity != d.Next.WriteSecurity {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "WriteSecurity", diff.Value{d.Prev.WriteSecurity}, diff.Value{d.Next.WriteSecurity}})
	}
	if d.Prev.CanLoad != d.Next.CanLoad {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "CanLoad", diff.Value{d.Prev.CanLoad}, diff.Value{d.Next.CanLoad}})
	}
	if d.Prev.CanSave != d.Next.CanSave {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "CanSave", diff.Value{d.Prev.CanSave}, diff.Value{d.Next.CanSave}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

type DiffFunction struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Function
}

func (d *DiffFunction) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if pv, nv := (diff.Value{d.Prev.GetParameters()}), (diff.Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.ReturnType != d.Next.ReturnType {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReturnType", diff.Value{d.Prev.ReturnType}, diff.Value{d.Next.ReturnType}})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", diff.Value{d.Prev.Security}, diff.Value{d.Next.Security}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

type DiffEvent struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Event
}

func (d *DiffEvent) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if pv, nv := (diff.Value{d.Prev.GetParameters()}), (diff.Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", diff.Value{d.Prev.Security}, diff.Value{d.Next.Security}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

type DiffCallback struct {
	Class      *rbxapijson.Class
	Prev, Next *rbxapijson.Callback
}

func (d *DiffCallback) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if pv, nv := (diff.Value{d.Prev.GetParameters()}), (diff.Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.ReturnType != d.Next.ReturnType {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReturnType", diff.Value{d.Prev.ReturnType}, diff.Value{d.Next.ReturnType}})
	}
	if d.Prev.Security != d.Next.Security {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", diff.Value{d.Prev.Security}, diff.Value{d.Next.Security}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

type DiffEnum struct {
	Prev, Next *rbxapijson.Enum
}

func (d *DiffEnum) Diff() (actions []patch.Action) {
	if d.Prev.Name != d.Next.Name {
		actions = append(actions, &patch.Enum{patch.Change, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Enum{patch.Change, d.Prev, "Tags", pv, nv})
	}
	{
		names := make(map[string]struct{}, len(d.Prev.Items))
		for _, p := range d.Prev.Items {
			names[p.GetName()] = struct{}{}
			n, _ := d.Next.GetItem(p.GetName()).(*rbxapijson.EnumItem)
			if n == nil {
				actions = append(actions, &patch.EnumItem{Type: patch.Remove, Enum: d.Prev, Item: p})
				continue
			}
			actions = append(actions, (&DiffEnumItem{d.Prev, p, n}).Diff()...)
		}
		for _, n := range d.Next.Items {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.EnumItem{Type: patch.Add, Enum: d.Prev, Item: n})
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
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Name", diff.Value{d.Prev.Name}, diff.Value{d.Next.Name}})
	}
	if d.Prev.Value != d.Next.Value {
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Value", diff.Value{d.Prev.Value}, diff.Value{d.Next.Value}})
	}
	if pv, nv := (diff.Value{d.Prev.GetTags()}), (diff.Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Tags", pv, nv})
	}
	return
}
