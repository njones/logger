//go:generate go run gen/gen.go

/*
Package logger is a full featured level logger that can track logging across microservices infrastructure
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
	Info  LogLevel // `gen.short:"INF" gen.show:"Green" gen.alias:"Print" gen.stdlog.compat:"true,Print"`
	Warn  LogLevel // `gen.short:"WRN" gen.show:"Yellow"`
	Error LogLevel // `gen.short:"ERR" gen.show:"Red"`
	Debug LogLevel // `gen.short:"DBG" gen.show:"Cyan"`
	Trace LogLevel // `gen.short:"TRC" gen.show:"Blue"`
	Fatal LogLevel // `gen.short:"FAT" gen.show:"Red" gen.stdlog.compat:"true"`
	Panic LogLevel // `gen.short:"PAN" gen.show:"Red" gen.stdlog.compat:"true"`
} // `gen-level:"*"`

type (
	// LogLevel is the int representation of the log level
	LogLevel int8

	// LogColor is the int representation of the log color
	LogColor int32

	// LogStyle is the int representation of the log styling
	LogStyle int32
)

// The VT-100 escape sequence for different formatting options
// Note that the code in the comments are used in the gen.go file to generate the logger_generated.go file
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

// The color number of the escape sequence, the sequence is dynamiclly generated
// Note that the code in the comments are used in the gen.go file to generate the logger_generated.go file
const (
	Black   LogColor = iota + 30 // `gen.color:"\x1b[30m"`
	Red                          // `gen.color:"\x1b[31m"`
	Green                        // `gen.color:"\x1b[32m"`
	Yellow                       // `gen.color:"\x1b[33m"`
	Blue                         // `gen.color:"\x1b[34m"`
	Magenta                      // `gen.color:"\x1b[35m"`
	Cyan                         // `gen.color:"\x1b[36m"`
	White                        // `gen.color:"\x1b[37m"`

	NoESCColor = "no-color"
)

// the different formats that can be printed out
const (
	formatNone = "-f"
	formatHave = "+f"
	formatLine = "ln"
)

// a buffer pool for maps that will print key/value pairs
var kvMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

// a buffer pool for the slice to print interface values
var ifSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0)
	},
}

// the internal deadline for timeout to a io.Writer
var filteredWriteDeadline = 5 * time.Second

// OptFunc are functions that add options to the *logger struct
type OptFunc func(*logger)

// nilLogger is used by the logger to log to a black hole.
type nilLogger struct{}

// colorLogger is used to color a log line with a specific color
type colorLogger struct {
	*logger

	color string // the ESC string value
}

// logger is the struct that holds the main loging constructs
type logger struct {
	sync.RWMutex
	w io.Writer

	ts          *time.Time
	tsIsUTC     bool
	tsFormat    string
	tsFormatted string // preformatted

	hide   bool
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

func rtnLogLevelStr(pfx LogLevel) string {
	return pfx.StringWithColon()
}

// New returns a new logger that can be used for logging.
func New(options ...OptFunc) Logger {

	// default logger values
	l := &logger{
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		marshal:  StdKVMarshal,
		tsFormat: "",
		prefix:   rtnLogLevelStr,
		ctxkv:    make(map[string]interface{}),
		fatal:    func(i int) { os.Exit(i) },
	}

	// apply all of the options to this logger
	for _, opt := range options {
		opt(l)
	}

	// the defaut writer
	l.w = l.stdout

	// set up all of the output writers and filters
	if len(l.output) > 0 || len(l.filter) > 0 {
		l.w = io.MultiWriter(append(l.output, l.filter...)...)
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
	return &colorLogger{l, color.ToESCColor()}

}

// NoColor removes color from the output
func (l *logger) NoColor() Logger {
	return &colorLogger{l, NoESCColor}
}

// Field overrides the default color
func (l *logger) Field(key string, value interface{}) Logger {
	l.ctxkv[key] = value
	return l
}

// Fields overrides the default color with key value pairs.
// The KVMap function can be used to add values from a map
func (l *logger) Fields(kvs ...keyValue) Logger {
	for i := range kvs {
		l.ctxkv[kvs[i].K] = kvs[i].V
	}
	return l
}

// print the internal function that prints non-formatted logging
func (l *logger) print(prefix LogLevel, color string, iface ...interface{}) {
	l.printx(formatNone, prefix, color, "", iface...)
}

// printf the internal function that prints formatted logging
func (l *logger) printf(prefix LogLevel, color string, format string, iface ...interface{}) {
	l.printx(formatHave, prefix, color, format, iface...)
}

// println the internal function that prints logging with a newline
func (l *logger) println(prefix LogLevel, color string, iface ...interface{}) {
	l.printx(formatLine, prefix, color, "", iface...)
}

// printx is the internal function that prints the log line to the output writer(s)
func (l *logger) printx(kind string, pfx LogLevel, color string, format string, iface ...interface{}) {

	switch pfx {
	case Level().Fatal:
		l.fatal(int(pfx))
	}

	// check to see if we should be writing to the io.Writers for this logger
	if l.hide {
		return
	}

	var kv = kvMapPool.Get().(map[string]interface{})

	var ts time.Time
	if l.ts == nil {
		ts = time.Now()
	} else {
		ts = *l.ts
	}

	if l.tsIsUTC {
		ts = ts.UTC()
	}

	var tsFormatted = l.tsFormatted
	if tsFormatted == "" {
		switch l.tsFormat {
		case "":
			// the standard  format
			var n int
			var buf = [...]byte{'J', 'a', 'n', '-', '1', '0', '-', '1', '9', '7', '0', ' ', '2', '4', ':', '0', '0', ':', '0', '0'}

			// month-day-year
			copy(buf[0:], ts.Month().String()[:3])
			buf[3] = '-'
			itoa(&buf, ts.Day(), -1, 4)
			if ts.Day() > 9 {
				n += 1
			}
			buf[5+n] = '-'
			itoa(&buf, ts.Year(), 4, 6+n)
			buf[10+n] = ' '

			// time
			itoa(&buf, ts.Hour(), 2, 11+n)
			buf[13+n] = ':'
			itoa(&buf, ts.Minute(), 2, 14+n)
			buf[16+n] = ':'
			itoa(&buf, ts.Second(), 2, 17+n)

			tsFormatted = string(buf[:19+n]) // Jan-2-2006 15:04:05
		default:
			// this is much more memory and is slower than the std formatting
			tsFormatted = ts.Format(l.tsFormat)
		}
	}

	// pre-allocate the memory for all of the data needed in the slices
	inlnif := ifSlicePool.Get().([]interface{})
	switch {
	case color == NoESCColor:
		inlnif = append(inlnif, tsFormatted+" "+l.prefix(pfx))
	default:
		inlnif = append(inlnif, tsFormatted+" "+color+" "+l.prefix(pfx))
	}

	switch kind {
	case formatNone:
		inlnif = append(inlnif, " ")
	}

	defer func() {
		for k := range kv {
			delete(kv, k)
		}
		kvMapPool.Put(kv)

		for i := range inlnif {
			inlnif[i] = nil
		}
		inlnif = inlnif[:0]

		ifSlicePool.Put(inlnif)
	}()

	// add any http keys to the internal structured kv logging map
	for k, v := range l.httpkv {
		kv[k] = *v
	}

	// add any context keys to the internal structured kv logging map
	for k, v := range l.ctxkv {
		kv[k] = v
	}

	// filter out all of the structured kv logging stucts and add to the map
	var ifaceCutPoints []int
	for i, iv := range iface {
		if val, ok := iv.(keyValue); ok {
			kv[val.K] = val.V
			ifaceCutPoints = append(ifaceCutPoints, i)
			continue
		}
		inlnif = append(inlnif, iv)
	}

	// add color codes
	if color != "" && color != NoESCColor {
		inlnif = append(inlnif, ResetCode)
	}

	// add kv with marshaled structured logging
	if len(kv) > 0 {
		// marshal the structred logging to the key=value representation. (usually JSON or k=v, style)
		msh, err := l.marshal(kv)
		if err != nil {
			// logs error to stderr and dumps data, so you shouldn't lose any, but it won't be formatted correctly
			fmt.Fprintln(l.stderr, "error marshaling:", err)
			msh = []byte(fmt.Sprintf("[ERR logger.go (marshal)]: %#v", kv))
		}
		if len(msh) > 0 {
			inlnif = append(inlnif, string(msh))
		}
	}

	if pfx == Level().Panic {
		switch kind {
		case formatNone:
			panic(fmt.Sprint(inlnif...))
		case formatHave:
			panic(fmt.Sprintf("%s "+format+" %s\n", inlnif...))
		case formatLine:
			panic(fmt.Sprintln(inlnif...))
		}
	}

	// send to all of the out io.Writers wait until they have been written to
	// to move on, otherwise we'll over write some stuff...
	l.RWMutex.Lock()
	switch kind {
	case formatNone:
		if _, err := fmt.Fprint(l.w, inlnif...); err != nil {
			fmt.Fprintln(l.stderr, "error writing print to log:", err)
		}
	case formatHave:
		if _, err := fmt.Fprintf(l.w, "%s "+format+" %s\n", inlnif...); err != nil {
			fmt.Fprintln(l.stderr, "error writing printf to log:", err)
		}
	case formatLine:
		if _, err := fmt.Fprintln(l.w, inlnif...); err != nil {
			fmt.Fprintln(l.stderr, "error writing println to log:", err)
		}
	}
	l.RWMutex.Unlock()

	// grab the raw data that we're using
	var fString string
	if len(l.filter) > 0 {
		for j, k := range ifaceCutPoints {
			i := k - j
			iface = append(iface[:i], iface[i+1:]...)
		}
		if kind == formatHave {
			fString = fmt.Sprintf(format, iface...)
		} else {
			fString = fmt.Sprint(iface...)
		}
	}

	// flush the filtered writers
	for _, out := range l.filter {
		if fo, ok := out.(filterFlusher); ok {
			if fs, ok := out.(filterOn); ok {
				fs.On(fString)
			}
			fo.Flush()
		}
	}

	if color != NoESCColor {
		color = "" // remove color, so we can be ready for the next one.
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

// itoa is pulled from the go stdlib (log.go) and slightly modified to work with an array
// instead of a slice
// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[20]byte, i int, wid int, start int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	copy((*buf)[start:], b[bp:])
}
