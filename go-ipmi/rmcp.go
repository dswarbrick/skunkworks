package main

import (
	"encoding/binary"
)

const (
	rmcpVersion1 = 0x06
	rmcpClassIPMI = 0x07
)

var (
	rmcpHeaderSize = binary.Size(rmcpHeader{})
)

type rmcpHeader struct {
	Version            uint8
	Reserved           uint8
	RMCPSequenceNumber uint8
	Class              uint8
}

func decodeRMCPHeader(buf []byte) *rmcpHeader {
	if len(buf) < rmcpHeaderSize {
		panic(nil)
	}

	return &rmcpHeader{buf[0], buf[1], buf[2], buf[3]}
}
