package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alanconway/trials/symnotify/pkg/symnotify"
	"github.com/pkg/profile"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	debugOn = flag.Bool("debug", false, "enable debug messages")
	address = flag.String("addr", ":2112", "HTTP scrape address")
)

func main() {
	defer profile.Start().Stop()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [FLAGS] PATH [PATH...]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	watcher, err := symnotify.NewWatcher()
	fatalErr(err)
	defer watcher.Close()
	for _, dir := range flag.Args() {
		fatalErr(watcher.Add(dir))
	}

	metrics := NewFileSizeCounterVec(prometheus.CounterOpts{
		Name: "bytes_written",
		Help: "total bytes written to file",
	})
	fatalErr(prometheus.Register(metrics))
	debug("watching: %v", flag.Args())
	go func() {
		for {
			event, err := watcher.Event()
			if err != nil && !os.IsNotExist(err) {
				debug("watch error: %v", err)
			} else {
				debug("watch event: %v", event)
				// Write (which includes truncate) is the only operation that can change file size.
				if event.Op == symnotify.Write {
					_, err = metrics.GetMetricWithLabelValues(event.Name)
				}
			}
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	fatalErr(http.ListenAndServe(*address, nil))
}

func debug(f string, x ...interface{}) {
	if *debugOn {
		log.Printf(f, x...)
	}
}

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
