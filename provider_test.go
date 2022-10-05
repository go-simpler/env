package env_test

import (
	"os"
	"testing"

	"github.com/junk1tm/env"
	"github.com/junk1tm/env/assert"
	. "github.com/junk1tm/env/assert/dotimport"
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

func Test_MultiProviders(t *testing.T) {
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

	m := env.Map{
		"FOO": "10",
		"BAR": "20",
		"BAZ": "30",
	}

	os.Setenv("FOO", "50")
	os.Setenv("LOREM", "100")

	defer func() {
		os.Unsetenv("FOO")
		os.Unsetenv("LOREM")
	}()

	o := env.OS

	var cfg struct {
		Foo   int `env:"FOO,required"`
		Bar   int `env:"BAR,required"`
		Baz   int `env:"BAZ,required"`
		LOREM int `env:"LOREM,required"`
	}

	// multi provider of 3 types
	provider := env.MultiProvider(f, m, o)

	err := env.LoadFrom(provider, &cfg)
	assert.NoErr[F](t, err)

	assert.Equal[E](t, cfg.Foo, 50)
	assert.Equal[E](t, cfg.Bar, 20)
	assert.Equal[E](t, cfg.Baz, 30)
	assert.Equal[E](t, cfg.LOREM, 100)
}
