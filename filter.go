package logger

import (
	"bytes"
	"io"
)

// FilteredByteFunc is the a function that takes a stream of bytes and returns a stream
type FilteredByteFunc func(*filteredByteWriter, []byte) []byte

// FilteredStringFunc is the a function that takes a string and determines if it should be logged or not
type FilteredStringFunc func(string) bool

// FilterFlusher is an interface that determines if a filter should be completed
type FilterFlusher interface {
	Flush() bool
}

// FilterString is an interface that determines if a filter should take the raw data to filter on
type FilterOn interface {
	On(string)
}

// filteredWriter is a struct that wraps a writer that will filter out writes based on a passed in function
type filteredStringWriter struct {
	w   io.Writer
	fn  FilteredStringFunc
	buf *bytes.Buffer

	fs string
}

// Write makes this a writer interface. This will buffer writes until an error or the Flush method is called.
func (fw *filteredStringWriter) Write(p []byte) (n int, err error) {
	if fw.buf == nil {
		fw.buf = new(bytes.Buffer)
	}
	return fw.buf.Write(p)
}

// Raw takes external data to stringize and do the flush check on if there
func (fw *filteredStringWriter) On(s string) {
	fw.fs = s
}

// Flush makes this a Filter interface. This will flush and reset the buffer. It returns if anything was written or not
func (fw *filteredStringWriter) Flush() bool {
	defer fw.buf.Reset()

	if fw.fs != "" {
		if fw.fn(fw.fs) {
			fw.w.Write(fw.buf.Bytes())
			return true
		}
		return false
	}

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

// Write makes this a writer interface. This will buffer writes until an error or the Flush method is called.
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
