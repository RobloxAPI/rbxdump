package legacy

import (
	"bufio"
	"bytes"
	"io"
	"strconv"

	"github.com/robloxapi/rbxapi"
)

// SyntaxError indicates that a syntax error occurred while decoding.
type SyntaxError interface {
	error
	// SyntaxError returns an error message and the line on which the error
	// occurred.
	SyntaxError() (msg string, line int)
}

// syntaxError implements the SyntaxError interface.
type syntaxError struct {
	Msg  string
	Line int
}

func (e *syntaxError) Error() string {
	return "error on line " + strconv.Itoa(e.Line) + ": " + e.Msg
}

type decoder struct {
	root  *rbxdump.Root
	r     io.ByteReader
	next  []byte
	buf   bytes.Buffer
	n     int64
	err   error
	line  int
	class *rbxdump.Class
	enum  *rbxdump.Enum
}

// Creates a syntaxError with the current line number.
func (d *decoder) syntaxError(msg string) {
	if d.err != nil && d.err != io.EOF {
		return
	}
	d.err = &syntaxError{Msg: msg, Line: d.line}
}

func (d *decoder) getc() (b byte, ok bool) {
	if d.err != nil {
		return 0, false
	}

	if len(d.next) > 0 {
		b, d.next = d.next[len(d.next)-1], d.next[:len(d.next)-1]
	} else {
		b, d.err = d.r.ReadByte()
		if d.err != nil {
			return 0, false
		}
		d.n++
	}
	if b == '\n' {
		d.line++
	}

	return b, true
}

func (d *decoder) ungetc(b byte) {
	if b == '\n' {
		d.line--
	}
	d.next = append(d.next, b)
}

// Get the next character and compare it. Unget if the comparison fails.
func (d *decoder) checkChar(c byte) bool {
	b, ok := d.getc()
	if !ok {
		return false
	}
	if b != c {
		d.ungetc(b)
		return false
	}
	return true
}

func (d *decoder) expectChar(c byte) {
	b, ok := d.getc()
	if !ok {
		return
	}
	if b != c {
		d.syntaxError("expected '" + string(c) + "'")
	}
}

// Decode characters that satisfy the isChar function. If a space satisfies
// isChar, then trailing spaces will be excluded from the result. Assumes that
// leading spaces have already been read.
func (d *decoder) decodeChars(check charCheck) string {
	if d.err != nil {
		return ""
	}
	d.buf.Reset()
	width := 0
	for {
		b, ok := d.getc()
		if !ok {
			goto finish
		}
		if !check.isChar(b) {
			d.ungetc(b)
			goto finish
		}
		if b == ' ' {
			width++
		} else {
			width = 0
		}
		d.buf.WriteByte(b)
	}
finish:
	b := d.buf.Bytes()
	if check.nofix {
		// Remove trailing spaces.
		for i := 0; i < width; i++ {
			d.ungetc(b[len(b)-1-i])
		}
		b = b[:len(b)-width]
	}
	return string(b)
}

// Decode characters from the given balanced brackets. Assumes the first
// opening bracket has already been decoded, and excludes the last closing
// bracket from the result. Spaces are treated the same as in decodeChars.
func (d *decoder) decodeNested(openChar, closeChar byte) string {
	if d.err != nil {
		return ""
	}
	d.buf.Reset()
	width := 0
	for depth := 1; ; {
		b, ok := d.getc()
		if !ok {
			goto finish
		}
		switch b {
		case openChar:
			depth++
		case closeChar:
			depth--
			if depth <= 0 {
				goto finish
			}
		}
		if b == ' ' {
			width++
		} else {
			width = 0
		}
		d.buf.WriteByte(b)
	}
finish:
	b := d.buf.Bytes()
	return string(b[:len(b)-width])
}

func (d *decoder) expectChars(check charCheck, msg string) (s string) {
	if s = d.decodeChars(check); s == "" {
		d.syntaxError("expected " + msg)
		return ""
	}
	return s
}

// Clear the parent class or enum.
func (d *decoder) clearParent() {
	d.class = nil
	d.enum = nil
}

