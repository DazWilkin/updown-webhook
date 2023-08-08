package updown

import (
	"fmt"
	"time"
)

// Push API (webhooks)
// https://updown.io/api
// Each Webhook POST may contain multiple Events

// Events is a type that represents multiple Event
type Events []Event

// Event is a type that represents an individual Webhook event
type Event struct {
	Event       string    `json:"event"`
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	Check       Check     `json:"check,omitempty"`

	// Only present with event 'check.performance_drop'
	ApdexDropped string `json:"apdex_dropped,omitempty"`

	// Union type
	Downtime    Downtime             `json:"downtime,omitempty"`
	SSL         SSL                  `json:"ssl,omitempty"`
	LastMetrics map[time.Time]Metric `json:"last_metrics,omitempty"`
}

func (e Event) Validate() error {
	switch e.Event {
	case "check.down":
		// Expects Downtime
		if e.Downtime == (Downtime{}) {
			return fmt.Errorf("expected 'check.down` event to contain 'downtime'")
		}
	case "check.up":
		// Expects Downtime
		if e.Downtime == (Downtime{}) {
			return fmt.Errorf("expected 'check.up` event to contain 'downtime'")
		}
	case "check.ssl_invalid":
		// Expects SSL
		if e.SSL == (SSL{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl'")
		}

		// Expects Cert
		if e.SSL.Cert == (Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}

		// Expects Error
		if e.SSL.Error == "" {
			return fmt.Errorf("expected 'check_ssl_invalid' to contain `ssl.error'")
		}
	case "check.ssl_valid":
		// Expects SSL
		if e.SSL == (SSL{}) {
			return fmt.Errorf("expected 'check.ssl_valid' to contain 'ssl'")
		}

		// Expects Cert
		if e.SSL.Cert == (Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}
	case "check.ssl_expiration":
		// Expects SSL
		if e.SSL == (SSL{}) {
			return fmt.Errorf("expected 'check.ssl_expiration' to contain 'ssl'")
		}

		// Expects Cert
		if e.SSL.Cert == (Cert{}) {
			return fmt.Errorf("expected 'check.ssl_invalid' to contain 'ssl.cert'")
		}

		// Expects DaysBeforeExpiration
		// Default numerical value is zero which is a valid value
	case "check.ssl_renewed":
		// Expects SSL
		if e.SSL == (SSL{}) {
			return fmt.Errorf("expected 'check.ssl_renewed' to contain 'ssl'")
		}

		// Expects NewCert+OldCert
		if e.SSL.NewCert == (Cert{}) && e.SSL.OldCert == (Cert{}) {
			return fmt.Errorf("expected 'check.ssl_renewed' to contain 'ssl.new_cert' and 'ssl.old_cert'")
		}
	case "check.performance_drop":
		// Expects ApdexDropped
		if e.ApdexDropped == "" {
			return fmt.Errorf("expected 'check.performance_drop' to contain 'apdex_dropped'")
		}

		// Expects LastMetrics
		if e.LastMetrics == nil {
			return fmt.Errorf("expected 'check.performance_drop' to contain 'last_metrics'")
		}
	default:
		return fmt.Errorf("unexpected event type")
	}

	return nil
}
