package env_test

import (
	"bytes"
	"testing"

	"github.com/junk1tm/env"
)

func TestUsage(t *testing.T) {
	const usage = `Usage:
  DB_PORT    int  required      database port
  HTTP_PORT  int  default 8080  http server port
`
	vars := []env.Var{
		{Name: "DB_PORT", Type: "int", Desc: "database port", Required: true},
		{Name: "HTTP_PORT", Type: "int", Desc: "http server port", Default: "8080"},
	}

	var buf bytes.Buffer
	env.Usage(&buf, vars)

	if got := buf.String(); got != usage {
		t.Logf("got:\n%s", got)
		t.Logf("want:\n%s", usage)
		t.Error("usage output mismatch")
	}
}
