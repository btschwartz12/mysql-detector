package main

import (
	"net"
	"testing"
	"time"
)

func startMockServer(response []byte, delay time.Duration) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	go func() {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			if delay > 0 {
				time.Sleep(delay)
			}
			conn.Write(response)
			conn.Close()
		}
	}()

	return ln.Addr().String(), nil
}

func TestDetectMySQLServerVersionSuccess(t *testing.T) {
	handshakePacket := []byte{
		0x5b, 0x00, 0x00, 0x00, 0x0a, '8', '.', '0', '.', '3', '6', '-', '0', 'u', 'b', 'u',
		'n', 't', 'u', '0', '.', '2', '2', '.', '0', '4', '.', '1', 0x00,
	}
	addr, err := startMockServer(handshakePacket, 0)
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	host, port, _ := net.SplitHostPort(addr)
	version, err := DetectMySQLServerVersion(host, port)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedVersion := "8.0.36-0ubuntu0.22.04.1"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestDetectMySQLServerVersionConnectionRefused(t *testing.T) {
	host := "127.0.0.1"
	port := "12345"
	_, err := DetectMySQLServerVersion(host, port)
	_, ok := err.(*ConnectionError)
	if !ok {
		t.Fatalf("Expected connection error, got %v", err)
	}
}

func TestDetectMySQLServerVersionReadTimeout(t *testing.T) {
	addr, err := startMockServer(nil, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	host, port, _ := net.SplitHostPort(addr)
	_, err = DetectMySQLServerVersion(host, port)
	_, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected parse error, got %v", err)
	}
}

func TestDetectMySQLServerVersionNonMySQLPacket(t *testing.T) {
	nonMySQLPacket := []byte{0x00, 0x00, 0x00, 0x00}
	addr, err := startMockServer(nonMySQLPacket, 0)
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	host, port, _ := net.SplitHostPort(addr)
	_, err = DetectMySQLServerVersion(host, port)
	_, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected parse error, got %v", err)
	}
}
