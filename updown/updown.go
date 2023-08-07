package updown

import (
	"time"
)

// Push API (webhooks)
// https://updown.io/api

// Each Webhook POST may contain multiple Events
type Events []Event

type Event struct {
	Event       string    `json:"event"`
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	Check       Check     `json:"check,omitempty"`

	// Only present with event 'check.performance_drop'
	ApdexDropped string `json:"apdex_dropped,omitempty"`

	// Union type
	Downtime    Downtime              `json:"downtime,omitempty"`
	SSL         SSL                   `json:"ssl,omitempty"`
	LastMetrics map[time.Time]Metrics `json:"last_metrics,omitempty"`
}
type Check struct {
	Token      string `json:"token"`
	URL        string `json:"url"`
	LastStatus uint16 `json:"last_status"`
}
type Downtime struct {
	ID        string      `json:"id"`
	Error     string      `json:"error"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   time.Time   `json:"ended_at"`
	Duration  int         `json:"duration"`
	Partial   interface{} `json:"partial"` // TODO:DazWilkin What type is this?
}

// SSL is a type that represents an Updown SSL certificate
// Depending on the event type either Cert or NewCert|OldCert are present
// Cert: check.ssl_invalid, check.ssl_valid, check.ssl_expiration
// NewCert|OldCert: check.ssl_renewed
type SSL struct {
	DaysBeforeExpiration uint   `json:"days_before_expiration,omitempty"`
	Error                string `json:"error,omitemtpy"`

	Cert Cert `json:"cert,omitempty"`

	NewCert Cert `json:"new_cert,omitempty"`
	OldCert Cert `json:"old_cert,omitempty"`
}
type Cert struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	From      time.Time `json:"from"`
	To        time.Time `json:"To"`
	Algorithm string    `json:"algorithm"`
}

type Metrics struct {
	Apdex float32 `json:"apdex"`
}