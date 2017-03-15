//go:generate go run gen/gen.go

/*
package logger is a full featured level logger that can track logging across microservices infrastructure
*/
package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// level holds the different log levels that are supported
type level struct {
	Info  LogLevel // `gen.short:"INF", gen.show:"Green""`
	Warn  LogLevel // `gen.short:"WRN", gen.show:"Yellow"`
	Error LogLevel // `gen.short:"ERR", gen.show:"Red"`
	Debug LogLevel // `gen.short:"DBG", gen.show:"Cyan"`
	Trace LogLevel // `gen.short:"TRC", gen.show:"Blue"`
	Fatal LogLevel // `gen.short:"FAT", gen.show:"Red"`
} // `gen-level:"*"`

type (
	// LogLevel is the int representation of the log level
	LogLevel int8

	// LogColor is the int representation of the log color
	LogColor int32

	// LogStyle is the int representation of the log styling
	LogStyle int32
)

// Note that the code in the comments are used in the gen.go file to generate the loggers.go file
const (
	ResetCode           = "\x1b[0m"
	Reset      LogStyle = iota // `gen.style:"\x1b[0m"`
	Bright                     // `gen.style:"\x1b[1m"`
	_                          //
	Dim                        // `gen.style:"\x1b[2m"`
	Underscore                 // `gen.style:"\x1b[4m"`
	_                          //
	Blink                      // `gen.style:"\x1b[5m"`
	Reverse                    // `gen.style:"\x1b[7m"`
	Hidden                     // `gen.style:"\x1b[08m"`
)

const (
	Black   LogColor = iota + 30 // `gen.color:"\x1b[30m"`
	Red                          // `gen.color:"\x1b[31m"`
	Green                        // `gen.color:"\x1b[32m"`
	Yellow                       // `gen.color:"\x1b[33m"`
	Blue                         // `gen.color:"\x1b[34m"`
	Magenta                      // `gen.color:"\x1b[35m"`
	Cyan                         // `gen.color:"\x1b[36m"`
	White                        // `gen.color:"\x1b[37m"`
)

var filteredWriteDeadline = 5 * time.Second

// OptFunc are functions that add options to the *Logger struct
type OptFunc func(*logger)

// FilteredFunc is the a function that takes a string and determines if it should be logged or not
type FilteredFunc func(string) bool

// nilLogger is used by the logger to log to a black hole.
type nilLogger struct{}

// logger is the struct that holds the main loging constructs
type logger struct {
	w io.Writer
	l sync.Mutex

	ts       *time.Time
	tsIsUTC  bool
	tsFormat string

	hide   bool
	color  string // the ESC string value
	prefix func(LogLevel) string
	output []io.Writer
	filter []io.Writer

	httpkv  map[string]*string
	ctxkv   map[string]interface{}
	marshal func(v interface{}) ([]byte, error)

	stderr io.Writer
	stdout io.Writer
	err    error

	fatal func(int)
}

// New returns a new logger that can be used for logging.
func New(options ...OptFunc) Logger {

	// default logger values
	l := &logger{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		marshal:  StdKVMarshal,
		tsFormat: "Jan-2-2006 15:04:05",
		prefix: func(pfx LogLevel) string {
			return pfx.String() + ":"
		},
		ctxkv: make(map[string]interface{}),
		fatal: func(i int) { os.Exit(i) },
	}

	// apply all of the options to this logger
	for _, opt := range options {
		opt(l)
	}

	// set up all of the output writers and filters
	if len(l.output) > 0 || len(l.filter) > 0 {
		l.w = io.MultiWriter(append(l.output, l.filter...)...)
	} else {
		l.w = l.stdout
	}

	return l
}

// Suppress stops logging
func (l *logger) Suppress() {
	l.hide = true
}

// UnSuppress starts logging
func (l *logger) UnSuppress() {
	l.hide = false
}

