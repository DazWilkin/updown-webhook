package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/DazWilkin/updown-webhook/webhook"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace string = "updown"
	subsystem string = "webhook"
)

var (
	// BuildTime is the time that this binary was built represented as a UNIX epoch
	BuildTime string
	// GitCommit is the git commit value and is expected to be set during build
	GitCommit string
	// GoVersion is the Golang runtime version
	GoVersion = runtime.Version()
	// OSVersion is the OS version (uname --kernel-release) and is expected to be set during build
	OSVersion string
	// StartTime is the start time of the exporter represented as a UNIX epoch
	StartTime = time.Now().Unix()
)

var (
	port = flag.Uint("port", 8888, "Port of HTTP server")
)

var (
	buildconfig = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "build_info",
			Namespace: namespace,
			Help:      "A metric with a constant '1' value labels by build|start time, git commit, OS and Go versions",
		}, []string{
			"subsystem",
			"build_time",
			"git_commit",
			"os_version",
			"go_version",
			"start_time",
		},
	)
	metrics = map[string]*prometheus.CounterVec{
		"PageFailures": promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "handler_failures",
				Namespace: namespace,
				Help:      "The number of times the handler has failed",
			}, []string{
				"subsystem",
				"handler",
				"event",
			},
		),
		"PageTotal": promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "handler_total",
				Namespace: namespace,
				Help:      "The number of times the handler has been invoked",
			}, []string{
				"subsystem",
				"handler",
				"event",
			},
		),
	}
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger = logger.With("function", "main")

	// Create Prometheus 'static' counter for build config
	logger.Info("Build config",
		"build_time", BuildTime,
		"git_commit", GitCommit,
		"os_version", OSVersion,
		"go_version", GoVersion,
		"start_time", strconv.FormatInt(StartTime, 10),
	)
	buildconfig.With(
		prometheus.Labels{
			"subsystem":  subsystem,
			"build_time": BuildTime,
			"git_commit": GitCommit,
			"os_version": OSVersion,
			"go_version": GoVersion,
			"start_time": strconv.FormatInt(StartTime, 10),
		}).Inc()

	flag.Parse()

	// Updown webhook
	hook, err := webhook.Handler(subsystem, metrics, logger)
	if err != nil {
		logger.Error("unable to create handlers",
			"error", err,
		)
		panic(err)
	}

	http.Handle("/", hook)

	// Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	logger.Info("Serving",
		"address", addr,
	)
	logger.Error(http.ListenAndServe(addr, nil).Error())
}
