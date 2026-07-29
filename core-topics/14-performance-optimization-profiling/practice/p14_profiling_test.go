package practice

import (
	"testing"
)

func TestFormatWithPool(t *testing.T) {
	res := FormatWithPool("INFO", "Server started")
	expected := "INFO: Server started"
	if res != expected {
		t.Fatalf("Expected %s, got %s", expected, res)
	}
}

func BenchmarkFormatWithPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FormatWithPool("ERROR", "Connection refused")
	}
}
