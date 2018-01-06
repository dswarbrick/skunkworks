package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func checksum(b ...uint8) uint8 {
	var c uint8
	for _, x := range b {
		c += x
	}
	return -c
}

func binaryWrite(w io.Writer, data interface{}) {
	if err := binary.Write(w, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}

func main() {
	var host = flag.String("host", "", "Target host and port")

	flag.Parse()

	if *host == "" {
		fmt.Println("Insufficient arguments:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp4", *host)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Connection established")

	foo := new(bytes.Buffer)

	// Write RMCP header
	rmcphdr := rmcpHeader{
		Version:            rmcpVersion1,
		RMCPSequenceNumber: 0xff,
		Class:              rmcpClassIPMI,
	}

	ipmisesshdr := ipmiSession{}

	binaryWrite(foo, rmcphdr)
	binaryWrite(foo, ipmisesshdr)

	// Construct and write IPMI header
	ipmihdr := ipmiHeader{
		MsgLen:     0x09, // Message len
		RsAddr:     0x20, // Target address
		NetFnRsLUN: 0x18, // NetFn, target LUN
		RqAddr:     0x81, // Source address
	}

	// Header checksum
	ipmihdr.Checksum = checksum(ipmihdr.RsAddr, ipmihdr.NetFnRsLUN)
	binaryWrite(foo, ipmihdr)

	buf := foo.Bytes()
	buf = append(buf, []byte{
		CmdGetChannelAuthCapabilities,
		0x8e, // IPMI v2.0+ extended data, current channel
		PrivLevelAdmin,
		0xb5, // Checksum
	}...)

	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		panic(err)
	}

	n, err := conn.Write(buf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d bytes written\n", n)

	inbuf := make([]byte, 512)
	n, err = conn.Read(inbuf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d bytes read: % x\n", n, inbuf[:n])

	hdr := decodeRMCPHeader(buf)
	fmt.Printf("%#v\n", hdr)

	if hdr.Class != rmcpClassIPMI {
		fmt.Printf("Unsupported class: %#x\n", hdr.Class)
	}
}
