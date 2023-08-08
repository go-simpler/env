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

// MultiSource combines multiple sources into a single one containing the union of all environment variables.
// The order of the given sources matters: if the same key occurs more than once, the later value takes precedence.
func MultiSource(ss ...Source) Source { return sources(ss) }

type sources []Source

func (ss sources) LookupEnv(key string) (string, bool) {
	var value string
	var found bool
	for _, s := range ss {
		if v, ok := s.LookupEnv(key); ok {
			value = v
			found = true
		}
	}
	return value, found
}
