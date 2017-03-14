package logger

import (
	"fmt"
	"io"
	"net"
	"time"
)

// StdKVMarshal is the structured logging which looks like key=value, Value is a string, int, float or goString.
func StdKVMarshal(in interface{}) ([]byte, error) {
	var rtn string
	switch val := in.(type) {
	case map[string]interface{}:
		for k, v := range val {
			switch vval := v.(type) {
			case string:
				rtn += fmt.Sprintf("%s=%s ", k, vval)
			case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64,
				float64:
				rtn += fmt.Sprintf("%s=%d ", k, vval)
			default:
				rtn += fmt.Sprintf("%s=%v ", k, vval)
			}
		}
	}
	return []byte(rtn), nil
}

// keyvalue is the struct that holds Key/Value pairs for structured logging
type keyValue struct {
	K string
	V interface{}
}

// KV is the publicly exposed function that returns a struct for structured logging
func KV(k string, v interface{}) keyValue {
	return keyValue{K: k, V: v}
}

// NetWriter is a helper function that will log writes to a TCP/UDP address
func NetWriter(network, address string) io.Writer {
	return netwriter{network: network, address: address}
}

// netwriter the underling struct that will write to the connection
type netwriter struct {
	network, address string
}

// Write pass writes to the connection as a io.Writer
func (nw netwriter) Write(p []byte) (int, error) {
	conn, err := net.Dial(nw.network, nw.address)
	if err != nil {
		return 0, err
	}
	go func() {
		conn.SetWriteDeadline(<-time.After(filteredWriteDeadline))
	}()
	defer conn.Close()

	return conn.Write(p)
}
