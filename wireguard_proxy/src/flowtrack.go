package main

import (
	"net"
	"strconv"
)

// ConnectionContext holds the core transport-layer metadata
type ConnectionContext struct {
	SrcIP    net.IP
	SrcPort  int
	DstIP    net.IP
	DstPort  int
	Protocol string
}

// NewConnectionContext initializes the context from an active net.Conn
func NewConnectionContext(conn net.Conn, detectedProtocol string) (*ConnectionContext, error) {
	srcStr, srcPortStr, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, err
	}
	
	dstStr, dstPortStr, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	return &ConnectionContext{
		SrcIP:    net.ParseIP(srcStr),
		SrcPort:  parsePort(srcPortStr),
		DstIP:    net.ParseIP(dstStr),
		DstPort:  parsePort(dstPortStr),
		Protocol: detectedProtocol,
	}, nil
}

// parsePort safely converts port strings to integers
func parsePort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}
