package rbxapijson

import (
	"encoding/json"
	"errors"
	"github.com/robloxapi/rbxapi"
	"strconv"
)

type ErrVersion int

func (err ErrVersion) Error() string {
	return "version " + strconv.FormatInt(int64(err), 10) + " is unsupported"
}

func (root *Root) UnmarshalJSON(b []byte) (err error) {
	var v struct{ Version int }
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch v.Version {
	case 1:
		r := struct {
			Classes []*Class
			Enums   []*Enum
		}{}
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}
		*root = Root(r)
	default:
		return ErrVersion(v.Version)
	}
	return nil
}

type jsonMember struct {
	MemberType string
	rbxapi.Member
}

func (jmember *jsonMember) UnmarshalJSON(b []byte) (err error) {
	var t struct{ MemberType string }
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	switch t.MemberType {
	case "Property":
		var member Property
		// Unmarshal matching fields.
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		// Unmarshal fields where the JSON structure differs.
		var extra struct {
			Security      struct{ Read, Write string }
			Serialization struct{ CanLoad, CanSave bool }
		}
		if err := json.Unmarshal(b, &extra); err != nil {
			return err
		}
		member.ReadSecurity = extra.Security.Read
		member.WriteSecurity = extra.Security.Write
		member.CanLoad = extra.Serialization.CanLoad
		member.CanSave = extra.Serialization.CanSave

		jmember.Member = &member

	case "Function":
		var member Function
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		jmember.Member = &member

	case "Event":
		var member Event
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		jmember.Member = &member

	case "Callback":
		var member Callback
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		jmember.Member = &member

	default:
		return errors.New("invalid member type \"" + t.MemberType + "\"")
	}
	return nil
}

func (class *Class) UnmarshalJSON(b []byte) (err error) {
	var c struct {
		Name           string
		Superclass     string
		MemoryCategory string
		Members        []jsonMember
		Tags
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}

	class.Name = c.Name
	class.Superclass = c.Superclass
	class.MemoryCategory = c.MemoryCategory
	class.Tags = c.Tags
	class.Members = make([]rbxapi.Member, len(c.Members))
	for i, m := range c.Members {
		class.Members[i] = m.Member
	}
	return nil
}
