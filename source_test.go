package env_test

import (
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/EF"
)

func TestSourceFunc_LookupEnv(t *testing.T) {
	fn := env.SourceFunc(func(key string) (string, bool) {
		switch key {
		case "FOO":
			return "1", true
		case "BAR":
			return "2", true
		case "BAZ":
			return "3", true
		default:
			return "", false
		}
	})

	testSource(t, fn)
}

func TestMap_LookupEnv(t *testing.T) {
	m := env.Map{
		"FOO": "1",
		"BAR": "2",
		"BAZ": "3",
	}

	testSource(t, m)
}

func testSource(t *testing.T, src env.Source) {
	var cfg struct {
		Foo int `env:"FOO"`
		Bar int `env:"BAR"`
		Baz int `env:"BAZ"`
	}
	err := env.Load(&cfg, env.WithSource(src))
	assert.NoErr[F](t, err)
	assert.Equal[E](t, cfg.Foo, 1)
	assert.Equal[E](t, cfg.Bar, 2)
	assert.Equal[E](t, cfg.Baz, 3)
}
