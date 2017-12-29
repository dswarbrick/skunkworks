// +build linux

// Bio Exporter - A Prometheus exporter for Linux block IO statistics.
//
// Copyright 2017 Daniel Swarbrick

package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/iovisor/gobpf/bcc"

	"github.com/prometheus/client_golang/prometheus"
)

type bioStats struct {
	readLat  map[float64]uint64
	writeLat map[float64]uint64
}

type exporter struct {
	bpfMod   *bcc.Module
	readLat  *bcc.Table
	writeLat *bcc.Table
	latency  *prometheus.Desc
}

func newExporter(m *bcc.Module) *exporter {
	e := exporter{
		bpfMod:   m,
		readLat:  bcc.NewTable(m.TableId("read_lat"), m),
		writeLat: bcc.NewTable(m.TableId("write_lat"), m),
		latency: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "request", "latency_usec"),
			"A histogram of bio request latencies in microseconds.",
			[]string{"device", "operation"},
			nil,
		),
	}

	return &e
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	// TODO: Implement asynchronous clearing of BPF tables to prevent data becoming stale if not
	// regularly polled.
	// TODO: Replace fmt.Println with logger functions.

	devStats := make(map[string]bioStats)

	for entry := range e.readLat.Iter() {
		devName, bucket := parseKey(entry.Key)

		stats, ok := devStats[devName]
		if !ok {
			// First time seeing this device, initialize new latency map
			stats = bioStats{make(map[float64]uint64), make(map[float64]uint64)}
			devStats[devName] = stats
		}

		if value, err := strconv.ParseUint(entry.Value, 0, 64); err == nil {
			if value > 0 {
				stats.readLat[math.Exp2(float64(bucket))] = value
			}
		}
	}

	// FIXME: Eliminate duplicated code
	for entry := range e.writeLat.Iter() {
		devName, bucket := parseKey(entry.Key)

		stats, ok := devStats[devName]
		if !ok {
			// First time seeing this device, initialize new latency map
			stats = bioStats{make(map[float64]uint64), make(map[float64]uint64)}
			devStats[devName] = stats
		}

		if value, err := strconv.ParseUint(entry.Value, 0, 64); err == nil {
			if value > 0 {
				stats.writeLat[math.Exp2(float64(bucket))] = value
			}
		}
	}

	// Clear table - depends on https://github.com/iovisor/gobpf/pull/91 because
	// table.Delete() does not seem to handle strings in the key.
	if err := e.readLat.DeleteAll(); err != nil {
		fmt.Println(err)
	}
	if err := e.writeLat.DeleteAll(); err != nil {
		fmt.Println(err)
	}

	// Walk devStats map and emit metrics to channel
	for devName, stats := range devStats {
		emit := func(buckets map[float64]uint64, reqOp string) {
			var sampleCount uint64
			var sampleSum float64

			for k, v := range buckets {
				sampleSum += float64(k) * float64(v)
				sampleCount += v
			}

			ch <- prometheus.MustNewConstHistogram(e.latency,
				sampleCount,
				sampleSum,
				buckets,
				devName, reqOp,
			)
		}

		emit(stats.readLat, "read")
		emit(stats.writeLat, "write")
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.latency
}

// parseKey parses a BPF hash key as created by the BPF program
func parseKey(s string) (string, uint64) {
	fields := strings.Fields(strings.Trim(s, "{ }"))
	label := strings.Trim(fields[0], "\"")
	bucket, _ := strconv.ParseUint(fields[1], 0, 64)
	return label, bucket
}
