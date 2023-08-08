package env_test

import (
	"testing"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/dotimport"
)

func TestSourceFunc_LookupEnv(t *testing.T) {
	f := env.SourceFunc(func(key string) (string, bool) {
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

	var cfg struct {
		Foo int `env:"FOO,required"`
		Bar int `env:"BAR,required"`
		Baz int `env:"BAZ,required"`
	}
	err := env.LoadFrom(f, &cfg)
	assert.NoErr[F](t, err)
	assert.Equal[E](t, cfg.Foo, 1)
	assert.Equal[E](t, cfg.Bar, 2)
	assert.Equal[E](t, cfg.Baz, 3)
}

func TestMap_LookupEnv(t *testing.T) {
	m := env.Map{
		"FOO": "1",
		"BAR": "2",
		"BAZ": "3",
	}

	var cfg struct {
		Foo int `env:"FOO,required"`
		Bar int `env:"BAR,required"`
		Baz int `env:"BAZ,required"`
	}
	err := env.LoadFrom(m, &cfg)
	assert.NoErr[F](t, err)
	assert.Equal[E](t, cfg.Foo, 1)
	assert.Equal[E](t, cfg.Bar, 2)
	assert.Equal[E](t, cfg.Baz, 3)
}

func TestMultiSource(t *testing.T) {
	m1 := env.Map{
		"FOO": "1",
		"BAR": "2",
	}
	m2 := env.Map{
		"BAR": "3", // overrides BAR from m1.
		"BAZ": "4",
	}
	p := env.MultiSource(m1, m2)

	var cfg struct {
		Foo int `env:"FOO,required"`
		Bar int `env:"BAR,required"`
		Baz int `env:"BAZ,required"`
	}
	err := env.LoadFrom(p, &cfg)
	assert.NoErr[F](t, err)
	assert.Equal[E](t, cfg.Foo, 1)
	assert.Equal[E](t, cfg.Bar, 3)
	assert.Equal[E](t, cfg.Baz, 4)
}