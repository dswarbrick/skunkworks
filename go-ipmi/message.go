package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	NetworkFunctionApp = 0x06
)

type ipmiSession struct {
	AuthType  uint8
	Sequence  uint32
	SessionID uint32
}

type ipmiHeader struct {
	MsgLen     uint8
	RsAddr     uint8
	NetFnRsLUN uint8
	Checksum   uint8
	RqAddr     uint8
	RqSeq      uint8
	Command    uint8
}

type message struct {
	*rmcpHeader
	*ipmiSession
	authCode [16]byte
	*ipmiHeader
	data []byte
}

func newMessageFromBytes(b []byte) (*message, error) {
	if len(b) < rmcpHeaderSize+ipmiSessionSize+ipmiHeaderSize {
		return nil, fmt.Errorf("Undersized packet")
	}

	m := &message{
		rmcpHeader:  &rmcpHeader{},
		ipmiSession: &ipmiSession{},
		ipmiHeader:  &ipmiHeader{},
	}

	r := bytes.NewReader(b)

	if err := binary.Read(r, binary.LittleEndian, m.rmcpHeader); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, m.ipmiSession); err != nil {
		return nil, err
	}

	if m.ipmiSession.AuthType != 0 {
		return nil, fmt.Errorf("AuthType not supported yet")
	}

	if err := binary.Read(r, binary.LittleEndian, m.ipmiHeader); err != nil {
		return nil, err
	}

	if m.headerChecksum() != m.Checksum {
		return nil, fmt.Errorf("Header checksum mismatch")
	}

	if m.MsgLen <= 0 {
		return nil, fmt.Errorf("Undersized packet")
	}

	m.data = make([]byte, int(m.ipmiHeader.MsgLen)-ipmiHeaderSize)
	if _, err := r.Read(m.data); err != nil {
		return nil, err
	}

	fmt.Printf("% x\n", m.data)

	// Checksum byte should be the last byte, immediately after the data
	csum, _ := r.ReadByte()

	if m.payloadChecksum() != csum {
		return nil, fmt.Errorf("Payload checksum mismatch")
	}

	return m, nil
}

func (m *message) headerChecksum() uint8 {
	return checksum(m.RsAddr, m.NetFnRsLUN)
}

func (m *message) payloadChecksum() uint8 {
	return checksum(m.RqAddr, m.RqSeq, m.Command) + checksum(m.data...)
}
