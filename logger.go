package logger

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	isatty "github.com/mattn/go-isatty"
)

type (
	logLevel   uint16
	colorType  uint8
	printType  uint8
	levelType  uint8
	formatType uint8
)

// HTTPLogFormatFunc is the type that is used for logging HTTP requests
type HTTPLogFormatFunc func(int, int64, string, *http.Request) string

// ESCStringer is an interface for values that write escape codes.
type ESCStringer interface {
	ESCStr() string
}

// Return is the return of the Logger interface methods that don't return Logger
type Return struct{ HasErr bool }

// The logLevel int representations of the different logging levels and related
// information that can be programmatically generated for colors and display values.
const (
	LevelPrint logLevel = 1 << iota //`short:"INF" long:"Info" color:"green"`
	LevelInfo                       //`short:"INF" color:"green"`
	LevelWarn                       //`short:"WRN" color:"yellow"`
	LevelDebug                      //`short:"DBG" color:"cyan"`
	LevelError                      //`short:"ERR" color:"magenta"`
	LevelTrace                      //`short:"TRC" color:"blue"`
	LevelFatal                      //`short:"FAT" color:"red"`
	LevelPanic                      //`short:"PAN" color:"red"`
	LevelHTTP                       //`short:"-" long:"-" color:"green" fn:"ln"`
)

// The levelType int repressentations which determine how the log level will be displayed.
const (
	LevelLongStr levelType = iota
	LevelShortStr
	LevelShortBracketStr
)

// The printType constants which deterimine how the line will be output to the
// writer.
const (
	asPrint printType = iota
	asPrintf
	asPrintln
)

