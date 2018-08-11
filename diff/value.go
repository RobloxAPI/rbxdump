package diff

import (
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/patch"
	"strconv"
	"strings"
)

// Value implements a patch.Value by wrapping around one of the several types
// that can appear in a rbxapi structure. The following types are implemented:
//
//     bool
//     int
//     string
//     rbxapi.Type
//     []string
//     []rbxapi.Parameter
type Value struct {
	Value interface{}
}

// String implements the Value interface.
func (v Value) String() string {
	switch v := v.Value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case rbxapi.Type:
		return v.String()
	case []string:
		return "[" + strings.Join(v, ", ") + "]"
	case []rbxapi.Parameter:
		ss := make([]string, len(v))
		for i, param := range v {
			ss[i] = param.GetType().String() + " " + param.GetName()
			if def, ok := param.GetDefault(); ok {
				ss[i] += " = " + def
			}
		}
		return "(" + strings.Join(ss, ", ") + ")"
	}
	return ""
}

// Equal implements the Value interface.
func (v Value) Equal(w patch.Value) bool {
	switch w := w.(type) {
	case Value:
		switch v := v.Value.(type) {
		case bool:
			w, ok := w.Value.(bool)
			if !ok {
				return false
			}
			return v == w
		case int:
			w, ok := w.Value.(int)
			if !ok {
				return false
			}
			return v == w
		case string:
			w, ok := w.Value.(string)
			if !ok {
				return false
			}
			return v == w
		case rbxapi.Type:
			w, ok := w.Value.(rbxapi.Type)
			if !ok {
				return false
			}
			return v == w
		case []string:
			w, ok := w.Value.([]string)
			if !ok {
				return false
			}
			if len(w) != len(v) {
				return false
			}
			for i, v := range v {
				if w[i] != v {
					return false
				}
			}
			return true
		case []rbxapi.Parameter:
			w, ok := w.Value.([]rbxapi.Parameter)
			if !ok {
				return false
			}
			if len(w) != len(v) {
				return false
			}
			for i, v := range v {
				w := w[i]
				vd, vk := v.GetDefault()
				wd, wk := w.GetDefault()
				switch {
				case v.GetType() != w.GetType(),
					v.GetName() != w.GetName(),
					vk != wk,
					!vk && !wk && vd != wd:
					return false
				}
			}
			return true
		}
	}
	return false
}
