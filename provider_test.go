package env_test

import (
	"testing"

	"github.com/junk1tm/env"
)

func TestProviderFunc_LookupEnv(t *testing.T) {
	f := env.ProviderFunc(func(key string) (value string, ok bool) {
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
	noerr[F](t, err)
	equal[E](t, cfg.Foo, 1)
	equal[E](t, cfg.Bar, 2)
	equal[E](t, cfg.Baz, 3)
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
	noerr[F](t, err)
	equal[E](t, cfg.Foo, 1)
	equal[E](t, cfg.Bar, 2)
	equal[E](t, cfg.Baz, 3)
}
