package practice

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// FormatWithPool uses sync.Pool to avoid allocations during string formatting.
func FormatWithPool(prefix, msg string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	buf.WriteString(prefix)
	buf.WriteString(": ")
	buf.WriteString(msg)
	return buf.String()
}
