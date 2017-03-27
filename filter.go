package logger

import (
	"bytes"
	"io"
)

// Filter is an interface that determines if a filter should be completed
type Filter interface {
	Flush() bool
}

// filteredWriter is a struct that wraps a writer that will filter out writes based on a passed in function
type filteredStringWriter struct {
	w   io.Writer
	fn  FilteredStringFunc
	buf *bytes.Buffer
}

// Write makes this a writer interface. This will buffer writes until an error or the Done method is called.
func (fw *filteredStringWriter) Write(p []byte) (n int, err error) {
	if fw.buf == nil {
		fw.buf = new(bytes.Buffer)
	}
	return fw.buf.Write(p)
}

// Flush makes this a Filter interface. This will flush and reset the buffer. It returns if anything was written or not
func (fw *filteredStringWriter) Flush() bool {
	defer fw.buf.Reset()

	if fw.fn(fw.buf.String()) {
		fw.w.Write(fw.buf.Bytes())
		return true
	}
	return false
}
