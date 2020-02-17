package legacy

type charCheck struct {
	// Whether leading/trailing spaces are valid.
	nofix bool
	// Whether a given byte is valid.
	isChar func(b byte) bool
}

var (
	isSpace = charCheck{nofix: false, isChar: func(b byte) bool {
		return b == ' ' || b == '\t' || b == '\f'
	}}
	isWord = charCheck{nofix: false, isChar: func(b byte) bool {
		return b == '_' ||
			('0' <= b && b <= '9') ||
			('A' <= b && b <= 'Z') ||
			('a' <= b && b <= 'z')
	}}
	isInt = charCheck{nofix: false, isChar: func(b byte) bool {
		return ('0' <= b && b <= '9') || b == '-'
	}}
	isName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isWord.isChar(b) || b == '<' || b == '>' || b == ' '
	}}
)

var (
	isClassName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isName.isChar(b)
	}}
	isMemberName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isName.isChar(b)
	}}
	isType = charCheck{nofix: false, isChar: func(b byte) bool {
		return isWord.isChar(b) || b == ':'
	}}
	isEnumName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isName.isChar(b)
	}}
	isEnumItemName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isName.isChar(b)
	}}
	isArgName = charCheck{nofix: true, isChar: func(b byte) bool {
		return isWord.isChar(b)
	}}
	isDefault = charCheck{nofix: true, isChar: func(b byte) bool {
		return b != ',' && b != ')'
	}}
)
