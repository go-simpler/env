package env_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/junk1tm/env"
)

func ExampleLoad() {
	os.Setenv("PORT", "8080")

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.Port) // 8080
}

func ExampleLoad_required() {
	// os.Setenv("PORT", "8080")

	var cfg struct {
		Port int `env:"PORT,required"`
	}
	if err := env.Load(&cfg); err != nil {
		var notSetErr *env.NotSetError
		if errors.As(err, &notSetErr) {
			fmt.Println(notSetErr.Names) // [PORT]
		}
	}
}

func ExampleLoad_defaultValue() {
	os.Setenv("PORT", "8081")

	cfg := struct {
		Port int `env:"PORT"`
	}{
		Port: 8080, // default value, will be overridden by PORT.
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.Port) // 8081
}

func ExampleLoad_nestedStruct() {
	os.Setenv("HTTP_PORT", "8080")

	var cfg struct {
		HTTP struct {
			Port int `env:"HTTP_PORT"`
		}
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.HTTP.Port) // 8080
}

func ExampleLoadFrom() {
	m := env.Map{"PORT": "8080"}

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.LoadFrom(m, &cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.Port) // 8080
}

func ExampleWithPrefix() {
	os.Setenv("APP_PORT", "8080")

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithPrefix("APP")); err != nil {
		// handle error
	}

	fmt.Println(cfg.Port) // 8080
}

func ExampleWithSliceSeparator() {
	os.Setenv("PORTS", "8080;8081;8082")

	var cfg struct {
		Ports []int `env:"PORTS"`
	}
	if err := env.Load(&cfg, env.WithSliceSeparator(";")); err != nil {
		// handle error
	}

	fmt.Println(cfg.Ports[0]) // 8080
	fmt.Println(cfg.Ports[1]) // 8081
	fmt.Println(cfg.Ports[2]) // 8082
}
