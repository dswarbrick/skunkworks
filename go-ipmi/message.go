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

func newMessageFromBytes(b []byte) error {
	if len(b) < rmcpHeaderSize+ipmiSessionSize+ipmiHeaderSize {
		return fmt.Errorf("Undersized packet")
	}

	m := message{
		rmcpHeader:  &rmcpHeader{},
		ipmiSession: &ipmiSession{},
		ipmiHeader:  &ipmiHeader{},
	}

	ipmiHeader := ipmiHeader{}

	r := bytes.NewReader(b)

	if err := binary.Read(r, binary.LittleEndian, m.rmcpHeader); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, m.ipmiSession); err != nil {
		return err
	}

	if m.ipmiSession.AuthType != 0 {
		return fmt.Errorf("AuthType not supported yet")
	}

	if err := binary.Read(r, binary.LittleEndian, m.ipmiHeader); err != nil {
		return err
	}

	fmt.Println("Message from bytes")
	fmt.Printf("%#v\n", m.rmcpHeader)
	fmt.Printf("%#v\n", m.ipmiSession)
	fmt.Printf("%#v\n", m.ipmiHeader)

	m.data = make([]byte, int(m.ipmiHeader.MsgLen)-ipmiHeaderSize)
	if _, err := r.Read(m.data); err != nil {
		return err
	}

	fmt.Printf("% x\n", m.data)

	// Checksum byte should be the last byte, immediately after the data
	csum, _ := r.ReadByte()
	fmt.Printf("csum: %x\n", csum)

	// Calculate payload checksum
	calcCsum := checksum(ipmiHeader.RqAddr, ipmiHeader.RqSeq, ipmiHeader.Command) + checksum(m.data...)
	fmt.Printf("calc csum: %x\n", calcCsum)

	res := AuthCapabilitiesResponse{}
	r = bytes.NewReader(m.data)
	binary.Read(r, binary.LittleEndian, &res)

	fmt.Printf("%#v\n", res)

	// Check for supported auth type in order of preference
	for _, t := range []uint8{AuthTypeMD5, AuthTypePassword, AuthTypeNone} {
		if (res.AuthTypeSupport & (1 << t)) != 0 {
			fmt.Println(t)
			break
		}
	}

	return nil
}
