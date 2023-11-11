// Package env provides an API for loading environment variables into structs.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Options are options for the [Load] function.
type Options struct {
	Source   Source // The source of environment variables. The default is [OS].
	SliceSep string // The separator used to parse slice values. The default is space.
}

// NotSetError is returned when environment variables are marked as required but not set.
type NotSetError struct {
	// The names of the missing required environment variables.
	Names []string
}

// Error implements the error interface.
func (e *NotSetError) Error() string {
	return fmt.Sprintf("env: %v are required but not set", e.Names)
}

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
func Load(cfg any, opts *Options) error {
	if opts == nil {
		opts = new(Options)
	}
	if opts.Source == nil {
		opts.Source = OS
	}
	if opts.SliceSep == "" {
		opts.SliceSep = " "
	}

	v := reflect.ValueOf(cfg)
	if !structPtr(v) {
		panic("env: cfg must be a non-nil struct pointer")
	}

	vars := parseVars(v.Elem())

	var notset []string
	for _, v := range vars {
		value, ok := lookupEnv(opts.Source, v.Name, v.Expand)
		if !ok {
			if v.Required {
				notset = append(notset, v.Name)
				continue
			}
			if !v.hasDefaultTag {
				continue // nothing to set.
			}
			value = v.Default
		}

		var err error
		if kindOf(v.structField, reflect.Slice) && !implements(v.structField, unmarshalerIface) {
			err = setSlice(v.structField, strings.Split(value, opts.SliceSep))
		} else {
			err = setValue(v.structField, value)
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

func parseVars(v reflect.Value) []Var {
	var vars []Var

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue // skip unexported fields.
		}

		// special case: a nested struct, parse its fields recursively.
		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			nested := parseVars(field)
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

		defValue, defSet := sf.Tag.Lookup("default")
		switch {
		case defSet && required:
			panic("env: `required` and `default` can't be used simultaneously")
		case !defSet && !required:
			defValue = fmt.Sprintf("%v", field.Interface())
		}

		vars = append(vars, Var{
			Name:          name,
			Type:          field.Type(),
			Usage:         sf.Tag.Get("usage"),
			Default:       defValue,
			Required:      required,
			Expand:        expand,
			structField:   field,
			hasDefaultTag: defSet,
		})
	}

	return vars
}

func lookupEnv(src Source, key string, expand bool) (string, bool) {
	value, ok := src.LookupEnv(key)
	if !ok {
		return "", false
	}
	if !expand {
		return value, true
	}
	mapping := func(key string) string {
		v, _ := src.LookupEnv(key)
		return v
	}
	return os.Expand(value, mapping), true
}