// OnErr returns the internal logger if the error is not nil, otherwise it returns a nilLogger which won't do any logging. This is used for logging only if there is an error present.
func (l *logger) OnErr(err error) Logger {
	if err != nil {
		return l
	}
	return new(nilLogger)
}

// Color overrides the default color
func (l *logger) Color(color LogColor) Logger {
	l.color = color2ESC(color)
	return l
}

// Field overrides the default color
func (l *logger) Field(key string, value interface{}) Logger {
	l.ctxkv[key] = value
	return l
}

// Field overrides the default color
func (l *logger) Fields(kvs ...keyValue) Logger {
	for i := range kvs {
		l.ctxkv[kvs[i].K] = kvs[i].V
	}
	return l
}

func (l *logger) println(prefix LogLevel, iface ...interface{}) {
	l.print("ln", prefix, "", iface...)
}

// printf the internal function that prints formatted logging
func (l *logger) printf(prefix LogLevel, format string, iface ...interface{}) {
	l.print("f", prefix, format, iface...)
}

// print is the internal function that prints the log line to the output writer(s)
func (l *logger) print(kind string, pfx LogLevel, format string, iface ...interface{}) {
	prefix := l.prefix(pfx)
	kv := make(map[string]interface{})

	// add any http keys to the internal structured kv logging map
	for k, v := range l.httpkv {
		kv[k] = *v
	}

	// add any context keys to the internal structured kv logging map
	for k, v := range l.ctxkv {
		kv[k] = v
	}

	// filter out all of the structured kv logging stucts and add to the map
	for i := 0; i < len(iface); i++ {
		n := iface[i]
		if val, ok := n.(keyValue); ok {
			kv[val.K] = val.V
			iface = append(iface[:i], iface[i+1:]...)
			i--
		}
	}

	// check to see if we should be writing to the io.Writers for this logger
	if !l.hide {

		var ts time.Time
		if l.ts == nil {
			ts = time.Now()
		} else {
			ts = *l.ts
		}

		if l.tsIsUTC {
			ts = ts.UTC()
		}

		// marshal the structred logging to the key=value representation. (usually JSON or k=v, style)
		msh, err := l.marshal(kv)
		if err != nil {
			// logs error to stderr and dumps data, so you shouldn't lose any, but it won't be formatted correctly
			fmt.Fprintln(l.stderr, "error marshaling:", err)
			msh = []byte(fmt.Sprintf("[ERR logger.go (marshal)]: %#v", kv))
		}

		// add color codes
		if l.color != "" {
			iface = append(iface, ResetCode)
		}

		// add kv with marshaled structured logging
		if len(msh) > 0 {
			iface = append(iface, string(msh))
		}

		// send to all of the out io.Writers
		switch kind {
		case "ln":
			if _, err := fmt.Fprintln(l.w, append([]interface{}{ts.Format(l.tsFormat), l.color, prefix}, iface...)...); err != nil {
				fmt.Fprintln(l.stderr, "error writting to log:", err)
			}
		case "f":
			if _, err := fmt.Fprintf(l.w, "%s %s %s "+format+" %s\n", append([]interface{}{ts.Format(l.tsFormat), l.color, prefix}, iface...)...); err != nil {
				fmt.Fprintln(l.stderr, "error writting to formatted log:", err)
			}
		}
		// close out the filtered writers
		for _, out := range l.filter {
			if fo, ok := out.(Filter); ok {
				fo.Flush()
			}
		}

		l.color = "" // remove color, so we can be ready for the next one.
	}

	if pfx == Level().Fatal {
		l.fatal(int(pfx))
	}
}

// HTTPMiddleware returns the standard HTTP handler middleware function that will capture headers for logging.
func (l *logger) HTTPMiddleware(next http.Handler) http.Handler {
	// lazy initialize
	if l.httpkv == nil {
		l.httpkv = make(map[string]*string)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k := range l.httpkv {
			v := r.Header.Get(k)
			l.httpkv[k] = &v
		}
		next.ServeHTTP(w, r)
	})
}
