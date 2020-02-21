package logger

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPLogFormatFunc is the type that is used for logging HTTP requests
type HTTPLogFormatFunc func(time.Time, int, int64, []byte, *http.Request) string

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
func (b *baseLogger) HTTPMiddleware(next http.Handler) http.Handler {
	// set the default http logger if it's nil
	if b.http.formatFn == nil {
		b.http.formatFn = CommonLogFormat
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &ResponseWriter{ResponseWriter: w}
		next.ServeHTTP(cw, r)

		for k := range b.http.headers {
			hval := r.Header.Get(string(k))
			b.http.headers[k] = hval
		}

		b.HTTPln(b.http.formatFn(b.ts.now, cw.status, cw.sent, b.ts.text, r), b.http.headers)
	})
}

func collapseSpace(s string) string {
	if len(s) == 0 {
		return ""
	}
	return " " + s
}

// CommonLogFormat is the Apache Common Logging format used for logging HTTP requests
func CommonLogFormat(now time.Time, status int, sent int64, tsText []byte, r *http.Request) string {
	// $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" | nginx
	// 127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326

	if tsText == nil {
		tsText = now.AppendFormat(tsText, "02/Jan/2006:15:04:05 -0700")
	}

	commonLog := struct {
		RemoteAddr    string `json:"remote_address"`
		RemoteID      string `json:"remote_identifier"`
		RemoteUser    string `json:"remote_user"`
		LocalTime     string `json:"time_local"`
		RequestString string `json:"request"`
		Status        int    `json:"status"`
		BytesSent     int64  `json:"body_bytes_sent"`
		Referer       string `json:"http_referer"`
		UserAgent     string `json:"http_user_agent"`
	}{
		RemoteAddr:    r.RemoteAddr,
		RemoteID:      "-",
		RemoteUser:    "-",
		LocalTime:     string(tsText),
		RequestString: fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto),
		Status:        status,
		BytesSent:     sent,
		Referer:       collapseSpace(r.Referer()),
		UserAgent:     collapseSpace(r.UserAgent()),
	}

	return fmt.Sprintf("%s %s %s [%s] \"%s\" %d %d%s%s", commonLog.RemoteAddr,
		commonLog.RemoteID, commonLog.RemoteUser,
		commonLog.LocalTime, commonLog.RequestString,
		commonLog.Status, commonLog.BytesSent,
		commonLog.Referer, commonLog.UserAgent)
}
