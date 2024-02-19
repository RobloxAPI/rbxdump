// The json package is used to serialize between rbxdump and Roblox JSON API
// dump format.
package json

import (
	"encoding/json"
	"strconv"

	"github.com/robloxapi/rbxdump"
)

// VersionError is an error indicating that the version of the JSON format is
// unsupported.
type VersionError interface {
	error
	// VersionError returns the unsupported version.
	VersionError() int
}

// errVersion implements the VersionError interface.
type errVersion int

func (err errVersion) Error() string {
	return "version " + strconv.FormatInt(int64(err), 10) + " is unsupported"
}

type jRoot struct {
	rbxdump.Root
}

func unmarshalTags(jtags []jTag) (tags []string, pd rbxdump.PreferredDescriptor) {
	tags = make([]string, 0, len(jtags))
	for _, jtag := range jtags {
		if jtag.Preferred == nil {
			tags = append(tags, jtag.Tag)
		} else {
			pd = rbxdump.PreferredDescriptor{
				Name:         jtag.Preferred.PreferredDescriptorName,
				ThreadSafety: jtag.Preferred.ThreadSafety,
			}
		}
	}
	return tags, pd
}

type jClass struct {
	Members        []jMember
	MemoryCategory string
	Name           string
	Superclass     string
	Tags           []jTag `json:",omitempty"`

	index int
}

type jMember struct {
	rbxdump.Member
	yields int // Used to sort YieldFunctions after Functions.
}

type jProperty struct {
	Category      string
	MemberType    string
	Name          string
	Security      struct{ Read, Write string }
	Serialization struct{ CanLoad, CanSave bool }
	ThreadSafety  string `json:",omitempty"`
	Tags          []jTag `json:",omitempty"`
	ValueType     rbxdump.Type
	Default       string
}

type jReturnType []rbxdump.Type

func (t *jReturnType) UnmarshalJSON(b []byte) error {
	var one rbxdump.Type
	if json.Unmarshal(b, &one) == nil {
		*t = []rbxdump.Type{one}
		return nil
	}

	var array []rbxdump.Type
	if err := json.Unmarshal(b, &array); err != nil {
		return err
	}
	*t = array
	return nil
}

func (t jReturnType) MarshalJSON() ([]byte, error) {
	if len(t) == 1 {
		return json.Marshal((t)[0])
	}
	return json.Marshal([]rbxdump.Type(t))
}

type jFunction struct {
	MemberType   string
	Name         string
	Parameters   []jParameter
	ReturnType   jReturnType
	Security     string
	ThreadSafety string `json:",omitempty"`
	Tags         []jTag `json:",omitempty"`
}

type jEvent struct {
	MemberType   string
	Name         string
	Parameters   []jBasicParameter
	Security     string
	ThreadSafety string `json:",omitempty"`
	Tags         []jTag `json:",omitempty"`
}

type jCallback struct {
	MemberType   string
	Name         string
	Parameters   []jBasicParameter
	ReturnType   jReturnType
	Security     string
	ThreadSafety string `json:",omitempty"`
	Tags         []jTag `json:",omitempty"`
}

type jEnum struct {
	Items []jEnumItem
	Name  string
	Tags  []jTag `json:",omitempty"`
}

type jEnumItem struct {
	Name        string
	Tags        []jTag   `json:",omitempty"`
	LegacyNames []string `json:",omitempty"`
	Value       int

	index int
}

type jParameter rbxdump.Parameter

type jBasicParameter struct {
	Name string
	Type rbxdump.Type
}

type jTag struct {
	Tag       string
	Preferred *jPreferredDescriptor
}

func (tag *jTag) UnmarshalJSON(b []byte) (err error) {
	var stringTag string
	if err := json.Unmarshal(b, &stringTag); err == nil {
		tag.Tag = stringTag
		return nil
	}
	var pdTag jPreferredDescriptor
	if err := json.Unmarshal(b, &pdTag); err != nil {
		return err
	}
	tag.Preferred = &pdTag
	return nil
}

func (tag *jTag) MarshalJSON() (b []byte, err error) {
	if tag.Preferred != nil {
		return json.Marshal(*tag.Preferred)
	}
	return json.Marshal(tag.Tag)
}

type jPreferredDescriptor struct {
	PreferredDescriptorName string
	ThreadSafety            string
}
