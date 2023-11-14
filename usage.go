package env

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

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
	v := reflect.ValueOf(cfg)
	if !structPtr(v) {
		panic("env: cfg must be a non-nil struct pointer")
	}

	vars := parseVars(v.Elem())

	if u, ok := cfg.(interface{ Usage([]Var, io.Writer) }); ok {
		u.Usage(vars, w)
	} else {
		defaultUsage(vars, w)
	}
}

func defaultUsage(vars []Var, w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "Usage:\n")
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
