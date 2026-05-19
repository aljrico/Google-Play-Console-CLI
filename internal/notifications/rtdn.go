package notifications

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type RTDNDecodeOptions struct {
	File string `json:"file,omitempty"`
	Data string `json:"data,omitempty"`
}

type PubSubEnvelope struct {
	Message      PubSubMessage `json:"message"`
	Subscription string        `json:"subscription,omitempty"`
}

type PubSubMessage struct {
	Attributes  map[string]string `json:"attributes,omitempty"`
	Data        string            `json:"data"`
	MessageID   string            `json:"messageId,omitempty"`
	PublishTime string            `json:"publishTime,omitempty"`
}

type DeveloperNotification struct {
	Version                    string                      `json:"version"`
	PackageName                string                      `json:"packageName"`
	EventTimeMillis            json.Number                 `json:"eventTimeMillis,omitempty"`
	SubscriptionNotification   *SubscriptionNotification   `json:"subscriptionNotification,omitempty"`
	OneTimeProductNotification *OneTimeProductNotification `json:"oneTimeProductNotification,omitempty"`
	VoidedPurchaseNotification *VoidedPurchaseNotification `json:"voidedPurchaseNotification,omitempty"`
	TestNotification           *TestNotification           `json:"testNotification,omitempty"`
}

type SubscriptionNotification struct {
	Version          string `json:"version"`
	NotificationType int64  `json:"notificationType"`
	PurchaseToken    string `json:"purchaseToken"`
	SubscriptionID   string `json:"subscriptionId"`
}

type OneTimeProductNotification struct {
	Version          string `json:"version"`
	NotificationType int64  `json:"notificationType"`
	PurchaseToken    string `json:"purchaseToken"`
	SKU              string `json:"sku"`
}

type VoidedPurchaseNotification struct {
	PurchaseToken string `json:"purchaseToken,omitempty"`
	OrderID       string `json:"orderId,omitempty"`
	ProductType   int64  `json:"productType,omitempty"`
	RefundType    int64  `json:"refundType,omitempty"`
}

type TestNotification struct {
	Version string `json:"version"`
}

type RTDNDecodeResult struct {
	Subscription string                `json:"subscription,omitempty"`
	MessageID    string                `json:"messageId,omitempty"`
	PublishTime  string                `json:"publishTime,omitempty"`
	Attributes   map[string]string     `json:"attributes,omitempty"`
	Kind         string                `json:"kind"`
	Notification DeveloperNotification `json:"notification"`
}

func DecodeRTDN(options RTDNDecodeOptions) (RTDNDecodeResult, error) {
	content, err := options.content()
	if err != nil {
		return RTDNDecodeResult{}, err
	}
	var envelope PubSubEnvelope
	decoder := json.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&envelope); err != nil {
		return RTDNDecodeResult{}, fmt.Errorf("parse Pub/Sub push payload: %w", err)
	}
	if strings.TrimSpace(envelope.Message.Data) == "" {
		return RTDNDecodeResult{}, fmt.Errorf("Pub/Sub message data is required")
	}
	decodedData, err := base64.StdEncoding.DecodeString(envelope.Message.Data)
	if err != nil {
		return RTDNDecodeResult{}, fmt.Errorf("decode Pub/Sub message data: %w", err)
	}
	var notification DeveloperNotification
	notificationDecoder := json.NewDecoder(bytes.NewReader(decodedData))
	notificationDecoder.UseNumber()
	if err := notificationDecoder.Decode(&notification); err != nil {
		return RTDNDecodeResult{}, fmt.Errorf("parse developer notification: %w", err)
	}
	if err := notification.Validate(); err != nil {
		return RTDNDecodeResult{}, err
	}
	kind, err := notification.Kind()
	if err != nil {
		return RTDNDecodeResult{}, err
	}
	return RTDNDecodeResult{
		Subscription: envelope.Subscription,
		MessageID:    envelope.Message.MessageID,
		PublishTime:  envelope.Message.PublishTime,
		Attributes:   sortedAttributes(envelope.Message.Attributes),
		Kind:         kind,
		Notification: notification,
	}, nil
}

func (o RTDNDecodeOptions) content() ([]byte, error) {
	switch {
	case strings.TrimSpace(o.File) != "" && strings.TrimSpace(o.Data) != "":
		return nil, fmt.Errorf("--file and --data cannot be used together")
	case strings.TrimSpace(o.File) != "":
		content, err := os.ReadFile(o.File)
		if err != nil {
			return nil, fmt.Errorf("read Pub/Sub payload file %s: %w", o.File, err)
		}
		return content, nil
	case strings.TrimSpace(o.Data) != "":
		return []byte(o.Data), nil
	default:
		return nil, fmt.Errorf("Pub/Sub payload requires --file or --data")
	}
}

func (n DeveloperNotification) Validate() error {
	if strings.TrimSpace(n.Version) == "" {
		return fmt.Errorf("developer notification version is required")
	}
	if strings.TrimSpace(n.PackageName) == "" {
		return fmt.Errorf("developer notification packageName is required")
	}
	if n.EventTimeMillis == "" {
		return fmt.Errorf("developer notification eventTimeMillis is required")
	}
	if _, err := n.EventTimeMillis.Int64(); err != nil {
		return fmt.Errorf("developer notification eventTimeMillis must be an integer: %w", err)
	}
	return nil
}

func (n DeveloperNotification) Kind() (string, error) {
	presentKinds := []string{}
	if n.SubscriptionNotification != nil {
		presentKinds = append(presentKinds, "subscription")
	}
	if n.OneTimeProductNotification != nil {
		presentKinds = append(presentKinds, "oneTimeProduct")
	}
	if n.VoidedPurchaseNotification != nil {
		presentKinds = append(presentKinds, "voidedPurchase")
	}
	if n.TestNotification != nil {
		presentKinds = append(presentKinds, "test")
	}
	if len(presentKinds) == 0 {
		return "", fmt.Errorf("developer notification kind is required")
	}
	if len(presentKinds) > 1 {
		return "", fmt.Errorf("developer notification kinds are mutually exclusive: %s", strings.Join(presentKinds, ","))
	}
	return presentKinds[0], nil
}

func sortedAttributes(attributes map[string]string) map[string]string {
	if len(attributes) == 0 {
		return nil
	}
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sorted := make(map[string]string, len(attributes))
	for _, key := range keys {
		sorted[key] = attributes[key]
	}
	return sorted
}
