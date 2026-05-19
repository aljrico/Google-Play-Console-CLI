package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second
const DefaultWebhookURLEnv = "GPC_NOTIFY_WEBHOOK_URL"

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

type SlackPayload struct {
	Text string `json:"text"`
}

type SendOptions struct {
	CommandPath    string   `json:"-"`
	WebhookURL     string   `json:"webhookUrl,omitempty"`
	WebhookURLEnv  string   `json:"webhookUrlEnv,omitempty"`
	WebhookURLFile string   `json:"webhookUrlFile,omitempty"`
	Title          string   `json:"title,omitempty"`
	Message        string   `json:"message"`
	Severity       string   `json:"severity,omitempty"`
	Fields         []string `json:"fields,omitempty"`
	Confirm        bool     `json:"confirm"`
	DryRun         bool     `json:"dryRun"`
}

type SendResult struct {
	Webhook    string  `json:"webhook"`
	Confirm    bool    `json:"confirm"`
	DryRun     bool    `json:"dryRun"`
	Delivered  bool    `json:"delivered"`
	StatusCode int     `json:"statusCode,omitempty"`
	Payload    Payload `json:"payload"`
}

type SlackSendResult struct {
	Webhook    string       `json:"webhook"`
	Confirm    bool         `json:"confirm"`
	DryRun     bool         `json:"dryRun"`
	Delivered  bool         `json:"delivered"`
	StatusCode int          `json:"statusCode,omitempty"`
	Payload    SlackPayload `json:"payload"`
}

type Sender interface {
	Send(ctx context.Context, webhookURL string, payload Payload) (int, error)
}

type SlackSender interface {
	SendSlack(ctx context.Context, webhookURL string, payload SlackPayload) (int, error)
}

type WebhookSender struct {
	Client *http.Client
}

