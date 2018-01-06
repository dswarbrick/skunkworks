package main

import (
	"encoding/binary"
)

const (
	rmcpVersion1  = 0x06
	rmcpClassIPMI = 0x07
)

var (
	rmcpHeaderSize  = binary.Size(rmcpHeader{})
	ipmiSessionSize = binary.Size(ipmiSession{})
	ipmiHeaderSize  = binary.Size(ipmiHeader{})
)

type rmcpHeader struct {
	Version            uint8
	Reserved           uint8
	RMCPSequenceNumber uint8
	Class              uint8
}

// TODO: Deprecate this function
func decodeRMCPHeader(buf []byte) *rmcpHeader {
	if len(buf) < rmcpHeaderSize {
		panic(nil)
	}

	return &rmcpHeader{buf[0], buf[1], buf[2], buf[3]}
}
