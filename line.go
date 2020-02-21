package logger

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
)

type deferFunc func()

type setize interface {
	set(*line)
}

type colorize []byte
type formatize string
type levelize []byte
type timeize []byte
type writeize struct{ io.Writer }

func (x colorize) set(ln *line) {
	if ln.color == nil { // a []byte{} == noColor
		ln.color = []byte(x)
	}
}

func (x formatize) set(ln *line) { ln.format = string(x) }
func (x levelize) set(ln *line)  { ln.prefixLevel = []byte(x) }
func (x timeize) set(ln *line)   { ln.time = []byte(x) }
func (w writeize) set(ln *line)  { ln.out.w = io.MultiWriter(ln.out.w, w) }

type dropCRWriter struct {
	w io.Writer
}

func (dw *dropCRWriter) Write(p []byte) (n int, err error) {
	var j = len(p)
	if p[j-1] == '\n' {
		return dw.w.Write(p[:j-1])
	}
	return dw.w.Write(p)
}

type line struct {
	do printKind

	flags int
	depth int

	time        []byte
	color       []byte
	prefixLevel []byte
	prefixUser  []byte

	format string
	v      []interface{}
	kv     string

	out struct {
		w  io.Writer
		cw io.Writer // color writer
		dw *dropCRWriter
	}
	err error
}

func (ln *line) write() error {
	defer ln.writeNewLine()
	defer ln.writeKVPairs()
	defer ln.writeColorEnd()

	ln.writeTime()
	ln.writeColor()
	ln.writeFilename()
	ln.writePrefixLevel()
	ln.writePrefixUser()
	ln.writePrint()
	return ln.err
}

func (ln *line) writeTime() {
	if ln.err != nil || ln.time == nil || len(ln.time) == 0 {
		return
	}
	_, ln.err = ln.out.w.Write(append(ln.time, ' '))
}

func (ln *line) writeColor() {
	if ln.err != nil || ln.color == nil || len(ln.color) == 0 {
		return
	}
	_, ln.err = ln.out.cw.Write(ln.color)
}

func (ln *line) writeColorEnd() {
	if ln.err != nil || ln.color == nil || len(ln.color) == 0 {
		return
	}
	_, ln.err = ln.out.cw.Write([]byte{0x1b, '[', '0', 'm'})
}

func (ln *line) writeFilename() {
	if ln.err != nil {
		return
	}

	if hasFlag(ln.flags, Llongfile, Lshortfile) {
		_, file, line, ok := runtime.Caller(ln.depth)
		if !ok {
			file, line = "???", 0
		}
		if hasFlag(ln.flags, Lshortfile) {
			file = filepath.Base(file)
		}

		fmt.Fprintf(ln.out.w, "%s:%d ", file, line)
	}
}

func (ln *line) writePrefixLevel() {
	if ln.err != nil || ln.prefixLevel == nil {
		return
	}
	_, ln.err = ln.out.w.Write(append(ln.prefixLevel, ' '))
}

func (ln *line) writePrefixUser() {
	if ln.err != nil || ln.prefixUser == nil {
		return
	}
	_, ln.err = ln.out.w.Write(append(ln.prefixUser, ' '))
}

func (ln *line) writePrint() {
	if ln.err != nil {
		return
	}

	switch ln.do {
	case bPrint:
		_, ln.err = fmt.Fprint(ln.out.w, ln.v...)
	case bPrintf:
		_, ln.err = fmt.Fprintf(ln.out.w, ln.format, ln.v...)
	case bPrintln:
		ln.out.dw.w = ln.out.w // this reduces an allocation we must add the writer each time...
		_, ln.err = fmt.Fprintln(ln.out.dw, ln.v...)
	}
}

func (ln *line) writeKVPairs() {
	if ln.err != nil || len(ln.kv) == 0 {
		return
	}

	_, ln.err = fmt.Fprint(ln.out.w, " "+ln.kv)
}

func (ln *line) writeNewLine() {
	if ln.err != nil {
		return
	}
	_, ln.err = fmt.Fprint(ln.out.w, "\n")
}
