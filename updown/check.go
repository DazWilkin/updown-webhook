package updown

import "time"

// Check is a type that represents an Updown check
type Check struct {
	Token             string              `json:"token"`
	URL               string              `json:"url"`
	Alias             string              `json:"alias"`
	LastStatus        uint16              `json:"last_status"`
	Uptime            float32             `json:"uptime"`
	Down              bool                `json:"down"`
	DownSince         time.Time           `json:"down_since"`
	Error             string              `json:"error"`
	Period            uint16              `json:"period"`
	Apdex             float32             `json:"apdex_t"`
	StringMatch       string              `json:"string_match"`
	Enabled           bool                `json:"enabled"`
	Published         bool                `json:"published"`
	DisabledLocations []string            `json:"disabled_locations"`
	Recipients        []string            `json:"recipients"`
	LastCheckAt       time.Time           `json:"last_check_at"`
	NextCheckAt       time.Time           `json:"next_check_at"`
	MuteUntil         time.Time           `json:"mute_until"`
	FaviconURL        string              `json:"favicon_url"`
	CustomHeaders     map[string][]string `json:"custom_headers"`
	HTTPVerb          string              `json:"http_verb"`
	HTTPBody          string              `json:"http_body"`
}
