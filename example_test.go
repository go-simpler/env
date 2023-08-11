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
		fmt.Println(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleLoad_defaultValue() {
	os.Unsetenv("PORT")

	var cfg struct {
		Port int `env:"PORT" default:"8080"`
	}
	if err := env.Load(&cfg); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleLoad_nestedStruct() {
	os.Setenv("HTTP_PORT", "8080")

	var cfg struct {
		HTTP struct {
			Port int `env:"HTTP_PORT"`
		}
	}
	if err := env.Load(&cfg); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.HTTP.Port)
	// Output: 8080
}

func ExampleLoad_required() {
	os.Unsetenv("HOST")
	os.Unsetenv("PORT")

	var cfg struct {
		Host string `env:"HOST,required"`
		Port int    `env:"PORT,required"`
	}
	if err := env.Load(&cfg); err != nil {
		var notSetErr *env.NotSetError
		if errors.As(err, &notSetErr) {
			fmt.Println(notSetErr.Names)
		}
	}

	// Output: [HOST PORT]
}

func ExampleLoad_expand() {
	os.Setenv("PORT", "8080")
	os.Setenv("ADDR", "localhost:${PORT}")

	var cfg struct {
		Addr string `env:"ADDR,expand"`
	}
	if err := env.Load(&cfg); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Addr)
	// Output: localhost:8080
}

func ExampleWithSource() {
	m := env.Map{"PORT": "8080"}

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithSource(m)); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleWithSource_multiple() {
	m := env.Map{"PORT": "8080"}

	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8081") // overrides PORT from m.

	var cfg struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithSource(m, env.OS)); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Host, cfg.Port)
	// Output: localhost 8081
}

func ExampleWithPrefix() {
	os.Setenv("APP_PORT", "8080")

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, env.WithPrefix("APP_")); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleWithSliceSeparator() {
	os.Setenv("PORTS", "8080;8081;8082")

	var cfg struct {
		Ports []int `env:"PORTS"`
	}
	if err := env.Load(&cfg, env.WithSliceSeparator(";")); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Ports)
	// Output: [8080 8081 8082]
}

func ExampleUsage() {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")

	var cfg struct {
		DB struct {
			Host string `env:"DB_HOST,required" desc:"database host"`
			Port int    `env:"DB_PORT,required" desc:"database port"`
		}
		HTTPPort int `env:"HTTP_PORT" default:"8080" desc:"http server port"`
	}
	if err := env.Load(&cfg); err != nil {
		fmt.Println(err)
		env.Usage(&cfg, os.Stdout)
	}

	// Output: env: [DB_HOST DB_PORT] are required but not set
	// Usage:
	//   DB_HOST    string  required      database host
	//   DB_PORT    int     required      database port
	//   HTTP_PORT  int     default 8080  http server port
}
