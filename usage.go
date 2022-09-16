package env

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

// Var contains information about the environment variable parsed from a struct
// field. It is exported as a part of the [Usage] function signature.
type Var struct {
	Name     string       // Name is the full name of the variable, including prefix.
	Type     reflect.Type // Type is the variable's type.
	Desc     string       // Desc is an optional description parsed from the `desc` tag.
	Default  string       // Default is the default value of the variable. If the variable is marked as required, it will be empty.
	Required bool         // Required is true, if the variable is marked as required.
	Expand   bool         // Expand is true, if the variable is marked to be expanded with [os.Expand].

	field reflect.Value // the original struct field.
}

// Usage prints a usage message documenting all defined environment variables.
// It will be called by [Load]/[LoadFrom] if the [WithUsageOnError] option is
// provided and an error occurs while loading environment variables. It is
// exported as a variable, so it can be changed to a custom implementation.
var Usage = func(w io.Writer, vars []Var) {
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
