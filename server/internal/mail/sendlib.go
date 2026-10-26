package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the Sendlib REST endpoint for transactional sends.
const DefaultBaseURL = "https://sendlib.samueltuoyo.com/api/send"

// Client sends transactional Email through the Sendlib REST API:
//
//	curl -X POST https://sendlib.samueltuoyo.com/api/send \
//	  -H "Authorization: Bearer YOUR_API_KEY" \
//	  -H "Content-Type: application/json" \
//	  -d '{"from":"sender@gmail.com","to":"recipient@example.com",
//	      "subject":"Welcome to Sendlib!",
//	      "html":"<p>This email was sent via Sendlib REST API.</p>"}'
type Client struct {
	apiKey  string
	baseURL string
	from    string
	http    *http.Client
}

// NewClient wires a Sendlib client. Empty apiKey disables sending:
// Send returns nil immediately so dev/test keep working on codes alone.
func NewClient(apiKey, baseURL, from string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		from:    from,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Enabled reports whether sends actually hit the network.
func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != "" && c.from != ""
}

type sendRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// Send delivers one HTML message. When disabled it succeeds silently.
func (c *Client) Send(ctx context.Context, to, subject, html string) error {
	if !c.Enabled() {
		return nil
	}
	body, err := json.Marshal(sendRequest{From: c.from, To: to, Subject: subject, HTML: html})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Best-effort: detail only enriches the error; the status code alone
		// is enough to report the failure.
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("sendlib send failed: status %d: %s", resp.StatusCode, string(detail))
	}
	return nil
}

// SendVerificationCode delivers the Email-verification code.
func (c *Client) SendVerificationCode(ctx context.Context, to, code string) error {
	return c.Send(ctx, to, "Verify your Chirp Email", VerifyEmailHTML(code))
}

// SendPasswordResetCode delivers the password-reset code.
func (c *Client) SendPasswordResetCode(ctx context.Context, to, code string) error {
	return c.Send(ctx, to, "Reset your Chirp password", ResetPasswordHTML(code))
}
