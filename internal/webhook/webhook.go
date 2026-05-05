package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	JobName   string    `json:"job_name"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exit_code"`
	Duration  float64   `json:"duration_seconds"`
	Output    string    `json:"output,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Notifier sends webhook notifications.
type Notifier struct {
	URL    string
	Client *http.Client
}

// New creates a Notifier with a sensible default HTTP client timeout.
func New(url string) *Notifier {
	return &Notifier{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals p and POSTs it to the configured URL.
// It returns an error if the request fails or the server responds with a
// non-2xx status code.
func (n *Notifier) Send(p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := n.Client.Post(n.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, n.URL)
	}
	return nil
}
