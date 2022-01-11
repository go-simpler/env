package env_test

import (
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/junk1tm/env"
)

func TestLoadFrom(t *testing.T) {
	t.Run("invalid argument error", func(t *testing.T) {
		test := func(name string, v interface{}) {
			t.Run(name, func(t *testing.T) {
				if err := env.LoadFrom(env.Map{}, v); !errors.Is(err, env.ErrInvalidArgument) {
					t.Errorf("got %v; want %v", err, env.ErrInvalidArgument)
				}
			})
		}

		test("nil", nil)
		test("not a pointer", struct{}{})
		test("not a struct pointer", new(int))
		test("nil struct pointer", (*struct{})(nil))
	})

	t.Run("not set error", func(t *testing.T) {
		var notSetErr *env.NotSetError

		var cfg struct {
			Port string `env:"PORT,required"`
		}
		if err := env.LoadFrom(env.Map{}, &cfg); !errors.As(err, &notSetErr) {
			t.Fatalf("got %T; want %T", err, notSetErr)
		}
		if !reflect.DeepEqual(notSetErr.Names, []string{"PORT"}) {
			t.Errorf("got %v; want [PORT]", notSetErr.Names)
		}
	})

	t.Run("unsupported type error", func(t *testing.T) {
		m := env.Map{"PORT": "8080"}

		var unsupportedTypeErr *env.UnsupportedTypeError

		var cfg struct {
			Port complex64 `env:"PORT,required"`
		}
		if err := env.LoadFrom(m, &cfg); !errors.As(err, &unsupportedTypeErr) {
			t.Fatalf("got %T; want %T", err, unsupportedTypeErr)
		}
		if unsupportedTypeErr.Type != reflect.TypeOf(cfg.Port) {
			t.Errorf("got %s; want %s", unsupportedTypeErr.Type, reflect.TypeOf(cfg.Port))
		}
	})

	t.Run("with default values", func(t *testing.T) {
		m := env.Map{"PORT": "8081"}

		cfg := struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		}{
			Host: "localhost", // must stay the same.
			Port: 8080,        // must be overridden with 8081.
		}
		if err := env.LoadFrom(m, &cfg); err != nil {
			t.Fatalf("got %v; want no error", err)
		}
		if cfg.Host != "localhost" {
			t.Errorf("got %s; want localhost", cfg.Host)
		}
		if cfg.Port != 8081 {
			t.Errorf("got %d; want 8080", cfg.Port)
		}
	})

	t.Run("with nested structs", func(t *testing.T) {
		m := env.Map{
			"DB_PORT":   "5432",
			"HTTP_PORT": "8080",
		}

		var cfg struct {
			DB struct {
				Port int `env:"DB_PORT,required"`
			}
			HTTP struct {
				Port int `env:"HTTP_PORT,required"`
			}
		}
		if err := env.LoadFrom(m, &cfg); err != nil {
			t.Fatalf("got %v; want no error", err)
		}
		if cfg.DB.Port != 5432 {
			t.Errorf("got %d; want 5432", cfg.DB.Port)
		}
		if cfg.HTTP.Port != 8080 {
			t.Errorf("got %d; want 8080", cfg.HTTP.Port)
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		m := env.Map{"APP_PORT": "8080"}

		var cfg struct {
			Port int `env:"PORT,required"`
		}
		if err := env.LoadFrom(m, &cfg, env.WithPrefix("APP_")); err != nil {
			t.Fatalf("got %v; want no error", err)
		}
		if cfg.Port != 8080 {
			t.Errorf("got %d; want 8080", cfg.Port)
		}
	})

	t.Run("with slice separator", func(t *testing.T) {
		m := env.Map{"PORTS": "8080;8081;8082"}

		var cfg struct {
			Ports []int `env:"PORTS,required"`
		}
		if err := env.LoadFrom(m, &cfg, env.WithSliceSeparator(";")); err != nil {
			t.Fatalf("got %v; want no error", err)
		}
		if want := []int{8080, 8081, 8082}; !reflect.DeepEqual(cfg.Ports, want) {
			t.Errorf("got %v; want %v", cfg.Ports, want)
		}
	})

	t.Run("skipped fields", func(t *testing.T) {
		m := env.Map{
			"UNEXPORTED":  "foo",
			"MISSING_TAG": "bar",
			"EMPTY_NAME":  "baz",
		}

		var cfg struct {
			unexported string `env:"UNEXPORTED,required"`
			MissingTag string `json:"missing_tag"`
			EmptyName  string `env:",required"`
		}
		if err := env.LoadFrom(m, &cfg); err != nil {
			t.Fatalf("got %v; want no error", err)
		}
		if cfg.unexported != "" {
			t.Errorf("got %s; want empty string", cfg.unexported)
		}
		if cfg.MissingTag != "" {
			t.Errorf("got %s; want empty string", cfg.MissingTag)
		}
		if cfg.EmptyName != "" {
			t.Errorf("got %s; want empty string", cfg.EmptyName)
		}
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
			Int       int             `env:"INT,required"`
			Ints      []int           `env:"INTS,required"`
			Int8      int8            `env:"INT8,required"`
			Int8s     []int8          `env:"INT8S,required"`
			Int16     int16           `env:"INT16,required"`
			Int16s    []int16         `env:"INT16S,required"`
			Int32     int32           `env:"INT32,required"`
			Int32s    []int32         `env:"INT32S,required"`
			Int64     int64           `env:"INT64,required"`
			Int64s    []int64         `env:"INT64S,required"`
			Uint      uint            `env:"UINT,required"`
			Uints     []uint          `env:"UINTS,required"`
			Uint8     uint8           `env:"UINT8,required"`
			Uint8s    []uint8         `env:"UINT8S,required"`
			Uint16    uint16          `env:"UINT16,required"`
			Uint16s   []uint16        `env:"UINT16S,required"`
			Uint32    uint32          `env:"UINT32,required"`
			Uint32s   []uint32        `env:"UINT32S,required"`
			Uint64    uint64          `env:"UINT64,required"`
			Uint64s   []uint64        `env:"UINT64S,required"`
			Float32   float32         `env:"FLOAT32,required"`
			Float32s  []float32       `env:"FLOAT32S,required"`
			Float64   float64         `env:"FLOAT64,required"`
			Float64s  []float64       `env:"FLOAT64S,required"`
			Bool      bool            `env:"BOOL,required"`
			Bools     []bool          `env:"BOOLS,required"`
			String    string          `env:"STRING,required"`
			Strings   []string        `env:"STRINGS,required"`
			Duration  time.Duration   `env:"DURATION,required"`
			Durations []time.Duration `env:"DURATIONS,required"`
			IP        net.IP          `env:"IP,required"`
			IPs       []net.IP        `env:"IPS,required"`
		}
		if err := env.LoadFrom(m, &cfg); err != nil {
			t.Fatalf("got %v; want no error", err)
		}

		test := func(name string, got, want interface{}) {
			t.Run(name, func(t *testing.T) {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %v; want %v", got, want)
				}
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
}
