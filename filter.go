package logger

import (
	"bytes"
	"io"
)

// FilteredByteFunc is the a function that takes a stream of bytes and returns a stream
type FilteredByteFunc func(*filteredByteWriter, []byte) []byte

// FilteredStringFunc is the a function that takes a string and determines if it should be logged or not
type FilteredStringFunc func(string) bool

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

// filteredWriter is a struct that wraps a writer that will filter out writes based on a passed in function
type filteredByteWriter struct {
	b   *bytes.Buffer
	w   io.Writer
	fn  FilteredByteFunc
	chk bool
	buf *bytes.Buffer
}

// Write makes this a writer interface. This will buffer writes until an error or the Done method is called.
func (fw *filteredByteWriter) Write(p []byte) (n int, err error) {
	if fw.buf == nil {
		fw.buf = new(bytes.Buffer)
	}

	if fw.b == nil {
		fw.b = new(bytes.Buffer)
	}

	n, err = fw.buf.Write(fw.fn(fw, p))
	if err == nil {
		n = len(p) // this prevents a short write error because this is what is supposed to happen
	}
	return
}

// Flush makes this a Filter interface. This will flush and reset the buffer. It returns if anything was written or not
func (fw *filteredByteWriter) Flush() bool {
	defer fw.buf.Reset()

	var err error
	if _, err = fw.w.Write(fw.buf.Bytes()); err == nil {
		_, err = fw.w.Write(fw.b.Bytes())
	}
	return err == nil
}
