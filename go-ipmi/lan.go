package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const ipmiBufSize = 1024

type lanConnection struct {
	conn      net.Conn // Socket connection
	priv      uint8    // Privilege level
	lun       uint8    // LUN
	sequence  uint32
	sessionID uint32
}

func newLanConnection(host string) (lanConnection, error) {
	l := lanConnection{
		priv: PrivLevelAdmin, // TODO
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	dialer := &net.Dialer{}
	if conn, err := dialer.DialContext(ctx, "udp4", host); err != nil {
		return l, err
	} else {
		l.conn = conn
	}

	// FIXME: Move to send / recv functions
	deadline, _ := ctx.Deadline()
	if err := l.conn.SetDeadline(deadline); err != nil {
		panic(err)
	}

	return l, nil
}

func (l *lanConnection) close() {
	l.conn.Close()
}

func (l *lanConnection) getAuthCapabilities() {
	req := Request{
		NetFnApp,
		CmdGetChannelAuthCapabilities,
		AuthCapabilitiesRequest{
			0x8e, // IPMI v2.0+ extended data, current channel
			l.priv,
		},
	}

	resp := AuthCapabilitiesResponse{}

	if err := l.send(req, &resp); err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", resp)

	// Check for supported auth type in order of preference
	for _, t := range []uint8{AuthTypeMD5, AuthTypePassword, AuthTypeNone} {
		if (resp.AuthTypeSupport & (1 << t)) != 0 {
			fmt.Println(t)
			break
		}
	}
}

func (l *lanConnection) message(req Request) []byte {
	buf := new(bytes.Buffer)

	// Write RMCP header
	rmcpHeader := rmcpHeader{
		Version:            rmcpVersion1,
		RMCPSequenceNumber: 0xff,
		Class:              rmcpClassIPMI,
	}

	ipmiSession := ipmiSession{
		Sequence:  l.nextSequence(),
		SessionID: l.sessionID,
	}

	binaryWrite(buf, rmcpHeader)
	binaryWrite(buf, ipmiSession)

	// Construct and write IPMI header
	ipmiHeader := ipmiHeader{
		MsgLen:     0x09,                                     // Message len
		RsAddr:     0x20,                                     // BMC slave address
		NetFnRsLUN: (req.NetworkFunction << 2) | (l.lun & 3), // NetFn, target LUN
		RqAddr:     0x81,                                     // Source address
		Command:    req.Command,
	}

	// Header checksum
	ipmiHeader.Checksum = checksum(ipmiHeader.RsAddr, ipmiHeader.NetFnRsLUN)

	binaryWrite(buf, ipmiHeader)

	data := new(bytes.Buffer)
	binaryWrite(data, req.Data)

	binaryWrite(buf, req.Data)

	payloadCsum := checksum(ipmiHeader.RqAddr, ipmiHeader.RqSeq, ipmiHeader.Command) + checksum(data.Bytes()...)
	buf.WriteByte(payloadCsum)

	return buf.Bytes()
}

func (l *lanConnection) nextSequence() uint32 {
	if l.sequence != 0 {
		l.sequence++
	}
	return l.sequence
}

func (l *lanConnection) recv() []byte {
	n, inbuf := l.recvPacket()
	fmt.Printf("%d bytes read: % x\n", n, inbuf[:n])

	hdr := decodeRMCPHeader(inbuf[:n])
	fmt.Printf("%#v\n", hdr)

	if hdr.Class != rmcpClassIPMI {
		fmt.Printf("Unsupported class: %#x\n", hdr.Class)
	}

	m, err := newMessageFromBytes(inbuf[:n])
	if err != nil {
		panic(err)
	}

	return m.data
}

func (l *lanConnection) recvPacket() (int, []byte) {
	buf := make([]byte, ipmiBufSize)
	n, err := l.conn.Read(buf)
	if err != nil {
		panic(err)
	}

	return n, buf
}

func (l *lanConnection) send(req Request, resp interface{}) error {
	buf := l.message(req)

	if _, err := l.sendPacket(buf); err != nil {
		return err
	}

	data := l.recv()

	r := bytes.NewReader(data)
	if err := binary.Read(r, binary.LittleEndian, resp); err != nil {
		panic(err)
	}

	return nil
}

func (l *lanConnection) sendPacket(b []byte) (int, error) {
	return l.conn.Write(b)
}
