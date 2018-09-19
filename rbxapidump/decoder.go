package rbxapidump

import (
	"bufio"
	"bytes"
	"github.com/robloxapi/rbxapi"
	"io"
	"strconv"
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
	root  *Root
	r     io.ByteReader
	next  []byte
	buf   bytes.Buffer
	n     int64
	err   error
	line  int
	class *Class
	enum  *Enum
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
func (d *decoder) addClass(class *Class) {
	if d.err != nil {
		return
	}
	d.root.Classes = append(d.root.Classes, class)
	d.class = class
}

// Add an enum to the API. Sets enum parent.
func (d *decoder) addEnum(enum *Enum) {
	if d.err != nil {
		return
	}
	d.root.Enums = append(d.root.Enums, enum)
	d.enum = enum
}

// Add a member to the parent class. Assumes the parent class exists.
func (d *decoder) addMember(member rbxapi.Member) {
	if d.err != nil {
		return
	}
	d.class.Members = append(d.class.Members, member)
}

// Add an  enum item to the parent enum. Assumes the parent enum exists.
func (d *decoder) addEnumItem(item *EnumItem) {
	if d.err != nil {
		return
	}
	d.enum.Items = append(d.enum.Items, item)
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
	return
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
		d.decodeFunction()
	case "YieldFunction":
		d.decodeYieldFunction()
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
	var class Class
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
	var member Property
	member.ValueType = Type(d.expectChars(isType, "value type"))
	d.expectWhitespace()
	member.Class = d.expectChars(isClassName, "member class")
	d.expectClass(member.Class)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeFunction() {
	var member Function
	member.ReturnType = Type(d.expectChars(isType, "return type"))
	d.expectWhitespace()
	member.Class = d.expectChars(isClassName, "member class")
	d.expectClass(member.Class)
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(true)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeYieldFunction() {
	var member YieldFunction
	member.ReturnType = Type(d.expectChars(isType, "return type"))
	d.expectWhitespace()
	member.Class = d.expectChars(isClassName, "member class")
	d.expectClass(member.Class)
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(true)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeEvent() {
	var member Event
	member.Class = d.expectChars(isClassName, "member class")
	d.expectClass(member.Class)
	d.expectChar('.')
	member.Name = d.expectChars(isMemberName, "member name")
	d.skipWhitespace()
	member.Parameters = d.decodeParameters(false)
	d.skipWhitespace()
	d.decodeTags(&member.Tags)
	d.addMember(&member)
}

func (d *decoder) decodeCallback() {
	var member Callback
	member.ReturnType = Type(d.expectChars(isType, "return type"))
	d.expectWhitespace()
	member.Class = d.expectChars(isClassName, "member class")
	d.expectClass(member.Class)
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

func (d *decoder) decodeParameters(canDefault bool) (params []Parameter) {
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

func (d *decoder) decodeParameter(canDefault bool) (param Parameter) {
	param.Type = Type(d.expectChars(isType, "type"))
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
	s := d.decodeDefault()
	param.Default = &s
	return param
}

// Decode default argument value. This includes any character that isn't ','
// or ')'. Trailing spaces are also excluded.
func (d *decoder) decodeDefault() string {
	return d.decodeChars(isDefault)
}

func (d *decoder) decodeEnum() {
	var enum Enum
	enum.Name = d.expectChars(isEnumName, "enum name")
	d.skipWhitespace()
	d.decodeTags(&enum.Tags)
	d.addEnum(&enum)
}

func (d *decoder) decodeEnumItem() {
	var item EnumItem
	item.Enum = d.expectChars(isEnumName, "enum name")
	d.expectEnum(item.Enum)
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

func (d *decoder) decodeTags(tags *Tags) {
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
func Decode(r io.Reader) (root *Root, err error) {
	br, ok := r.(io.ByteReader)
	if !ok {
		br = bufio.NewReader(r)
	}
	d := decoder{
		root: &Root{},
		r:    br,
		next: make([]byte, 0, 9),
		line: 1,
	}
	err = d.decode()
	root = d.root
	return
}
