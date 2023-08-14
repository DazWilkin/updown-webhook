package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/DazWilkin/updown-webhook/updown"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	updownUserAgent string = "updown.io"
	updownWhitelist string = "ips.updown.io"
)

// Handlers is a type that represents a webhook handler for Updown
type Handlers struct {
	Subsystem string
	IPs       []net.IP
	Metrics   map[string]*prometheus.CounterVec
	Logger    *slog.Logger
}

// Handler is a function that returns new Handlers
func Handler(
	subsystem string,
	metrics map[string]*prometheus.CounterVec,
	logger *slog.Logger,
) (*Handlers, error) {
	// On initialization enumerate valid updown.io IPs
	ips, err := net.LookupIP(updownWhitelist)
	if err != nil {
		msg := "unable to enumerate updown.io IPs"
		logger.Error(msg, "error", err)
		return &Handlers{}, err
	}

	logger.Info("Caching updown.io IP whitelist",
		"ips", ips,
	)

	return &Handlers{
		Subsystem: subsystem,
		IPs:       ips,
		Metrics:   metrics,
		Logger:    logger,
	}, nil
}

func (h *Handlers) validUserAgent(values []string) bool {
	for _, value := range values {
		if strings.Contains(value, updownUserAgent) {
			return true
		}
	}
	return false
}
func (h *Handlers) permittedIP(ip net.IP) bool {
	for _, x := range h.IPs {
		if x.Equal(ip) {
			return true
		}
	}
	return false
}

// ServeHTTP is a method that implements the http.Handler interface
// Calling the (sole) method ServeHTTP permits using http.Handle("/path",Handler())
func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := "webhook"
	logger := h.Logger.With("handler", handler)

	// Debug: enumerate headers
	logger.Info("Debugging",
		"headers", r.Header,
	)

	// Must include User-Agent header for updown.io
	if values, ok := r.Header["User-Agent"]; !ok {
		logger.Error("expected 'User-Agent' header")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		if !h.validUserAgent(values) {
			logger.Error(
				fmt.Sprintf("expected User-Agent header to contain %s", updownUserAgent),
				"values", values,
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Must include X-Forward-For header with permitted IP
	if values, ok := r.Header["X-Forwarded-For"]; !ok {
		logger.Error("expected 'X-Forwarded-For' header")
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		// Original client should be first address (if any)
		if len(values) == 0 {
			// X-Forwarded-For exists but is empty
			logger.Error("expected 'X-Forwarded-For' header to contain at least one value")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// There are multiple X-Forward-For values
		// First should be (!) the originating (updown.io) host
		ip := net.ParseIP(values[0])
		if !h.permittedIP(ip) {
			logger.Error("unable to match IP to permitted IP list",
				"ip", ip,
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Webhook must be POST'ed
	if r.Method != "POST" {
		logger.Error("unexpected method",
			"method", r.Method,
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("unable to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var events updown.Events
	if err := json.Unmarshal(body, &events); err != nil {
		logger.Error("unable to parse request body as Event",
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.processEvents(events); err != nil {
		logger.Error("unable to process events",
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// processEvents is a method that processes multiple Updown events
func (h *Handlers) processEvents(events []updown.Event) error {
	handler := "processEvents"
	logger := h.Logger.With("handler", handler)

	// Accumulate errors
	errors := []error{}

	for _, event := range events {
		logger.Info("Event",
			"event", event.Event,
		)
		h.Metrics["PageTotal"].With(
			prometheus.Labels{
				"subsystem": h.Subsystem,
				"handler":   handler,
				"event":     event.Event,
			}).Inc()

		// If the event doesn't validate, record the error
		// and skip process'ing the invalid event
		if err := event.Validate(); err != nil {
			logger.Error(err.Error())
			h.Metrics["PageFailures"].With(
				prometheus.Labels{
					"subsystem": h.Subsystem,
					"handler":   handler,
					"event":     event.Event,
				}).Inc()
			errors = append(errors, err)
		} else {
			h.processEvent(event)
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf("%d event%s invalid",
			len(errors),
			func(i int) string {
				if i == 1 {
					return " is"
				}

				return "s are"
			}(len(errors)),
		)
	}

	return nil
}

// processEvent is a method that processes one Updown event
func (h *Handlers) processEvent(event updown.Event) {
	handler := "processEvent"
	logger := h.Logger.With("handler", handler)

	switch event.Event {
	case "check.down":
		logger.Info("Received",
			"downtime", event.Downtime,
		)
	case "check.up":
		logger.Info("Received",
			"downtime", event.Downtime,
		)
	case "check.ssl_invalid":
		logger.Info("Received",
			"cert", event.SSL.Cert,
			"error", event.SSL.Error,
		)
	case "check.ssl_valid":
		logger.Info("Received",
			"cert", event.SSL.Cert,
		)
	case "check.ssl_expiration":
		logger.Info("Received",
			"cert", event.SSL.Cert,
		)
	case "check.ssl_renewed":
		logger.Info("Received",
			"new_cert", event.SSL.NewCert,
			"old_cert", event.SSL.OldCert,
		)
	case "check.performance_drop":
		logger.Info("Received",
			"apdex_dropped", event.ApdexDropped,
			"last_metrics", event.LastMetrics,
		)
	default:
		logger.Info("unexpected event type")
	}
}