func Send(ctx context.Context, sender Sender, options SendOptions) (SendResult, error) {
	resolvedURL, err := options.ResolvedWebhookURL()
	if err != nil {
		return SendResult{}, err
	}
	if err := options.ValidateWebhookURL(resolvedURL); err != nil {
		return SendResult{}, err
	}
	payload, err := options.Payload()
	if err != nil {
		return SendResult{}, err
	}
	result := SendResult{
		Webhook:   RedactedURL(resolvedURL),
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
	statusCode, err := sender.Send(ctx, resolvedURL, payload)
	result.StatusCode = statusCode
	if err != nil {
		return result, err
	}
	result.Delivered = true
	return result, nil
}

func SendSlack(ctx context.Context, sender SlackSender, options SendOptions) (SlackSendResult, error) {
	resolvedURL, err := options.ResolvedWebhookURL()
	if err != nil {
		return SlackSendResult{}, err
	}
	if err := options.ValidateWebhookURL(resolvedURL); err != nil {
		return SlackSendResult{}, err
	}
	payload, err := options.SlackPayload()
	if err != nil {
		return SlackSendResult{}, err
	}
	result := SlackSendResult{
		Webhook:   RedactedURL(resolvedURL),
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
	statusCode, err := sender.SendSlack(ctx, resolvedURL, payload)
	result.StatusCode = statusCode
	if err != nil {
		return result, err
	}
	result.Delivered = true
	return result, nil
}

func (o SendOptions) ResolvedWebhookURL() (string, error) {
	if o.Confirm && o.DryRun {
		return "", fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return "", fmt.Errorf("%s requires --confirm or --dry-run", o.commandPath())
	}
	if strings.TrimSpace(o.WebhookURLFile) != o.WebhookURLFile {
		return "", fmt.Errorf("webhook URL file cannot have leading or trailing whitespace")
	}
	webhookURLEnv := o.webhookURLEnv()
	if strings.TrimSpace(webhookURLEnv) != webhookURLEnv {
		return "", fmt.Errorf("webhook URL environment variable cannot have leading or trailing whitespace")
	}
	if o.WebhookURL != "" && o.WebhookURLFile != "" {
		return "", fmt.Errorf("--webhook-url and --webhook-url-file cannot be used together")
	}
	switch {
	case o.WebhookURL != "":
		return strings.TrimSpace(o.WebhookURL), nil
	case o.WebhookURLFile != "":
		content, err := os.ReadFile(o.WebhookURLFile)
		if err != nil {
			return "", fmt.Errorf("read webhook URL file: %w", err)
		}
		return strings.TrimSpace(string(content)), nil
	case webhookURLEnv != "":
		return strings.TrimSpace(os.Getenv(webhookURLEnv)), nil
	default:
		return "", fmt.Errorf("webhook URL is required")
	}
}

func (o SendOptions) ValidateWebhookURL(webhookURL string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook URL is required")
	}
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("parse webhook URL: %s", RedactedURL(webhookURL))
	}
	if parsedURL.Scheme != "https" {
		if parsedURL.Scheme != "http" || !isLoopbackHost(parsedURL.Hostname()) {
			return fmt.Errorf("webhook URL must use https unless the host is loopback")
		}
	}
	if parsedURL.User != nil {
		return fmt.Errorf("webhook URL must not contain userinfo")
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

func (o SendOptions) commandPath() string {
	if o.CommandPath == "" {
		return "notify send"
	}
	return o.CommandPath
}

func (o SendOptions) webhookURLEnv() string {
	if o.WebhookURLEnv == "" {
		return DefaultWebhookURLEnv
	}
	return o.WebhookURLEnv
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

func (o SendOptions) SlackPayload() (SlackPayload, error) {
	payload, err := o.Payload()
	if err != nil {
		return SlackPayload{}, err
	}
	return SlackPayload{Text: SlackText(payload)}, nil
}

func SlackText(payload Payload) string {
	var builder strings.Builder
	if payload.Title != "" {
		builder.WriteString("*")
		builder.WriteString(payload.Title)
		builder.WriteString("*\n")
	}
	builder.WriteString(payload.Message)
	if payload.Severity != "" {
		builder.WriteString("\nSeverity: ")
		builder.WriteString(payload.Severity)
	}
	for _, field := range payload.Fields {
		builder.WriteString("\n")
		builder.WriteString(field.Name)
		builder.WriteString(": ")
		builder.WriteString(field.Value)
	}
	return builder.String()
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
	return s.sendJSON(ctx, webhookURL, payload)
}

func (s WebhookSender) SendSlack(ctx context.Context, webhookURL string, payload SlackPayload) (int, error) {
	return s.sendJSON(ctx, webhookURL, payload)
}

func (s WebhookSender) sendJSON(ctx context.Context, webhookURL string, payload any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("encode notification payload: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create notification request for %s", RedactedURL(webhookURL))
	}
	request.Header.Set("Content-Type", "application/json")
	client := noRedirectClient(s.Client)
	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send notification webhook to %s: %s", RedactedURL(webhookURL), RedactedError(webhookURL, err))
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
		return "redacted-url"
	}
	redactedURL := &url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host}
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		redactedURL.Path = "/redacted"
	}
	if parsedURL.RawQuery != "" {
		redactedURL.RawQuery = "redacted=true"
	}
	if parsedURL.Fragment != "" {
		redactedURL.Fragment = "redacted"
	}
	return redactedURL.String()
}

func RedactedError(rawURL string, err error) string {
	if err == nil {
		return ""
	}
	var urlError *url.Error
	if errors.As(err, &urlError) {
		return redactText(rawURL, urlError.Err.Error())
	}
	return redactText(rawURL, err.Error())
}

func redactText(rawURL string, text string) string {
	text = strings.ReplaceAll(text, rawURL, RedactedURL(rawURL))
	if parsedURL, parseErr := url.Parse(rawURL); parseErr == nil {
		text = strings.ReplaceAll(text, parsedURL.String(), RedactedURL(rawURL))
	}
	return text
}

func noRedirectClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &copyClient
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
