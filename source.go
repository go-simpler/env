package env

import "os"

// Source represents a source of environment variables.
type Source interface {
	// LookupEnv retrieves the value of the environment variable named by the key.
	LookupEnv(key string) (value string, ok bool)
}

// SourceFunc is an adapter that allows using a function as a [Source].
type SourceFunc func(key string) (value string, ok bool)

// LookupEnv implements the [Source] interface.
func (fn SourceFunc) LookupEnv(key string) (string, bool) { return fn(key) }

// OS is the main [Source] that uses [os.LookupEnv].
var OS Source = SourceFunc(os.LookupEnv)

// Map is an in-memory [Source] implementation useful in tests.
type Map map[string]string

// LookupEnv implements the [Source] interface.
func (m Map) LookupEnv(key string) (string, bool) {
	value, ok := m[key]
	return value, ok
}

type multiSource []Source

func (ms multiSource) LookupEnv(key string) (string, bool) {
	for _, src := range ms {
		if value, ok := src.LookupEnv(key); ok {
			return value, true
		}
	}
	return "", false
}
