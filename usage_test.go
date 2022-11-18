package env_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/go-simpler/env"
)

func TestUsage(t *testing.T) {
	const usage = `Usage:
  DB_HOST    string  default <empty>  database host
  DB_PORT    int     required         database port
  HTTP_PORT  int     default 8080     http server port
`
	vars := []env.Var{
		{Name: "DB_HOST", Type: reflect.TypeOf(""), Desc: "database host", Default: ""},
		{Name: "DB_PORT", Type: reflect.TypeOf(0), Desc: "database port", Required: true},
		{Name: "HTTP_PORT", Type: reflect.TypeOf(0), Desc: "http server port", Default: "8080"},
	}

	var buf bytes.Buffer
	env.Usage(&buf, vars)

	if got := buf.String(); got != usage {
		t.Logf("got:\n%s", got)
		t.Logf("want:\n%s", usage)
		t.Error("usage output mismatch")
	}
}
