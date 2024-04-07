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
	if err := env.Load(&cfg, nil); err != nil {
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
	if err := env.Load(&cfg, nil); err != nil {
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
	if err := env.Load(&cfg, nil); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.HTTP.Port)
	// Output: 8080
}

func ExampleLoad_nestedStructPrefixed() {
	os.Setenv("DBHOST", "localhost")
	os.Setenv("DBPORT", "5432")

	var cfg struct {
		DB struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		} `env:"DB"`
	}
	if err := env.Load(&cfg, nil); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.DB.Host)
	fmt.Println(cfg.DB.Port)
	// Output:
	// localhost
	// 5432
}

func ExampleLoad_nestedStructPrefixedWithSeparator() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")

	var cfg struct {
		DB struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		} `env:"DB"`
	}
	if err := env.Load(&cfg, &env.Options{NameSep: "_"}); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.DB.Host)
	fmt.Println(cfg.DB.Port)
	// Output:
	// localhost
	// 5432
}

func ExampleLoad_required() {
	os.Unsetenv("PORT")

	var cfg struct {
		Port int `env:"PORT,required"`
	}
	if err := env.Load(&cfg, nil); err != nil {
		var notSetErr *env.NotSetError
		if errors.As(err, &notSetErr) {
			fmt.Println(notSetErr)
		}
	}

	// Output: env: PORT is required but not set
}

func ExampleLoad_expand() {
	os.Setenv("PORT", "8080")
	os.Setenv("ADDR", "localhost:${PORT}")

	var cfg struct {
		Addr string `env:"ADDR,expand"`
	}
	if err := env.Load(&cfg, nil); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Addr)
	// Output: localhost:8080
}

func ExampleLoad_source() {
	m := env.Map{"PORT": "8080"}

	var cfg struct {
		Port int `env:"PORT"`
	}
	if err := env.Load(&cfg, &env.Options{Source: m}); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Port)
	// Output: 8080
}

func ExampleLoad_sliceSeparator() {
	os.Setenv("PORTS", "8080,8081,8082")

	var cfg struct {
		Ports []int `env:"PORTS"`
	}
	if err := env.Load(&cfg, &env.Options{SliceSep: ","}); err != nil {
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
			Host string `env:"DB_HOST,required" usage:"database host"`
			Port int    `env:"DB_PORT,required" usage:"database port"`
		}
		HTTPPort int `env:"HTTP_PORT" default:"8080" usage:"http server port"`
	}
	if err := env.Load(&cfg, nil); err != nil {
		fmt.Println(err)
		fmt.Println("Usage:")
		env.Usage(&cfg, os.Stdout, nil)
	}

	// Output: env: DB_HOST DB_PORT are required but not set
	// Usage:
	//   DB_HOST    string  required      database host
	//   DB_PORT    int     required      database port
	//   HTTP_PORT  int     default 8080  http server port
}
