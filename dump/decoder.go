package dump

import (
	"bufio"
	"bytes"
	"github.com/robloxapi/rbxapi"
	"io"
	"strconv"
)

type SyntaxError struct {
	Msg  string
	Line int
}

func (e *SyntaxError) Error() string {
	return "error on line " + strconv.Itoa(e.Line) + ": " + e.Msg
}

type decoder struct {
	api   *rbxapi.API
	r     io.ByteReader
	next  []byte
	buf   bytes.Buffer
	n     int64
	err   error
	line  int
	class *rbxapi.Class
	enum  *rbxapi.Enum
}

// Creates a SyntaxError with the current line number.
func (d *decoder) syntaxError(msg string) {
	if d.err != nil && d.err != io.EOF {
		return
	}
	d.err = &SyntaxError{Msg: msg, Line: d.line}
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

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\f'
}

func isName(b byte) bool {
	return isWord(b) || b == '<' || b == '>' || b == ' '
}

func isWord(b byte) bool {
	return b == '_' ||
		('0' <= b && b <= '9') ||
		('A' <= b && b <= 'Z') ||
		('a' <= b && b <= 'z')
}

func isInt(b byte) bool {
	return ('0' <= b && b <= '9')
}

func isType(b byte) bool {
	return isWord(b)
}

func isDefault(b byte) bool {
	return b != ',' && b != ')'
}

func isTag(b byte) bool {
	return b != ']'
}

