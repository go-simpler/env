package env_test

import (
	"bytes"
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/dotimport"
)

func TestUsage(t *testing.T) {
	const usage = `Usage:
  DB_HOST    string  default <empty>  database host
  DB_PORT    int     required         database port
  HTTP_PORT  int     default 8080     http server port
`
	cfg := struct {
		DB struct {
			Host string `env:"DB_HOST" desc:"database host"`
			Port int    `env:"DB_PORT,required" desc:"database port"`
		}
		HTTPPort int `env:"HTTP_PORT" desc:"http server port"`
	}{
		HTTPPort: 8080,
	}

	var buf bytes.Buffer
	env.Usage(&cfg, &buf)
	assert.Equal[E](t, buf.String(), usage)
}
