<div align="center">

![logo](logo.svg)

🔍 Load environment variables into a config struct

[![awesome-go](https://awesome.re/badge.svg)](https://github.com/avelino/awesome-go#configuration)
[![checks](https://github.com/go-simpler/env/actions/workflows/checks.yml/badge.svg)](https://github.com/go-simpler/env/actions/workflows/checks.yml)
[![pkg.go.dev](https://pkg.go.dev/badge/go-simpler.org/env.svg)](https://pkg.go.dev/go-simpler.org/env)
[![goreportcard](https://goreportcard.com/badge/go-simpler.org/env)](https://goreportcard.com/report/go-simpler.org/env)
[![codecov](https://codecov.io/gh/go-simpler/env/branch/main/graph/badge.svg)](https://codecov.io/gh/go-simpler/env)

</div>

## 📌 About

This package is made for apps that [store config in environment variables][1].
Its purpose is to replace fragmented `os.Getenv` calls in `main.go` with a single struct definition,
which simplifies config management and improves code readability.

## 🚀 Features

* Support for all common types and user-defined types
* Options: [required](#required), [expand](#expand), [slice separator](#slice-separator)
* Configurable [source](#source) of environment variables
* Auto-generated [usage message](#usage-message)

## 📦 Install

Go 1.20+
```shell
go get go-simpler.org/env
```

## 📋 Usage

`Load` is the main function of the package.
It loads environment variables into the given struct.

The struct fields must have the `env:"VAR"` struct tag,
where VAR is the name of the corresponding environment variable.
Unexported fields are ignored.

```go
os.Setenv("PORT", "8080")

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg, nil); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port) // 8080
```

### Supported types

* `int` (any kind)
* `float` (any kind)
* `bool`
* `string`
* `time.Duration`
* `encoding.TextUnmarshaler`
* slices of any type above
* nested structs of any depth

See the `strconv.Parse*` functions for the parsing rules.
User-defined types can be used by implementing the `encoding.TextUnmarshaler` interface.

### Nested structs

Nested struct of any depth level are supported,
allowing grouping of related environment variables.

```go
os.Setenv("HTTP_PORT", "8080")

var cfg struct {
    HTTP struct {
        Port int `env:"HTTP_PORT"`
    }
}
if err := env.Load(&cfg, nil); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.HTTP.Port) // 8080
```

A nested struct can have the optional `env:"PREFIX"` tag.
In this case, the environment variables declared by its fields are prefixed with PREFIX.
This rule is applied recursively to all nested structs.

```go
os.Setenv("DB_HOST", "localhost")
os.Setenv("DB_PORT", "5432")

var cfg struct {
    DB struct {
        Host string `env:"HOST"`
        Port int    `env:"PORT"`
    } `env:"DB_"`
}
if err := env.Load(&cfg, nil); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.DB.Host) // localhost
fmt.Println(cfg.DB.Port) // 5432
```

### Default values

Default values can be specified using the `default:"VALUE"` struct tag:

```go
os.Unsetenv("PORT")

var cfg struct {
    Port int `env:"PORT" default:"8080"`
}
if err := env.Load(&cfg, nil); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port) // 8080
```

### Required

Use the `required` option to mark an environment variable as required.
If it is not set, an error of type `NotSetError` is returned.

```go
os.Unsetenv("PORT")

var cfg struct {
    Port int `env:"PORT,required"`
}
if err := env.Load(&cfg, nil); err != nil {
    var notSetErr *env.NotSetError
    if errors.As(err, &notSetErr) {
        fmt.Println(notSetErr) // env: PORT is required but not set
    }
}
```

### Expand

Use the `expand` option to automatically expand the value of an environment variable using `os.Expand`.

```go
os.Setenv("PORT", "8080")
os.Setenv("ADDR", "localhost:${PORT}")

var cfg struct {
    Addr string `env:"ADDR,expand"`
}
if err := env.Load(&cfg, nil); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Addr) // localhost:8080
```

### Slice separator

Space is the default separator used to parse slice values.
It can be changed with `Options.SliceSep`:

```go
os.Setenv("PORTS", "8080,8081,8082")

var cfg struct {
    Ports []int `env:"PORTS"`
}
if err := env.Load(&cfg, &env.Options{SliceSep: ","}); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Ports) // [8080 8081 8082]
```

### Source

By default, `Load` retrieves environment variables directly from OS.
To use a different source, pass an implementation of the `Source` interface via `Options.Source`.

```go
type Source interface {
    LookupEnv(key string) (value string, ok bool)
}
```

Here's an example of using `Map`, a `Source` implementation useful in tests:

```go
m := env.Map{"PORT": "8080"}

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg, &env.Options{Source: m}); err != nil {
    fmt.Println(err)
}

fmt.Println(cfg.Port) // 8080
```

### Usage message

The `Usage` function prints a usage message documenting all defined environment variables.
An optional usage string can be added for each environment variable via the `usage:"STRING"` struct tag:

```go
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
    env.Usage(&cfg, os.Stdout)
}
```

```
Usage:
  DB_HOST    string  required      database host
  DB_PORT    int     required      database port
  HTTP_PORT  int     default 8080  http server port
```

The format of the message can be customized by implementing the `Usage([]env.Var, io.Writer)` method:

```go
type config struct{ ... }

func (config) Usage(vars []env.Var, w io.Writer) {
    for v := range vars {
        // write to w.
    }
}
```

[1]: https://12factor.net/config
