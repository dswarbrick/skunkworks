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
)

func main() {
	// This seems to leave behind a `sys-$PID` directory after exiting, which does not occur when
	// running the equivalent from a C program.
	if C.umad_init() < 0 {
		fmt.Println("Error initialising umad library")
		os.Exit(1)
	}

	C.umad_done()
}
