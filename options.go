package logger

import (
	"io"
)

// WithOutput adds a new writer to the logging output
func WithOutput(w io.Writer) OptFunc {
	return func(l *logger) {
		l.output = append(l.output, w)
	}
}

// WithFilteredOutput adds a new writer that is filtered by the function presented. This returns a func which takes the same input as the WithOutput function.
func WithFilteredOutput(fn FilteredFunc, w io.Writer) OptFunc {
	return func(l *logger) {
		l.filter = append(l.filter, &filteredWriter{w: w, fn: fn})
	}
}

// WithHTTPHeader adds a HTTP header to be captured for the structured output
func WithHTTPHeader(header string) OptFunc {
	return func(l *logger) {
		// lazy initialize
		if l.httpkv == nil {
			l.httpkv = make(map[string]*string)
		}
		l.httpkv[header] = (*string)(nil)
	}
}

// WithKVMarshaler allows a custom marshaler for the structured logging. This follows the Marshaler standard
func WithKVMarshaler(fn func(interface{}) ([]byte, error)) OptFunc {
	return func(l *logger) {
		l.marshal = fn
	}
}

// WithTimeAsUTC sets the logging time to be UTC, otherwise it is the same as the OS timezone
func WithTimeAsUTC(l *logger) {
	l.tsIsUTC = true
}

// WithShortPrefix uses the three letter prefix
func WithShortPrefix(l *logger) {
	l.prefix = func(pfx LogLevel) string {
		return "[" + pfx.Short() + "]"
	}
}

// WithTimeFormat sets the logging time to be formatted using the Go time formatting options
func WithTimeFormat(fmt string) OptFunc {
	return func(l *logger) {
		l.tsFormat = fmt
	}
}
