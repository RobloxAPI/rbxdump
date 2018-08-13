package rbxapijson

import (
	"encoding/json"
	"errors"
	"github.com/robloxapi/rbxapi"
	"io"
)

func Decode(r io.Reader) (root *Root, err error) {
	jd := json.NewDecoder(r)
	root = &Root{}
	err = jd.Decode(root)
	return root, err
}

func Encode(w io.Writer, root *Root) (err error) {
	je := json.NewEncoder(w)
	je.SetIndent("", "\t")
	je.SetEscapeHTML(false)
	return je.Encode(root)
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
// package under the name "json". Returns whether or not the registration was
// successful.
func Register() bool {
	return rbxapi.Register("json", codec{})
}
