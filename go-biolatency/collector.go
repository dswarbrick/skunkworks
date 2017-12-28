package main

import (
	"github.com/iovisor/gobpf/bcc"

	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	bpfMod  *bcc.Module
	latency *prometheus.Desc
}

func newExporter(m *bcc.Module) *exporter {
	e := exporter{
		bpfMod: m,
		latency: prometheus.NewDesc(
			"bio_request_latency_usec",
			"A histogram of bio request latencies in microseconds.",
			[]string{"device", "operation"},
			nil,
		),
	}

	return &e
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	buckets := make(map[float64]uint64)

	ch <- prometheus.MustNewConstHistogram(e.latency,
		1234,
		5678,
		buckets,
		"sda1", "read", // Dummy values, to be filled in later
	)
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.latency
}
