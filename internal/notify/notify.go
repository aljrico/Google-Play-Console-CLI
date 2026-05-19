package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Payload struct {
	Title    string  `json:"title,omitempty"`
	Message  string  `json:"message"`
	Severity string  `json:"severity,omitempty"`
	Fields   []Field `json:"fields,omitempty"`
}

type SendOptions struct {
	WebhookURL string   `json:"webhookUrl"`
	Title      string   `json:"title,omitempty"`
	Message    string   `json:"message"`
	Severity   string   `json:"severity,omitempty"`
	Fields     []string `json:"fields,omitempty"`
	Confirm    bool     `json:"confirm"`
	DryRun     bool     `json:"dryRun"`
}

type SendResult struct {
	Webhook    string  `json:"webhook"`
	Confirm    bool    `json:"confirm"`
	DryRun     bool    `json:"dryRun"`
	Delivered  bool    `json:"delivered"`
	StatusCode int     `json:"statusCode,omitempty"`
	Payload    Payload `json:"payload"`
}

type Sender interface {
	Send(ctx context.Context, webhookURL string, payload Payload) (int, error)
}

type WebhookSender struct {
	Client *http.Client
}

func Send(ctx context.Context, sender Sender, options SendOptions) (SendResult, error) {
	if err := options.Validate(); err != nil {
		return SendResult{}, err
	}
	payload, err := options.Payload()
	if err != nil {
		return SendResult{}, err
	}
	result := SendResult{
		Webhook:   RedactedURL(options.WebhookURL),
		Confirm:   options.Confirm,
		DryRun:    options.DryRun,
		Delivered: false,
		Payload:   payload,
	}
	if options.DryRun {
		return result, nil
	}
	if sender == nil {
		sender = WebhookSender{}
	}
	statusCode, err := sender.Send(ctx, options.WebhookURL, payload)
	result.StatusCode = statusCode
	if err != nil {
		return result, err
	}
	result.Delivered = true
	return result, nil
}

func (o SendOptions) Validate() error {
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("notify send requires --confirm or --dry-run")
	}
	if strings.TrimSpace(o.WebhookURL) == "" {
		return fmt.Errorf("webhook URL is required")
	}
	parsedURL, err := url.Parse(o.WebhookURL)
	if err != nil {
		return fmt.Errorf("parse webhook URL: %w", err)
	}
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("webhook URL must use http or https")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL host is required")
	}
	if strings.TrimSpace(o.Message) == "" {
		return fmt.Errorf("message is required")
	}
	if strings.TrimSpace(o.Title) != o.Title {
		return fmt.Errorf("title cannot have leading or trailing whitespace")
	}
	if strings.TrimSpace(o.Message) != o.Message {
		return fmt.Errorf("message cannot have leading or trailing whitespace")
	}
	if strings.TrimSpace(o.Severity) != o.Severity {
		return fmt.Errorf("severity cannot have leading or trailing whitespace")
	}
	return nil
}

func (o SendOptions) Payload() (Payload, error) {
	fields := make([]Field, 0, len(o.Fields))
	for _, rawField := range o.Fields {
		field, err := ParseField(rawField)
		if err != nil {
			return Payload{}, err
		}
		fields = append(fields, field)
	}
	return Payload{
		Title:    o.Title,
		Message:  o.Message,
		Severity: o.Severity,
		Fields:   fields,
	}, nil
}

func ParseField(rawField string) (Field, error) {
	name, value, ok := strings.Cut(rawField, "=")
	if !ok {
		return Field{}, fmt.Errorf("field %q must use name=value", rawField)
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" {
		return Field{}, fmt.Errorf("field name is required")
	}
	if value == "" {
		return Field{}, fmt.Errorf("field %q value is required", name)
	}
	return Field{Name: name, Value: value}, nil
}

func (s WebhookSender) Send(ctx context.Context, webhookURL string, payload Payload) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("encode notification payload: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create notification request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send notification webhook: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.StatusCode, fmt.Errorf("notification webhook returned status %d", response.StatusCode)
	}
	return response.StatusCode, nil
}

func RedactedURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if parsedURL.User != nil {
		parsedURL.User = url.User("redacted")
	}
	if parsedURL.RawQuery != "" {
		parsedURL.RawQuery = "redacted=true"
	}
	return parsedURL.String()
}
