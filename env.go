// Package env provides an API for loading environment variables into structs.
package env

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Load loads environment variables into the provided struct using the [OS] [Provider] as their source.
// To specify a custom [Provider], use the [LoadFrom] function.
// dst must be a non-nil struct pointer, otherwise Load panics.
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
// In addition to the per-variable options, [env] also supports global options that apply to all variables:
//
//   - [WithPrefix]: sets prefix for each environment variable
//   - [WithSliceSeparator]: sets custom separator to parse slice values
//   - [WithUsageOnError]: enables a usage message printing when an error occurs
//
// See their documentation for details.
func Load(dst any, opts ...Option) error {
	return newLoader(OS, opts...).loadVars(dst)
}

// LoadFrom loads environment variables into the provided struct using the specified [Provider] as their source.
// See [Load] documentation for more details.
func LoadFrom(p Provider, dst any, opts ...Option) error {
	return newLoader(p, opts...).loadVars(dst)
}

// Option allows to configure the behaviour of the [Load]/[LoadFrom] functions.
type Option func(*loader)

// WithPrefix configures [Load]/[LoadFrom] to automatically add the provided prefix to each environment variable.
// By default, no prefix is configured.
func WithPrefix(prefix string) Option {
	return func(l *loader) { l.prefix = prefix }
}

// WithSliceSeparator configures [Load]/[LoadFrom] to use the provided separator when parsing slice values.
// The default one is space.
func WithSliceSeparator(sep string) Option {
	return func(l *loader) { l.sliceSep = sep }
}

// WithUsageOnError configures [Load]/[LoadFrom] to write an auto-generated usage message to the provided [io.Writer],
// if an error occurs while loading environment variables.
// The message format can be changed by assigning the global [Usage] variable to a custom implementation.
func WithUsageOnError(w io.Writer) Option {
	return func(l *loader) { l.usageOutput = w }
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
	provider    Provider
	prefix      string
	sliceSep    string
	usageOutput io.Writer
}

func newLoader(p Provider, opts ...Option) *loader {
	l := loader{
		provider:    p,
		prefix:      "",
		sliceSep:    " ",
		usageOutput: nil,
	}
	for _, opt := range opts {
		opt(&l)
	}
	return &l
}

func (l *loader) loadVars(dst any) (err error) {
	rv := reflect.ValueOf(dst)
	if !structPtr(rv) {
		panic("env: argument must be a non-nil struct pointer")
	}

	vars := l.parseVars(rv.Elem())
	defer func() {
		if err != nil && l.usageOutput != nil {
			Usage(l.usageOutput, vars)
		}
	}()

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
