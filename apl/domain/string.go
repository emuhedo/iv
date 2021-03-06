package domain

import (
	"github.com/ktye/iv/apl"
)

// IsString accepts strings
func IsString(child SingleDomain) SingleDomain {
	return stringtype{child}
}

type stringtype struct{ child SingleDomain }

func (s stringtype) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	if v, ok := V.(apl.String); ok {
		if s.child == nil {
			return v, true
		}
		return s.child.To(a, v)
	}
	return V, false
}
func (s stringtype) String(a *apl.Apl) string {
	if s.child == nil {
		return "string"
	}
	return "string" + " " + s.child.String(a)
}

// IsStringArray accepts uniform.Strings
func IsStringArray(child SingleDomain) SingleDomain {
	return stringstype{child, false}
}

func ToStringArray(child SingleDomain) SingleDomain {
	return stringstype{child, true}
}

type stringstype struct {
	child   SingleDomain
	convert bool
}

func (s stringstype) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	if _, ok := V.(apl.StringArray); ok {
		return propagate(a, V, s.child)
	} else {
		if s.convert == false {
			return V, false
		}
		if str, ok := V.(apl.String); ok {
			return propagate(a, apl.StringArray{
				Dims:    []int{1},
				Strings: []string{string(str)},
			}, s.child)

		} else if ar, ok := V.(apl.Array); ok {
			str := make([]string, ar.Size())
			for i := range str {
				if sv, ok := ar.At(i).(apl.String); ok {
					str[i] = string(sv)
				} else {
					return V, false
				}
			}
			return propagate(a, apl.StringArray{
				Dims:    apl.CopyShape(ar),
				Strings: str,
			}, s.child)
		} else {
			return V, false
		}
	}
}
func (s stringstype) String(a *apl.Apl) string {
	name := "string array"
	if s.convert {
		name = "to string array"
	}
	if s.child == nil {
		return name
	}
	return name + " " + s.child.String(a)
}
