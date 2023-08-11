<div align="center">

![logo](logo.svg)

A lightweight package for loading environment variables into structs

[![awesome-go](https://awesome.re/badge.svg)](https://github.com/avelino/awesome-go#configuration)
[![checks](https://github.com/go-simpler/env/actions/workflows/checks.yml/badge.svg)](https://github.com/go-simpler/env/actions/workflows/checks.yml)
[![pkg.go.dev](https://pkg.go.dev/badge/go-simpler.org/env.svg)](https://pkg.go.dev/go-simpler.org/env)
[![goreportcard](https://goreportcard.com/badge/go-simpler.org/env)](https://goreportcard.com/report/go-simpler.org/env)
[![codecov](https://codecov.io/gh/go-simpler/env/branch/main/graph/badge.svg)](https://codecov.io/gh/go-simpler/env)

</div>

## 📌 About

This package is made for apps that [store config in environment variables][1].
Its purpose is to replace multiple fragmented `os.Getenv` calls in `main.go`
with a single struct definition, which simplifies config management and improves
code readability.

## 📦 Install

```shell
go get go-simpler.org/env
```

## 🚀 Features

* Simple API
* Dependency-free
* Per-variable options: [required](#required), [expand](#expand)
* Global options: [source](#source), [prefix](#prefix), [slice separator](#slice-separator)
* Auto-generated [usage message](#usage-message)

## 📋 Usage

`Load` is the main function of this package. It loads environment variables into
the provided struct.

The struct fields must have the `env:"VAR"` struct tag, where `VAR` is the name
of the corresponding environment variable. Unexported fields are ignored.

```go
os.Setenv("PORT", "8080")

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port)
// Output: 8080
```

### Supported types

* `int` (any kind)
* `float` (any kind)
* `bool`
* `string`
* `time.Duration`
* `encoding.TextUnmarshaler`
* slices of any type above

See the `strconv.Parse*` functions for parsing rules.

Nested structs of any depth level are supported, only the leaves of the config
tree must have the `env` tag.

```go
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
```

### Default values

Default values can be specified using the `default` struct tag:

```go
os.Unsetenv("PORT")

var cfg struct {
    Port int `env:"PORT" default:"8080"`
}
if err := env.Load(&cfg); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port)
// Output: 8080
```

### Per-variable options

The name of the environment variable can be followed by comma-separated options
in the form of `env:"VAR,option1,option2,..."`.

#### Required

Use the `required` option to mark the environment variable as required. In case
no such variable is found, an error of type `NotSetError` will be returned.

```go
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
```

#### Expand

Use the `expand` option to automatically expand the value of the environment
variable using `os.Expand`.

```go
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
```

### Global options

`Load` also accepts global options that apply to all environment variables.

#### Source

By default, `Load` retrieves environment variables values directly from OS.
To use a different source, provide an implementation of the `Source` interface via the `WithSource` option.

```go
// Source represents a source of environment variables.
type Source interface {
    // LookupEnv retrieves the value of the environment variable named by the key.
    LookupEnv(key string) (value string, ok bool)
}
```

Here's an example of using `Map`, a builtin `Source` implementation useful in tests:

```go
m := env.Map{"PORT": "8080"}

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg, env.WithSource(m)); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port)
// Output: 8080
```

#### Prefix

It is a common practice to prefix app's environment variables with some string
(e.g., its name). Such a prefix can be set using the `WithPrefix` option:

```go
os.Setenv("APP_PORT", "8080")

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg, env.WithPrefix("APP_")); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port)
// Output: 8080
```

#### Slice separator

Space is the default separator when parsing slice values. It can be changed
using the `WithSliceSeparator` option:

```go
os.Setenv("PORTS", "8080;8081;8082")

var cfg struct {
    Ports []int `env:"PORTS"`
}
if err := env.Load(&cfg, env.WithSliceSeparator(";")); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Ports)
// Output: [8080 8081 8082]
```

### Usage message

`env` supports printing an auto-generated usage message the same way the `flag` package does it.

```go
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
```

[1]: https://12factor.net/config
