package rbxapijson

import (
	"encoding/json"
	"io"
)

func (root *Root) MarshalJSON() (b []byte, err error) {
	r := struct {
		Version int
		Classes []*Class
		Enums   []*Enum
	}{1, root.Classes, root.Enums}
	return json.Marshal(&r)
}

func (class *Class) MarshalJSON() (b []byte, err error) {
	var c struct {
		Name           string
		Superclass     string
		MemoryCategory string
		Members        []interface{}
		Tags           `json:",omitempty"`
	}
	c.Name = class.Name
	c.Superclass = class.Superclass
	c.MemoryCategory = class.MemoryCategory
	c.Tags = class.Tags
	c.Members = make([]interface{}, len(class.Members))
	for i, m := range class.Members {
		switch m := m.(type) {
		case *Property:
			type security struct {
				Read  string
				Write string
			}
			type serialization struct {
				CanLoad bool
				CanSave bool
			}
			c.Members[i] = struct {
				MemberType    string
				Name          string
				ValueType     Type
				Category      string
				Security      security
				Serialization serialization
				Tags          `json:",omitempty"`
			}{
				MemberType:    "Property",
				Name:          m.Name,
				ValueType:     m.ValueType,
				Category:      m.Category,
				Security:      security{Read: m.ReadSecurity, Write: m.WriteSecurity},
				Serialization: serialization{CanLoad: m.CanLoad, CanSave: m.CanSave},
				Tags:          m.Tags,
			}
		case *Function:
			c.Members[i] = struct {
				MemberType string
				*Function
			}{m.GetMemberType(), m}
		case *Event:
			c.Members[i] = struct {
				MemberType string
				*Event
			}{m.GetMemberType(), m}
		case *Callback:
			c.Members[i] = struct {
				MemberType string
				*Callback
			}{m.GetMemberType(), m}
		}
	}
	return json.Marshal(&c)
}

func Encode(w io.Writer, root *Root) (err error) {
	je := json.NewEncoder(w)
	je.SetIndent("", "\t")
	je.SetEscapeHTML(false)
	return je.Encode(root)
}
