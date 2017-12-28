package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	bpf "github.com/iovisor/gobpf/bcc"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/golang/protobuf/proto"
)

const source string = `
#include <uapi/linux/ptrace.h>
#include <linux/blkdev.h>

typedef struct disk_key {
	char disk[DISK_NAME_LEN];
	u64 slot;
} disk_key_t;

BPF_HASH(start, struct request *);
BPF_HISTOGRAM(dist, disk_key_t);

// Record start time of a request
int trace_req_start(struct pt_regs *ctx, struct request *req)
{
	u64 ts = bpf_ktime_get_ns();
	start.update(&req, &ts);
	return 0;
}

// Calculate request duration and store in appropriate histogram bucket
int trace_req_completion(struct pt_regs *ctx, struct request *req)
{
	u64 *tsp, delta;

	// Fetch timestamp and calculate delta
	tsp = start.lookup(&req);
	if (tsp == 0) {
		return 0;   // missed issue
	}
	delta = bpf_ktime_get_ns() - *tsp;

	// Convert to microseconds
	delta /= 1000;

	// Store as histogram
	disk_key_t key = {.slot = bpf_log2l(delta)};
	bpf_probe_read(&key.disk, sizeof(key.disk), req->rq_disk->disk_name);
	dist.increment(key);

	start.delete(&req);
	return 0;
}
`

var (
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9123").String()
)

// parseKey parses a BPF hash key as created by the BPF program
func parseKey(s string) (string, uint64) {
	fields := strings.Fields(strings.Trim(s, "{ }"))
	label := strings.Trim(fields[0], "\"")
	bucket, _ := strconv.ParseUint(fields[1], 0, 64)
	return label, bucket
}

// log2Histogram prints a simple ASCII-based histogram to stdout
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

	prometheus.MustRegister(newExporter())

	http.Handle("/metrics", promhttp.Handler())

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}

	m := bpf.NewModule(source, []string{})
	defer m.Close()

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

	table := bpf.NewTable(m.TableId("dist"), m)

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	desc := prometheus.NewDesc(
		"http_request_duration_seconds",
		"A histogram of the HTTP request durations.",
		[]string{"device", "operation"},
		nil,
	)

	for t := range ticker.C {
		hist := make([]uint64, 64)

		// TODO: We're currently ignoring label (i.e., block dev name)
		for entry := range table.Iter() {
			_, bucket := parseKey(entry.Key)

			if value, err := strconv.ParseUint(entry.Value, 0, 64); err == nil {
				hist[bucket] = value
			}
		}

		// Clear table - depends on https://github.com/iovisor/gobpf/pull/91 because
		// table.Delete() does not seem to handle strings in the key.
		if err := table.DeleteAll(); err != nil {
			fmt.Println(err)
		}

		fmt.Println(t)
		log2Histogram(hist, 40)
		fmt.Println()

		buckets := make(map[float64]uint64)
		var sampleCount uint64
		var sampleSum float64

		for i, v := range hist {
			if v > 0 {
				buckets[math.Exp2(float64(i))] = v
				sampleCount += v
				sampleSum += float64(v) * float64(i) // FIXME: This is not correct
			}
		}

		promHist := prometheus.MustNewConstHistogram(desc,
			sampleCount,
			sampleSum,
			buckets,
			"sda1", "read", // Dummy values, to be filled in later
		)

		metric := &dto.Metric{}
		promHist.Write(metric)
		fmt.Println(proto.MarshalTextString(metric))
	}
}
