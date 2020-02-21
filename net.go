package logger

import (
	"io"
	"net"
	"sync"
	"time"
)

// nwOpt defines a typed functional option interface
type nwOpt interface {
	setOption(*netWriter)
}

// tod timeout duration will wrap a time duration and allow for it to be passed in as an option
type tod time.Duration

// setOption satisfies the functional option interface for a NetWriter
func (t tod) setOption(nw *netWriter) { nw.timeout = time.Duration(t) }

// NetWriter is a helper function that will log writes to a TCP/UDP address. Any errors will be written to stderr.
func NetWriter(network, address string, opts ...nwOpt) io.Writer {
	nw := &netWriter{m: new(sync.Mutex), network: network, address: address, timeout: 5 * time.Second}
	for _, opt := range opts {
		opt.setOption(nw)
	}
	return nw
}

// netWriter the underling struct that will write to the connection
type netWriter struct {
	network, address string
	conn             net.Conn
	timeout          time.Duration

	m   *sync.Mutex // keeps the writes and close's in sync
	err error
}

// Write passes writes to the network connection from a io.Writer
func (nw *netWriter) Write(p []byte) (n int, err error) {
	nw.m.Lock() // protects creating a new connection on nil...
	defer nw.m.Unlock()

	if nw.conn == nil {
		nw.conn, nw.err = net.DialTimeout(nw.network, nw.address, nw.timeout)
		if nw.err != nil {
			return 0, nw.err
		}
	}
	defer nw.conn.SetWriteDeadline(time.Now().Add(nw.timeout)) // keep extending the deadline timeout

	return nw.conn.Write(p)
}

// Close closes the connection and removes it, so it can be opened on the
// next write if need be.
func (nw *netWriter) Close() error {
	nw.m.Lock()
	defer nw.m.Unlock()

	if nw.conn != nil {
		nw.err = nw.conn.Close()
	}
	nw.conn = nil
	return nw.err
}

func (nw *netWriter) Err() error { return nw.err }

func (nw *netWriter) NoColor() {}
