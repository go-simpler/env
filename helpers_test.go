package env_test

import (
	"errors"
	"reflect"
	"testing"
)

type (
	E *testing.T
	F *testing.T
)

func noerr[T E | F](t T, err error) {
	(*testing.T)(t).Helper()
	if err != nil {
		f(t)("got %v; want no error", err)
	}
}

func iserr[T E | F](t T, err, target error) {
	(*testing.T)(t).Helper()
	if !errors.Is(err, target) {
		f(t)("got %v; want %v", err, target)
	}
}

func aserr[T E | F](t T, err error, target any) {
	(*testing.T)(t).Helper()
	if !errors.As(err, target) {
		f(t)("got %T; want %T", err, target)
	}
}

func equal[T E | F](t T, got, want any) {
	(*testing.T)(t).Helper()
	if !reflect.DeepEqual(got, want) {
		if got == "" {
			got = "<empty>"
		}
		if want == "" {
			want = "<empty>"
		}
		f(t)("got %v; want %v", got, want)
	}
}

func f[T E | F](t T) func(format string, args ...any) {
	switch any(t).(type) {
	case E:
		return (*testing.T)(t).Errorf
	case F:
		return (*testing.T)(t).Fatalf
	default:
		panic("unreachable")
	}
}
