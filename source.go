package env

import "os"

// Source represents a source of environment variables.
type Source interface {
	// LookupEnv retrieves the value of the environment variable named by the key.
	LookupEnv(key string) (value string, ok bool)
}

// OS is the main [Source] that uses [os.LookupEnv].
var OS Source = sourceFunc(os.LookupEnv)

// Map is a [Source] implementation useful in tests.
type Map map[string]string

// LookupEnv implements the [Source] interface.
func (m Map) LookupEnv(key string) (string, bool) {
	value, ok := m[key]
	return value, ok
}

type sourceFunc func(key string) (string, bool)

func (fn sourceFunc) LookupEnv(key string) (string, bool) { return fn(key) }
