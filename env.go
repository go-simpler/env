// Package env implements loading environment variables into a config struct.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Options are the options for the [Load] and [Usage] functions.
type Options struct {
	Source   Source // The source of environment variables. The default is [OS].
	SliceSep string // The separator used to parse slice values. The default is space.
	NameSep  string // The separator used to concatenate environment variable names from nested struct tags. The default is an empty string.
}

// NotSetError is returned when required environment variables are not set.
type NotSetError struct {
	Names []string
}

// Error implements the error interface.
func (e *NotSetError) Error() string {
	if len(e.Names) == 1 {
		return fmt.Sprintf("env: %s is required but not set", e.Names[0])
	}
	return fmt.Sprintf("env: %s are required but not set", strings.Join(e.Names, " "))
}

// Load loads environment variables into the given struct.
// cfg must be a non-nil struct pointer, otherwise Load panics.
// If opts is nil, the default [Options] are used.
//
// The struct fields must have the `env:"VAR"` struct tag,
// where VAR is the name of the corresponding environment variable.
// Unexported fields are ignored.
//
// The following types are supported:
//   - int (any kind)
//   - float (any kind)
//   - bool
//   - string
//   - [time.Duration]
//   - [encoding.TextUnmarshaler]
//   - slices of any type above
//   - nested structs of any depth
//
// See the [strconv].Parse* functions for the parsing rules.
// User-defined types can be used by implementing the [encoding.TextUnmarshaler] interface.
//
// Nested struct of any depth level are supported,
// allowing grouping of related environment variables.
// If a nested struct has the optional `env:"PREFIX"` tag,
// the environment variables declared by its fields are prefixed with PREFIX.
//
// Default values can be specified using the `default:"VALUE"` struct tag.
//
// The name of an environment variable can be followed by comma-separated options:
//   - required: marks the environment variable as required
//   - expand: expands the value of the environment variable using [os.Expand]
func Load(cfg any, opts *Options) error {
	pv := reflect.ValueOf(cfg)
	if !structPtr(pv) {
		panic("env: cfg must be a non-nil struct pointer")
	}

	opts = setDefaultOptions(opts)

	v := pv.Elem()
	vars := parseVars(v, opts)
	cache[v.Type()] = vars

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

func setDefaultOptions(opts *Options) *Options {
	if opts == nil {
		opts = new(Options)
	}
	if opts.Source == nil {
		opts.Source = OS
	}
	if opts.SliceSep == "" {
		opts.SliceSep = " "
	}
	return opts
}

func parseVars(v reflect.Value, opts *Options) []Var {
	var vars []Var

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		tags := v.Type().Field(i).Tag

		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			var prefix string
			if value, ok := tags.Lookup("env"); ok {
				prefix = value + opts.NameSep
			}
			for _, v := range parseVars(field, opts) {
				v.Name = prefix + v.Name
				vars = append(vars, v)
			}
			continue
		}

		value, ok := tags.Lookup("env")
		if !ok {
			continue
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

		defValue, defSet := tags.Lookup("default")
		switch {
		case defSet && required:
			panic("env: `required` and `default` can't be used simultaneously")
		case !defSet && !required:
			defValue = fmt.Sprintf("%v", field.Interface())
		}

		vars = append(vars, Var{
			Name:          name,
			Type:          field.Type(),
			Usage:         tags.Get("usage"),
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
