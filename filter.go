package logger

import (
	"io"
	"regexp"
)

func FilterWriter(w io.Writer, filters ...Filter) io.Writer {
	return &filterWriter{w: w, filters: filters}
}

// NotFilter negates the check on the wrapped filter
func NotFilter(filter Filter) Filter {
	return &filterNot{filter}
}

// filterNot is a type to define a function that accepts a filter and negates the check.
type filterNot struct{ Filter }

// Check satisfies the Filter interface and runs the passed in function.
func (n *filterNot) Check(data string) bool { return !n.Filter.Check(data) }

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

// filterRegex is a type to define a function that accepts a regular expression to filter log lines.
type filterRegex struct{ regexp *regexp.Regexp }

// Check satisfies the Filter interface and matches against a regular expression
func (r *filterRegex) Check(data string) bool { return r.regexp.MatchString(data) }

// filterWriter runs the w io.Writer through the filters
type filterWriter struct {
	filters []Filter
	w       io.Writer
}

func (fw filterWriter) Write(p []byte) (n int, err error) {
	return fw.w.Write(p)
}
