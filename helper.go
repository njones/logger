package logger

import (
	"strings"
)

func hasFlag(has int, flags ...int) bool {
	for _, flag := range flags {
		if flag&has != 0 {
			return true
		}
	}
	return false
}

func KV(k K, v V) KeyVal {
	return KeyVal{k, v}
}

// convertStamp converts a timestamp from the std libary format of (i.e. 2006-02-01) to
// the internal single byte representation (i.e. \x06-\x02-\x01)
func convertStamp(format string) string {
	for _, k := range tsMap {
		format = strings.Replace(format, k, string(tsFmtMap[k]), -1)
	}
	return format
}

// tsMap is the string lookup for maping timestamps using the `convertStamp` function
// NOTE: the string needs to be storted from longest string to shortest, so that the
// lookup works as expected
var tsMap = []string{"05.000000", "Monday", "January", "2006", "-0700", "MST", "Mon", "Jan", "02", "15", "04", "05", "01", "06", "03", "pm"}