// Decode characters that satisfy the isChar function. If a space satisfies
// isChar, then trailing spaces will be excluded from the result.
func (d *decoder) decodeChars(isChar func(byte) bool, notrail bool) string {
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
		if !isChar(b) {
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
	if notrail {
		// Remove trailing spaces.
		for i := 0; i < width; i++ {
			d.ungetc(b[len(b)-1-i])
		}
		b = b[:len(b)-width]
	}
	return string(b)
}

func (d *decoder) expectChars(isChar func(byte) bool, notrail bool, msg string) (s string) {
	if s = d.decodeChars(isChar, notrail); s == "" {
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
	d.clearParent()
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
	if d.decodeChars(isSpace, false) == "" {
		d.syntaxError("expected whitespace")
	}
}

// Expect zero or more whitespace characters.
func (d *decoder) skipWhitespace() {
	d.decodeChars(isSpace, false)
}

// Expect an integer.
func (d *decoder) expectInt() int {
	i, err := strconv.Atoi(d.decodeChars(isInt, false))
	if err != nil {
		d.syntaxError("expected integer")
	}
	return i
}

// Add a class to the API. Sets class parent.
func (d *decoder) addClass(class *rbxapi.Class) {
	if d.err != nil {
		return
	}
	if _, exists := d.api.Classes[class.Name]; exists {
		d.syntaxError("class '" + class.Name + "' already exists")
		return
	}
	d.api.Classes[class.Name] = class
	d.class = class
}

// Add an enum to the API. Sets enum parent.
func (d *decoder) addEnum(enum *rbxapi.Enum) {
	if d.err != nil {
		return
	}
	if _, exists := d.api.Enums[enum.Name]; exists {
		d.syntaxError("enum '" + enum.Name + "' already exists")
		return
	}
	d.api.Enums[enum.Name] = enum
	d.enum = enum
}

// Add a member to the parent class. Assumes the parent class exists.
func (d *decoder) addMember(member rbxapi.Member) {
	if d.err != nil {
		return
	}
	if _, exists := d.class.Members[member.Name()]; exists {
		d.syntaxError("member '" + member.Name() + "' already exists")
		return
	}
	d.class.Members[member.Name()] = member
}

// Add an  enum item to the parent enum. Assumes the parent enum exists.
func (d *decoder) addEnumItem(item *rbxapi.EnumItem) {
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
	word := d.expectChars(isWord, false, "item type")
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
	class := rbxapi.NewClass("")
	class.Name = d.expectChars(isName, true, "class name")
	d.skipWhitespace()
	if d.checkChar(':') {
		d.skipWhitespace()
		class.Superclass = d.expectChars(isName, true, "superclass name")
		d.skipWhitespace()
	}
	d.decodeTags(class)
	d.addClass(class)
}

func (d *decoder) decodeProperty() {
	member := rbxapi.NewProperty("", "")
	member.ValueType = d.expectChars(isType, false, "value type")
	d.expectWhitespace()
	member.MemberClass = d.expectChars(isName, true, "member class")
	d.expectClass(member.MemberClass)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	member.MemberName = d.expectChars(isName, true, "member name")
	d.skipWhitespace()
	d.decodeTags(member)
	d.addMember(member)
}

func (d *decoder) decodeFunction() {
	member := rbxapi.NewFunction("", "")
	member.ReturnType = d.expectChars(isType, false, "return type")
	d.expectWhitespace()
	member.MemberClass = d.expectChars(isName, true, "member class")
	d.expectClass(member.MemberClass)
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	member.MemberName = d.expectChars(isName, true, "member name")
	d.skipWhitespace()
	member.Arguments = d.decodeArguments(true)
	d.skipWhitespace()
	d.decodeTags(member)
	d.addMember(member)
}

func (d *decoder) decodeYieldFunction() {
	member := rbxapi.NewYieldFunction("", "")
	member.ReturnType = d.expectChars(isType, false, "return type")
	d.expectWhitespace()
	member.MemberClass = d.expectChars(isName, true, "member class")
	d.expectClass(member.MemberClass)
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	member.MemberName = d.expectChars(isName, true, "member name")
	d.skipWhitespace()
	member.Arguments = d.decodeArguments(true)
	d.skipWhitespace()
	d.decodeTags(member)
	d.addMember(member)
}

func (d *decoder) decodeEvent() {
	member := rbxapi.NewEvent("", "")
	member.MemberClass = d.expectChars(isName, true, "member class")
	d.expectClass(member.MemberClass)
	d.expectChar('.')
	member.MemberName = d.expectChars(isName, true, "member name")
	d.skipWhitespace()
	member.Arguments = d.decodeArguments(false)
	d.skipWhitespace()
	d.decodeTags(member)
	d.addMember(member)
}

func (d *decoder) decodeCallback() {
	member := rbxapi.NewCallback("", "")
	member.ReturnType = d.expectChars(isType, false, "return type")
	d.expectWhitespace()
	member.MemberClass = d.expectChars(isName, true, "member class")
	d.expectClass(member.MemberClass)
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	member.MemberName = d.expectChars(isName, true, "member name")
	d.skipWhitespace()
	member.Arguments = d.decodeArguments(false)
	d.skipWhitespace()
	d.decodeTags(member)
	d.addMember(member)
}

func (d *decoder) decodeArguments(canDefault bool) (args []rbxapi.Argument) {
	d.expectChar('(')
	d.skipWhitespace()
	if d.checkChar(')') {
		return args
	}
	for {
		args = append(args, d.decodeArgument(canDefault))
		d.skipWhitespace()
		if d.checkChar(')') {
			break
		} else if !d.checkChar(',') {
			d.syntaxError("expected argument separator")
			return nil
		}
		d.skipWhitespace()
	}
	return args
}

func (d *decoder) decodeArgument(canDefault bool) (arg rbxapi.Argument) {
	arg.Type = d.expectChars(isType, false, "type")
	d.expectWhitespace()
	arg.Name = d.expectChars(isName, true, "argument name")
	if !canDefault {
		return arg
	}
	d.skipWhitespace()
	if !d.checkChar('=') {
		return arg
	}
	d.skipWhitespace()
	s := d.decodeDefault()
	arg.Default = &s
	return arg
}

// Decode default argument value. This includes any character that isn't ','
// or ')'. Trailing spaces are also excluded.
func (d *decoder) decodeDefault() string {
	return d.decodeChars(isDefault, true)
}

func (d *decoder) decodeEnum() {
	enum := rbxapi.NewEnum("")
	enum.Name = d.expectChars(isName, true, "enum name")
	d.skipWhitespace()
	d.decodeTags(enum)
	d.addEnum(enum)
}

func (d *decoder) decodeEnumItem() {
	item := rbxapi.NewEnumItem("", "")
	item.Enum = d.expectChars(isName, true, "enum name")
	d.skipWhitespace()
	d.expectChar('.')
	d.skipWhitespace()
	item.Name = d.expectChars(isName, true, "enum item name")
	d.skipWhitespace()
	d.expectChar(':')
	d.skipWhitespace()
	item.Value = d.expectInt()
	d.skipWhitespace()
	d.decodeTags(item)
	d.addEnumItem(item)
}

func (d *decoder) decodeTags(t rbxapi.Taggable) {
	for d.checkChar('[') {
		d.skipWhitespace()
		t.SetTag(d.decodeTag())
		d.skipWhitespace()
	}
}

func (d *decoder) decodeTag() string {
	s := d.decodeChars(isTag, true)
	d.expectChar(']')
	return s
}

func Decode(r io.Reader) (api *rbxapi.API, err error) {
	br, ok := r.(io.ByteReader)
	if !ok {
		br = bufio.NewReader(r)
	}
	d := decoder{
		api:  rbxapi.NewAPI(),
		r:    br,
		next: make([]byte, 0, 9),
		line: 1,
	}
	err = d.decode()
	api = d.api
	return
}