// Expect a class as the parent item.
func (d *decoder) expectClass(name string) {
	if d.err != nil {
		return
	}
	if d.class == nil {
		d.syntaxError("expected previously-declared class")
		return
	}
	if d.class.Name != name {
		d.syntaxError("member of class '" + name + "' does not match class '" + d.class.Name + "'")
		return
	}
}

// Expect an enum as the parent item.
func (d *decoder) expectEnum(name string) {
	if d.err != nil {
		return
	}
	if d.enum == nil {
		d.syntaxError("expected previously-declared enum")
		return
	}
	if d.enum.Name != name {
		d.syntaxError("enum item of enum '" + name + "' does not match enum '" + d.enum.Name + "'")
		return
	}
}

// Expect a decoded string to be non-empty.
func (d *decoder) expectString(v string, msg string) {
	if v == "" {
		d.syntaxError("expected " + msg)
	}
}

// Expect at least one whitespace character.
func (d *decoder) expectWhitespace() {
	if d.decodeChars(isSpace) == "" {
		d.syntaxError("expected whitespace")
	}
}

// Expect zero or more whitespace characters.
func (d *decoder) skipWhitespace() {
	d.decodeChars(isSpace)
}

// Expect an integer.
func (d *decoder) expectInt() int {
	i, err := strconv.Atoi(d.decodeChars(isInt))
	if err != nil {
		d.syntaxError("expected integer")
	}
	return i
}

// Add a class to the API. Sets class parent.
func (d *decoder) addClass(class *rbxdump.Class) {
	if d.err != nil {
		return
	}
	d.root.Classes[class.Name] = class
	d.class = class
}

// Add an enum to the API. Sets enum parent.
func (d *decoder) addEnum(enum *rbxdump.Enum) {
	if d.err != nil {
		return
	}
	d.root.Enums[enum.Name] = enum
	d.enum = enum
}

// Add a member to the parent class. Assumes the parent class exists.
func (d *decoder) addMember(member rbxdump.Member) {
	if d.err != nil {
		return
	}
	d.class.Members[member.MemberName()] = member
}

// Add an  enum item to the parent enum. Assumes the parent enum exists.
func (d *decoder) addEnumItem(item *rbxdump.EnumItem) {
	if d.err != nil {
		return
	}
	d.enum.Items[item.Name] = item
}

func (d *decoder) decode() error {
	d.decodeLine()
	if d.err == io.EOF {
		return nil
	}

	for d.err == nil {
		d.decodeItem()
		// Skip over whitespace between items. Expect at least one EOL, but
		// only if we aren't at EOF.
		if !d.decodeLine() && d.err != io.EOF {
			d.syntaxError("expected end-of-line")
		}
	}
	if d.err != io.EOF {
		return d.err
	}
	return nil
}

// Skips any whitespace and lines. Returns whether at least one line was
// decoded.
func (d *decoder) decodeLine() (line bool) {
	if d.err != nil {
		return
	}
	for {
		b, ok := d.getc()
		if !ok {
			return
		}
		switch b {
		case ' ', '\t', '\f', '\r':
			// continue
		case '\n':
			line = true
		default:
			d.ungetc(b)
			return
		}
	}
}

func (d *decoder) decodeItem() {
	word := d.expectChars(isWord, "item type")
	d.expectWhitespace()
	switch word {
	case "Class":
		d.decodeClass()
	case "Enum":
		d.decodeEnum()
	case "Property":
		d.decodeProperty()
	case "Function":
		d.decodeFunction(false)
	case "YieldFunction":
		d.decodeFunction(true)
	case "Event":
		d.decodeEvent()
	case "Callback":
		d.decodeCallback()
	case "EnumItem":
		d.decodeEnumItem()
	default:
		d.syntaxError("unknown item type")
	}
}

func (d *decoder) decodeClass() {
	d.clearParent()
	class := rbxdump.Class{Members: make(map[string]rbxdump.Member)}
	class.Name = d.expectChars(isClassName, "class name")
	d.skipWhitespace()
	if d.checkChar(':') {
		d.skipWhitespace()
		class.Superclass = d.expectChars(isClassName, "superclass name")
		d.skipWhitespace()
	}
	d.decodeTags(&class.Tags)
	d.addClass(&class)
}

