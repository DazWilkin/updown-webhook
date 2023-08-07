package webhook

import (
	"encoding/json"
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

		// Because Event includes a union of types: Check, Downtime, SSL
		// Necessary to validate that the event has the expected type
		switch event.Event {
		case "check.down":
			// Expects Downtime
			if event.Downtime == (updown.Downtime{}) {
				logger.Error("expected 'check.down` event to contain 'downtime'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"downtime", event.Downtime,
			)
		case "check.up":
			// Expects Downtime
			if event.Downtime == (updown.Downtime{}) {
				logger.Error("expected 'check.up` event to contain 'downtime'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"downtime", event.Downtime,
			)
		case "check.ssl_invalid":
			// Expects SSL
			if event.SSL == (updown.SSL{}) {
				logger.Error("expected 'check.ssl_invalid' to contain 'ssl'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects Cert
			if event.SSL.Cert == (updown.Cert{}) {
				logger.Error("expected 'check.ssl_invalid' to contain 'ssl.cert'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects Error
			if event.SSL.Error == "" {
				logger.Error("expected 'check_ssl_invalid' to contain `ssl.error'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"cert", event.SSL.Cert,
				"error", event.SSL.Error,
			)
		case "check.ssl_valid":
			// Expects SSL
			if event.SSL == (updown.SSL{}) {
				logger.Error("expected 'check.ssl_valid' to contain 'ssl'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects Cert
			if event.SSL.Cert == (updown.Cert{}) {
				logger.Error("expected 'check.ssl_invalid' to contain 'ssl.cert'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"cert", event.SSL.Cert,
			)
		case "check.ssl_expiration":
			// Expects SSL
			if event.SSL == (updown.SSL{}) {
				logger.Error("expected 'check.ssl_expiration' to contain 'ssl'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects Cert
			if event.SSL.Cert == (updown.Cert{}) {
				logger.Error("expected 'check.ssl_invalid' to contain 'ssl.cert'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects DaysBeforeExpiration
			// Default numerical value is zero which is a valid value

			logger.Info("Received",
				"cert", event.SSL.Cert,
			)
		case "check.ssl_renewed":
			// Expects SSL
			if event.SSL == (updown.SSL{}) {
				logger.Error("expected 'check.ssl_renewed' to contain 'ssl'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects NewCert+OldCert
			if event.SSL.NewCert == (updown.Cert{}) && event.SSL.OldCert == (updown.Cert{}) {
				logger.Error("expected 'check.ssl_renewed' to contain 'ssl.new_cert' and 'ssl.old_cert'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"new_cert", event.SSL.NewCert,
				"old_cert", event.SSL.OldCert,
			)
		case "check.performance_drop":
			// Expects ApdexDropped
			if event.ApdexDropped == "" {
				logger.Error("expected 'check.performance_drop' to contain 'apdex_dropped'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Expects LastMetrics
			if event.LastMetrics == nil {
				logger.Error("expected 'check.performance_drop' to contain 'last_metrics'")
				h.Metrics["PageTotal"].With(
					prometheus.Labels{
						"subsystem": h.Subsystem,
						"handler":   handler,
						"event":     event.Event,
					}).Inc()
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logger.Info("Received",
				"apdex_dropped", event.ApdexDropped,
				"last_metrics", event.LastMetrics,
			)
		default:
			logger.Error("unexpected event type")
		}
	}
}
