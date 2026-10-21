package solutions

import (
	"testing"
)

type Config struct {
	Host string `json:"host_name"`
	Port int    `json:"port_number"`
}

func TestZeroCopyStringToBytes(t *testing.T) {
	str := "Go Performance"
	b := ZeroCopyStringToBytes(str)
	if len(b) != len(str) {
		t.Fatalf("Expected length %d, got %d", len(str), len(b))
	}
	if string(b) != str {
		t.Fatalf("Expected %s, got %s", str, string(b))
	}
}

func TestGetStructFieldTag(t *testing.T) {
	cfg := Config{}
	tag := GetStructFieldTag(cfg, "Host")
	if tag != "host_name" {
		t.Fatalf("Expected host_name, got %s", tag)
	}
}
