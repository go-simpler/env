# env

[![ci-img]][ci]
[![docs-img]][docs]
[![report-img]][report]
[![codecov-img]][codecov]
[![license-img]][license]
[![release-img]][release]

> A lightweight package for loading environment variables into structs

## About

This package is made for apps that [store config in environment variables][12factor]. Its purpose is to replace multiple
fragmented `os.Getenv` calls in `main.go` with a single struct definition, which simplifies config management and
improves code readability.

## Install

```
go get github.com/junk1tm/env
```

## Features

* Simple API
* Dependency-free
* `required` validation
* Lightweight yet [customizable](customization)

## Supported types

The following types are supported as struct fields:

* `int` (any kind)
* `float` (any kind)
* `bool`
* `string`
* `time.Duration`
* `encoding.TextUnmarshaler`
* slices of any type above

See the `strconv` package from the standard library for parsing rules.

## Usage

`Load` is the main function of this package. It loads environment variables into the provided struct.

The struct fields must have the `env:"VAR"` struct tag, where `VAR` is the name of the corresponding environment
variable. Any unexported fields, fields without this tag (except nested structs) or fields with empty name are ignored.

```go
os.Setenv("PORT", "8080")

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg); err != nil {
    // handle error
}

fmt.Println(cfg.Port) // 8080
```

### Required

If a field has the tag in the form of `env:"VAR,required"`, it will be marked as required and an error of type
`NotSetError` will be returned in case no such environment variable is found.

```go
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
```

### Default values

Default values can be specified using basic struct initialization. They will be left untouched, if no corresponding
environment variables are found.

```go
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
```

### Nested structs

Nested structs of any depth level are supported, but only non-struct fields are considered as targets for parsing.

```go
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
```

See [example_test.go][examples] for more examples.

## Customization

### Provider

`Load` retrieves environment variables values directly from OS. To use a different source, try `LoadFrom` that accepts
an implementation of the `Provider` interface as the first argument.

```go
// Provider represents an entity that is able to provide environment variables.
type Provider interface {
    // LookupEnv retrieves the value of the environment variable named by the
    // key. If it is not found, the boolean will be false.
    LookupEnv(key string) (value string, ok bool)
}
```

`Map` is a builtin `Provider` implementation that might be useful in tests.

```go
m := env.Map{"PORT": "8080"}

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.LoadFrom(m, &cfg); err != nil {
    // handle error
}

fmt.Println(cfg.Port) // 8080
```

### Options

`Load`'s behavior can be customized using various options:

#### Prefix

It is a common practise to prefix app's environment variables with some string (e.g., its name). Such a prefix can be
set using `WithPrefix` option:

```go
os.Setenv("APP_PORT", "8080")

var cfg struct {
    Port int `env:"PORT"`
}
if err := env.Load(&cfg, env.WithPrefix("APP")); err != nil {
    // handle error
}

fmt.Println(cfg.Port) // 8080
```

#### Slice separator

Space is the default separator when parsing slice values. It can be changed using `WithSliceSeparator` option:

```go
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
```

[ci]: https://github.com/junk1tm/env/actions/workflows/go.yml
[ci-img]: https://github.com/junk1tm/env/actions/workflows/go.yml/badge.svg
[docs]: https://pkg.go.dev/github.com/junk1tm/env
[docs-img]: https://pkg.go.dev/badge/github.com/junk1tm/env.svg
[report]: https://goreportcard.com/report/github.com/junk1tm/env
[report-img]: https://goreportcard.com/badge/github.com/junk1tm/env
[codecov]: https://codecov.io/gh/junk1tm/env
[codecov-img]: https://codecov.io/gh/junk1tm/env/branch/main/graph/badge.svg
[license]: https://github.com/junk1tm/env/blob/main/LICENSE
[license-img]: https://img.shields.io/github/license/junk1tm/env
[release]: https://github.com/junk1tm/env/releases
[release-img]: https://img.shields.io/github/v/release/junk1tm/env
[12factor]: https://12factor.net/config
[examples]: https://github.com/junk1tm/env/blob/main/example_test.go
