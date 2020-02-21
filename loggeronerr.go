package logger

import "io"

type Return struct{ err error }

func (rtn Return) Err() error   { return rtn.err }
func (rtn Return) HasErr() bool { return rtn.err != nil }

type ErrSub struct{ error }

type onErrLogger struct {
	b   *baseLogger
	err error
}

func (e *onErrLogger) Flags() int {
	return e.b.flags
}

func (e *onErrLogger) Output(calldepth int, s string) error {
	e.b.depth = calldepth
	e.b.Println(s)
	return nil //TODO(njones): pass the error along...
}

func (e *onErrLogger) Prefix() string { return e.b.Prefix() }

func (e *onErrLogger) SetFlags(flags int) { e.b.SetFlags(flags) }

func (e *onErrLogger) SetOutput(w io.Writer) { e.b.SetOutput(w) }

func (e *onErrLogger) SetPrefix(s string) { e.b.SetPrefix(s) }

func (e *onErrLogger) Writer() io.Writer { return e.b.Writer() }

// fillOnErr checks to see if error is not nil, if so then fill any OnErr structs
// passed along in the log level function with the non-nil error value
func (e onErrLogger) popOnErr(v []interface{}) bool {
	b := e.err != nil
	if b {
		for i, val := range v {
			if vv, ok := val.(ErrSub); ok {
				vv.error = e.err
				v[i] = vv
			}
		}
	}
	return b
}
