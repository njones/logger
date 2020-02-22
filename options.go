package logger

import (
	"io"
	"os"

	"github.com/njones/logger/color"
)

// All of the flags from the std log pkg
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// WithColor adds color to the logged output (overriding any level colors)
func WithColor(v color.Foreground) optFunc {
	var escColor = []byte(v.ToESC())
	return func(b *baseLogger) {
		b.color = escColor
	}
}

// WithHTTPHeader takes headers and addes them as a structured K/V pair to the logged output
func WithHTTPHeader(headers ...string) optFunc {
	return func(b *baseLogger) {
		b.http.headers = make(KVMap)
		for _, header := range headers {
			if _, ok := b.http.headers[K(header)]; !ok {
				b.http.headers[K(header)] = ""
			}
		}
	}
}

// WithKVMarshaler takes a encoding.Marshaler interface and uses it to marshal kv values
func WithKVMarshaler(fn func(interface{}) ([]byte, error)) optFunc {
	return func(b *baseLogger) {
		b.kv.marshal = fn
	}
}

// WithOutput adds the ws writers to the logged output. This can be overridden by using
// the logger.Output function. A nil or empty ws []io.Writer uses os.Stdout as the writer
func WithOutput(ws ...io.Writer) optFunc {
	if len(ws) == 0 {
		ws = []io.Writer{os.Stdout}
	}
	return func(b *baseLogger) {
		b.writers(ws)
	}
}

// WithTimeFormat formats the time according to the format value. Once the time has been
// formatted, it can apply any func(string)string function to the format, this can be
// used to uppercase the string (i.e. strings.ToUpper) for example. A empty value removes
// the time from the log output
func WithTimeFormat(format string, fns ...func(string) string) optFunc {
	return func(b *baseLogger) {
		b.ts.fns = fns
		b.ts.stamp = convertStamp(format)
	}
}

// WithTimeText replaces the timestamp value with the provided text, a empty value does
// not replace any timestamp formatted value
func WithTimeText(text string) optFunc {
	return func(b *baseLogger) {
		b.ts.text = []byte(text)
	}
}
