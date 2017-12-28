package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/iovisor/gobpf/bcc"

	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	bpfMod  *bcc.Module
	bpfHist *bcc.Table
	latency *prometheus.Desc
}

func newExporter(m *bcc.Module) *exporter {
	e := exporter{
		bpfMod:  m,
		bpfHist: bcc.NewTable(m.TableId("dist"), m),
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

	buckets := make(map[float64]uint64)

	// TODO: We're currently ignoring label (i.e., block dev name)
	for entry := range e.bpfHist.Iter() {
		_, bucket := parseKey(entry.Key)

		if value, err := strconv.ParseUint(entry.Value, 0, 64); err == nil {
			if value > 0 {
				buckets[math.Exp2(float64(bucket))] = value
			}
		}
	}

	// Clear table - depends on https://github.com/iovisor/gobpf/pull/91 because
	// table.Delete() does not seem to handle strings in the key.
	if err := e.bpfHist.DeleteAll(); err != nil {
		fmt.Println(err)
	}

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
		"sda1", "read", // Dummy values, to be filled in later
	)
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
