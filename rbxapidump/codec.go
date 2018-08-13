package rbxapidump

import (
	"bufio"
	"errors"
	"github.com/robloxapi/rbxapi"
	"io"
)

// Decode parses an API dump from r. The resulting API structure is a *Root.
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

// Encode encodes root, writing the results to w in the API dump format.
func Encode(w io.Writer, root *Root) (err error) {
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

type codec struct{}

func (codec) Decode(r io.Reader) (root rbxapi.Root, err error) {
	return Decode(r)
}

func (codec) Encode(w io.Writer, root rbxapi.Root) (err error) {
	switch root := root.(type) {
	case *Root:
		err = Encode(w, root)
	default:
		err = errors.New("cannot encode root: unknown implementation")
	}
	return
}

// Register function automatically registers Decode and Encode with the rbxapi
// package under the name "dump". Returns whether or not the registration was
// successful.
func Register() bool {
	return rbxapi.Register("dump", codec{})
}
