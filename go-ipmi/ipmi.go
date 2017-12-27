package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp4", "127.0.0.1:9001")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Connection established")

	buf := []byte{
		0x06,                   // Version
		0x00,                   // Reserved
		0xff,                   // Sequence
		0x07,                   // Class, message type
		0x00,                   // Auth type
		0x00, 0x00, 0x00, 0x00, // Session sequence number
		0x00, 0x00, 0x00, 0x00, // Session ID
		0x09, // Message len
		0x20, // Target address
		0x18, // NetFn, target LUN
		0xc8, // Checksum
		0x81, // Source address
		0x00, // Source LUN, sequence number
		CmdGetChannelAuthCapabilities,
		0x8e, // IPMI v2.0+ extended data, current channel
		PrivLevelAdmin,
		0xb5, // Checksum
	}

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
