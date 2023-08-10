package env_test

import (
	"errors"
	"fmt"
	"os"

	"go-simpler.org/env"
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

func ExampleLoad_defaultValue() {
	cfg := struct {
		Host string `env:"HOST" default:"localhost"` // either use the `default` tag...
		Port int    `env:"PORT"`
	}{
		Port: 8080, // ...or initialize the struct field directly.
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.Host) // localhost
	fmt.Println(cfg.Port) // 8080
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

//nolint:gocritic //commentedOutCode
func ExampleLoad_required() {
	// os.Setenv("HOST", "localhost")
	// os.Setenv("PORT", "8080")

	var cfg struct {
		Host string `env:"HOST,required"`
		Port int    `env:"PORT,required"`
	}
	if err := env.Load(&cfg); err != nil {
		var notSetErr *env.NotSetError
		if errors.As(err, &notSetErr) {
			fmt.Println(notSetErr.Names) // [HOST PORT]
		}
	}
}

func ExampleLoad_expand() {
	os.Setenv("PORT", "8080")
	os.Setenv("ADDR", "localhost:${PORT}")

	var cfg struct {
		Addr string `env:"ADDR,expand"`
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
	}

	fmt.Println(cfg.Addr) // localhost:8080
}

func ExampleWithSource() {
	m := env.Map{"PORT": "8080"}

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithSource(m)); err != nil {
		// handle error
	}

	fmt.Println(cfg.Port) // 8080
}

func ExampleWithSource_multiple() {
	m := env.Map{"PORT": "8080"}

	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8081") // overrides PORT from m.

	var cfg struct {
		Host string `env:"HOST,required"`
		Port int    `env:"PORT,required"`
	}
	if err := env.Load(&cfg, env.WithSource(m, env.OS)); err != nil {
		// handle error
	}

	fmt.Println(cfg.Host) // localhost
	fmt.Println(cfg.Port) // 8081
}

func ExampleWithPrefix() {
	os.Setenv("APP_PORT", "8080")

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithPrefix("APP_")); err != nil {
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

//nolint:gocritic //commentedOutCode
func ExampleUsage() {
	// os.Setenv("DB_HOST", "localhost")
	// os.Setenv("DB_PORT", "5432")

	var cfg struct {
		DB struct {
			Host string `env:"DB_HOST,required" desc:"database host"`
			Port int    `env:"DB_PORT,required" desc:"database port"`
		}
		HTTPPort int `env:"HTTP_PORT" default:"8080" desc:"http server port"`
	}
	if err := env.Load(&cfg); err != nil {
		// handle error
		env.Usage(&cfg, os.Stdout)
	}

	// Output:
	// Usage:
	//   DB_HOST    string  required      database host
	//   DB_PORT    int     required      database port
	//   HTTP_PORT  int     default 8080  http server port
}
