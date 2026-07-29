package practice

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
