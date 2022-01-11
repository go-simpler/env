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
	if err := env.LoadFrom(f, &cfg); err != nil {
		t.Fatalf("got %v; want no error", err)
	}
	if cfg.Foo != 1 {
		t.Errorf("got %d; want 1", cfg.Foo)
	}
	if cfg.Bar != 2 {
		t.Errorf("got %d; want 2", cfg.Bar)
	}
	if cfg.Baz != 3 {
		t.Errorf("got %d; want 3", cfg.Baz)
	}
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
	if err := env.LoadFrom(m, &cfg); err != nil {
		t.Fatalf("got %v; want no error", err)
	}
	if cfg.Foo != 1 {
		t.Errorf("got %d; want 1", cfg.Foo)
	}
	if cfg.Bar != 2 {
		t.Errorf("got %d; want 2", cfg.Bar)
	}
	if cfg.Baz != 3 {
		t.Errorf("got %d; want 3", cfg.Baz)
	}
}
