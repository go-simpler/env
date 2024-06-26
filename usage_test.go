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
		env.Usage(&cfg, &buf, nil)
		assert.Equal[E](t, buf.String(), "  FOO  string  default <empty>\n")
	})

	t.Run("with Options.NameSep", func(t *testing.T) {
		var buf bytes.Buffer
		var cfg struct {
			A struct {
				Foo int `env:"FOO"`
			} `env:"A"`
		}
		env.Usage(&cfg, &buf, &env.Options{NameSep: "_"})
		assert.Equal[E](t, buf.String(), "  A_FOO  int  default 0\n")
	})

	t.Run("custom usage message", func(t *testing.T) {
		var buf bytes.Buffer
		var cfg Config
		env.Usage(&cfg, &buf, nil)
		assert.Equal[E](t, buf.String(), "custom")
	})

	t.Run("vars cache", func(t *testing.T) {
		m := env.Map{"FOO": "1"}

		var cfg struct {
			Foo int `env:"FOO"`
		}
		err := env.Load(&cfg, &env.Options{Source: m})
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Foo, 1)

		var buf bytes.Buffer
		env.Usage(&cfg, &buf, nil)
		assert.Equal[E](t, buf.String(), "  FOO  int  default 0\n")
	})
}

type Config struct{}

func (Config) Usage(_ []env.Var, w io.Writer, _ *env.Options) {
	_, _ = w.Write([]byte("custom"))
}
