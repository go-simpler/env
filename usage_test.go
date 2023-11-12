package env_test

import (
	"bytes"
	"io"
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/EF"
)

func TestUsage(t *testing.T) {
	t.Run("empty string as default", func(t *testing.T) {
		var buf bytes.Buffer
		var cfg struct {
			Foo string `env:"FOO" default:""`
		}
		env.Usage(&cfg, &buf)
		assert.Equal[E](t, buf.String(), "Usage:\n  FOO  string  default <empty>\n")
	})

	t.Run("custom usage message", func(t *testing.T) {
		var buf bytes.Buffer
		var cfg config
		env.Usage(&cfg, &buf)
		assert.Equal[E](t, buf.String(), "custom")
	})
}

type config struct{}

func (config) Usage(_ []env.Var, w io.Writer) {
	_, _ = w.Write([]byte("custom"))
}
