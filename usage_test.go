package env_test

import (
	"bytes"
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/dotimport"
)

func TestUsage(t *testing.T) {
	t.Run("default empty string", func(t *testing.T) {
		var cfg struct {
			Foo string `env:"FOO"`
		}
		var buf bytes.Buffer
		env.Usage(&cfg, &buf)
		assert.Equal[E](t, buf.String(), "Usage:\n  FOO  string  default <empty>\n")
	})
}
