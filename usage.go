package env

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"
)

// Usage writes a usage message to the given [io.Writer], documenting all defined environment variables.
func Usage(vars []Var, w io.Writer) {
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
