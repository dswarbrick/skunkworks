/*
 * Pure Go devicemapper proof of concept
 * Copyright 2017 Daniel Swarbrick
 *
 * This package contains some alternatives to functions in
 * https://github.com/docker/docker/tree/master/pkg/devicemapper, which uses cgo and requires the
 * actual libdevmapper to build.
 */

package main

import (
	"fmt"
	"os"
)

func main() {
	dm, err := NewDevMapper()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer dm.Close()

	fmt.Printf("Kernel devmapper version: %s\n", dm.Version())
}
