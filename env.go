// Package env provides an API for loading environment variables into structs.
package env

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrInvalidArgument is returned when the argument provided to Load/LoadFrom is
// invalid.
var ErrInvalidArgument = errors.New("env: argument must be a non-nil struct pointer")

// NotSetError is returned when environment variables are marked as required but
// not set.
type NotSetError struct {
	// Names is a slice of the names of the missing required environment
	// variables.
	Names []string
}

// Error implements the error interface.
func (e *NotSetError) Error() string {
	return fmt.Sprintf("env: %v are required but not set", e.Names)
}

// UnsupportedTypeError is returned when the provided struct contains a field of
// an unsupported type.
type UnsupportedTypeError struct {
	// Type is the found unsupported type.
	Type reflect.Type
}

// Error implements the error interface.
func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("env: unsupported type %s", e.Type)
}

// Load loads environment variables into the provided struct using the OS
// Provider as their source. To specify a custom Provider, use LoadFrom
// function. dst must be a non-nil struct pointer, otherwise Load returns
// ErrInvalidArgument.
//
// The struct fields must have the `env:"VAR"` struct tag, where VAR is the name
// of the corresponding environment variable. Any unexported fields, fields
// without this tag (except nested structs) or fields with empty name are
// ignored. If a field has the tag in the form of `env:"VAR,required"`, it will
// be marked as required and an error of type NotSetError will be returned in
// case no such environment variable is found. Default values can be specified
// using basic struct initialization. They will be left untouched, if no
// corresponding environment variables are found.
//
// The following types are supported as struct fields:
//  int (any kind)
//  float (any kind)
//  bool
//  string
//  time.Duration
//  encoding.TextUnmarshaler
//  slices of any type above (space is the default separator for values)
// See the strconv package from the standard library for parsing rules.
// Implementing the encoding.TextUnmarshaler interface is enough to use any
// user-defined type. Nested structs of any depth level are supported, but only
// non-struct fields are considered as targets for parsing. If a field of an
// unsupported type is found, the error will be of type UnsupportedTypeError.
//
// Load's behavior can be customized using various options:
//  WithPrefix
//  WithSliceSeparator
// See their documentation for details.
func Load(dst interface{}, opts ...Option) error {
	return newLoader(OS, opts...).loadVars(dst)
}

// LoadFrom loads environment variables into the provided struct using the
// specified Provider as their source. See Load documentation for more details.
func LoadFrom(p Provider, dst interface{}, opts ...Option) error {
	return newLoader(p, opts...).loadVars(dst)
}

// Option allows to customize the behaviour of Load/LoadFrom functions.
type Option func(*loader)

// WithPrefix configures Load/LoadFrom to automatically add the provided prefix
// to each environment variable. By default, no prefix is configured.
func WithPrefix(prefix string) Option {
	return func(l *loader) { l.prefix = prefix }
}

// WithSliceSeparator configures Load/LoadFrom to use the provided separator
// when parsing slice values. The default one is space.
func WithSliceSeparator(sep string) Option {
	return func(l *loader) { l.sliceSep = sep }
}

// loader is an environment variables loader.
type loader struct {
	provider Provider
	prefix   string
	sliceSep string
}

// newLoader creates a new loader with the specified Provider and applies the
// provided options, which override the default settings.
func newLoader(p Provider, opts ...Option) *loader {
	l := loader{
		provider: p,
		prefix:   "",
		sliceSep: " ",
	}
	for _, opt := range opts {
		opt(&l)
	}
	return &l
}

// loadVars loads environment variables into the provided struct.
func (l *loader) loadVars(dst interface{}) error {
	rv := reflect.ValueOf(dst)
	if !structPtr(rv) {
		return ErrInvalidArgument
	}

	// accumulate missing required variables
	// to return NotSetError after the iteration is finished.
	var notset []string

	for _, v := range l.parseVars(rv.Elem()) {
		value, ok := l.provider.LookupEnv(v.name)
		if !ok {
			if v.required {
				notset = append(notset, v.name)
			}
			continue
		}

		var err error
		if kindOf(v.field, reflect.Slice) && !implements(v.field, unmarshalerIface) {
			err = setSlice(v.field, strings.Split(value, l.sliceSep))
		} else {
			err = setValue(v.field, value)
		}
		if err != nil {
			return err
		}
	}

	if len(notset) > 0 {
		return &NotSetError{Names: notset}
	}

	return nil
}

// parseVars parses environment variables from the fields of the provided
// struct.
func (l *loader) parseVars(v reflect.Value) []variable {
	var vars []variable

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			// skip unexported fields.
			continue
		}

		// special case: a nested struct, parse its fields recursively.
		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			vars = append(vars, l.parseVars(field)...)
			continue
		}

		sf := v.Type().Field(i)
		value, ok := sf.Tag.Lookup("env")
		if !ok {
			// skip fields without the `env` tag.
			continue
		}

		parts := strings.Split(value, ",")
		name := parts[0]
		if name == "" {
			// skip fields with empty name.
			// TODO(junk1tm): return an error instead?
			continue
		}

		// a variable named VAR is required when
		// the `env:"VAR,required"` tag is specified.
		required := len(parts) > 1 && parts[1] == "required"

		vars = append(vars, variable{
			name:     l.prefix + name,
			required: required,
			field:    field,
		})
	}

	return vars
}

// variable contains information about an environment variable parsed from a
// struct field.
type variable struct {
	name     string
	required bool
	field    reflect.Value // the original struct field.
}
