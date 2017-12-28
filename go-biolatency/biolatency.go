package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iovisor/gobpf/bcc"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const namespace = "bio"

var (
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9123").String()
)

// log2Histogram prints a simple ASCII-based histogram to stdout
// DEPRECATED
func log2Histogram(hist []uint64, width int) {
	var (
		idxMax uint
		valMax uint64
	)

	for i, v := range hist {
		if v > 0 {
			idxMax = uint(i)
		}

		if v > valMax {
			valMax = v
		}
	}

	for i := uint(1); i <= idxMax; i++ {
		low := 1 << (i - 1)
		high := (1 << i) - 1

		if low == high {
			low -= 1
		}

		// Fill string with asterisks according to current value's proportion of max
		stars := strings.Repeat("*", int(float64(hist[i])/float64(valMax)*float64(width)))

		fmt.Printf("%20d -> %-20d : %-8d |%-*s|\n", low, high, hist[i], width, stars)
	}
}

func main() {
	allowedLevel := promlog.AllowedLevel{}
	flag.AddFlags(kingpin.CommandLine, &allowedLevel)
	kingpin.Version(version.Print("blat_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(allowedLevel)

	level.Info(logger).Log("msg", "Starting blat_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", version.BuildContext())

	// Compile BPF code and return new module
	m := bcc.NewModule(bpfSource, []string{})
	defer m.Close()

	// Load and attach kprobes
	startKprobe, err := m.LoadKprobe("trace_req_start")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load trace_req_start: %s\n", err)
		os.Exit(1)
	}

	err = m.AttachKprobe("blk_account_io_start", startKprobe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach trace_req_start: %s\n", err)
		os.Exit(1)
	}

	endKprobe, err := m.LoadKprobe("trace_req_completion")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load trace_req_completion: %s\n", err)
		os.Exit(1)
	}

	err = m.AttachKprobe("blk_account_io_completion", endKprobe)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach trace_req_completion: %s\n", err)
		os.Exit(1)
	}

	prometheus.MustRegister(newExporter(m))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
	<head>
	<title>Blat Exporter</title>
	<style>html { font-family: sans-serif; }</style>
	</head>
	<body>
	<h1>Blat Exporter</h1>
	<p><a href="/metrics">Metrics</a></p>
	</body>
</html>`))
	})
	http.Handle("/metrics", promhttp.Handler())

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
