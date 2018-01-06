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

// AuthCapabilitiesResponse per section 22.13
type AuthCapabilitiesResponse struct {
	CompletionCode  uint8
	ChannelNumber   uint8
	AuthTypeSupport uint8
	Status          uint8
	Reserved        uint8
	OEMID           uint16
	OEMAux          uint8
}

func newMessageFromBytes(b []byte) error {
	if len(b) < rmcpHeaderSize+ipmiSessionSize+ipmiHeaderSize {
		return fmt.Errorf("Undersized packet")
	}

	rmcpHeader := rmcpHeader{}
	ipmiSession := ipmiSession{}
	ipmiHeader := ipmiHeader{}

	r := bytes.NewReader(b)

	if err := binary.Read(r, binary.LittleEndian, &rmcpHeader); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &ipmiSession); err != nil {
		return err
	}

	if ipmiSession.AuthType != 0 {
		return fmt.Errorf("AuthType not supported yet")
	}

	if err := binary.Read(r, binary.LittleEndian, &ipmiHeader); err != nil {
		return err
	}

	fmt.Printf("%#v\n", rmcpHeader)
	fmt.Printf("%#v\n", ipmiSession)
	fmt.Printf("%#v\n", ipmiHeader)

	data := make([]byte, int(ipmiHeader.MsgLen)-ipmiHeaderSize)
	if _, err := r.Read(data); err != nil {
		return err
	}

	fmt.Printf("% x\n", data)

	// Checksum byte should be the last byte, immediately after the data
	csum, _ := r.ReadByte()
	fmt.Printf("csum: %x\n", csum)

	// Calculate payload checksum
	calcCsum := checksum(ipmiHeader.RqAddr, ipmiHeader.RqSeq, ipmiHeader.Command) + checksum(data...)
	fmt.Printf("calc csum: %x\n", calcCsum)

	res := AuthCapabilitiesResponse{}
	r = bytes.NewReader(data)
	binary.Read(r, binary.LittleEndian, &res)

	fmt.Printf("%#v\n", res)

	return nil
}
