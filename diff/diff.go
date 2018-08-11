// The diff package provides an implementation of the patch package for the
// generic rbxapi types.
package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
	"strconv"
	"strings"
)

// Value implements a patch.Value by wrapping around one of the several types
// that can appear in a rbxapi structure. The following types are implemented:
//
//     bool
//     int
//     string
//     rbxapi.Type
//     []string
//     []rbxapi.Parameter
type Value struct {
	Value interface{}
}

// String implements the Value interface.
func (v Value) String() string {
	switch v := v.Value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case rbxapi.Type:
		return v.String()
	case []string:
		return "[" + strings.Join(v, ", ") + "]"
	case []rbxapi.Parameter:
		ss := make([]string, len(v))
		for i, param := range v {
			ss[i] = param.GetType().String() + " " + param.GetName()
			if def, ok := param.GetDefault(); ok {
				ss[i] += " = " + def
			}
		}
		return "(" + strings.Join(ss, ", ") + ")"
	}
	return ""
}

// Equal implements the Value interface.
func (v Value) Equal(w patch.Value) bool {
	switch w := w.(type) {
	case Value:
		switch v := v.Value.(type) {
		case bool:
			w, ok := w.Value.(bool)
			if !ok {
				return false
			}
			return v == w
		case int:
			w, ok := w.Value.(int)
			if !ok {
				return false
			}
			return v == w
		case string:
			w, ok := w.Value.(string)
			if !ok {
				return false
			}
			return v == w
		case rbxapi.Type:
			w, ok := w.Value.(rbxapi.Type)
			if !ok {
				return false
			}
			return v == w
		case []string:
			w, ok := w.Value.([]string)
			if !ok {
				return false
			}
			if len(w) != len(v) {
				return false
			}
			for i, v := range v {
				if w[i] != v {
					return false
				}
			}
			return true
		case []rbxapi.Parameter:
			w, ok := w.Value.([]rbxapi.Parameter)
			if !ok {
				return false
			}
			if len(w) != len(v) {
				return false
			}
			for i, v := range v {
				w := w[i]
				vd, vk := v.GetDefault()
				wd, wk := w.GetDefault()
				switch {
				case v.GetType() != w.GetType(),
					v.GetName() != w.GetName(),
					vk != wk,
					!vk && !wk && vd != wd:
					return false
				}
			}
			return true
		}
	}
	return false
}

// Diff is a patch.Differ that finds differences between two rbxapi.Root
// values.
type Diff struct {
	Prev, Next rbxapi.Root
}

// Diff implements the patch.Differ interface.
func (d *Diff) Diff() (actions []patch.Action) {
	{
		classes := d.Prev.GetClasses()
		names := make(map[string]struct{}, len(classes))
		for _, p := range classes {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetClass(p.GetName())
			if n == nil {
				actions = append(actions, &patch.Class{Type: patch.Remove, Class: p})
				continue
			}
			actions = append(actions, (&DiffClass{p, n}).Diff()...)
		}
		for _, n := range d.Next.GetClasses() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.Class{Type: patch.Add, Class: n})
			}
		}
	}
	{
		enums := d.Prev.GetEnums()
		names := make(map[string]struct{}, len(enums))
		for _, p := range enums {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetEnum(p.GetName())
			if n == nil {
				actions = append(actions, &patch.Enum{Type: patch.Remove, Enum: p})
				continue
			}
			actions = append(actions, (&DiffEnum{p, n}).Diff()...)
		}
		for _, n := range d.Next.GetEnums() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.Enum{Type: patch.Add, Enum: n})
			}
		}
	}
	return
}

// DiffClass is a patch.Differ that finds differences between two rbxapi.Class
// values.
type DiffClass struct {
	Prev, Next rbxapi.Class
}

// Diff implements the patch.Differ interface.
func (d *DiffClass) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if d.Prev.GetSuperclass() != d.Next.GetSuperclass() {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Superclass", Value{d.Prev.GetSuperclass()}, Value{d.Next.GetSuperclass()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Class{patch.Change, d.Prev, "Tags", pv, nv})
	}
	{
		members := d.Prev.GetMembers()
		names := make(map[string]struct{}, len(members))
		for _, p := range members {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetMember(p.GetName())
			if n == nil {
				actions = append(actions, &patch.Member{Type: patch.Remove, Class: d.Prev, Member: p})
				continue
			}
			switch p := p.(type) {
			case rbxapi.Property:
				if n, ok := n.(rbxapi.Property); ok {
					actions = append(actions, (&DiffProperty{d.Prev, p, n}).Diff()...)
					continue
				}
			case rbxapi.Function:
				if n, ok := n.(rbxapi.Function); ok {
					actions = append(actions, (&DiffFunction{d.Prev, p, n}).Diff()...)
					continue
				}
			case rbxapi.Event:
				if n, ok := n.(rbxapi.Event); ok {
					actions = append(actions, (&DiffEvent{d.Prev, p, n}).Diff()...)
					continue
				}
			case rbxapi.Callback:
				if n, ok := n.(rbxapi.Callback); ok {
					actions = append(actions, (&DiffCallback{d.Prev, p, n}).Diff()...)
					continue
				}
			}
			actions = append(actions, &patch.Member{Type: patch.Remove, Class: d.Prev, Member: p})
			actions = append(actions, &patch.Member{Type: patch.Add, Class: d.Prev, Member: p})
		}
		for _, n := range d.Next.GetMembers() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.Member{Type: patch.Add, Class: d.Prev, Member: n})
			}
		}
	}
	return
}

