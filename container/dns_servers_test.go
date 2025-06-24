package main

import (
	"fmt"
	"testing"
)

func TestDNSServer_String(t *testing.T) {
	server := DNSServer{
		Name:    "TestServer",
		Address: "127.0.0.1",
		Port:    5353,
	}
	expected := "TestServer (127.0.0.1:5353)"
	result := server.String()
	if result != expected {
		t.Errorf("String() = %q, want %q", result, expected)
	}
}

func TestDNSServer_AddressString(t *testing.T) {
	server := DNSServer{
		Name:    "TestServer",
		Address: "192.168.1.1",
		Port:    53,
	}
	expected := fmt.Sprintf("%s:%d", server.Address, server.Port)
	result := server.AddressString()
	if result != expected {
		t.Errorf("AddressString() = %q, want %q", result, expected)
	}
}
