// Package env implements loading environment variables into a config struct.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Options are the options for the [Load] function.
type Options struct {
	Source   Source // The source of environment variables. The default is [OS].
	SliceSep string // The separator used to parse slice values. The default is space.
	NameSep  string // The separator used to join nested struct names. The default is underscore.
}

// NotSetError is returned when environment variables are marked as required but not set.
type NotSetError struct {
	Names []string // The names of the missing environment variables.
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
// Default values can be specified using the `default:"VALUE"` struct tag.
//
// Names of environment variables for nested structs are built by joining
// the tags using the value defined `NameSep` option.
//
// The name of an environment variable can be followed by comma-separated options:
//   - required: marks the environment variable as required
//   - expand: expands the value of the environment variable using [os.Expand]
//
// These additional options are not supported for nested structs.
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

	pv := reflect.ValueOf(cfg)
	if !structPtr(pv) {
		panic("env: cfg must be a non-nil struct pointer")
	}

	v := pv.Elem()
	vars := parseVars(v, opts.NameSep)
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

func parseVars(v reflect.Value, nameSep string) []Var {
	var vars []Var

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		// special case: a nested struct, parse its fields recursively.
		if kindOf(field, reflect.Struct) && !implements(field, unmarshalerIface) {
			// first check the `env` tag of the struct field.
			var prefix string
			sf := v.Type().Field(i)
			value, ok := sf.Tag.Lookup("env")
			if ok {
				prefix = value
			}

			for _, v := range parseVars(field, nameSep) {
				v.Name = prefix + nameSep + v.Name
				vars = append(vars, v)
			}
			continue
		}

		sf := v.Type().Field(i)
		value, ok := sf.Tag.Lookup("env")
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
