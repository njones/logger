package logger

import (
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"time"
)

// StdKVMarshal is the structured logging which looks like key=value, Value is a string, int, float or goString.
func StdKVMarshal(in interface{}) ([]byte, error) {
	var rtns []string
	switch val := in.(type) {
	case map[string]interface{}:
		for k, v := range val {
			switch vval := v.(type) {
			case string:
				rtns = append(rtns, fmt.Sprintf("%s=%s", k, vval))
			case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64,
				float64:
				rtns = append(rtns, fmt.Sprintf("%s=%d", k, vval))
			default:
				rtns = append(rtns, fmt.Sprintf("%s=%v", k, vval))
			}
		}
	}

	sort.Strings(rtns)
	return []byte(strings.Join(rtns, " ")), nil
}

type KeyValues map[string]interface{}

// keyValue is the struct that holds Key/Value pairs for structured logging
type keyValue struct {
	K string
	V interface{}
}

// KV is the publicly exposed function that returns a struct for structured logging
func KV(k string, v interface{}) keyValue {
	return keyValue{K: k, V: v}
}

// KVMap is the publicly exposed function that returns a slice of structs for structured logging of a map
func KVMap(m KeyValues) []keyValue {
	rtn := make([]keyValue, len(m))
	i := 0
	for k, v := range m {
		rtn[i] = keyValue{K: k, V: v}
		i++
	}
	return rtn
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
