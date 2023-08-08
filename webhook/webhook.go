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
		logger.Error("unable to parse request body as Event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.processEvents(events); err != nil {
		logger.Error("unable to process events")
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

		if err := h.processEvent(event); err != nil {
			logger.Error(err.Error())
			h.Metrics["PageFailures"].With(
				prometheus.Labels{
					"subsystem": h.Subsystem,
					"handler":   handler,
					"event":     event.Event,
				}).Inc()
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		return fmt.Errorf("multiple (%d) errors were encountered", len(errors))
	}

	return nil
}

// processEvent is a method that processes one Updown event
func (h *Handlers) processEvent(event updown.Event) error {
	handler := "processEvent"
	logger := h.Logger.With("handler", handler)

	// Because Event includes a union of types: Check, Downtime, SSL
	// Necessary to validate that the event has the expected type
	switch event.Event {
	case "check.down":
		// Expects Downtime
		if event.Downtime == (updown.Downtime{}) {
			return fmt.Errorf("expected 'check.down` event to contain 'downtime'")
		}

		logger.Info("Received",
			"downtime", event.Downtime,
		)
	case "check.up":
		// Expects Downtime
		if event.Downtime == (updown.Downtime{}) {
			return fmt.Errorf("expected 'check.up` event to contain 'downtime'")
		}

		logger.Info("Received",
			"downtime", event.Downtime,
		)
	case "check.ssl_invalid":
		// Expects SSL
		if event.SSL == (updown.SSL{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl'")
		}

		// Expects Cert
		if event.SSL.Cert == (updown.Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}

		// Expects Error
		if event.SSL.Error == "" {
			return fmt.Errorf("expected 'check_ssl_invalid' to contain `ssl.error'")
		}

		logger.Info("Received",
			"cert", event.SSL.Cert,
			"error", event.SSL.Error,
		)
	case "check.ssl_valid":
		// Expects SSL
		if event.SSL == (updown.SSL{}) {
			return fmt.Errorf("expected 'check.ssl_valid' to contain 'ssl'")
		}

		// Expects Cert
		if event.SSL.Cert == (updown.Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}

		logger.Info("Received",
			"cert", event.SSL.Cert,
		)
	case "check.ssl_expiration":
		// Expects SSL
		if event.SSL == (updown.SSL{}) {
			return fmt.Errorf("expected 'check.ssl_expiration' to contain 'ssl'")
		}

		// Expects Cert
		if event.SSL.Cert == (updown.Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}

		// Expects DaysBeforeExpiration
		// Default numerical value is zero which is a valid value

		logger.Info("Received",
			"cert", event.SSL.Cert,
		)
	case "check.ssl_renewed":
		// Expects SSL
		if event.SSL == (updown.SSL{}) {
			return fmt.Errorf("expected 'check.ssl_renewed' to contain 'ssl'")
		}

		// Expects NewCert+OldCert
		if event.SSL.NewCert == (updown.Cert{}) && event.SSL.OldCert == (updown.Cert{}) {
			return fmt.Errorf("expected 'check.ssl_renewed' to contain 'ssl.new_cert' and 'ssl.old_cert'")
		}

		logger.Info("Received",
			"new_cert", event.SSL.NewCert,
			"old_cert", event.SSL.OldCert,
		)
	case "check.performance_drop":
		// Expects ApdexDropped
		if event.ApdexDropped == "" {
			return fmt.Errorf("expected 'check.performance_drop' to contain 'apdex_dropped'")
		}

		// Expects LastMetrics
		if event.LastMetrics == nil {
			return fmt.Errorf("expected 'check.performance_drop' to contain 'last_metrics'")
		}

		logger.Info("Received",
			"apdex_dropped", event.ApdexDropped,
			"last_metrics", event.LastMetrics,
		)
	default:
		return fmt.Errorf("unexpected event type")
	}

	return nil
}
