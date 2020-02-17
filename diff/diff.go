package diff

import (
	"github.com/robloxapi/rbxapi"
)

// compareTags compares two string slices.
func compareTags(prev, next []string) bool {
	if len(prev) != len(next) {
		return false
	}
	for i, s := range prev {
		if next[i] != s {
			return false
		}
	}
	return true
}

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

// Diff is a Differ that finds differences between two rbxdump.Root values.
type Diff struct {
	Prev, Next *rbxdump.Root
}

// Diff implements the Differ interface.
func (d *Diff) Diff() (actions []Action) {
	{
		var names map[string]struct{}
		if d.Prev != nil {
			classes := d.Prev.GetClasses()
			names = make(map[string]struct{}, len(classes))
			if d.Next == nil {
				for _, p := range classes {
					names[p.Name] = struct{}{}
					actions = append(actions, Action{
						Type:    Remove,
						Element: Class,
						Primary: p.Name,
					})
				}
			} else {
				for _, p := range classes {
					names[p.Name] = struct{}{}
					n := d.Next.Classes[p.Name]
					if n == nil {
						actions = append(actions, Action{
							Type:    Remove,
							Element: Class,
							Primary: p.Name,
						})
						continue
					}
					actions = append(actions, (&DiffClass{p, n, false}).Diff()...)
				}
			}
		}
		if d.Next != nil {
			for _, n := range d.Next.GetClasses() {
				if _, ok := names[n.Name]; !ok {
					actions = append(actions, Action{
						Type:    Add,
						Element: Class,
						Primary: n.Name,
					})
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
					names[p.Name] = struct{}{}
					actions = append(actions, Action{
						Type:    Remove,
						Element: Enum,
						Primary: p.Name,
					})
				}
			} else {
				for _, p := range enums {
					names[p.Name] = struct{}{}
					n := d.Next.Enums[p.Name]
					if n == nil {
						actions = append(actions, Action{
							Type:    Remove,
							Element: Enum,
							Primary: p.Name,
						})
						continue
					}
					actions = append(actions, (&DiffEnum{p, n, false}).Diff()...)
				}
			}
		}
		if d.Next != nil {
			for _, n := range d.Next.GetEnums() {
				if _, ok := names[n.Name]; !ok {
					actions = append(actions, Action{
						Type:    Add,
						Element: Enum,
						Primary: n.Name,
					})
				}
			}
		}
	}
	return
}

// DiffClass is a Differ that finds differences between two rbxdump.Class
// values.
type DiffClass struct {
	Prev, Next *rbxdump.Class
	// ExcludeMembers indicates whether members should be diffed.
	ExcludeMembers bool
}

// Diff implements the Differ interface.
func (d *DiffClass) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:    Add,
			Element: Class,
			Primary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:    Remove,
			Element: Class,
			Primary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:    Change,
			Element: Class,
			Primary: d.Prev.Name,
			Fields:  rbxdump.Fields{"Name": n},
		})
	}
	if p, n := d.Prev.Superclass, d.Next.Superclass; p != n {
		actions = append(actions, Action{
			Type:    Change,
			Element: Class,
			Primary: d.Prev.Name,
			Fields:  rbxdump.Fields{"Superclass": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:    Change,
			Element: Class,
			Primary: d.Prev.Name,
			Fields:  rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	if !d.ExcludeMembers {
		members := d.Prev.GetMembers()
		names := make(map[string]struct{}, len(members))
		for _, p := range members {
			names[p.MemberName()] = struct{}{}
			n := d.Next.Members[p.MemberName()]
			if n == nil {
				actions = append(actions, Action{
					Type:      Remove,
					Element:   FromElement(p),
					Primary:   d.Prev.Name,
					Secondary: p.MemberName(),
				})
				continue
			}
			switch p := p.(type) {
			case *rbxdump.Property:
				if n, ok := n.(*rbxdump.Property); ok {
					actions = append(actions, (&DiffProperty{d.Prev.Name, p, n}).Diff()...)
					continue
				}
			case *rbxdump.Function:
				if n, ok := n.(*rbxdump.Function); ok {
					actions = append(actions, (&DiffFunction{d.Prev.Name, p, n}).Diff()...)
					continue
				}
			case *rbxdump.Event:
				if n, ok := n.(*rbxdump.Event); ok {
					actions = append(actions, (&DiffEvent{d.Prev.Name, p, n}).Diff()...)
					continue
				}
			case *rbxdump.Callback:
				if n, ok := n.(*rbxdump.Callback); ok {
					actions = append(actions, (&DiffCallback{d.Prev.Name, p, n}).Diff()...)
					continue
				}
			}
			actions = append(actions, Action{
				Type:      Remove,
				Element:   FromElement(p),
				Primary:   d.Prev.Name,
				Secondary: p.MemberName(),
			})
			actions = append(actions, Action{
				Type:      Add,
				Element:   FromElement(p),
				Primary:   d.Prev.Name,
				Secondary: p.MemberName(),
			})
		}
		for _, n := range d.Next.GetMembers() {
			if _, ok := names[n.MemberName()]; !ok {
				actions = append(actions, Action{
					Type:      Add,
					Element:   FromElement(n),
					Primary:   d.Prev.Name,
					Secondary: n.MemberName(),
				})
			}
		}
	}
	return
}

// DiffProperty is a Differ that finds differences between two
// rbxdump.Property values.
type DiffProperty struct {
	// Class is the name of the outer structure of the Prev value.
	Class      string
	Prev, Next *rbxdump.Property
}

// Diff implements the Differ interface.
func (d *DiffProperty) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Name": n},
		})
	}
	if p, n := (d.Prev.ValueType), d.Next.ValueType; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"ValueType": n},
		})
	}
	pr, pw := d.Prev.ReadSecurity, d.Prev.WriteSecurity
	nr, nw := d.Next.ReadSecurity, d.Next.WriteSecurity
	if pr != nr {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"ReadSecurity": nr},
		})
	}
	if pw != nw {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"WriteSecurity": nw},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	return
}

