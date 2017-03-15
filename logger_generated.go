// go generate
// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT THIS FILE
// ~~ This file is not generated by hand ~~
// ~~ generated on: 2017-03-15 00:19:00.457335162 +0000 UTC ~~

package logger

import (
	"fmt"
	"net/http"
)

// String is the string representation of the color
func (lc LogColor) String() string {
	switch lc {
	case Black:
		return "Black"
	case Blue:
		return "Blue"
	case Cyan:
		return "Cyan"
	case Green:
		return "Green"
	case Magenta:
		return "Magenta"
	case Red:
		return "Red"
	case White:
		return "White"
	case Yellow:
		return "Yellow"
	}

	return "unknown"
}

// color2ESC returns the VT100 escape codes for a color
func color2ESC(color LogColor) string {
	return fmt.Sprintf("\x1b[%dm", int32(color))
}

// Level returns the log level used
func Level() (lvl level) {
	lvl.Debug = 8
	lvl.Error = 4
	lvl.Fatal = 32
	lvl.Info = 1
	lvl.Trace = 16
	lvl.Warn = 2
	return lvl
}

// String is the string representation of the log level
func (ll LogLevel) String() string {
	switch ll {
	case 8:
		return "Debug"
	case 4:
		return "Error"
	case 32:
		return "Fatal"
	case 1:
		return "Info"
	case 16:
		return "Trace"
	case 2:
		return "Warn"
	}

	return "unknown"
}

// Short is the short three letter abbreviation of the log level
func (ll LogLevel) Short() string {
	switch ll {
	case 8:
		return "DBG"
	case 4:
		return "ERR"
	case 32:
		return "FAT"
	case 1:
		return "INF"
	case 16:
		return "TRC"
	case 2:
		return "WRN"
	}

	return "unknown"
}

// Logger is the main interface that is presented as a logger
type Logger interface {
	Color(LogColor) Logger
	OnErr(error) Logger
	HTTPMiddleware(next http.Handler) http.Handler
	Suppress()
	UnSuppress()

	Debug(...interface{})
	Debugf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Trace(...interface{})
	Tracef(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
}

// Debug is the generated logger function to satisfy the interface
func (l *logger) Debug(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Cyan)
	}
	l.println(8, iface...)
}

// Debugf is the generated logger function to satisfy the interface
func (l *logger) Debugf(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Cyan)
	}
	l.printf(8, fmt, iface...)
}

// Error is the generated logger function to satisfy the interface
func (l *logger) Error(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Red)
	}
	l.println(4, iface...)
}

// Errorf is the generated logger function to satisfy the interface
func (l *logger) Errorf(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Red)
	}
	l.printf(4, fmt, iface...)
}

// Fatal is the generated logger function to satisfy the interface
func (l *logger) Fatal(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Red)
	}
	l.println(32, iface...)
}

// Fatalf is the generated logger function to satisfy the interface
func (l *logger) Fatalf(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Red)
	}
	l.printf(32, fmt, iface...)
}

// Info is the generated logger function to satisfy the interface
func (l *logger) Info(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Green)
	}
	l.println(1, iface...)
}

// Infof is the generated logger function to satisfy the interface
func (l *logger) Infof(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Green)
	}
	l.printf(1, fmt, iface...)
}

// Trace is the generated logger function to satisfy the interface
func (l *logger) Trace(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Blue)
	}
	l.println(16, iface...)
}

// Tracef is the generated logger function to satisfy the interface
func (l *logger) Tracef(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Blue)
	}
	l.printf(16, fmt, iface...)
}

// Warn is the generated logger function to satisfy the interface
func (l *logger) Warn(iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Yellow)
	}
	l.println(2, iface...)
}

// Warnf is the generated logger function to satisfy the interface
func (l *logger) Warnf(fmt string, iface ...interface{}) {
	l.l.Lock() // locks in the color change
	defer l.l.Unlock()
	if l.color == "" {
		l.color = color2ESC(Yellow)
	}
	l.printf(2, fmt, iface...)
}

// HTTPMiddleware is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) HTTPMiddleware(next http.Handler) (r http.Handler) {
	return next
}

// Suppress is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Suppress() {}

// UnSuppress is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) UnSuppress() {}

// Color is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Color(x LogColor) Logger { return l }

// OnErr is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) OnErr(x error) Logger { return l }

// Debug is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Debug(iface ...interface{}) { return }

// Debugf is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Debugf(fmt string, iface ...interface{}) { return }

// Error is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Error(iface ...interface{}) { return }

// Errorf is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Errorf(fmt string, iface ...interface{}) { return }

// Fatal is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Fatal(iface ...interface{}) { return }

// Fatalf is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Fatalf(fmt string, iface ...interface{}) { return }

// Info is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Info(iface ...interface{}) { return }

// Infof is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Infof(fmt string, iface ...interface{}) { return }

// Trace is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Trace(iface ...interface{}) { return }

// Tracef is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Tracef(fmt string, iface ...interface{}) { return }

// Warn is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Warn(iface ...interface{}) { return }

// Warnf is the nilLogger function to satisfy the interface. It does nothing.
func (l *nilLogger) Warnf(fmt string, iface ...interface{}) { return }
