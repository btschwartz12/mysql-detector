package main

import (
	"fmt"
	"net"
	"time"
)

type ConnectionError struct {
	OriginalError error
}
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.OriginalError)
}

type ParseError struct {
	OriginalError error
}
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %v", e.OriginalError)
}

func initiateHandshake(host, port string, packetSize, timeoutSec int) ([]byte, error) {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		return nil, &ConnectionError{OriginalError: err}
	}
	defer conn.Close()

  	err = conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	if err != nil {
    return nil, fmt.Errorf("failed to set read deadline: %v", err)
	}

	buffer := make([]byte, packetSize)
	_, err = conn.Read(buffer)
	if err != nil {
		return nil, &ParseError{OriginalError: err}
	}
	return buffer, nil
}

const DocumentationURL = "https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_handshake.html"

func DetectMySQLServerVersion(host, port string) (string, error) {
	/* 
		See DocumentationURL above.
		
		You may think that the version information starts at the second byte, but take a look at a raw packet:

		00000000  5b 00 00 00 0a 38 2e 30  2e 33 36 2d 30 75 62 75  |[....8.0.36-0ubu|
		00000010  6e 74 75 30 2e 32 32 2e  30 34 2e 31 00 1e 00 00  |ntu0.22.04.1....|
	
		The version info actually starts after the 0x0a byte, which is not the second byte.
	*/
	buffer, err := initiateHandshake(host, port, 1024, 1)
	if err != nil {
		return "", err
	}
	startIndex := 1
	for startIndex < len(buffer) && buffer[startIndex] != 0x0a {
		startIndex++
	}
	if startIndex >= len(buffer) {
		return "", &ParseError{OriginalError: fmt.Errorf("start index out of bounds")}
	}
	startIndex++
	endIndex := startIndex
	for endIndex < len(buffer) && buffer[endIndex] != 0x00 {
		endIndex++
	}
	if endIndex >= len(buffer) {
		return "", &ParseError{OriginalError: fmt.Errorf("end index out of bounds")}
	}
	serverVersion := string(buffer[startIndex:endIndex])
	return serverVersion, nil
}


