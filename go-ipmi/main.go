package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
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

	lc.getAuthCapabilities()
}
