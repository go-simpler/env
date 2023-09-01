package env

// Option allows to configure the behaviour of the [Load] function.
type Option func(*loader)

// WithSource configures [Load] to retrieve environment variables from the provided [Source].
// If multiple sources are provided, they will be merged into a single one containing the union of all environment variables.
// The order of the sources matters: if the same key occurs more than once, the later value takes precedence.
// The default source is [OS].
func WithSource(src Source, srcs ...Source) Option {
	// reverse the slice first, since the later value should take precedence.
	for i, j := 0, len(srcs)-1; i < j; i, j = i+1, j-1 {
		srcs[i], srcs[j] = srcs[j], srcs[i]
	}
	return func(l *loader) { l.source = multiSource(append(srcs, src)) }
}

// WithPrefix configures [Load] to automatically add the provided prefix to each environment variable.
// By default, no prefix is configured.
func WithPrefix(prefix string) Option {
	return func(l *loader) { l.prefix = prefix }
}

// WithSliceSeparator configures [Load] to use the provided separator when parsing slice values.
// The default separator is space.
func WithSliceSeparator(sep string) Option {
	return func(l *loader) { l.sliceSep = sep }
}
