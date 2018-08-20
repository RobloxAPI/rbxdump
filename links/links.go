// The links package is used to associate items in one API structure with
// items in another. Two linked items may have fields with differing values,
// but can ultimately be identified as representing the same entity across the
// two API structures.
package links

import (
	"github.com/robloxapi/rbxapi"
)

// Links represents a set of links between a previous and next API structure.
type Links struct {
	prev map[interface{}]interface{}
	next map[interface{}]interface{}
}

// Next returns the item in the next API structure that is associated with the
// given item in the previous structure. Returns nil if there is no
// association. Otherwise, the returned value is guaranteed to implement the
// same rbxapi interface as the input item.
func (l *Links) Next(v interface{}) interface{} {
	return l.next[v]
}

// Prev returns the item in the previous API structure that is associated with
// the given item in the next structure. Returns nil if there is no
// association. Otherwise, the returned value is guaranteed to implement the
// same rbxapi interface as the input item.
func (l *Links) Prev(v interface{}) interface{} {
	return l.prev[v]
}

func (l *Links) appendClasses(prev, next []rbxapi.Class) {
	// List of per-ID indices.
	pidx := make([]int, len(prev))
	{
		// Map ID to frequency of ID.
		ids := make(map[string]int)
		for i, v := range prev {
			id := v.GetName()
			pidx[i] = ids[id]
			ids[id]++
		}
	}
	nidx := make([]int, len(next))
	{
		ids := make(map[string]int)
		for i, v := range next {
			id := v.GetName()
			nidx[i] = ids[id]
			ids[id]++
		}
	}
	for pi, p := range prev {
		pid := p.GetName()
		for ni, n := range next {
			nid := n.GetName()
			if pid == nid && pidx[pi] == nidx[ni] {
				l.next[p] = n
				l.prev[n] = p
				l.appendMembers(p.GetMembers(), n.GetMembers())
			}
		}
	}
}

func (l *Links) appendMembers(prev, next []rbxapi.Member) {
	type ID struct{ typ, name string }
	pidx := make([]int, len(prev))
	{
		ids := make(map[ID]int)
		for i, v := range prev {
			id := ID{v.GetMemberType(), v.GetName()}
			pidx[i] = ids[id]
			ids[id]++
		}
	}
	nidx := make([]int, len(next))
	{
		ids := make(map[ID]int)
		for i, v := range next {
			id := ID{v.GetMemberType(), v.GetName()}
			nidx[i] = ids[id]
			ids[id]++
		}
	}
	for pi, p := range prev {
		pid := ID{p.GetMemberType(), p.GetName()}
		for ni, n := range next {
			nid := ID{n.GetMemberType(), n.GetName()}
			if pid == nid && pidx[pi] == nidx[ni] {
				l.next[p] = n
				l.prev[n] = p
			}
		}
	}
}

func (l *Links) appendEnums(prev, next []rbxapi.Enum) {
	pidx := make([]int, len(prev))
	{
		ids := make(map[string]int)
		for i, v := range prev {
			id := v.GetName()
			pidx[i] = ids[id]
			ids[id]++
		}
	}
	nidx := make([]int, len(next))
	{
		ids := make(map[string]int)
		for i, v := range next {
			id := v.GetName()
			nidx[i] = ids[id]
			ids[id]++
		}
	}
	for pi, p := range prev {
		pid := p.GetName()
		for ni, n := range next {
			nid := n.GetName()
			if pid == nid && pidx[pi] == nidx[ni] {
				l.next[p] = n
				l.prev[n] = p
				l.appendEnumItems(p.GetItems(), n.GetItems())
			}
		}
	}
}

func (l *Links) appendEnumItems(prev, next []rbxapi.EnumItem) {
	pidx := make([]int, len(prev))
	{
		ids := make(map[string]int)
		for i, v := range prev {
			id := v.GetName()
			pidx[i] = ids[id]
			ids[id]++
		}
	}
	nidx := make([]int, len(next))
	{
		ids := make(map[string]int)
		for i, v := range next {
			id := v.GetName()
			nidx[i] = ids[id]
			ids[id]++
		}
	}
	for pi, p := range prev {
		pid := p.GetName()
		for ni, n := range next {
			nid := n.GetName()
			if pid == nid && pidx[pi] == nidx[ni] {
				l.next[p] = n
				l.prev[n] = p
			}
		}
	}
}

// Append performs the same operation as Make, except the results are included
// in the current set of links.
func (l *Links) Append(prev, next rbxapi.Root) *Links {
	if prev == nil || next == nil {
		return l
	}
	l.appendClasses(prev.GetClasses(), next.GetClasses())
	l.appendEnums(prev.GetEnums(), next.GetEnums())
	return l
}

// Make creates a set of links between a previous and next API structure. A
// link is made between an item in the previous API and an item in the next
// API when the two items are determined to be representing the same entity.
//
// An entity has several identifiers used to determine its distinctiveness.
// First and foremost is the list in which the entity appears. For example, A
// class is always distinct from an enum, because they appear in different
// lists. For the same reason, a member in one class is always distinct from a
// member in another class.
//
// Secondly are the identifying fields of the entity, usually the Name.
// However, the uniqueness of these identifiers is not enforced. Such an
// ambiguity is resolved by the location in which the entity appears in its
// list. For example, two classes in the same structure that are otherwise the
// same can always be distinguished, because one class is the first of its
// kind, while the other is the second.
//
// The following fields are used to identify items:
//
//     Class:    Class.GetName
//     Member:   Member.GetName, Member.GetMemberType
//     Enum:     Enum.GetName
//     EnumItem: EnumItem.GetName
//
func Make(prev, next rbxapi.Root) *Links {
	links := &Links{
		prev: map[interface{}]interface{}{},
		next: map[interface{}]interface{}{},
	}
	return links.Append(prev, next)
}
