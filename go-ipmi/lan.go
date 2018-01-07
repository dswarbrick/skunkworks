package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

const ipmiBufSize = 1024

type lanConnection struct {
	conn net.Conn
}

func newLanConnection(host string) (lanConnection, error) {
	l := lanConnection{}

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
	// TBC
}

func (l *lanConnection) message(req Request) []byte {
	buf := new(bytes.Buffer)

	// Write RMCP header
	rmcpHeader := rmcpHeader{
		Version:            rmcpVersion1,
		RMCPSequenceNumber: 0xff,
		Class:              rmcpClassIPMI,
	}

	ipmiSession := ipmiSession{}

	binaryWrite(buf, rmcpHeader)
	binaryWrite(buf, ipmiSession)

	// Construct and write IPMI header
	ipmiHeader := ipmiHeader{
		MsgLen:     0x09, // Message len
		RsAddr:     0x20, // Target address
		NetFnRsLUN: 0x18, // NetFn, target LUN
		RqAddr:     0x81, // Source address
		Command:    req.Command,
	}

	// Header checksum
	ipmiHeader.Checksum = checksum(ipmiHeader.RsAddr, ipmiHeader.NetFnRsLUN)

	binaryWrite(buf, ipmiHeader)

	data := new(bytes.Buffer)
	binaryWrite(data, req.Data)

	binaryWrite(buf, req.Data)

	calcCsum := checksum(ipmiHeader.RqAddr, ipmiHeader.RqSeq, ipmiHeader.Command) + checksum(data.Bytes()...)
	fmt.Printf("calc csum: %x\n", calcCsum)
	buf.WriteByte(calcCsum)

	return buf.Bytes()
}

func (l *lanConnection) recv() (int, []byte) {
	buf := make([]byte, ipmiBufSize)
	n, err := l.conn.Read(buf)
	if err != nil {
		panic(err)
	}

	return n, buf
}

func (l *lanConnection) send(req Request) (int, error) {
	buf := l.message(req)
	return l.sendPacket(buf)
}

func (l *lanConnection) sendPacket(b []byte) (int, error) {
	return l.conn.Write(b)
}
