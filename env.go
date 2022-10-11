// Package env provides an API for loading environment variables into structs.
// See the [Load] function documentation for details.
package env

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// ErrInvalidArgument is returned when the argument provided to
// [Load]/[LoadFrom] is invalid.
var ErrInvalidArgument = errors.New("env: argument must be a non-nil struct pointer")

// ErrEmptyTagName is returned when the `env` tag is found but the name of the
// environment variable is empty.
var ErrEmptyTagName = errors.New("env: empty tag name is not allowed")

// ErrUnsupportedType is returned when the provided struct contains a field of
// an unsupported type.
var ErrUnsupportedType = errors.New("env: unsupported type")

// ErrInvalidTagOption is returned when the `env` tag contains an invalid
// option, e.g. `env:"VAR,invalid"`.
var ErrInvalidTagOption = errors.New("env: invalid tag option")

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

// Load loads environment variables into the provided struct using the [OS]
// [Provider] as their source. To specify a custom [Provider], use the
// [LoadFrom] function. dst must be a non-nil struct pointer, otherwise Load
// returns [ErrInvalidArgument].
//
// The struct fields must have the `env:"VAR"` struct tag, where VAR is the name
// of the corresponding environment variable. Unexported fields are ignored. If
// the tag is found but the name of the environment variable is empty, the
// error will be [ErrEmptyTagName].
//
// # Supported types
//
//   - int (any kind)
//   - float (any kind)
//   - bool
//   - string
//   - [time.Duration]
//   - [encoding.TextUnmarshaler]
//   - slices of any type above (space is the default separator for values)
//
// See the [strconv].Parse* functions for parsing rules. Implementing the
// [encoding.TextUnmarshaler] interface is enough to use any user-defined type.
// Nested structs of any depth level are supported, only the leaves of the
// config tree must have the `env` tag. If a field of an unsupported type is
// found, the error will be [ErrUnsupportedType].
//
// # Default values
//
// Default values can be specified either using the `default` struct tag (has a
// higher priority) or by initializing the struct fields directly.
//
// # Per-variable options
//
// The name of the environment variable can be followed by comma-separated
// options in the form of `env:"VAR,option1,option2,..."`:
//
//   - required: marks the environment variable as required
//   - expand: expands the value of the environment variable using [os.Expand]
//
// If environment variables are marked as required but not set, an error of type
// [NotSetError] will be returned. If the tag contains an invalid option, the
// error will be [ErrInvalidTagOption].
//
// # Global options
//
// In addition to the per-variable options, [env] also supports global options
// that apply to all variables:
//
//   - [WithPrefix]: sets prefix for each environment variable
//   - [WithSliceSeparator]: sets custom separator to parse slice values
//   - [WithStrictMode]: enables strict mode: no `default` tag == required
//   - [WithUsageOnError]: enables a usage message printing when an error occurs
//
// See their documentation for details.
func Load(dst any, opts ...Option) error {
	return newLoader(OS, opts...).loadVars(dst)
}

// LoadFrom loads environment variables into the provided struct using the
// specified [Provider] as their source. See [Load] documentation for more
// details.
func LoadFrom(p Provider, dst any, opts ...Option) error {
	return newLoader(p, opts...).loadVars(dst)
}

// Option allows to configure the behaviour of the [Load]/[LoadFrom] functions.
type Option func(*loader)

// WithPrefix configures [Load]/[LoadFrom] to automatically add the provided
// prefix to each environment variable. By default, no prefix is configured.
func WithPrefix(prefix string) Option {
	return func(l *loader) { l.prefix = prefix }
}

// WithSliceSeparator configures [Load]/[LoadFrom] to use the provided separator
// when parsing slice values. The default one is space.
func WithSliceSeparator(sep string) Option {
	return func(l *loader) { l.sliceSep = sep }
}

// WithStrictMode configures [Load]/[LoadFrom] to treat all environment
// variables without the `default` tag as required. By default, strict mode is
// disabled.
func WithStrictMode() Option {
	return func(l *loader) { l.strictMode = true }
}

