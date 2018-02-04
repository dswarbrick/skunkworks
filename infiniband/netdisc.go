package main

// #cgo CFLAGS: -I/usr/include/infiniband
// #cgo LDFLAGS: -libmad -libumad -libnetdisc
// #include <stdlib.h>
// #include <umad.h>
// #include <ibnetdisc.h>
import "C"

import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

func umadGetCANames() []string {
	var (
		buf  [C.UMAD_CA_NAME_LEN][C.UMAD_MAX_DEVICES]byte
		hcas = make([]string, 0, C.UMAD_MAX_DEVICES)
	)

	// Call umad_get_cas_names with pointer to first element in our buffer
	numHCAs := C.umad_get_cas_names((*[C.UMAD_CA_NAME_LEN]C.char)(unsafe.Pointer(&buf[0])), C.UMAD_MAX_DEVICES)

	for x := 0; x < int(numHCAs); x++ {
		hcas = append(hcas, strings.TrimRight(string(buf[x][:]), "\x00"))
	}

	return hcas
}

func main() {

	// Application will not see the simulated sysfs directory from ibsim unless umad_init() is
	// called first.
	// This seems to leave behind a `sys-$PID` directory after exiting, which does not occur when
	// running the equivalent from a C program.
	if C.umad_init() < 0 {
		fmt.Println("Error initialising umad library")
		os.Exit(1)
	}

	for _, ca := range umadGetCANames() {
		fmt.Println(ca)
	}

	C.umad_done()
}