// DiffFunction is a Differ that finds differences between two
// rbxdump.Function values.
type DiffFunction struct {
	// Class is the name of the outer structure of the Prev value.
	Class      string
	Prev, Next *rbxdump.Function
}

// Diff implements the Differ interface.
func (d *DiffFunction) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Primary:   d.Class,
			Secondary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions,
			Action{
				Type:      Remove,
				Primary:   d.Class,
				Secondary: d.Prev.Name,
			})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Name": n},
		})
	}
	if !compareParams(d.Prev.Parameters, d.Next.Parameters) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Parameters": rbxdump.CopyParams(d.Next.Parameters)},
		})
	}
	if p, n := (d.Prev.ReturnType), d.Next.ReturnType; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"ReturnType": n},
		})
	}
	if p, n := (d.Prev.Security), d.Next.Security; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Security": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	return
}

// DiffEvent is a Differ that finds differences between two rbxdump.Event
// values.
type DiffEvent struct {
	// Class is the name of the outer structure of the Prev value.
	Class      string
	Prev, Next *rbxdump.Event
}

// Diff implements the Differ interface.
func (d *DiffEvent) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Primary:   d.Class,
			Secondary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Name": n},
		})
	}
	if !compareParams(d.Prev.Parameters, d.Next.Parameters) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Parameters": rbxdump.CopyParams(d.Next.Parameters)},
		})
	}
	if p, n := (d.Prev.Security), d.Next.Security; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Security": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	return
}

// DiffCallback is a Differ that finds differences between two
// rbxdump.Callback values.
type DiffCallback struct {
	// Class is the name of the outer structure of the Prev value.
	Class      string
	Prev, Next *rbxdump.Callback
}

// Diff implements the Differ interface.
func (d *DiffCallback) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Primary:   d.Class,
			Secondary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Name": n},
		})
	}
	if !compareParams(d.Prev.Parameters, d.Next.Parameters) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Parameters": rbxdump.CopyParams(d.Next.Parameters)},
		})
	}
	if p, n := (d.Prev.ReturnType), d.Next.ReturnType; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"ReturnType": n},
		})
	}
	if p, n := (d.Prev.Security), d.Next.Security; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Security": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Property,
			Primary:   d.Class,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	return
}

// DiffEnum is a Differ that finds differences between two rbxdump.Enum
// values.
type DiffEnum struct {
	Prev, Next *rbxdump.Enum
	// ExcludeEnumItems indicates whether enum items should be diffed.
	ExcludeEnumItems bool
}

// Diff implements the Differ interface.
func (d *DiffEnum) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:    Add,
			Element: Enum,
			Primary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:    Remove,
			Element: Enum,
			Primary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:    Change,
			Element: Enum,
			Primary: d.Prev.Name,
			Fields:  rbxdump.Fields{"Name": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Element:   Enum,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	if !d.ExcludeEnumItems {
		items := d.Prev.GetEnumItems()
		names := make(map[string]struct{}, len(items))
		for _, p := range items {
			names[p.Name] = struct{}{}
			n := d.Next.Items[p.Name]
			if n == nil {
				actions = append(actions, Action{
					Type:      Remove,
					Element:   EnumItem,
					Primary:   d.Prev.Name,
					Secondary: p.Name,
				})
				continue
			}
			actions = append(actions, (&DiffEnumItem{d.Prev.Name, p, n}).Diff()...)
		}
		for _, n := range d.Next.GetEnumItems() {
			if _, ok := names[n.Name]; !ok {
				actions = append(actions, Action{
					Type:      Add,
					Element:   EnumItem,
					Primary:   d.Prev.Name,
					Secondary: n.Name,
				})
			}
		}
	}
	return
}

// DiffEnumItem is a Differ that finds differences between two
// rbxdump.EnumItem values.
type DiffEnumItem struct {
	// Enum is the name of the outer structure of the Prev value.
	Enum       string
	Prev, Next *rbxdump.EnumItem
}

// Diff implements the Differ interface.
func (d *DiffEnumItem) Diff() (actions []Action) {
	if d.Prev == nil && d.Next == nil {
		return
	} else if d.Prev == nil {
		actions = append(actions, Action{
			Type:      Add,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Next.Name,
		})
		return
	} else if d.Next == nil {
		actions = append(actions, Action{
			Type:      Remove,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Prev.Name,
		})
		return
	}
	if p, n := (d.Prev.Name), d.Next.Name; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Name": n},
		})
	}
	if p, n := (d.Prev.Value), d.Next.Value; p != n {
		actions = append(actions, Action{
			Type:      Change,
			Element:   EnumItem,
			Primary:   d.Enum,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Value": n},
		})
	}
	if !compareTags(d.Prev.Tags, d.Next.Tags) {
		actions = append(actions, Action{
			Type:      Change,
			Primary:   d.Enum,
			Secondary: d.Prev.Name,
			Fields:    rbxdump.Fields{"Tags": d.Next.GetTags()},
		})
	}
	return
}
