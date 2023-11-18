package env

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

// cache maps the struct type to the [Var] slice parsed from it.
// It is primarily needed to fix the following bug:
//
//	var cfg struct {
//		Port int `env:"PORT"`
//	}
//	env.Load(&cfg, nil)        // 1. sets cfg.Port to 8080
//	env.Usage(&cfg, os.Stdout) // 2. prints cfg.Port's default == 8080 (instead of 0)
//
// It also speeds up [Usage], since there is no need to parse the struct again.
var cache = make(map[reflect.Type][]Var)

// Var holds the information about an environment variable parsed from the struct field.
type Var struct {
	Name     string       // The name of the variable.
	Type     reflect.Type // The type of the variable.
	Usage    string       // The usage string parsed from the `usage` tag (if exists).
	Default  string       // The default value of the variable. Empty, if the variable is required.
	Required bool         // True, if the variable is marked as required.
	Expand   bool         // True, if the variable is marked to be expanded with [os.Expand].

	structField   reflect.Value
	hasDefaultTag bool
}

// Usage writes a usage message documenting all defined environment variables to the given [io.Writer].
// An optional usage string can be added for each environment variable via the `usage:"STRING"` struct tag.
// The format of the message can be customized by implementing the Usage([]env.Var, io.Writer) method on the cfg's type.
func Usage(cfg any, w io.Writer) {
	pv := reflect.ValueOf(cfg)
	if !structPtr(pv) {
		panic("env: cfg must be a non-nil struct pointer")
	}

	v := pv.Elem()
	vars, ok := cache[v.Type()]
	if !ok {
		vars = parseVars(v)
	}

	if u, ok := cfg.(interface{ Usage([]Var, io.Writer) }); ok {
		u.Usage(vars, w)
	} else {
		defaultUsage(vars, w)
	}
}

func defaultUsage(vars []Var, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	for _, v := range vars {
		fmt.Fprintf(tw, "\t%s\t%s", v.Name, v.Type)
		if v.Required {
			fmt.Fprintf(tw, "\trequired")
		} else {
			if v.Type.Kind() == reflect.String && v.Default == "" {
				v.Default = "<empty>"
			}
			fmt.Fprintf(tw, "\tdefault %s", v.Default)
		}
		if v.Usage != "" {
			fmt.Fprintf(tw, "\t%s", v.Usage)
		}
		fmt.Fprintf(tw, "\n")
	}
}
