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

	devices, err := dm.ListDevices()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, device := range devices {
		fmt.Printf("\nDevice %d,%d (%s) targets:\n", major(device.Dev), minor(device.Dev), device.Name)

		targets, err := dm.TableStatus(device.Dev)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, t := range targets {
			fmt.Printf("Type: %s, start: %d, length: %d, params: %s\n",
				t.Type, t.Start, t.Length, t.Params)

			if t.Type == "cache" {
				d := unmarshallParams(t.Params)
				fmt.Printf("Parsed data: %#v\n", d)

				fmt.Printf("\nMetadata block size: %d bytes    Usage: %f %%\n", d.mdataBlockSize, d.mdataUsedPerc())
				fmt.Printf("Cache block size: %d bytes    Usage: %f %%\n", d.cacheBlockSize, d.cacheUsedPerc())
				fmt.Printf("Read hit ratio: %f %%\n", d.readHitRatio()*100)
				fmt.Printf("Write hit ratio: %f %%\n", d.writeHitRatio()*100)
				fmt.Printf("Demotions: %d    Promotions: %d    Dirty: %d\n", d.demotions, d.promotions, d.dirty)
			}
		}
	}
}
