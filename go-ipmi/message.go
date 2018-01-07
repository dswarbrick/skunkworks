package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// Network function codes per section 5.1
	NetFnChassis     = 0x00
	NetFnSensorEvent = 0x04
	NetFnApp         = 0x06
	NetFnStorage     = 0x0a
	NetFnGroupExtn   = 0x2c
)

type ipmiSession struct {
	AuthType  uint8
	Sequence  uint32
	SessionID uint32
}

type ipmiHeader struct {
	MsgLen     uint8
	RsAddr     uint8 // Responder slave address
	NetFnRsLUN uint8 // Network function, responder LUN
	Checksum   uint8
	RqAddr     uint8 // Requester address
	RqSeq      uint8 // Requester sequence number
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
		return nil, ErrShortPacket
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
		return nil, ErrInvalidPacket
	}

	if m.MsgLen <= 0 {
		return nil, ErrInvalidPacket
	}

	m.data = make([]byte, int(m.ipmiHeader.MsgLen)-ipmiHeaderSize)
	if _, err := r.Read(m.data); err != nil {
		return nil, err
	}

	fmt.Printf("% x\n", m.data)

	// Checksum byte should be the last byte, immediately after the data
	csum, _ := r.ReadByte()

	if m.payloadChecksum() != csum {
		return nil, ErrInvalidPacket
	}

	return m, nil
}

func (m *message) headerChecksum() uint8 {
	return checksum(m.RsAddr, m.NetFnRsLUN)
}

func (m *message) payloadChecksum() uint8 {
	return checksum(m.RqAddr, m.RqSeq, m.Command) + checksum(m.data...)
}

func binaryWrite(w io.Writer, data interface{}) {
	if err := binary.Write(w, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}

func checksum(b ...uint8) uint8 {
	var c uint8
	for _, x := range b {
		c += x
	}
	return -c
}
