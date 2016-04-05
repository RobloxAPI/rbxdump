package dump

import (
	"bufio"
	"errors"
	"github.com/robloxapi/rbxapi"
	"io"
	"sort"
	"strconv"
)

type encoder struct {
	w      *bufio.Writer
	api    *rbxapi.API
	n      int64
	err    error
	line   string
	indent string
	prefix string
}

func (e *encoder) setError(msg string) {
	if e.err != nil {
		return
	}
	e.err = errors.New(msg)
}

func (e *encoder) write(p []byte) bool {
	if e.err != nil {
		return false
	}
	n, err := e.w.Write(p)
	e.n += int64(n)
	if err != nil {
		e.err = err
		return false
	}
	return true
}

func (e *encoder) writeByte(b byte) bool {
	if e.err != nil {
		return false
	}
	if err := e.w.WriteByte(b); err != nil {
		e.err = err
		return false
	}
	e.n += 1
	return true
}

func (e *encoder) writeString(s string) bool {
	if e.err != nil {
		return false
	}
	n, err := e.w.WriteString(s)
	e.n += int64(n)
	if err != nil {
		e.err = err
		return false
	}
	return true
}

func (e *encoder) flush() bool {
	if e.err != nil {
		return false
	}
	if err := e.w.Flush(); err != nil {
		e.err = err
		return false
	}
	return true
}

func (e *encoder) checkChars(check charCheck, noempty bool, s, msg string) {
	if len(s) == 0 {
		if noempty {
			e.setError(msg + ": unexpected empty string")
		}
		return
	}
	if check.nofix {
		if s[0] == ' ' {
			e.setError(msg + ": unexpected leading space")
		} else if s[len(s)-1] == ' ' {
			e.setError(msg + ": unexpected trailing space")
		}
		return
	}
	for _, b := range []byte(s) {
		if !check.isChar(b) {
			e.setError(msg + ": unexpected character")
			return
		}
	}
}

func (e *encoder) encode() (n int64, err error) {
	for _, class := range rbxapi.TreeList(e.api.ClassTree()) {
		e.encodeClass(class)
		if e.err != nil {
			goto finish
		}
	}
	for _, enum := range e.api.EnumList() {
		e.encodeEnum(enum)
		if e.err != nil {
			goto finish
		}
	}
finish:
	e.flush()
	return e.n, e.err
}

func (e *encoder) encodeClass(class *rbxapi.Class) {
	e.checkChars(isName, true, class.Name, "Class.Name")
	e.checkChars(isName, false, class.Superclass, "Class.Superclass")

	e.writeString(e.prefix)
	e.writeString("Class ")
	e.writeString(class.Name)
	if class.Superclass != "" {
		e.writeString(" : ")
		e.writeString(class.Superclass)
	}
	e.encodeTags(class)
	e.writeString(e.line)

	members := class.MemberList()
	sort.Sort(rbxapi.SortMembersByType(members))
	for _, member := range members {
		e.encodeMember(class, member)
		if e.err != nil {
			return
		}
	}
}

func (e *encoder) encodeMember(class *rbxapi.Class, member rbxapi.Member) {
	e.checkChars(isName, true, member.Name(), "Member.Name")
	e.checkChars(isName, true, member.Class(), "Member.Class")
	if member.Class() != class.Name {
		e.setError("member class does not match parent class")
		return
	}
	e.writeString(e.prefix)
	e.writeString(e.indent)
	e.writeString(member.Type())
	e.writeString(" ")

	switch member := member.(type) {
	case *rbxapi.Property:
		e.checkChars(isName, true, member.ValueType, "Property.ValueType")
		e.writeString(member.ValueType)
		e.writeString(" ")
		e.writeString(member.MemberClass)
		e.writeString(".")
		e.writeString(member.MemberName)
	case *rbxapi.Function:
		e.checkChars(isName, true, member.ReturnType, "Function.ReturnType")
		e.writeString(member.ReturnType)
		e.writeString(" ")
		e.writeString(member.MemberClass)
		e.writeString(":")
		e.writeString(member.MemberName)
		e.encodeArguments(member.Arguments, true)
	case *rbxapi.YieldFunction:
		e.checkChars(isName, true, member.ReturnType, "YieldFunction.ReturnType")
		e.writeString(member.ReturnType)
		e.writeString(" ")
		e.writeString(member.MemberClass)
		e.writeString(":")
		e.writeString(member.MemberName)
		e.encodeArguments(member.Arguments, true)
	case *rbxapi.Event:
		e.writeString(member.MemberClass)
		e.writeString(".")
		e.writeString(member.MemberName)
		e.encodeArguments(member.Arguments, false)
	case *rbxapi.Callback:
		e.checkChars(isName, true, member.ReturnType, "Callback.ReturnType")
		e.writeString(member.ReturnType)
		e.writeString(" ")
		e.writeString(member.MemberClass)
		e.writeString(".")
		e.writeString(member.MemberName)
		e.encodeArguments(member.Arguments, false)
	default:
		e.setError("unknown member type")
	}
	e.encodeTags(member)
	e.writeString(e.line)
}

