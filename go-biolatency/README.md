# bio_exporter

bio_exporter is an experimental Prometheus exporter which uses eBPF kprobes to
efficiently record a histogram of bio request latencies for each block device.
