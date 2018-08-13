package rbxapi

import (
	"io"
	"sync"
)

// Codecs holds a number of Codec values, each referred to by a unique name.
// Codecs is safe for use by multiple goroutines.
type Codecs struct {
	mutex sync.Mutex
	set   map[string]Codec
}

// Codec is the interface that is implemented by types that can decode and
// encode an API structure.
//
// Codecs must not retain state between calls to Decode or Encode, and must be
// safe for use by multiple goroutines.
type Codec interface {
	// Decode reads from r, decoding into an API structure.
	Decode(r io.Reader) (root Root, err error)
	// Encode encodes an API structure, writing to w.
	Encode(w io.Writer, root Root) (err error)
}

// UnknownCodecError is an error indicating that a given codec name has not
// been registered.
type UnknownCodecError interface {
	error
	// UnknownCodec returns the name of the unknown codec.
	UnknownCodec() string
}

// errUnknownCodec implements UnknownCodecError.
type errUnknownCodec string

func (err errUnknownCodec) Error() string {
	return "unknown codec '" + string(err) + "'"
}

func (err errUnknownCodec) UnknownCodec() string {
	return string(err)
}

// Register adds a codec under the given name. Returns false if registration
// fails, which can occur if the codec is nil, or if a codec with the name has
// already been registered.
//
// Codecs that normally allow encoding and decoding to be configured should be
// initialized with a sensible default configuration before being registered.
func (codecs Codecs) Register(name string, codec Codec) bool {
	codecs.mutex.Lock()
	defer codecs.mutex.Unlock()

	if _, ok := codecs.set[name]; ok {
		return false
	}
	if codec == nil {
		return false
	}
	codecs.set[name] = codec
	return true
}

// Decode decodes r into an API structure, using the codec registered under
// the given name. Returns UnknownCodecError if the name has not been
// registered.
func (codecs Codecs) Decode(name string, r io.Reader) (root Root, err error) {
	codecs.mutex.Lock()
	codec, ok := codecs.set[name]
	codecs.mutex.Unlock()
	if !ok {
		err = errUnknownCodec(name)
		return
	}
	return codec.Decode(r)
}

// Encode encodes an API structure into w, using the codec registered under
// the given name. Returns UnknownCodecError if the name has not been
// registered.
func (codecs Codecs) Encode(name string, w io.Writer, root Root) (err error) {
	codecs.mutex.Lock()
	codec, ok := codecs.set[name]
	codecs.mutex.Unlock()
	if !ok {
		err = errUnknownCodec(name)
		return
	}
	return codec.Encode(w, root)
}

var defaultCodecs = Codecs{set: map[string]Codec{}}

// Register adds a codec under the given name. Returns false if registration
// fails, which can occur if the codec is nil, or if a codec with the name has
// already been registered.
func Register(name string, codec Codec) bool {
	return defaultCodecs.Register(name, codec)
}

// Decode decodes r into an API structure, using the codec registered under
// the given name. Returns UnknownCodecError if the name has not been
// registered.
func Decode(name string, r io.Reader) (root Root, err error) {
	return defaultCodecs.Decode(name, r)
}

// Encode encodes an API structure into w, using the codec registered under
// the given name. Returns UnknownCodecError if the name has not been
// registered.
func Encode(name string, w io.Writer, root Root) (err error) {
	return defaultCodecs.Encode(name, w, root)
}
