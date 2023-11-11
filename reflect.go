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

func typeOf(v reflect.Value, types ...reflect.Type) bool {
	for _, t := range types {
		if t == v.Type() {
			return true
		}
	}
	return false
}

func kindOf(v reflect.Value, kinds ...reflect.Kind) bool {
	for _, k := range kinds {
		if k == v.Kind() {
			return true
		}
	}
	return false
}

func implements(v reflect.Value, ifaces ...reflect.Type) bool {
	for _, iface := range ifaces {
		if t := v.Type(); t.Implements(iface) || reflect.PtrTo(t).Implements(iface) {
			return true
		}
	}
	return false
}

func structPtr(v reflect.Value) bool {
	return v.IsValid() && v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct && !v.IsNil()
}

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
		panic(fmt.Sprintf("env: unsupported type `%s`", v.Type()))
	}
}

func setInt(v reflect.Value, s string) error {
	i, err := strconv.ParseInt(s, 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("parsing int: %w", err)
	}
	v.SetInt(i)
	return nil
}

func setUint(v reflect.Value, s string) error {
	u, err := strconv.ParseUint(s, 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("parsing uint: %w", err)
	}
	v.SetUint(u)
	return nil
}

func setFloat(v reflect.Value, s string) error {
	f, err := strconv.ParseFloat(s, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("parsing float: %w", err)
	}
	v.SetFloat(f)
	return nil
}

func setBool(v reflect.Value, s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("parsing bool: %w", err)
	}
	v.SetBool(b)
	return nil
}

func setString(v reflect.Value, s string) error {
	v.SetString(s)
	return nil
}

func setDuration(v reflect.Value, s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("parsing duration: %w", err)
	}
	v.Set(reflect.ValueOf(d))
	return nil
}

func setUnmarshaler(v reflect.Value, s string) error {
	u := v.Addr().Interface().(encoding.TextUnmarshaler)
	if err := u.UnmarshalText([]byte(s)); err != nil {
		return fmt.Errorf("unmarshaling text: %w", err)
	}
	return nil
}

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