// Escape codes integer values for formatting and colors - the escape sequences
// are auto-generated.
const (
	SeqUnk   formatType = 255
	SeqReset formatType = iota
	SeqBright
	SeqDim
	_
	SeqUnderscore
	SeqBlink
	_
	SeqReverse
	SeqHidden

	ColorUnk   colorType = 255
	ColorBlack colorType = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

const (
	// The hard-coded deadline for writes to a network connection.
	filteredWriteDeadline = 5 * time.Second
)

// context is a struct that represents all of the log line data.
type context struct {
	is     printType
	colors [3]ESCStringer

	tsStrCh chan string

	level     logLevel
	levelStr  string
	formatStr string
	values    []interface{}
	kvMap     map[string]interface{}

	panicCh chan string
	wg      *sync.WaitGroup
}

// nilLogger is a logger that does nothing...
type nilLogger struct{}

// baseLogger is base logger type.
type baseLogger struct {
	o []io.Writer // multiple outputs

	stdout io.Writer
	stderr io.Writer

	to chan context

	ts       *time.Time
	tsIsUTC  bool
	tsText   string
	tsFormat string
	logLevel levelType

	skip      logLevel
	hasFilter bool
	kv        map[string]interface{}
	color     colorType

	httpk         []string
	httpLogFormat HTTPLogFormatFunc

	marshal func(interface{}) ([]byte, error)
	fatal   func(i int)
	fatali  int

	hasErr bool
}

// KVStruct is a public struct that represents key value pairs, which can
// be added to a log line and marshaled to different data structures (i.e.
// JSON, XML, etc) for structured logging.
type KVStruct struct {
	key   string
	value interface{}
}

// optFunc represents an option to change the defaults of the
// baseLogger, this is not exported because we don't expose changing
// options to the end user.
type optFunc func(*baseLogger)

// New returns a Logger interface that exposes the same as the standard logger
// along with level logging and other options.
func New(opts ...optFunc) Logger {
	l := new(baseLogger)
	l.o = make([]io.Writer, 0, 5)
	l.tsFormat = "std"
	l.stdout = os.Stdout
	l.stderr = os.Stderr
	l.logLevel = LevelLongStr
	l.marshal = StdKVMarshal
	l.color = ColorUnk - 1
	l.fatal = func(i int) { os.Exit(i) } // so we can test fatal
	l.fatali = 1

	// apply defaults
	for _, opt := range opts {
		opt(l)
	}

	// update the output writers
	if len(l.o) > 0 {
		l.stdout = io.MultiWriter(l.o...)
	}

	// create the channel that will write the logs to the output io.Writers
	if l.to == nil {
		l.to = make(chan context, 1000)
	}

	// the go routine that manages writing out the log state (one line at a time)
	go l.out()

	return l
}

func (l *baseLogger) rtn() Return { return Return{HasErr: l.hasErr} }

func (l *baseLogger) out() {
	var err error

	buf := new(bytes.Buffer)
	var lenPoint1, lenPoint2 int
	for logg := range l.to {

		// marshal and format KV data
		var kv []byte
		if l.marshal != nil && logg.kvMap != nil && len(logg.kvMap) > 0 {
			kv, err = l.marshal(logg.kvMap)
			if err != nil {
				fmt.Fprintf(l.stderr, "error marshaling: %v\n", err)
				kv = []byte(fmt.Sprintf("[ERR logger.go (marshal)]: %#v", logg.kvMap))
			}
			kv = append([]byte{' '}, kv...)
		}

		// add color info
		colorIdx := 0
		if logg.colors[1] != ColorUnk-1 {
			colorIdx = 1
		}

		// write the time slug and log level
		buf.WriteString(<-logg.tsStrCh)
		tsChanPool.Put(logg.tsStrCh)

		buf.WriteString(string(logg.colors[colorIdx].ESCStr()))
		buf.WriteString(logg.levelStr)
		lenPoint1 = buf.Len()

		// write the log line
		switch logg.is {
		case asPrint:
			buf.WriteString(fmt.Sprint(logg.values...))
		case asPrintf:
			buf.WriteString(fmt.Sprintf(logg.formatStr, logg.values...))
		case asPrintln:
			// spacing was added manually...
			buf.WriteString(fmt.Sprint(logg.values...))
		}
		lenPoint2 = buf.Len()

		// add color information
		buf.WriteString(string(logg.colors[2].ESCStr()))

		// add marshaled KV data
		buf.Write(kv)

		// panic if it's the correct log type
		if logg.level == LevelPanic {
			logg.panicCh <- buf.String()
		}

		// always end on a newline
		buf.WriteByte([]byte("\n")[0])

		// write the log line to all io.Writers
		if _, err := fmt.Fprint(l.stdout, buf.String()); err != nil {
			fmt.Fprintln(l.stderr, "error writing to log:", err)
		}

		// if there are filters then call the Callback method to write
		// data to a filtered io.Writer
		if l.hasFilter {
			logg.wg.Add(1)
			go func(logln string, p1, p2 int) {
				for _, w := range l.o {
					if fw, ok := w.(filterwriter); ok {
						fw.Callback(logln, p1, p2)
					}
				}
				logg.wg.Done()
			}(buf.String(), lenPoint1, lenPoint2)
		}

		// reset things and move on!
		buf.Reset()
		logg.wg.Done()
	}
}

// Color changes the color escape codes of a single log line. Use the colorType constants as the
// method paramteter. Save to a new log variable to keep the color options across log calls.
func (l *baseLogger) Color(color colorType) Logger {
	// return a copy of the base logger, but with the color filled in
	cl := copyBaseLogger(l)
	cl.color = color
	return cl
}

func (l *baseLogger) FatalInt(i int) Logger {
	// return a copy of the base logger, but with the fatal int filled in
	fil := copyBaseLogger(l)
	fil.fatali = i
	return fil
}

// Field adds a single key, value pair to a single log line.
func (l *baseLogger) Field(key string, value interface{}) Logger {
	if l.kv != nil {
		l.kv[key] = value // if we already have this filled out, then just add
		return l
	}
	kvl := copyBaseLogger(l)
	kvl.kv = map[string]interface{}{key: value}
	return kvl
}

// Fields adds a key, value map to a single log line.
func (l *baseLogger) Fields(kv map[string]interface{}) Logger {
	if l.kv != nil {
		for k, v := range kv {
			l.kv[k] = v
		}
		return l
	}
	kvl := copyBaseLogger(l)
	kvl.kv = kv
	return kvl
}

// NoColor is a convenience method that removes the color escape codes from a log line.
func (l *baseLogger) NoColor() Logger { return l.Color(ColorUnk) }

// OnErr displays the log line only if the error err is not nil.
func (l *baseLogger) OnErr(err error) Logger {
	if err == nil {
		return nilLogger{}
	}
	l.hasErr = true
	return l
}

// Suppress takes a bitwise OR (|) of the different levels to suppress
// to Unsupress set to 0,
func (l *baseLogger) Suppress(loglevels logLevel) Logger {
	sl := copyBaseLogger(l)
	sl.skip = loglevels
	return sl
}

// With adds options to a logger and returns a new logger with those options
func (l *baseLogger) With(options ...optFunc) Logger {
	wl := copyBaseLogger(l)
	for _, opt := range options {
		opt(wl)
	}
	return wl
}

// tsChanPool is a sync pool for sending formatted time back
// formatting the time takes a while, so try to concurrently
// format to help things look faster
var tsChanPool = sync.Pool{
	New: func() interface{} {
		return make(chan string, 1)
	},
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

//tsChantsChanEmpty is a function that returns the passed in text string as the time
func tsChanText(text string) chan string {
	ch := tsChanPool.Get().(chan string)
	go func() { ch <- text }()
	return ch
}

// tsChan is a function that returns a formated current timestamp, unless
// the format is overwritten, or a standard time is used
func tsChan(text string, format string, tss *time.Time, isUTC bool) chan string {
	ch := tsChanPool.Get().(chan string)

	go func() {
		// overwrite the formatting to be static text
		if len(text) > 0 {
			ch <- text + " "
			return
		}

		// check to see if the current time is overwritten
		var ts time.Time
		if tss != nil {
			ts = *tss
		} else {
			ts = time.Now()
		}

		if isUTC {
			ts = ts.UTC()
		}

		switch format {
		case "std":
			// the standard  format
			var n int
			var buf = [...]byte{'J', 'a', 'n', '-', '1', '0', '-', '1', '9', '7', '0', ' ', '2', '4', ':', '0', '0', ':', '0', '0'}

			// month-day-year
			copy(buf[0:], ts.Month().String()[:3])
			buf[3] = '-'
			itoa(&buf, ts.Day(), -1, 4)
			if ts.Day() > 9 {
				n++
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

			ch <- string(buf[:19+n]) + " " // Jan-2-2006 15:04:05
		default:
			// this is much more memory and is slower than the std formatting
			ch <- ts.Format(format) + " "
		}
	}()
	return ch
}

// KV is a convenience function for returning a KV struct, if a log line contains this value,
// then it is pulled out of the parameter list and structured.
func KV(key string, value interface{}) (kv KVStruct) {
	kv.key = key
	kv.value = value
	return
}

// StdKVMarshal is the structured logging which looks like key=value, Values can
// be displayed for a string, int, float or GoString.
func StdKVMarshal(in interface{}) ([]byte, error) {
	var rtns []string
	switch val := in.(type) {
	case map[string]interface{}:
		for k, v := range val {
			switch vt := v.(type) {
			case string:
				rtns = append(rtns, fmt.Sprintf("%s=%s", k, vt))
			case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64,
				float64:
				rtns = append(rtns, fmt.Sprintf("%s=%d", k, vt))
			default:
				rtns = append(rtns, fmt.Sprintf("%s=%v", k, vt))
			}
		}
	}

	sort.Strings(rtns)
	return []byte(strings.Join(rtns, " ")), nil
}

// WithPrefix changes the prefix of the log level to be different.
func WithPrefix(lt levelType) optFunc {
	return func(l *baseLogger) {
		l.logLevel = lt
	}
}

// WithTimeText overwrites the current timestamp slug with a static string.
func WithTimeText(text string) optFunc {
	return func(l *baseLogger) {
		l.tsText = text
	}
}

// WithTimeAsUTC makes sure that the time stamp is in UTC time.
func WithTimeAsUTC() optFunc {
	return func(l *baseLogger) {
		l.tsIsUTC = true
	}
}

// WithTimeFormat changes the format of the log line time to the format, using standard Go
// date formating rules.
func WithTimeFormat(format string) optFunc {
	return func(l *baseLogger) {
		l.tsFormat = format
	}
}

// WriterToFileDescriptor is an interface to define a TTY interface
type WriterToFileDescriptor interface {
	Fd() uintptr
}

// WithOutput adds a writer to be output, it overwrites sdtout the first time used, but then
// appends after.
func WithOutput(w io.Writer) optFunc {
	return func(l *baseLogger) {
		if ofd, ok := w.(WriterToFileDescriptor); ok {
			if !isatty.IsTerminal(ofd.Fd()) && !isatty.IsCygwinTerminal(ofd.Fd()) {
				w = StripWriter(w)
			}
		}
		l.o = append(l.o, w)
	}
}

// WithHTTPHeader sets a HTTP header to be captured from the HTTPMiddleware handler
func WithHTTPHeader(header ...string) optFunc {
	return func(l *baseLogger) {
		l.httpk = append(l.httpk, header...)
	}
}

// WithKVMarshaler uses the standard marshal interface to format the structured logging values,
// it works with standard JSON and XML marshalers.
func WithKVMarshaler(fn func(interface{}) ([]byte, error)) optFunc {
	return func(l *baseLogger) {
		l.marshal = fn
	}
}

// Filter is a type that can be used to filter log lines.
type Filter interface {
	Check(string) bool
}

// WithFilterOutput adds a writer that will be filtered with the added Filter functions, in order until one is satisfied.
func WithFilterOutput(w io.Writer, filters ...Filter) optFunc {
	return func(l *baseLogger) {
		l.hasFilter = true
		l.o = append(l.o, filterwriter{
			w:       w,
			filters: filters,
		})
	}
}

// NotFilter negates the check on the wrapped filter
func NotFilter(filter Filter) Filter {
	return &filterNot{f: filter}
}

// filterNot is a type to define a function that accepts a filter and negates the check.
type filterNot struct{ f Filter }

// Check satisfies the Filter interface and runs the passed in function.
func (n *filterNot) Check(data string) bool { return !n.f.Check(data) }

// StringFuncFilter is a filter function that takes the function func(string)bool with any returned true
// value, filtering out the log line, so it will not display.
func StringFuncFilter(fn func(string) bool) Filter {
	return &filterStrFunc{fn: fn}
}

// filterStrFunc is a type to define a function that accepts a string to filter log lines.
type filterStrFunc struct{ fn func(string) bool }

// Check satisfies the Filter interface and runs the passed in function.
func (sf *filterStrFunc) Check(data string) bool { return sf.fn(data) }

// RegexFilter is a filter function that takes a regular expresson and if it is matched by the logline
// then that line is filtered out
func RegexFilter(pattern string) Filter {
	return &filterRegex{regexp: regexp.MustCompile(pattern)}
}

// filterRegex is a type to define a function that accepts a regualar expression to filter log lines.
type filterRegex struct{ regexp *regexp.Regexp }

// Check satisfies the Filter interface and matches against a regular expression
func (r *filterRegex) Check(data string) bool { return r.regexp.MatchString(data) }

// filterwriter the underling struct that will filter input to the supplied witer. This happens
// because writes come through the callback and not through the io.Writer interface.
type filterwriter struct {
	filters []Filter
	w       io.Writer
}

// Write doesn't propgate Filter writers to the downstream writer, that's done through
// the Filter Callback method.
func (filterwriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Callback is the method called by the logger, and will filter out log lines
// based on some criteria.
func (fw filterwriter) Callback(logln string, pre, line int) {
	data := logln[pre:line]
	for _, filter := range fw.filters {
		if !filter.Check(data) {
			fw.w.Write([]byte(logln))
			return
		}
	}
}

// NetWriter is a helper function that will log writes to a TCP/UDP address. Any errors will be written to stderr.
func NetWriter(network, address string) io.Writer {
	return netwriter{network: network, address: address}
}

// netwriter the underling struct that will write to the connection
type netwriter struct {
	network, address string
}

// Write passes writes to the network connection from a io.Writer
func (nw netwriter) Write(p []byte) (int, error) {
	conn, err := net.Dial(nw.network, nw.address)
	if err != nil {
		return 0, err
	}
	go func() {
		conn.SetWriteDeadline(<-time.After(filteredWriteDeadline))
	}()
	defer conn.Close()

	return conn.Write(p)
}

// StripWriter wraps a writer that will strip out some VT100 escape codes
func StripWriter(w io.Writer) io.Writer {
	return stripescwriter{w: w}
}

// stripescwriter the underling struct that will strip out escape characters
type stripescwriter struct {
	state func(*stripescwriter, byte) bool
	w     io.Writer
}

// StBracket returns the next states after a bracket is found
func (se *stripescwriter) StBracket(b byte) bool {
	switch b {
	case '[':
		se.state = (*stripescwriter).StCode
		return false
	}
	se.state = nil
	return false
}

// StBracket returns the next states after a number or 'm' is found
func (se *stripescwriter) StCode(b byte) bool {
	switch b {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return false
	case 'm':
		se.state = (*stripescwriter).StDone
		return false
	}
	se.state = nil
	return false
}

// StDone stops the machine and returns that its completed successfully
func (se *stripescwriter) StDone(b byte) bool {
	se.state = nil
	return true
}

// Write looks for the escape character and strips out any codes via a simple state machine.
// This may flush to the underlining writer more than once.
func (se stripescwriter) Write(p []byte) (n int, err error) {
	var start int
	for i, b := range p {
		if se.state == nil && b == 0x1b {
			se.state = (*stripescwriter).StBracket
			n2, err2 := se.Write(p[start:i])
			n, err = n+n2, err2
			continue
		}
		if se.state == nil {
			continue
		}
		drop := se.state(&se, b)
		if se.state == nil && drop {
			start = i
			continue
		}
	}
	if se.state == nil {
		n2, err2 := se.w.Write(p[start:])
		n, err = n+n2, err2
	}
	return n, err
}

// ResponseWriter holds an embedded HTTP ResponseWriter but will capture the status
// and number of bytes sent so they can be logged.
type ResponseWriter struct {
	http.ResponseWriter
	status int
	sent   int64
}

// Write writes to the underlining write, while counting the number of bytes that pass through
func (c *ResponseWriter) Write(p []byte) (n int, err error) {
	if c.status == 0 {
		c.WriteHeader(http.StatusOK) // is so that it acts like the http.ResponseWriter Write([]byte): https://golang.org/pkg/net/http/#ResponseWriter
	}
	n, err = c.ResponseWriter.Write(p)
	c.sent += int64(n)
	return
}

// WriteHeader captures the status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
func (c *ResponseWriter) WriteHeader(code int) {
	c.status = code
	c.ResponseWriter.WriteHeader(code)
}

// HTTPMiddleware is a middleware handler that will log HTTP server requests
func (l *baseLogger) HTTPMiddleware(next http.Handler) http.Handler {
	// set the default http logger if it's nil
	if l.httpLogFormat == nil {
		l.httpLogFormat = CommonLogFormat
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &ResponseWriter{ResponseWriter: w}
		next.ServeHTTP(cw, r)

		if len(l.httpk) > 0 {
			httpkv := make(map[string]interface{})
			for _, k := range l.httpk {
				httpkv[k] = r.Header.Get(k)
			}
			l.Fields(httpkv).HTTPln(l.httpLogFormat(cw.status, cw.sent, l.tsText, r))
			return
		}

		l.HTTPln(l.httpLogFormat(cw.status, cw.sent, l.tsText, r))
	})
}

// CommonLogFormat is the Apache Common Logging format used for logging HTTP requests
func CommonLogFormat(status int, sent int64, tsText string, r *http.Request) string {
	// $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" | nginx
	// 127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326

	if len(tsText) == 0 {
		tsText = time.Now().Format("02/Jan/2006:15:04:05 -0700")
	}

	commonLog := struct {
		RemoteAddr    string `json:"remote_address"`
		RemoteId      string `json:"remote_identifier"`
		RemoteUser    string `json:"remote_user"`
		LocalTime     string `json:"time_local"`
		RequestString string `json:"request"`
		Status        int    `json:"status"`
		BytesSent     int64  `json:"body_bytes_sent"`
		Referer       string `json:"http_referer"`
		UserAgent     string `json:"http_user_agent"`
	}{
		RemoteAddr:    r.RemoteAddr,
		RemoteId:      "-",
		RemoteUser:    "-",
		LocalTime:     tsText,
		RequestString: fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto),
		Status:        status,
		BytesSent:     sent,
		Referer:       r.Referer(),
		UserAgent:     r.UserAgent(),
	}

	return fmt.Sprintf("%s %s %s [%s] \"%s\" %d %d %s %s", commonLog.RemoteAddr,
		commonLog.RemoteId, commonLog.RemoteUser,
		commonLog.LocalTime, commonLog.RequestString,
		commonLog.Status, commonLog.BytesSent,
		commonLog.Referer, commonLog.UserAgent)
}
