package env

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	durationType     = reflect.TypeOf(new(time.Duration)).Elem()
	unmarshalerIface = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()
)

// typeOf reports whether v's type is one of the provided types.
func typeOf(v reflect.Value, types ...reflect.Type) bool {
	for _, t := range types {
		if t == v.Type() {
			return true
		}
	}
	return false
}

// kindOf reports whether v's kind is one of the provided kinds.
func kindOf(v reflect.Value, kinds ...reflect.Kind) bool {
	for _, k := range kinds {
		if k == v.Kind() {
			return true
		}
	}
	return false
}

// implements reports whether v's type implements one of the provided
// interfaces.
func implements(v reflect.Value, ifaces ...reflect.Type) bool {
	for _, iface := range ifaces {
		if t := v.Type(); t.Implements(iface) || reflect.PtrTo(v.Type()).Implements(iface) {
			return true
		}
	}
	return false
}

// structPtr reports whether v is a non-nil struct pointer.
func structPtr(v reflect.Value) bool {
	return v.IsValid() && v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct && !v.IsNil()
}

// setValue parses s based on v's type/kind and sets v's underlying field to the
// result.
func setValue(v reflect.Value, s string) error {
	switch {
	case typeOf(v, durationType):
		return setDuration(v, s)
	case implements(v, unmarshalerIface):
		return setUnmarshaler(v, s)
	case kindOf(v, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64):
		return setInt(v, s)
	case kindOf(v, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64):
		return setUint(v, s)
	case kindOf(v, reflect.Float32, reflect.Float64):
		return setFloat(v, s)
	case kindOf(v, reflect.Bool):
		return setBool(v, s)
	case kindOf(v, reflect.String):
		return setString(v, s)
	default:
		return fmt.Errorf("%w %q", ErrUnsupportedType, v.Type())
	}
}

// setInt parses an int field from s and sets v's underlying field to it.
func setInt(v reflect.Value, s string) error {
	bits := v.Type().Bits()
	i, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		return fmt.Errorf("parsing int: %w", err)
	}
	v.SetInt(i)
	return nil
}

// setUint parses an uint field from s and sets v's underlying field to it.
func setUint(v reflect.Value, s string) error {
	bits := v.Type().Bits()
	u, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		return fmt.Errorf("parsing uint: %w", err)
	}
	v.SetUint(u)
	return nil
}

// setFloat parses a float field from s and sets v's underlying field to it.
func setFloat(v reflect.Value, s string) error {
	bits := v.Type().Bits()
	f, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return fmt.Errorf("parsing float: %w", err)
	}
	v.SetFloat(f)
	return nil
}

// setBool parses a bool field from s and sets v's underlying field to it.
func setBool(v reflect.Value, s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("parsing bool: %w", err)
	}
	v.SetBool(b)
	return nil
}

// setString sets v's underlying field to s.
func setString(v reflect.Value, s string) error {
	v.SetString(s)
	return nil
}

// setDuration parses a duration field from s and sets v's underlying field to
// it.
func setDuration(v reflect.Value, s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("parsing duration: %w", err)
	}
	v.Set(reflect.ValueOf(d))
	return nil
}

// setUnmarshaler calls v's UnmarshalText method with s as the text argument.
func setUnmarshaler(v reflect.Value, s string) error {
	u := v.Addr().Interface().(encoding.TextUnmarshaler)
	if err := u.UnmarshalText([]byte(s)); err != nil {
		return fmt.Errorf("unmarshaling text: %w", err)
	}
	return nil
}

// setSlice creates a new slice of the values parsed from s and sets v's
// underlying field to it.
func setSlice(v reflect.Value, s []string) error {
	slice := reflect.MakeSlice(v.Type(), len(s), cap(s))
	for i := 0; i < slice.Len(); i++ {
		if err := setValue(slice.Index(i), s[i]); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}
