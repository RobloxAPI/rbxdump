// The json package is used to serialize between rbxdump and Roblox JSON API
// dump format.
package json

import (
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

type jClass struct {
	Members        []jMember
	MemoryCategory string
	Name           string
	Superclass     string
	Tags           []string `json:",omitempty"`

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
	ThreadSafety  string   `json:",omitempty"`
	Tags          []string `json:",omitempty"`
	ValueType     rbxdump.Type
}

type jFunction struct {
	MemberType   string
	Name         string
	Parameters   []jParameter
	ReturnType   rbxdump.Type
	Security     string
	ThreadSafety string   `json:",omitempty"`
	Tags         []string `json:",omitempty"`
}

type jEvent struct {
	MemberType   string
	Name         string
	Parameters   []jBasicParameter
	Security     string
	ThreadSafety string   `json:",omitempty"`
	Tags         []string `json:",omitempty"`
}

type jCallback struct {
	MemberType   string
	Name         string
	Parameters   []jBasicParameter
	ReturnType   rbxdump.Type
	Security     string
	ThreadSafety string   `json:",omitempty"`
	Tags         []string `json:",omitempty"`
}

type jEnum struct {
	Items []jEnumItem
	Name  string
	Tags  []string `json:",omitempty"`
}

type jEnumItem struct {
	Name        string
	Tags        []string `json:",omitempty"`
	LegacyNames []string `json:",omitempty"`
	Value       int

	index int
}

type jParameter rbxdump.Parameter

type jBasicParameter struct {
	Name string
	Type rbxdump.Type
}
