// Package env provides an API for loading environment variables into structs.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Load loads environment variables into the provided struct using the [OS] [Source].
// cfg must be a non-nil struct pointer, otherwise Load panics.
//
// The struct fields must have the `env:"VAR"` struct tag, where VAR is the name of the corresponding environment variable.
// Unexported fields are ignored.
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
// See the [strconv].Parse* functions for parsing rules.
// Implementing the [encoding.TextUnmarshaler] interface is enough to use any user-defined type.
// Nested structs of any depth level are supported, only the leaves of the config tree must have the `env` tag.
//
// # Default values
//
// Default values can be specified either using the `default` struct tag (has a higher priority) or by initializing the struct fields directly.
//
// # Per-variable options
//
// The name of the environment variable can be followed by comma-separated options in the form of `env:"VAR,option1,option2,..."`:
//
//   - required: marks the environment variable as required
//   - expand: expands the value of the environment variable using [os.Expand]
//
// If environment variables are marked as required but not set, an error of type [NotSetError] will be returned.
//
// # Global options
//
// Load also accepts global options that apply to all environment variables, see the With* functions for details.
func Load(cfg any, opts ...Option) error {
	return newLoader(opts).loadVars(cfg)
}

// Option allows to configure the behaviour of the [Load] function.
type Option func(*loader)

// WithSource configures [Load] to retrieve environment variables from the provided [Source].
// If multiple sources are provided, they will be merged into a single one containing the union of all environment variables.
// The order of the sources matters: if the same key occurs more than once, the later value takes precedence.
// The default source is [OS].
func WithSource(src Source, srcs ...Source) Option {
	// reverse the slice first, since the later value should take precedence.
	for i, j := 0, len(srcs)-1; i < j; i, j = i+1, j-1 {
		srcs[i], srcs[j] = srcs[j], srcs[i]
	}
	return func(l *loader) { l.source = multiSource(append(srcs, src)) }
}

// WithPrefix configures [Load] to automatically add the provided prefix to each environment variable.
// By default, no prefix is configured.
func WithPrefix(prefix string) Option {
	return func(l *loader) { l.prefix = prefix }
}

// WithSliceSeparator configures [Load] to use the provided separator when parsing slice values.
// The default separator is space.
func WithSliceSeparator(sep string) Option {
	return func(l *loader) { l.sliceSep = sep }
}

// NotSetError is returned when environment variables are marked as required but not set.
type NotSetError struct {
	// Names is a slice of the names of the missing required environment variables.
	Names []string
}

// Error implements the error interface.
func (e *NotSetError) Error() string {
	return fmt.Sprintf("env: %v are required but not set", e.Names)
}

type loader struct {
	source   Source
	prefix   string
	sliceSep string
}

func newLoader(opts []Option) *loader {
	l := loader{
		source:   OS,
		prefix:   "",
		sliceSep: " ",
	}
	for _, opt := range opts {
		opt(&l)
	}
	return &l
}

func (l *loader) loadVars(cfg any) error {
	v := reflect.ValueOf(cfg)
	if !structPtr(v) {
		panic("env: argument must be a non-nil struct pointer")
	}

	vars := l.parseVars(v.Elem())

	// accumulate missing required variables to return NotSetError after the loop is finished.
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
			value = v.Default
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

func (l *loader) parseVars(v reflect.Value) []Var {
	var vars []Var

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue // skip unexported fields.
		}

		// special case: a nested struct, parse its fields recursively.
		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			nested := l.parseVars(field)
			vars = append(vars, nested...)
			continue
		}

		sf := v.Type().Field(i)
		value, ok := sf.Tag.Lookup("env")
		if !ok {
			continue // skip fields without the `env` tag.
		}

		parts := strings.Split(value, ",")
		name, options := parts[0], parts[1:]
		if name == "" {
			panic("env: empty tag name is not allowed")
		}

		var required, expand bool
		for _, option := range options {
			switch option {
			case "required":
				required = true
			case "expand":
				expand = true
			default:
				panic(fmt.Sprintf("env: invalid tag option `%s`", option))
			}
		}

		// the value from the `default` tag has a higher priority.
		defValue, ok := sf.Tag.Lookup("default")
		if !ok {
			defValue = fmt.Sprintf("%v", field.Interface())
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

	return vars
}

func (l *loader) lookupEnv(key string, expand bool) (string, bool) {
	value, ok := l.source.LookupEnv(key)
	if !ok {
		return "", false
	}

	if !expand {
		return value, true
	}

	mapping := func(key string) string {
		v, _ := l.source.LookupEnv(key)
		return v
	}

	return os.Expand(value, mapping), true
}
