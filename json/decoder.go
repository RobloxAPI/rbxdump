package json

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/robloxapi/rbxdump"
)

func (root *jRoot) UnmarshalJSON(b []byte) (err error) {
	var v struct{ Version int }
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch v.Version {
	case 1:
		r := struct {
			Classes []jClass
			Enums   []jEnum
		}{}
		if err := json.Unmarshal(b, &r); err != nil {
			return err
		}

		root.Classes = make(map[string]*rbxdump.Class, len(r.Classes))
		for _, jclass := range r.Classes {
			tags, pd := unmarshalTags(jclass.Tags)
			class := rbxdump.Class{
				Name:                jclass.Name,
				Superclass:          jclass.Superclass,
				MemoryCategory:      jclass.MemoryCategory,
				Members:             make(map[string]rbxdump.Member, len(jclass.Members)),
				PreferredDescriptor: pd,
				Tags:                tags,
			}
			for _, jmember := range jclass.Members {
				class.Members[jmember.MemberName()] = jmember.Member
			}
			root.Classes[class.Name] = &class
		}

		root.Enums = make(map[string]*rbxdump.Enum, len(r.Enums))
		for _, jenum := range r.Enums {
			tags, pd := unmarshalTags(jenum.Tags)
			enum := rbxdump.Enum{
				Name:                jenum.Name,
				Items:               make(map[string]*rbxdump.EnumItem, len(jenum.Items)),
				PreferredDescriptor: pd,
				Tags:                tags,
			}
			for i, jitem := range jenum.Items {
				tags, pd := unmarshalTags(jitem.Tags)
				enum.Items[jitem.Name] = &rbxdump.EnumItem{
					Name:                jitem.Name,
					Value:               jitem.Value,
					Index:               i,
					PreferredDescriptor: pd,
					Tags:                tags,
					LegacyNames:         jitem.LegacyNames,
				}
			}
			root.Enums[enum.Name] = &enum
		}
	default:
		return errVersion(v.Version)
	}
	return nil
}

func (jmember *jMember) UnmarshalJSON(b []byte) (err error) {
	var t struct{ MemberType string }
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	switch t.MemberType {
	case "Property":
		var member jProperty
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		tags, pd := unmarshalTags(member.Tags)
		jmember.Member = &rbxdump.Property{
			Name:                member.Name,
			ValueType:           rbxdump.Type(member.ValueType),
			Default:             member.Default,
			Category:            member.Category,
			ReadSecurity:        member.Security.Read,
			WriteSecurity:       member.Security.Write,
			CanLoad:             member.Serialization.CanLoad,
			CanSave:             member.Serialization.CanSave,
			ThreadSafety:        member.ThreadSafety,
			PreferredDescriptor: pd,
			Tags:                tags,
		}

	case "Function":
		var member jFunction
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		params := make([]rbxdump.Parameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = rbxdump.Parameter(param)
		}
		tags, pd := unmarshalTags(member.Tags)
		jmember.Member = &rbxdump.Function{
			Name:                member.Name,
			Parameters:          params,
			ReturnType:          member.ReturnType,
			Security:            member.Security,
			ThreadSafety:        member.ThreadSafety,
			PreferredDescriptor: pd,
			Tags:                tags,
		}

	case "Event":
		var member jEvent
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		params := make([]rbxdump.Parameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = rbxdump.Parameter{Type: rbxdump.Type(param.Type), Name: param.Name}
		}
		tags, pd := unmarshalTags(member.Tags)
		jmember.Member = &rbxdump.Event{
			Name:                member.Name,
			Parameters:          params,
			Security:            member.Security,
			ThreadSafety:        member.ThreadSafety,
			PreferredDescriptor: pd,
			Tags:                tags,
		}

	case "Callback":
		var member jCallback
		if err := json.Unmarshal(b, &member); err != nil {
			return err
		}
		params := make([]rbxdump.Parameter, len(member.Parameters))
		for i, param := range member.Parameters {
			params[i] = rbxdump.Parameter{Type: rbxdump.Type(param.Type), Name: param.Name}
		}
		tags, pd := unmarshalTags(member.Tags)
		jmember.Member = &rbxdump.Callback{
			Name:                member.Name,
			Parameters:          params,
			ReturnType:          member.ReturnType,
			Security:            member.Security,
			ThreadSafety:        member.ThreadSafety,
			PreferredDescriptor: pd,
			Tags:                tags,
		}

	default:
		return errors.New("invalid member type \"" + t.MemberType + "\"")
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (param *jParameter) UnmarshalJSON(b []byte) (err error) {
	var p struct {
		Default *string `json:",omitempty"`
		Name    string
		Type    jType
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	param.Type = rbxdump.Type(p.Type)
	param.Name = p.Name
	if p.Default != nil {
		param.Optional = true
		param.Default = *p.Default
	}
	return nil
}

// Decode parses an API dump from r in JSON format.
func Decode(r io.Reader) (root *rbxdump.Root, err error) {
	jroot := &jRoot{}
	if err = json.NewDecoder(r).Decode(jroot); err != nil {
		return nil, err
	}
	return &jroot.Root, nil
}
