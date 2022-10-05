package env

import "os"

// Provider represents an entity that is able to provide environment variables.
type Provider interface {
	// LookupEnv retrieves the value of the environment variable named by the
	// key. If it is not found, the boolean will be false.
	LookupEnv(key string) (value string, ok bool)
}

// ProviderFunc is an adapter that allows using functions as [Provider].
type ProviderFunc func(key string) (value string, ok bool)

// LookupEnv implements the [Provider] interface.
func (f ProviderFunc) LookupEnv(key string) (string, bool) { return f(key) }

// OS is the main [Provider] that uses [os.LookupEnv].
var OS Provider = ProviderFunc(os.LookupEnv)

// Map is an in-memory [Provider] implementation useful in tests.
type Map map[string]string

// LookupEnv implements the [Provider] interface.
func (m Map) LookupEnv(key string) (string, bool) {
	value, ok := m[key]
	return value, ok
}

// MultiProvider combines multiple providers into a single one, which will
// contain the union of their environment variables. The order of the providers
// matters: if the same key exists in more than one provider, the value from
// the last one will be used.
func MultiProvider(ps ...Provider) Provider { return providers(ps) }

// providers wraps a slice of providers so it can be used as [Provider].
type providers []Provider

// LookupEnv implements the [Provider] interface.
func (ps providers) LookupEnv(key string) (string, bool) {
	var value string
	var found bool
	for _, p := range ps {
		if v, ok := p.LookupEnv(key); ok {
			value = v
			found = true
		}
	}
	return value, found
}
