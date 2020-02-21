package kv

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

var bPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var sPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 16)
	},
}

func Marshal(v interface{}) (out []byte, err error) {
	// cheat and only accept maps for now
	// TODO(njones): make this accept structs as well.

	var comma string
	var sb = bPool.Get().(*bytes.Buffer)
	defer func() { sb.Reset(); bPool.Put(sb) }()

	switch kvs := v.(type) {
	case map[string]interface{}:
		var ks = sPool.Get().([]string)
		defer func() { ks = ks[:0]; sPool.Put(ks) }()

		for key := range kvs {
			ks = append(ks, key)
		}

		sort.Strings(ks)
		for _, k := range ks {
			fmt.Fprintf(sb, "%s%s=%v", comma, k, kvs[k])
			comma = ", "
		}
	}

	return sb.Bytes(), err
}
