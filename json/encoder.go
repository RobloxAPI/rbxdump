package json

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/robloxapi/rbxdump"
)

// treeNode represents one node of a class inheritance tree.
type treeNode struct {
	name  string
	sup   string
	subs  treeNodes
	index int
}

// treeNodes sorts by name.
type treeNodes []*treeNode

func (a treeNodes) Len() int           { return len(a) }
func (a treeNodes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a treeNodes) Less(i, j int) bool { return a[i].name < a[j].name }

// setTreeIndex recursively traverses treeNodes depth-first, setting each index
// field.
func setTreeIndex(classes []jClass, nodes treeNodes, i int) int {
	for _, node := range nodes {
		classes[node.index].index = i
		if len(node.subs) > 0 {
			i = setTreeIndex(classes, node.subs, i+1)
		} else {
			i++
		}
	}
	return i
}

// jClasses sorts classes by index.
type jClasses []jClass

func (a jClasses) Len() int           { return len(a) }
func (a jClasses) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a jClasses) Less(i, j int) bool { return a[i].index < a[j].index }

// Sorts the list as an inheritance tree traversed depth-first.
func sortByInheritance(classes []jClass) {
	nodes := make(map[string]*treeNode, len(classes))
	for i, class := range classes {
		nodes[class.Name] = &treeNode{
			name:  class.Name,
			sup:   class.Superclass,
			index: i,
		}
	}

	top := treeNodes{}
	for _, node := range nodes {
		if sup := nodes[node.sup]; sup != nil {
			sup.subs = append(sup.subs, node)
		} else {
			top = append(top, node)
		}
	}

	sort.Sort(top)
	for _, node := range nodes {
		sort.Sort(node.subs)
	}

	setTreeIndex(classes, top, 0)
	sort.Sort(jClasses(classes))
}

// memberTypeOrder returns the preferred order of each type of member.
func memberTypeOrder(member rbxdump.Member) int {
	switch member.(type) {
	case *rbxdump.Callback:
		return 4
	case *rbxdump.Event:
		return 3
	case *rbxdump.Function:
		return 1
	case *rbxdump.Property:
		return 0
	}
	return -1
}

// jMembers sorts members by member type, then by name.
type jMembers []jMember

func (a jMembers) Len() int      { return len(a) }
func (a jMembers) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a jMembers) Less(i, j int) bool {
	ti := memberTypeOrder(a[i].Member)
	tj := memberTypeOrder(a[j].Member)
	if ti == tj {
		if a[i].yields == a[j].yields {
			return a[i].MemberName() < a[j].MemberName()
		}
		return a[i].yields < a[j].yields
	}
	return ti < tj
}

// jEnums sorts enums by name.
type jEnums []jEnum

func (a jEnums) Len() int           { return len(a) }
func (a jEnums) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a jEnums) Less(i, j int) bool { return a[i].Name < a[j].Name }

// jEnumItems sorts enum items by index, then name, then value.
type jEnumItems []jEnumItem

func (a jEnumItems) Len() int      { return len(a) }
func (a jEnumItems) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a jEnumItems) Less(i, j int) bool {
	if a[i].index == a[j].index {
		if a[i].Name == a[j].Name {
			return a[i].Value < a[j].Value
		}
		return a[i].Name < a[j].Name
	}
	return a[i].index < a[j].index
}

func (root jRoot) MarshalJSON() (b []byte, err error) {
	var r struct {
		Version int
		Classes []jClass
		Enums   []jEnum
	}
	r.Version = 1

	r.Classes = make([]jClass, 0, len(root.Classes))
	for _, class := range root.Classes {
		members := make([]jMember, 0, len(class.Members))
		for _, member := range class.Members {
			jmember := jMember{Member: member}
			if f, ok := member.(*rbxdump.Function); ok {
				if f.GetTag("Yields") {
					jmember.yields = 1
				}
			}
			members = append(members, jmember)
		}
		sort.Sort(jMembers(members))
		r.Classes = append(r.Classes, jClass{
			Name:           class.Name,
			Superclass:     class.Superclass,
			MemoryCategory: class.MemoryCategory,
			Members:        members,
			Tags:           class.Tags,
		})
	}
	sortByInheritance(r.Classes)

	r.Enums = make([]jEnum, 0, len(root.Enums))
	for _, enum := range root.Enums {
		items := make([]jEnumItem, 0, len(enum.Items))
		for _, item := range enum.Items {
			items = append(items, jEnumItem{
				Name:        item.Name,
				Value:       item.Value,
				Tags:        item.Tags,
				LegacyNames: item.LegacyNames,
				index:       item.Index,
			})
		}
		sort.Sort(jEnumItems(items))
		r.Enums = append(r.Enums, jEnum{
			Name:  enum.Name,
			Items: items,
			Tags:  enum.Tags,
		})
	}
	sort.Sort(jEnums(r.Enums))

	return json.Marshal(&r)
}

func (member jMember) MarshalJSON() (b []byte, err error) {
	var jmember interface{}
	switch member := member.Member.(type) {
	case *rbxdump.Property:
		m := jProperty{
			MemberType:   "Property",
			Name:         member.Name,
			ValueType:    member.ValueType,
			Category:     member.Category,
			ThreadSafety: member.ThreadSafety,
			Tags:         member.Tags,
		}
		m.Security.Read = member.ReadSecurity
		m.Security.Write = member.WriteSecurity
		m.Serialization.CanLoad = member.CanLoad
		m.Serialization.CanSave = member.CanSave
		jmember = m
	case *rbxdump.Function:
		params := make([]jParameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = jParameter(param)
		}
		m := jFunction{
			MemberType:   "Function",
			Name:         member.Name,
			Parameters:   params,
			ReturnType:   member.ReturnType,
			Security:     member.Security,
			ThreadSafety: member.ThreadSafety,
			Tags:         member.Tags,
		}
		jmember = m
	case *rbxdump.Event:
		params := make([]jBasicParameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = jBasicParameter{Type: param.Type, Name: param.Name}
		}
		m := jEvent{
			MemberType:   "Event",
			Name:         member.Name,
			Parameters:   params,
			Security:     member.Security,
			ThreadSafety: member.ThreadSafety,
			Tags:         member.Tags,
		}
		jmember = m
	case *rbxdump.Callback:
		params := make([]jBasicParameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = jBasicParameter{Type: param.Type, Name: param.Name}
		}
		m := jCallback{
			MemberType:   "Callback",
			Name:         member.Name,
			Parameters:   params,
			ReturnType:   member.ReturnType,
			Security:     member.Security,
			ThreadSafety: member.ThreadSafety,
			Tags:         member.Tags,
		}
		jmember = m
	}
	return json.Marshal(&jmember)
}

func (param jParameter) MarshalJSON() (b []byte, err error) {
	var p struct {
		Default *string `json:",omitempty"`
		Name    string
		Type    rbxdump.Type
	}
	p.Type = param.Type
	p.Name = param.Name
	if param.Optional {
		p.Default = &param.Default
	}
	return json.Marshal(&p)
}

// Encode encodes root, writing the results to w in the API dump JSON format.
func Encode(w io.Writer, root *rbxdump.Root) (err error) {
	je := json.NewEncoder(w)
	je.SetIndent("", "\t")
	je.SetEscapeHTML(false)
	return je.Encode(&jRoot{*root})
}
