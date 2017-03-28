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

// WithFilteredStringOutput adds a new writer that is filtered by the function presented. This returns a func which takes the same input as the WithOutput function.
func WithFilteredStringOutput(fn FilteredStringFunc, w io.Writer) OptFunc {
	return func(l *logger) {
		l.filter = append(l.filter, &filteredStringWriter{w: w, fn: fn})
	}
}

// WithFilteredByteOutput adds a new writer that is filtered by the function presented. This returns a func which takes the same input as the WithOutput function.
func WithFilteredByteOutput(fn FilteredByteFunc, w io.Writer) OptFunc {
	return func(l *logger) {
		l.filter = append(l.filter, &filteredByteWriter{w: w, fn: fn})
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

// WithNoColorOutput adds a new writer that is strips the color from the output.
func WithNoColorOutput(w io.Writer) OptFunc {
	return WithFilteredByteOutput(noColorOutputFunc, w)
}

const x0len = len("\x1b[0m")
const x3len = len("\x1b[31m")

func noColorOutputFunc(fw *filteredByteWriter, p []byte) (r []byte) {
	if fw.b.Len() > 0 {
		p = append(fw.b.Bytes(), p...)
		fw.b.Reset()
	}

	for i := 0; i < len(p); i++ {
		if fw.chk {
			fw.chk = false
			if p[i] == ' ' {
				continue
			}
		}

		if p[i] == '\x1b' {
			if len(p[i:]) >= x3len {
				switch string(p[i : i+x3len]) {
				case "\x1b[31m", "\x1b[32m",
					"\x1b[33m", "\x1b[34m",
					"\x1b[35m", "\x1b[36m":
					i += (x3len - 1)
					fw.chk = true
					continue
				}
			}
			if len(p[i:]) >= x0len {
				switch string(p[i : i+x0len]) {
				case "\x1b[0m":
					i += (x0len - 1)
					fw.chk = true
					continue
				}
			}
			if len(p[i:]) < x0len || len(p[i:]) < x3len {
				fw.b.Write(p[i:])
				return
			}
		}
		r = append(r, p[i])
	}
	return r
}