func (e *encoder) encodeArguments(args []rbxapi.Argument, canDefault bool) {
	e.writeString("(")
	if len(args) > 0 {
		e.encodeArgument(args[0], canDefault)
		for i := 1; i < len(args); i++ {
			e.writeString(", ")
			e.encodeArgument(args[i], canDefault)
		}
	}
	e.writeString(")")
}

func (e *encoder) encodeArgument(arg rbxapi.Argument, canDefault bool) {
	if !canDefault && arg.Default != nil {
		e.setError("member cannot have default argument")
		return
	}

	e.checkChars(isType, true, arg.Type, "Argument.Type")
	e.checkChars(isName, true, arg.Name, "Argument.Name")
	e.writeString(arg.Type)
	e.writeString(" ")
	e.writeString(arg.Name)
	if arg.Default != nil {
		e.checkChars(isDefault, false, *arg.Default, "Argument.Default")
		e.writeString(" = ")
		e.writeString(*arg.Default)
	}
}

func (e *encoder) encodeEnum(enum *rbxapi.Enum) {
	e.checkChars(isName, true, enum.Name, "Enum.Name")

	e.writeString(e.prefix)
	e.writeString("Enum ")
	e.writeString(enum.Name)
	e.encodeTags(enum)
	e.writeString(e.line)

	for _, item := range enum.Items {
		e.encodeEnumItem(enum, item)
		if e.err != nil {
			return
		}
	}
}

func (e *encoder) encodeEnumItem(enum *rbxapi.Enum, item *rbxapi.EnumItem) {
	e.checkChars(isName, true, item.Name, "EnumItem.Name")
	e.checkChars(isName, true, item.Enum, "EnumItem.Enum")
	if item.Enum != enum.Name {
		e.setError("enum item enum does not match parent enum")
		return
	}
	e.writeString(e.prefix)
	e.writeString(e.indent)
	e.writeString("EnumItem ")
	e.writeString(item.Enum)
	e.writeString(".")
	e.writeString(item.Name)
	e.writeString(" : ")
	e.writeString(strconv.Itoa(item.Value))
	e.encodeTags(item)
	e.writeString(e.line)
}

func (e *encoder) encodeTags(t rbxapi.Taggable) {
	if t.TagCount() == 0 {
		return
	}

	tags := []string{}
	unsorted := map[string]bool{}
	for _, tag := range t.Tags() {
		unsorted[tag] = true
	}
loop:
	for _, group := range rbxapi.GroupOrder {
		// Filter groups not in the correct context.
		switch group.Name {
		case "MetadataItem":
		case "MetadataClass":
			if _, ok := t.(*rbxapi.Class); !ok {
				continue loop
			}
		case "MetadataProperty":
			if _, ok := t.(*rbxapi.Property); !ok {
				continue loop
			}
		case "MetadataCallback":
			if _, ok := t.(*rbxapi.Callback); !ok {
				continue loop
			}
		case "MemberSecurity":
			if _, ok := t.(rbxapi.Member); !ok {
				continue loop
			}
		}

		// Add known tags, sorted by group.
		for _, tag := range group.Tags {
			if t.Tag(tag) {
				tags = append(tags, tag)
				delete(unsorted, tag)
			}
		}
	}

	// Add unknown tags, sorted by name.
	sorted := make([]string, len(unsorted))
	i := 0
	for tag := range unsorted {
		sorted[i] = tag
		i++
	}
	sort.Strings(sorted)
	tags = append(tags, sorted...)

	for _, tag := range tags {
		e.checkChars(isTag, false, tag, "Tag")
		e.writeString(" [")
		e.writeString(tag)
		e.writeString("]")
	}
}

func Encode(w io.Writer, api *rbxapi.API) (n int64, err error) {
	e := &encoder{
		w:      bufio.NewWriter(w),
		api:    api,
		prefix: "",
		indent: "\t",
		line:   "\n",
	}
	return e.encode()
}
