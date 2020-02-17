package legacy

import (
	"bufio"
	"errors"
	"io"
	"strconv"

	"github.com/robloxapi/rbxapi"
)

type encoder struct {
	w      *bufio.Writer
	root   *rbxdump.Root
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
	for _, class := range e.root.Classes {
		e.encodeClass(class)
		if e.err != nil {
			goto finish
		}
	}
	for _, enum := range e.root.Enums {
		e.encodeEnum(enum)
		if e.err != nil {
			goto finish
		}
	}
finish:
	e.flush()
	return e.n, e.err
}

func (e *encoder) encodeClass(class *rbxdump.Class) {
	e.checkChars(isName, true, class.Name, "Class.Name")
	e.checkChars(isName, false, class.Superclass, "Class.Superclass")

	e.writeString(e.prefix)
	e.writeString("Class ")
	e.writeString(class.Name)
	if class.Superclass != "" {
		e.writeString(" : ")
		e.writeString(class.Superclass)
	}
	e.encodeTags(class.Tags)
	e.writeString(e.line)

	for _, member := range class.Members {
		e.encodeMember(class, member)
		if e.err != nil {
			return
		}
	}
}

func (e *encoder) encodeMember(class *rbxdump.Class, member rbxdump.Member) {
	e.checkChars(isName, true, member.MemberName(), "Member.Name")
	e.writeString(e.prefix)
	e.writeString(e.indent)

	switch member := member.(type) {
	case *rbxdump.Property:
		e.writeString("Property ")
		e.checkChars(isName, true, member.ValueType.String(), "Property.ValueType")
		e.writeString(member.ValueType.String())
		e.writeString(" ")
		e.writeString(class.Name)
		e.writeString(".")
		e.writeString(member.Name)
		e.encodeTags(member.Tags)
	case *rbxdump.Function:
		if member.Tags.GetTag("Yields") {
			e.writeString("YieldFunction ")
		} else {
			e.writeString("Function ")
		}
		e.checkChars(isName, true, member.ReturnType.String(), "Function.ReturnType")
		e.writeString(member.ReturnType.String())
		e.writeString(" ")
		e.writeString(class.Name)
		e.writeString(":")
		e.writeString(member.Name)
		e.encodeParameters(member.Parameters, true)
		e.encodeTags(member.Tags, "Yields")
	case *rbxdump.Event:
		e.writeString("Event ")
		e.writeString(class.Name)
		e.writeString(".")
		e.writeString(member.Name)
		e.encodeParameters(member.Parameters, false)
		e.encodeTags(member.Tags)
	case *rbxdump.Callback:
		e.writeString("Callback ")
		e.checkChars(isName, true, member.ReturnType.String(), "Callback.ReturnType")
		e.writeString(member.ReturnType.String())
		e.writeString(" ")
		e.writeString(class.Name)
		e.writeString(".")
		e.writeString(member.Name)
		e.encodeParameters(member.Parameters, false)
		e.encodeTags(member.Tags)
	default:
		e.setError("unknown member type")
	}
	e.writeString(e.line)
}

func (e *encoder) encodeParameters(params []rbxdump.Parameter, canDefault bool) {
	e.writeString("(")
	if len(params) > 0 {
		e.encodeParameter(params[0], canDefault)
		for i := 1; i < len(params); i++ {
			e.writeString(", ")
			e.encodeParameter(params[i], canDefault)
		}
	}
	e.writeString(")")
}

func (e *encoder) encodeParameter(param rbxdump.Parameter, canDefault bool) {
	if !canDefault && param.Optional {
		e.setError("member cannot have default argument")
		return
	}

	e.checkChars(isType, true, param.Type.String(), "Argument.Type")
	e.checkChars(isName, true, param.Name, "Argument.Name")
	e.writeString(param.Type.String())
	e.writeString(" ")
	e.writeString(param.Name)
	if param.Optional {
		e.checkChars(isDefault, false, param.Default, "Argument.Default")
		e.writeString(" = ")
		e.writeString(param.Default)
	}
}

func (e *encoder) encodeEnum(enum *rbxdump.Enum) {
	e.checkChars(isName, true, enum.Name, "Enum.Name")

	e.writeString(e.prefix)
	e.writeString("Enum ")
	e.writeString(enum.Name)
	e.encodeTags(enum.Tags)
	e.writeString(e.line)

	for _, item := range enum.Items {
		e.encodeEnumItem(enum, item)
		if e.err != nil {
			return
		}
	}
}

func (e *encoder) encodeEnumItem(enum *rbxdump.Enum, item *rbxdump.EnumItem) {
	e.checkChars(isName, true, item.Name, "EnumItem.Name")
	e.writeString(e.prefix)
	e.writeString(e.indent)
	e.writeString("EnumItem ")
	e.writeString(enum.Name)
	e.writeString(".")
	e.writeString(item.Name)
	e.writeString(" : ")
	e.writeString(strconv.Itoa(item.Value))
	e.encodeTags(item.Tags)
	e.writeString(e.line)
}

func (e *encoder) encodeTags(tags rbxdump.Tags, exclude ...string) {
loop:
	for _, tag := range tags {
		for _, ex := range exclude {
			if tag == ex {
				continue loop
			}
		}
		e.encodeTag(tag)
	}
}

func isBalanced(s string, op, cl rune) bool {
	n := 0
	for _, c := range s {
		switch c {
		case op:
			n++
		case cl:
			n--
			if n < 0 {
				return false
			}
		}
	}
	return n == 0
}

func (e *encoder) encodeTag(tag string) {
	if !isBalanced(tag, '[', ']') {
		e.setError("unbalanced tag brackets")
		return
	}
	e.writeString(" [")
	e.writeString(tag)
	e.writeString("]")
}

// Encode encodes root, writing the results to w in the API dump format.
func Encode(w io.Writer, root *rbxdump.Root) (err error) {
	e := &encoder{
		w:      bufio.NewWriter(w),
		root:   root,
		prefix: "",
		indent: "\t",
		line:   "\n",
	}
	_, err = e.encode()
	return err
}
