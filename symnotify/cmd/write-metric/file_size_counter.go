package main

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

// NewFileSizeCounter creates a CounterFunc to track the size of a file
func NewFileSizeCounter(filename string, opts prometheus.CounterOpts) prometheus.CounterFunc {
	return prometheus.NewCounterFunc(opts, func() float64 {
		stat, err := os.Stat(filename)
		if err == nil && !stat.IsDir() {
			return float64(stat.Size())
		}
		return 0
	})
}

type FileSizeCounterVec struct {
	*prometheus.MetricVec
}

// TODO For now we use "filename" as the sole label, may need review.
var labelNames = []string{"filename"}

func NewFileSizeCounterVec(opts prometheus.CounterOpts) *FileSizeCounterVec {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		labelNames,
		opts.ConstLabels,
	)
	return &FileSizeCounterVec{
		MetricVec: prometheus.NewMetricVec(desc,
			func(lvs ...string) prometheus.Metric {
				if len(lvs) != len(labelNames) {
					panic(fmt.Errorf("Inconsistent cardinality: %v, %v", desc.String(), lvs))
				}
				return NewFileSizeCounter(lvs[0], opts)
			}),
	}
}
