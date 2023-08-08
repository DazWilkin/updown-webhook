package updown

import (
	"time"
)

// Push API (webhooks)
// https://updown.io/api
// Each Webhook POST may contain multiple Events

// Downtime is a type that represents the downtime of an Updown check
type Downtime struct {
	ID        string      `json:"id"`
	Error     string      `json:"error"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   time.Time   `json:"ended_at"`
	Duration  int         `json:"duration"`
	Partial   interface{} `json:"partial"` // TODO(dazwilkin) What type is this?
}

// SSL is a type that represents an Updown SSL record
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

// Cert is a type that represents an Updown SSL certificate
type Cert struct {
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	From      time.Time `json:"from"`
	To        time.Time `json:"To"`
	Algorithm string    `json:"algorithm"`
}

// Metric is a type that represents an Apdex metric
type Metric struct {
	Apdex float32 `json:"apdex"`
}
