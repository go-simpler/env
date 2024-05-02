package env_test

import (
	"errors"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/EF"
)

//go:generate go run -tags=cp go-simpler.org/assert/cmd/cp@v0.8.0 -dir=internal

func TestLoad(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		tests := map[string]any{
			"nil":                  nil,
			"not a pointer":        struct{}{},
			"not a struct pointer": new(int),
			"nil struct pointer":   (*struct{})(nil),
		}

		for name, cfg := range tests {
			t.Run(name, func(t *testing.T) {
				const panicMsg = "env: cfg must be a non-nil struct pointer"

				load := func() { _ = env.Load(cfg, nil) }
				assert.Panics[E](t, load, panicMsg)

				usage := func() { env.Usage(cfg, io.Discard, nil) }
				assert.Panics[E](t, usage, panicMsg)
			})
		}
	})

	t.Run("empty tag name", func(t *testing.T) {
		var cfg struct {
			Foo int `env:""`
		}
		load := func() { _ = env.Load(&cfg, nil) }
		assert.Panics[E](t, load, "env: empty tag name is not allowed")
	})

	t.Run("invalid tag option", func(t *testing.T) {
		var cfg struct {
			Foo int `env:"FOO,?"`
		}
		load := func() { _ = env.Load(&cfg, nil) }
		assert.Panics[E](t, load, "env: invalid tag option `?`")
	})

	t.Run("required with default", func(t *testing.T) {
		var cfg struct {
			Foo int `env:"FOO,required" default:"1"`
		}
		load := func() { _ = env.Load(&cfg, nil) }
		assert.Panics[E](t, load, "env: `required` and `default` can't be used simultaneously")
	})

	t.Run("nested struct w/ and w/o tag", func(t *testing.T) {
		m := env.Map{"A_FOO": "1", "BAR": "2"}

		var cfg struct {
			A struct {
				Foo int `env:"FOO"`
			} `env:"A"`
			B struct {
				Bar int `env:"BAR"`
			}
		}
		err := env.Load(&cfg, &env.Options{Source: m, NameSep: "_"})
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.A.Foo, 1)
		assert.Equal[E](t, cfg.B.Bar, 2)
	})

	t.Run("unsupported type", func(t *testing.T) {
		m := env.Map{"FOO": "1+2i"}

		var cfg struct {
			Foo complex64 `env:"FOO"`
		}
		load := func() { _ = env.Load(&cfg, &env.Options{Source: m}) }
		assert.Panics[E](t, load, "env: unsupported type `complex64`")
	})

	t.Run("ignored fields", func(t *testing.T) {
		m := env.Map{"FOO": "1", "BAR": "2"}

		var cfg struct {
			foo int `env:"FOO"`
			Bar int
		}
		err := env.Load(&cfg, &env.Options{Source: m})
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.foo, 0)
		assert.Equal[E](t, cfg.Bar, 0)
	})

	t.Run("all supported types", func(t *testing.T) {
		m := env.Map{
			"INT": "-1", "INTS": "-1 0",
			"INT8": "-8", "INT8S": "-8 0",
			"INT16": "-16", "INT16S": "-16 0",
			"INT32": "-32", "INT32S": "-32 0",
			"INT64": "-64", "INT64S": "-64 0",
			"UINT": "1", "UINTS": "0 1",
			"UINT8": "8", "UINT8S": "0 8",
			"UINT16": "16", "UINT16S": "0 16",
			"UINT32": "32", "UINT32S": "0 32",
			"UINT64": "64", "UINT64S": "0 64",
			"FLOAT32": "0.32", "FLOAT32S": "0.32 0.64",
			"FLOAT64": "0.64", "FLOAT64S": "0.64 0.32",
			"BOOL": "true", "BOOLS": "true false",
			"STRING": "foo", "STRINGS": "foo bar",
			"DURATION": "1s", "DURATIONS": "1s 1m",
			"IP": "0.0.0.0", "IPS": "0.0.0.0 255.255.255.255",
		}

		var cfg struct {
			Int       int             `env:"INT"`
			Ints      []int           `env:"INTS"`
			Int8      int8            `env:"INT8"`
			Int8s     []int8          `env:"INT8S"`
			Int16     int16           `env:"INT16"`
			Int16s    []int16         `env:"INT16S"`
			Int32     int32           `env:"INT32"`
			Int32s    []int32         `env:"INT32S"`
			Int64     int64           `env:"INT64"`
			Int64s    []int64         `env:"INT64S"`
			Uint      uint            `env:"UINT"`
			Uints     []uint          `env:"UINTS"`
			Uint8     uint8           `env:"UINT8"`
			Uint8s    []uint8         `env:"UINT8S"`
			Uint16    uint16          `env:"UINT16"`
			Uint16s   []uint16        `env:"UINT16S"`
			Uint32    uint32          `env:"UINT32"`
			Uint32s   []uint32        `env:"UINT32S"`
			Uint64    uint64          `env:"UINT64"`
			Uint64s   []uint64        `env:"UINT64S"`
			Float32   float32         `env:"FLOAT32"`
			Float32s  []float32       `env:"FLOAT32S"`
			Float64   float64         `env:"FLOAT64"`
			Float64s  []float64       `env:"FLOAT64S"`
			Bool      bool            `env:"BOOL"`
			Bools     []bool          `env:"BOOLS"`
			String    string          `env:"STRING"`
			Strings   []string        `env:"STRINGS"`
			Duration  time.Duration   `env:"DURATION"`
			Durations []time.Duration `env:"DURATIONS"`
			IP        net.IP          `env:"IP"`
			IPs       []net.IP        `env:"IPS"`
		}
		err := env.Load(&cfg, &env.Options{Source: m})
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Int, -1)
		assert.Equal[E](t, cfg.Ints, []int{-1, 0})
		assert.Equal[E](t, cfg.Int8, -8)
		assert.Equal[E](t, cfg.Int8s, []int8{-8, 0})
		assert.Equal[E](t, cfg.Int16, -16)
		assert.Equal[E](t, cfg.Int16s, []int16{-16, 0})
		assert.Equal[E](t, cfg.Int32, -32)
		assert.Equal[E](t, cfg.Int32s, []int32{-32, 0})
		assert.Equal[E](t, cfg.Int64, -64)
		assert.Equal[E](t, cfg.Int64s, []int64{-64, 0})
		assert.Equal[E](t, cfg.Uint, 1)
		assert.Equal[E](t, cfg.Uints, []uint{0, 1})
		assert.Equal[E](t, cfg.Uint8, 8)
		assert.Equal[E](t, cfg.Uint8s, []uint8{0, 8})
		assert.Equal[E](t, cfg.Uint16, 16)
		assert.Equal[E](t, cfg.Uint16s, []uint16{0, 16})
		assert.Equal[E](t, cfg.Uint32, 32)
		assert.Equal[E](t, cfg.Uint32s, []uint32{0, 32})
		assert.Equal[E](t, cfg.Uint64, 64)
		assert.Equal[E](t, cfg.Uint64s, []uint64{0, 64})
		assert.Equal[E](t, cfg.Float32, 0.32)
		assert.Equal[E](t, cfg.Float32s, []float32{0.32, 0.64})
		assert.Equal[E](t, cfg.Float64, 0.64)
		assert.Equal[E](t, cfg.Float64s, []float64{0.64, 0.32})
		assert.Equal[E](t, cfg.Bool, true)
		assert.Equal[E](t, cfg.Bools, []bool{true, false})
		assert.Equal[E](t, cfg.String, "foo")
		assert.Equal[E](t, cfg.Strings, []string{"foo", "bar"})
		assert.Equal[E](t, cfg.Duration, time.Second)
		assert.Equal[E](t, cfg.Durations, []time.Duration{time.Second, time.Minute})
		assert.Equal[E](t, cfg.IP, net.IPv4zero)
		assert.Equal[E](t, cfg.IPs, []net.IP{net.IPv4zero, net.IPv4bcast})
	})

	t.Run("parsing errors", func(t *testing.T) {
		tests := map[string]struct {
			src      env.Source
			checkErr func(error)
		}{
			"invalid int": {
				src:      env.Map{"INT": "-"},
				checkErr: func(err error) { assert.IsErr[E](t, err, strconv.ErrSyntax) },
			},
			"invalid uint": {
				src:      env.Map{"UINT": "-"},
				checkErr: func(err error) { assert.IsErr[E](t, err, strconv.ErrSyntax) },
			},
			"invalid float64": {
				src:      env.Map{"FLOAT64": "-"},
				checkErr: func(err error) { assert.IsErr[E](t, err, strconv.ErrSyntax) },
			},
			"invalid bool": {
				src:      env.Map{"BOOL": "-"},
				checkErr: func(err error) { assert.IsErr[E](t, err, strconv.ErrSyntax) },
			},
			"invalid time.Duration": {
				src:      env.Map{"DURATION": "-"},
				checkErr: func(err error) { assert.Equal[E](t, errors.Unwrap(err).Error(), `time: invalid duration "-"`) },
			},
			"invalid encoding.TextUnmarshaler": {
				src:      env.Map{"IP": "-"},
				checkErr: func(err error) { assert.AsErr[E](t, err, new(*net.ParseError)) },
			},
			"invalid slice": {
				src:      env.Map{"IPS": "-"},
				checkErr: func(err error) { assert.AsErr[E](t, err, new(*net.ParseError)) },
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				var cfg struct {
					Int      int           `env:"INT"`
					Uint     uint          `env:"UINT"`
					Float64  float64       `env:"FLOAT64"`
					Bool     bool          `env:"BOOL"`
					Duration time.Duration `env:"DURATION"`
					IP       net.IP        `env:"IP"`
					IPs      []net.IP      `env:"IPS"`
				}
				err := env.Load(&cfg, &env.Options{Source: test.src})
				test.checkErr(err)
			})
		}
	})
}
