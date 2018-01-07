package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
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

	lc, err := newLanConnection(*host)
	if err != nil {
		panic(err)
	}
	defer lc.close()

	fmt.Printf("Connection established: %#v\n", lc)

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
		Command:    CmdGetChannelAuthCapabilities,
	}

	// Header checksum
	ipmiHeader.Checksum = checksum(ipmiHeader.RsAddr, ipmiHeader.NetFnRsLUN)
	binaryWrite(buf, ipmiHeader)

	req := AuthCapabilitiesRequest{
		0x8e, // IPMI v2.0+ extended data, current channel
		PrivLevelAdmin,
	}

	binaryWrite(buf, req)

	calcCsum := checksum(ipmiHeader.RqAddr, ipmiHeader.RqSeq, ipmiHeader.Command, req.ChannelNumber, req.PrivLevel)
	fmt.Printf("calc csum: %x\n", calcCsum)
	buf.WriteByte(calcCsum)

	fmt.Printf("%d bytes written\n", lc.send(buf.Bytes()))

	n, inbuf := lc.recv()
	fmt.Printf("%d bytes read: % x\n", n, inbuf[:n])

	hdr := decodeRMCPHeader(inbuf[:n])
	fmt.Printf("%#v\n", hdr)

	if hdr.Class != rmcpClassIPMI {
		fmt.Printf("Unsupported class: %#x\n", hdr.Class)
	}

	newMessageFromBytes(inbuf[:n])
}
