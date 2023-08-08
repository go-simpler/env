package env

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

// Var contains the information about the environment variable parsed from a struct field.
type Var struct {
	Name     string       // The full name of the variable, including the prefix.
	Type     reflect.Type // The type of the variable.
	Desc     string       // The description parsed from the `desc` tag (if exists).
	Default  string       // The default value of the variable. Empty, if the variable is required.
	Required bool         // True, if the variable is marked as required.
	Expand   bool         // True, if the variable is marked to be expanded with [os.Expand].

	field reflect.Value // the original struct field.
}

// Usage writes a usage message to the given [io.Writer], documenting all defined environment variables.
func Usage(cfg any, w io.Writer) {
	v := reflect.ValueOf(cfg)
	if !structPtr(v) {
		panic("env: argument must be a non-nil struct pointer")
	}

	vars := newLoader(nil).parseVars(v.Elem())

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
		if v.Desc != "" {
			fmt.Fprintf(tw, "\t%s", v.Desc)
		}
		fmt.Fprintf(tw, "\n")
	}
}
