package logger

import (
	"io"
	"net/http"
)

type NoColorWriter interface {
	NoColor()
}

type printKind int

const (
	bPrint printKind = iota
	bPrintf
	bPrintln
)

type optFunc func(*baseLogger)

// Filter defines the intterface for checking if a log line should be
// logged based on a true value from the filter Check(string) function
type Filter interface {
	Check(string) bool
}

// Logger the interface that defines the logger package functions
type Logger interface {
	ExtendedLogger
	HTTPLogger

	OnErr(error) OnErrLogger

	FatalInt(int) Logger
	Field(string, interface{}) Logger
	Fields(map[string]interface{}) Logger
	Suppress(logLevel) Logger
	With(...optFunc) Logger
}

// StandardOptions the interface that matches the std library log package
// functions outside of Fatal(x), Print(x) and Panic(x).
type StandardOptions interface {
	Flags() int
	Output(calldepth int, s string) error
	Prefix() string
	SetFlags(flag int)
	SetOutput(w io.Writer)
	SetPrefix(prefix string)
	Writer() io.Writer
}

// StandardLogger the interface that matches the std library log package
type StandardLogger interface {
	StandardOptions

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// ExtendedLogger the interface that defines the extra log levels outside of
// the std log package logging.
type ExtendedLogger interface {
	StandardLogger

	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Traceln(v ...interface{})
}

// HTTPLogger the interface that defines HTTP logging that
// can be inserted as middleware in HTTP routers
type HTTPLogger interface {
	HTTPMiddleware(http.Handler) http.Handler
}

type OnErrLogger interface {
	StandardOptions

	// from the standard logger
	Print(v ...interface{}) Return
	Printf(format string, v ...interface{}) Return
	Println(v ...interface{}) Return
	Fatal(v ...interface{}) Return
	Fatalf(format string, v ...interface{}) Return
	Fatalln(v ...interface{}) Return
	Panic(v ...interface{}) Return
	Panicf(format string, v ...interface{}) Return
	Panicln(v ...interface{}) Return

	// from the extended logger
	Info(v ...interface{}) Return
	Infof(format string, v ...interface{}) Return
	Infoln(v ...interface{}) Return
	Warn(v ...interface{}) Return
	Warnf(format string, v ...interface{}) Return
	Warnln(v ...interface{}) Return
	Error(v ...interface{}) Return
	Errorf(format string, v ...interface{}) Return
	Errorln(v ...interface{}) Return
	Debug(v ...interface{}) Return
	Debugf(format string, v ...interface{}) Return
	Debugln(v ...interface{}) Return
	Trace(v ...interface{}) Return
	Tracef(format string, v ...interface{}) Return
	Traceln(v ...interface{}) Return
}

type Errer interface{ Err() error }

// The Key-Value types
type (
	K string
	V interface{}

	KVMap map[K]V

	KeyVal struct {
		Key   K `json:"key"`
		Value V `json:"value"`
	}
)
