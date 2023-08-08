package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DazWilkin/updown-webhook/updown"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
)

// Handlers is a type that represents a webhook handler for Updown
type Handlers struct {
	Subsystem string
	Metrics   map[string]*prometheus.CounterVec
	Logger    *slog.Logger
}

// Handler is a function that returns new Handlers
func Handler(
	subsystem string,
	metrics map[string]*prometheus.CounterVec,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		Subsystem: subsystem,
		Metrics:   metrics,
		Logger:    logger,
	}
}

// ServeHTTP is a method that implements the http.Handler interface
// Calling the (sole) method ServeHTTP permits using http.Handle("/path",Handler())
func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := "webhook"
	logger := h.Logger.With("handler", handler)

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
