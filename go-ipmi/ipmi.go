package main

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
		0x06, 0x00, 0xff, 0x07, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x20, 0x18,
		0xc8, 0x81, 0x00, 0x38, 0x8e, 0x04, 0xb5,
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
