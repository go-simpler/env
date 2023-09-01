package env_test

import (
	"bytes"
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/EF"
)

func TestUsage(t *testing.T) {
	t.Run("empty string as default", func(t *testing.T) {
		var cfg struct {
			Foo string `env:"FOO" default:""`
		}
		vars, err := env.Load(&cfg, env.WithSource(env.Map{}))
		assert.NoErr[F](t, err)

		var buf bytes.Buffer
		env.Usage(vars, &buf)
		assert.Equal[E](t, buf.String(), "Usage:\n  FOO  string  default <empty>\n")
	})
}
