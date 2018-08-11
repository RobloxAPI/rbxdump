// The rbxapi package is used to represent information about the Roblox Lua
// API.
package rbxapi

// Root represents of the top-level structure of an API.
type Root interface {
	// GetClasses returns a list of class descriptors present in the API.
	GetClasses() []Class

	// GetClass returns the first class descriptor of the given name, or nil
	// if no class of the given name is present.
	GetClass(name string) Class

	// GetEnums returns a list of enum descriptors present in the API.
	GetEnums() []Enum

	// GetEnum returns the first enum descriptor of the given name, or nil if
	// no enum of the given name is present.
	GetEnum(name string) Enum
}

// Class represents a class descriptor.
type Class interface {
	// GetName returns the class name.
	GetName() string

	// GetSuperclass returns the class name of the superclass.
	GetSuperclass() string

	// GetMembers returns a list of member descriptors belonging to the class.
	GetMembers() []Member

	// GetMember returns the first member descriptor of the given name, or nil
	// if no member of the given name is present.
	GetMember(name string) Member

	Taggable
}

// Member represents a class member descriptor. A Member can be casted to a
// more specific type, depending on the member type. These are Property,
// Function, Event, and Callback.
type Member interface {
	// GetMemberType returns the type of member.
	GetMemberType() string

	// GetName returns the name of the member.
	GetName() string

	Taggable
}

// Property represents a class member of the Property member type.
type Property interface {
	Member

	// GetSecurity returns the security context associated with the property's
	// read and write access.
	GetSecurity() (read, write string)

	// GetValueType returns the type of value stored in the property.
	GetValueType() Type
}

// Function represents a class member of the Function member type.
type Function interface {
	Member

	// GetSecurity returns the security context of the member's access.
	GetSecurity() string

	// GetParameters returns the list of parameters describing the arguments
	// passed to the function. Parameters may have default values.
	GetParameters() []Parameter

	// GetReturnType returns the type of value returned by the function.
	GetReturnType() Type
}

// Event represents a class member of the Event member type.
type Event interface {
	Member

	// GetSecurity returns the security context of the member's access.
	GetSecurity() string

	// GetParameters returns the list of parameters describing the arguments
	// received from the event. Parameters will not have default values.
	GetParameters() []Parameter
}

// Callback represents a class member of the Callback member type.
type Callback interface {
	Member

	// GetSecurity returns the security context of the member's access.
	GetSecurity() string

	// GetParameters returns the list of parameters describing the arguments
	// passed to the callback. Parameters will not have default values.
	GetParameters() []Parameter

	// GetReturnType returns the type of value that should be returned by the
	// callback.
	GetReturnType() Type
}

// Parameter represents a parameter of a function, event, or callback.
type Parameter interface {
	// GetType returns the type of the parameter value.
	GetType() Type

	// GetName returns a name describing the parameter.
	GetName() string

	// GetDefault returns a string representing the default value of the
	// parameter, and whether a default value is present.
	GetDefault() (value string, ok bool)
}

// Enum represents an enum descriptor.
type Enum interface {
	// GetName returns the name of the enum.
	GetName() string

	// GetItems returns a list of items of the enum.
	GetItems() []EnumItem

	// GetItem returns the first item of the given name, or nil if no item of
	// the given name is present.
	GetItem(name string) EnumItem

	Taggable
}

// EnumItem represents an enum item descriptor.
type EnumItem interface {
	// GetName returns the name of the enum item.
	GetName() string

	// GetValue returns the value of the enum item.
	GetValue() int

	Taggable
}

// Taggable indicates that a descriptor is capable of having tags.
type Taggable interface {
	// GetTag returns whether the given tag is present in the descriptor.
	GetTag(tag string) bool

	// GetTags returns a list of all tags present in the descriptor.
	GetTags() []string
}

// Type represents a value type.
type Type interface {
	// GetName returns the name of the type.
	GetName() string

	// GetCategory returns the category of the type. This may be empty where a
	// type category is inapplicable or unavailable.
	GetCategory() string

	// String returns a string representation of the entire type.
	String() string
}