// DiffProperty is a patch.Differ that finds differences between two
// rbxapi.Property values.
type DiffProperty struct {
	// Class is the outer structure of the Prev value. It is used only for
	// context, so it may be omitted.
	Class      rbxapi.Class
	Prev, Next rbxapi.Property
}

// Diff implements the patch.Differ interface.
func (d *DiffProperty) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if d.Prev.GetValueType() != d.Next.GetValueType() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ValueType", Value{d.Prev.GetValueType()}, Value{d.Next.GetValueType()}})
	}
	pr, pw := d.Prev.GetSecurity()
	nr, nw := d.Next.GetSecurity()
	if pr != nr {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReadSecurity", Value{pr}, Value{nr}})
	}
	if pw != nw {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "WriteSecurity", Value{pw}, Value{nw}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

// DiffFunction is a patch.Differ that finds differences between two
// rbxapi.Function values.
type DiffFunction struct {
	// Class is the outer structure of the Prev value. It is used only for
	// context, so it may be omitted.
	Class      rbxapi.Class
	Prev, Next rbxapi.Function
}

// Diff implements the patch.Differ interface.
func (d *DiffFunction) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if pv, nv := (Value{d.Prev.GetParameters()}), (Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.GetReturnType() != d.Next.GetReturnType() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReturnType", Value{d.Prev.GetReturnType()}, Value{d.Next.GetReturnType()}})
	}
	if d.Prev.GetSecurity() != d.Next.GetSecurity() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", Value{d.Prev.GetSecurity()}, Value{d.Next.GetSecurity()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

// DiffEvent is a patch.Differ that finds differences between two rbxapi.Event
// values.
type DiffEvent struct {
	// Class is the outer structure of the Prev value. It is used only for
	// context, so it may be omitted.
	Class      rbxapi.Class
	Prev, Next rbxapi.Event
}

// Diff implements the patch.Differ interface.
func (d *DiffEvent) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if pv, nv := (Value{d.Prev.GetParameters()}), (Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.GetSecurity() != d.Next.GetSecurity() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", Value{d.Prev.GetSecurity()}, Value{d.Next.GetSecurity()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

// DiffCallback is a patch.Differ that finds differences between two
// rbxapi.Callback values.
type DiffCallback struct {
	// Class is the outer structure of the Prev value. It is used only for
	// context, so it may be omitted.
	Class      rbxapi.Class
	Prev, Next rbxapi.Callback
}

// Diff implements the patch.Differ interface.
func (d *DiffCallback) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if pv, nv := (Value{d.Prev.GetParameters()}), (Value{d.Next.GetParameters()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Parameters", pv, nv})
	}
	if d.Prev.GetReturnType() != d.Next.GetReturnType() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "ReturnType", Value{d.Prev.GetReturnType()}, Value{d.Next.GetReturnType()}})
	}
	if d.Prev.GetSecurity() != d.Next.GetSecurity() {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Security", Value{d.Prev.GetSecurity()}, Value{d.Next.GetSecurity()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Member{patch.Change, d.Class, d.Prev, "Tags", pv, nv})
	}
	return
}

// DiffEnum is a patch.Differ that finds differences between two rbxapi.Enum
// values.
type DiffEnum struct {
	Prev, Next rbxapi.Enum
}

// Diff implements the patch.Differ interface.
func (d *DiffEnum) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.Enum{patch.Change, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.Enum{patch.Change, d.Prev, "Tags", pv, nv})
	}
	{
		items := d.Prev.GetItems()
		names := make(map[string]struct{}, len(items))
		for _, p := range items {
			names[p.GetName()] = struct{}{}
			n := d.Next.GetItem(p.GetName())
			if n == nil {
				actions = append(actions, &patch.EnumItem{Type: patch.Remove, Enum: d.Prev, Item: p})
				continue
			}
			actions = append(actions, (&DiffEnumItem{d.Prev, p, n}).Diff()...)
		}
		for _, n := range d.Next.GetItems() {
			if _, ok := names[n.GetName()]; !ok {
				actions = append(actions, &patch.EnumItem{Type: patch.Add, Enum: d.Prev, Item: n})
			}
		}
	}
	return
}

// DiffEnumItem is a patch.Differ that finds differences between two
// rbxapi.EnumItem values.
type DiffEnumItem struct {
	// Enum is the outer structure of the Prev value. It is used only for
	// context, so it may be omitted.
	Enum       rbxapi.Enum
	Prev, Next rbxapi.EnumItem
}

// Diff implements the patch.Differ interface.
func (d *DiffEnumItem) Diff() (actions []patch.Action) {
	if d.Prev.GetName() != d.Next.GetName() {
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Name", Value{d.Prev.GetName()}, Value{d.Next.GetName()}})
	}
	if d.Prev.GetValue() != d.Next.GetValue() {
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Value", Value{d.Prev.GetValue()}, Value{d.Next.GetValue()}})
	}
	if pv, nv := (Value{d.Prev.GetTags()}), (Value{d.Next.GetTags()}); !pv.Equal(nv) {
		actions = append(actions, &patch.EnumItem{patch.Change, d.Enum, d.Prev, "Tags", pv, nv})
	}
	return
}