func (d *decoder) decodeProperty() {
	var member rbxdump.Property
	member.ValueType = rbxdump.Type{Name: d.expectChars(isType, "value type")}
	d.expectWhitespace()
	class := d.expectChars(isClassName, "member class")
	d.expectClass(class)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeFunction(yields bool) {
	var member rbxdump.Function
	member.ReturnType = rbxdump.Type{Name: d.expectChars(isType, "return type")}
	d.expectWhitespace()
	class := d.expectChars(isClassName, "member class")
	d.expectClass(class)
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(true)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	if yields {
		member.Tags.SetTag("Yields")
	} else {
		member.Tags.UnsetTag("Yields")
	}
	d.addMember(&member)
}

func (d *decoder) decodeEvent() {
	var member rbxdump.Event
	class := d.expectChars(isClassName, "member class")
	d.expectClass(class)
	d.expectChar('.')
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(false)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeCallback() {
	var member rbxdump.Callback
	member.ReturnType = rbxdump.Type{Name: d.expectChars(isType, "return type")}
	d.expectWhitespace()
	class := d.expectChars(isClassName, "member class")
	d.expectClass(class)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(false)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeParameters(canDefault bool) (params []rbxdump.Parameter) {
	d.expectChar('(')
	d.skipWhitespace()
	if d.checkChar(')') {
		return params
	}
	for {
		params = append(params, d.decodeParameter(canDefault))
		d.skipWhitespace()
		if d.checkChar(')') {
			break
		} else if !d.checkChar(',') {
			d.syntaxError("expected parameter separator")
			return nil
		}
		d.skipWhitespace()
	}
	return params
}

func (d *decoder) decodeParameter(canDefault bool) (param rbxdump.Parameter) {
	param.Type = rbxdump.Type{Name: d.expectChars(isType, "type")}
	d.expectWhitespace()
	param.Name = d.expectChars(isArgName, "argument name")
	if !canDefault {
		return param
	}
	d.skipWhitespace()
	if !d.checkChar('=') {
		return param
	}
	d.skipWhitespace()
	param.Optional = true
	param.Default = d.decodeDefault()
	return param
}

// Decode default argument value. This includes any character that isn't ','
// or ')'. Trailing spaces are also excluded.
func (d *decoder) decodeDefault() string {
	return d.decodeChars(isDefault)
}

func (d *decoder) decodeEnum() {
	enum := rbxdump.Enum{Items: make(map[string]*rbxdump.EnumItem)}
	enum.Name = d.expectChars(isEnumName, "enum name")
	d.skipWhitespace()
	d.decodeTags(&enum.Tags)
	d.addEnum(&enum)
}

func (d *decoder) decodeEnumItem() {
	var item rbxdump.EnumItem
	enum := d.expectChars(isEnumName, "enum name")
	d.expectEnum(enum)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	item.Name = d.expectChars(isEnumItemName, "enum item name")
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	item.Value = d.expectInt()
	d.skipWhitespace()
	d.decodeTags(&item.Tags)
	d.addEnumItem(&item)
}

func (d *decoder) decodeTags(tags *rbxdump.Tags) {
	for d.checkChar('[') {
		d.skipWhitespace()
		tags.SetTag(d.decodeTag())
		d.skipWhitespace()
	}
}

func (d *decoder) decodeTag() string {
	return d.decodeNested('[', ']')
}

// Decode parses an API dump from r.
func Decode(r io.Reader) (root *rbxdump.Root, err error) {
	br, ok := r.(io.ByteReader)
	if !ok {
		br = bufio.NewReader(r)
	}
	d := decoder{
		root: &rbxdump.Root{
			Classes: make(map[string]*rbxdump.Class),
			Enums:   make(map[string]*rbxdump.Enum),
		},
		r:    br,
		next: make([]byte, 0, 9),
		line: 1,
	}
	err = d.decode()
	root = d.root
	return
}
