// Package slack posts messages to a Slack incoming webhook.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Notifier sends notifications. A nil or unconfigured notifier is a no-op.
type Notifier struct {
	WebhookURL string
	HTTP       *http.Client
}

// Enabled reports whether messages are actually sent.
func (n *Notifier) Enabled() bool { return n != nil && n.WebhookURL != "" }

// Send posts text (Slack mrkdwn).
func (n *Notifier) Send(ctx context.Context, text string) error {
	if !n.Enabled() {
		return nil
	}
	client := n.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	body, err := json.Marshal(map[string]any{"text": text, "unfurl_links": false})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("slack webhook: %d %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

// Escape escapes text for Slack mrkdwn.
func Escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// Link formats a mrkdwn link.
func Link(url, text string) string {
	return "<" + url + "|" + Escape(text) + ">"
}
