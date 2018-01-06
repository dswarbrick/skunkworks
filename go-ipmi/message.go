package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

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

	return nil
}
