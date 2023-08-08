package env_test

import (
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"go-simpler.org/env"
	"go-simpler.org/env/internal/assert"
	. "go-simpler.org/env/internal/assert/dotimport"
)

//go:generate go run -tags=copier go-simpler.org/assert/cmd/copier@v0.6.0 internal

func TestLoad(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		test := func(name string, cfg any) {
			t.Run(name, func(t *testing.T) {
				assert.Panics[E](t,
					func() { _ = env.Load(cfg, env.WithSource(env.Map{})) },
					"env: argument must be a non-nil struct pointer",
				)
			})
		}

		test("nil", nil)
		test("not a pointer", struct{}{})
		test("not a struct pointer", new(int))
		test("nil struct pointer", (*struct{})(nil))
	})

	t.Run("empty tag name", func(t *testing.T) {
		var cfg struct {
			Port string `env:""`
		}
		assert.Panics[E](t,
			func() { _ = env.Load(&cfg, env.WithSource(env.Map{})) },
			"env: empty tag name is not allowed",
		)
	})

	t.Run("unsupported type", func(t *testing.T) {
		m := env.Map{"PORT": "8080"}

		var cfg struct {
			Port complex64 `env:"PORT"`
		}
		assert.Panics[E](t,
			func() { _ = env.Load(&cfg, env.WithSource(m)) },
			"env: unsupported type `complex64`",
		)
	})

	t.Run("ignored fields", func(t *testing.T) {
		m := env.Map{
			"UNEXPORTED":  "foo",
			"MISSING_TAG": "bar",
		}

		var cfg struct {
			unexported string `env:"UNEXPORTED"`
			MissingTag string
		}
		err := env.Load(&cfg, env.WithSource(m))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.unexported, "")
		assert.Equal[E](t, cfg.MissingTag, "")
	})

	t.Run("default values", func(t *testing.T) {
		cfg := struct {
			Host string `env:"HOST" default:"localhost"`
			Port int    `env:"PORT" default:"8080"`
		}{
			Port: 8000, // must be overridden with 8080 (from the `default` tag).
		}
		err := env.Load(&cfg, env.WithSource(env.Map{}))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Host, "localhost")
		assert.Equal[E](t, cfg.Port, 8080)
	})

	t.Run("nested structs", func(t *testing.T) {
		m := env.Map{
			"DB_PORT":   "5432",
			"HTTP_PORT": "8080",
		}

		var cfg struct {
			DB struct {
				Port int `env:"DB_PORT"`
			}
			HTTP struct {
				Port int `env:"HTTP_PORT"`
			}
		}
		err := env.Load(&cfg, env.WithSource(m))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.DB.Port, 5432)
		assert.Equal[E](t, cfg.HTTP.Port, 8080)
	})

	t.Run("required tag option", func(t *testing.T) {
		var notSetErr *env.NotSetError

		var cfg struct {
			Host string `env:"HOST,required"`
			Port int    `env:"PORT,required"`
		}
		err := env.Load(&cfg, env.WithSource(env.Map{}))
		assert.AsErr[F](t, err, &notSetErr)
		assert.Equal[E](t, notSetErr.Names, []string{"HOST", "PORT"})

		// more coverage!
		_ = notSetErr.Error()
	})

	t.Run("expand tag option", func(t *testing.T) {
		m := env.Map{
			"HOST": "localhost",
			"PORT": "8080",
			"ADDR": "$HOST:${PORT}", // try both $VAR and ${VAR} forms.
		}

		var cfg struct {
			Addr string `env:"ADDR,expand"`
		}
		err := env.Load(&cfg, env.WithSource(m))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Addr, "localhost:8080")
	})

	t.Run("invalid tag option", func(t *testing.T) {
		var cfg struct {
			HTTP struct {
				Port string `env:"HTTP_PORT,foo"`
			}
		}
		assert.Panics[E](t,
			func() { _ = env.Load(&cfg, env.WithSource(env.Map{})) },
			"env: invalid tag option `foo`",
		)
	})

	t.Run("with source", func(t *testing.T) {
		m1 := env.Map{"FOO": "1", "BAR": "2"}
		m2 := env.Map{"FOO": "2", "BAZ": "3"}
		m3 := env.Map{"BAR": "3", "BAZ": "4"}

		var cfg struct {
			Foo int `env:"FOO,required"`
			Bar int `env:"BAR,required"`
			Baz int `env:"BAZ,required"`
		}
		err := env.Load(&cfg, env.WithSource(m1, m2, m3))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Foo, 2)
		assert.Equal[E](t, cfg.Bar, 3)
		assert.Equal[E](t, cfg.Baz, 4)
	})

	t.Run("with prefix", func(t *testing.T) {
		m := env.Map{"APP_PORT": "8080"}

		var cfg struct {
			Port int `env:"PORT"`
		}
		err := env.Load(&cfg, env.WithSource(m), env.WithPrefix("APP_"))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Port, 8080)
	})

	t.Run("with slice separator", func(t *testing.T) {
		m := env.Map{"PORTS": "8080;8081;8082"}

		var cfg struct {
			Ports []int `env:"PORTS"`
		}
		err := env.Load(&cfg, env.WithSource(m), env.WithSliceSeparator(";"))
		assert.NoErr[F](t, err)
		assert.Equal[E](t, cfg.Ports, []int{8080, 8081, 8082})
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
			"FLOAT32": "0.1", "FLOAT32S": "0.1 0.2 0.3",
			"FLOAT64": "0.2", "FLOAT64S": "0.2 0.4 0.6",
			"BOOL": "true", "BOOLS": "true false",
			"STRING": "foo", "STRINGS": "foo bar baz",
			"DURATION": "1s", "DURATIONS": "1s 1m 1h",
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
		err := env.Load(&cfg, env.WithSource(m))
		assert.NoErr[F](t, err)

		test := func(name string, got, want any) {
			t.Run(name, func(t *testing.T) {
				assert.Equal[E](t, got, want)
			})
		}

		test("int", cfg.Int, -1)
		test("ints", cfg.Ints, []int{-1, 0})
		test("int8", cfg.Int8, int8(-8))
		test("int8s", cfg.Int8s, []int8{-8, 0})
		test("int16", cfg.Int16, int16(-16))
		test("int16s", cfg.Int16s, []int16{-16, 0})
		test("int32", cfg.Int32, int32(-32))
		test("int32s", cfg.Int32s, []int32{-32, 0})
		test("int64", cfg.Int64, int64(-64))
		test("int64s", cfg.Int64s, []int64{-64, 0})
		test("uint", cfg.Uint, uint(1))
		test("uints", cfg.Uints, []uint{0, 1})
		test("uint8", cfg.Uint8, uint8(8))
		test("uint8s", cfg.Uint8s, []uint8{0, 8})
		test("uint16", cfg.Uint16, uint16(16))
		test("uint16s", cfg.Uint16s, []uint16{0, 16})
		test("uint32", cfg.Uint32, uint32(32))
		test("uint32s", cfg.Uint32s, []uint32{0, 32})
		test("uint64", cfg.Uint64, uint64(64))
		test("uint64s", cfg.Uint64s, []uint64{0, 64})
		test("float32", cfg.Float32, float32(0.1))
		test("float32s", cfg.Float32s, []float32{0.1, 0.2, 0.3})
		test("float64", cfg.Float64, 0.2)
		test("float64s", cfg.Float64s, []float64{0.2, 0.4, 0.6})
		test("bool", cfg.Bool, true)
		test("bools", cfg.Bools, []bool{true, false})
		test("string", cfg.String, "foo")
		test("strings", cfg.Strings, []string{"foo", "bar", "baz"})
		test("duration", cfg.Duration, time.Second)
		test("durations", cfg.Durations, []time.Duration{time.Second, time.Minute, time.Hour})
		test("unmarshaler", cfg.IP, net.IPv4zero)
		test("unmarshalers", cfg.IPs, []net.IP{net.IPv4zero, net.IPv4bcast})
	})

	t.Run("parsing errors", func(t *testing.T) {
		test := func(name, envName string, checkErr func(error) bool) {
			t.Run(name, func(t *testing.T) {
				// "-" is an invalid value for all the following types,
				// it causes an error for strconv.Parse*, time.ParseDuration and net.ParseIP.
				m := env.Map{envName: "-"}

				var cfg struct {
					Int         int           `env:"INT"`
					Uint        uint          `env:"UINT"`
					Float       float64       `env:"FLOAT"`
					Bool        bool          `env:"BOOL"`
					Duration    time.Duration `env:"DURATION"`
					Unmarshaler net.IP        `env:"UNMARSHALER"`
					Slice       []net.IP      `env:"SLICE"`
				}
				err := env.Load(&cfg, env.WithSource(m))
				assert.Equal[E](t, checkErr(err), true)
			})
		}

		isErrSyntax := func(err error) bool {
			return errors.Is(err, strconv.ErrSyntax)
		}
		isInvalidDuration := func(err error) bool {
			return errors.Unwrap(err).Error() == `time: invalid duration "-"`
		}
		asParseError := func(err error) bool {
			return errors.As(err, new(*net.ParseError))
		}

		test("invalid int", "INT", isErrSyntax)
		test("invalid uint", "UINT", isErrSyntax)
		test("invalid float", "FLOAT", isErrSyntax)
		test("invalid bool", "BOOL", isErrSyntax)
		test("invalid duration", "DURATION", isInvalidDuration)
		test("invalid unmarshaler", "UNMARSHALER", asParseError)
		test("invalid slice", "SLICE", asParseError)
	})
}