// WithUsageOnError configures [Load]/[LoadFrom] to write an auto-generated
// usage message to the provided [io.Writer], if an error occurs while loading
// environment variables. The message format can be changed by assigning the
// global [Usage] variable to a custom implementation.
func WithUsageOnError(w io.Writer) Option {
	return func(l *loader) { l.usageOutput = w }
}

// loader is an environment variables loader.
type loader struct {
	provider    Provider
	prefix      string
	sliceSep    string
	strictMode  bool
	usageOutput io.Writer
}

// newLoader creates a new loader with the specified [Provider] and applies the
// provided options, which override the default settings.
func newLoader(p Provider, opts ...Option) *loader {
	l := loader{
		provider:    p,
		prefix:      "",
		sliceSep:    " ",
		strictMode:  false,
		usageOutput: nil,
	}
	for _, opt := range opts {
		opt(&l)
	}
	return &l
}

// loadVars loads environment variables into the provided struct.
func (l *loader) loadVars(dst any) (err error) {
	rv := reflect.ValueOf(dst)
	if !structPtr(rv) {
		return ErrInvalidArgument
	}

	vars, err := l.parseVars(rv.Elem())
	if err != nil {
		return err
	}

	defer func() {
		if err != nil && l.usageOutput != nil {
			Usage(l.usageOutput, vars)
		}
	}()

	// accumulate missing required variables
	// to return NotSetError after the loop is finished.
	var notset []string

	for _, v := range vars {
		value, ok := l.lookupEnv(v.Name, v.Expand)
		if !ok {
			// if the variable is required, mark it as missing and skip the iteration...
			if v.Required {
				notset = append(notset, v.Name)
				continue
			}
			// ...otherwise, use the default value.
			// TODO(junk1tm): actually, there is no need to set a default value
			//                if it has been obtained from the initialized struct field.
			value = v.Default
		}

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
func (l *loader) parseVars(v reflect.Value) ([]Var, error) {
	var vars []Var

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			// skip unexported fields.
			continue
		}

		// special case: a nested struct, parse its fields recursively.
		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			nested, err := l.parseVars(field)
			if err != nil {
				return nil, err
			}
			vars = append(vars, nested...)
			continue
		}

		sf := v.Type().Field(i)
		value, ok := sf.Tag.Lookup("env")
		if !ok {
			// skip fields without the `env` tag.
			continue
		}

		parts := strings.Split(value, ",")
		name, options := parts[0], parts[1:]
		if name == "" {
			return nil, ErrEmptyTagName
		}

		var required, expand bool
		for _, option := range options {
			switch option {
			case "required":
				required = true
			case "expand":
				expand = true
			default:
				return nil, fmt.Errorf("%w %q", ErrInvalidTagOption, option)
			}
		}

		// the value from the `default` tag has a higher priority.
		defValue, defSet := sf.Tag.Lookup("default")
		if !defSet {
			defValue = fmt.Sprintf("%v", field.Interface())
		}

		// strict mode only: no `default` tag means the variable is required.
		if l.strictMode && !defSet {
			required = true
		}

		// the variable is either required or has a default value, but not both.
		if required {
			defValue = ""
		}

		vars = append(vars, Var{
			Name:     l.prefix + name,
			Type:     field.Type(),
			Desc:     sf.Tag.Get("desc"),
			Default:  defValue,
			Required: required,
			Expand:   expand,
			field:    field,
		})
	}

	return vars, nil
}

// lookupEnv retrieves the value of the environment variable named by the key
// using the internal [Provider]. It replaces $VAR or ${VAR} in the result
// using [os.Expand] if expand is true.
func (l *loader) lookupEnv(key string, expand bool) (string, bool) {
	value, ok := l.provider.LookupEnv(key)
	if !ok {
		return "", false
	}

	if !expand {
		return value, true
	}

	mapping := func(key string) string {
		v, _ := l.provider.LookupEnv(key)
		return v
	}

	return os.Expand(value, mapping), true
}
