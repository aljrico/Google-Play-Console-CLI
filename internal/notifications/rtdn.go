package notifications

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type RTDNDecodeOptions struct {
	File string `json:"file,omitempty"`
	Data string `json:"data,omitempty"`
}

type PubSubEnvelope struct {
	DeliveryAttempt *Int64        `json:"deliveryAttempt,omitempty"`
	Message         PubSubMessage `json:"message"`
	Subscription    string        `json:"subscription,omitempty"`
}

type PubSubMessage struct {
	Attributes  map[string]string `json:"attributes,omitempty"`
	Data        string            `json:"data"`
	MessageID   string            `json:"messageId,omitempty"`
	OrderingKey string            `json:"orderingKey,omitempty"`
	PublishTime string            `json:"publishTime,omitempty"`
}

type DeveloperNotification struct {
	Version                    string                      `json:"version"`
	PackageName                string                      `json:"packageName"`
	EventTimeMillis            *Int64                      `json:"eventTimeMillis,omitempty"`
	SubscriptionNotification   *SubscriptionNotification   `json:"subscriptionNotification,omitempty"`
	OneTimeProductNotification *OneTimeProductNotification `json:"oneTimeProductNotification,omitempty"`
	VoidedPurchaseNotification *VoidedPurchaseNotification `json:"voidedPurchaseNotification,omitempty"`
	TestNotification           *TestNotification           `json:"testNotification,omitempty"`
}

type SubscriptionNotification struct {
	Version          string `json:"version"`
	NotificationType *Int64 `json:"notificationType,omitempty"`
	PurchaseToken    string `json:"purchaseToken"`
}

type OneTimeProductNotification struct {
	Version          string `json:"version"`
	NotificationType *Int64 `json:"notificationType,omitempty"`
	PurchaseToken    string `json:"purchaseToken"`
	SKU              string `json:"sku"`
}

type VoidedPurchaseNotification struct {
	PurchaseToken string `json:"purchaseToken"`
	OrderID       string `json:"orderId"`
	ProductType   *Int64 `json:"productType,omitempty"`
	RefundType    *Int64 `json:"refundType,omitempty"`
}

type TestNotification struct {
	Version string `json:"version"`
}

type RTDNDecodeResult struct {
	DeliveryAttempt *Int64                `json:"deliveryAttempt,omitempty"`
	Subscription    string                `json:"subscription,omitempty"`
	MessageID       string                `json:"messageId,omitempty"`
	OrderingKey     string                `json:"orderingKey,omitempty"`
	PublishTime     string                `json:"publishTime,omitempty"`
	Attributes      map[string]string     `json:"attributes,omitempty"`
	Kind            string                `json:"kind"`
	Notification    DeveloperNotification `json:"notification"`
}

func DecodeRTDN(options RTDNDecodeOptions) (RTDNDecodeResult, error) {
	content, err := options.content()
	if err != nil {
		return RTDNDecodeResult{}, err
	}
	var envelope PubSubEnvelope
	if err := decodeSingleJSON(content, &envelope, false); err != nil {
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
	if err := decodeSingleJSON(decodedData, &notification, true); err != nil {
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
		DeliveryAttempt: envelope.DeliveryAttempt,
		Subscription:    envelope.Subscription,
		MessageID:       envelope.Message.MessageID,
		OrderingKey:     envelope.Message.OrderingKey,
		PublishTime:     envelope.Message.PublishTime,
		Attributes:      sortedAttributes(envelope.Message.Attributes),
		Kind:            kind,
		Notification:    notification,
	}, nil
}

func decodeSingleJSON(content []byte, target any, useNumber bool) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	if useNumber {
		decoder.UseNumber()
	}
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errorsIsEOF(err) {
		if err != nil {
			return err
		}
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func errorsIsEOF(err error) bool {
	return err == io.EOF
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
	if n.EventTimeMillis == nil {
		return fmt.Errorf("developer notification eventTimeMillis is required")
	}
	return n.validateKind()
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

func (n DeveloperNotification) validateKind() error {
	kind, err := n.Kind()
	if err != nil {
		return err
	}
	switch kind {
	case "subscription":
		return n.SubscriptionNotification.Validate()
	case "oneTimeProduct":
		return n.OneTimeProductNotification.Validate()
	case "voidedPurchase":
		return n.VoidedPurchaseNotification.Validate()
	case "test":
		return n.TestNotification.Validate()
	default:
		return nil
	}
}

func (n SubscriptionNotification) Validate() error {
	if strings.TrimSpace(n.Version) == "" {
		return fmt.Errorf("subscription notification version is required")
	}
	if n.NotificationType == nil {
		return fmt.Errorf("subscription notification notificationType is required")
	}
	if strings.TrimSpace(n.PurchaseToken) == "" {
		return fmt.Errorf("subscription notification purchaseToken is required")
	}
	return nil
}

func (n OneTimeProductNotification) Validate() error {
	if strings.TrimSpace(n.Version) == "" {
		return fmt.Errorf("one-time product notification version is required")
	}
	if n.NotificationType == nil {
		return fmt.Errorf("one-time product notification notificationType is required")
	}
	if strings.TrimSpace(n.PurchaseToken) == "" {
		return fmt.Errorf("one-time product notification purchaseToken is required")
	}
	if strings.TrimSpace(n.SKU) == "" {
		return fmt.Errorf("one-time product notification sku is required")
	}
	return nil
}

func (n VoidedPurchaseNotification) Validate() error {
	if strings.TrimSpace(n.PurchaseToken) == "" {
		return fmt.Errorf("voided purchase notification purchaseToken is required")
	}
	if strings.TrimSpace(n.OrderID) == "" {
		return fmt.Errorf("voided purchase notification orderId is required")
	}
	if n.ProductType == nil {
		return fmt.Errorf("voided purchase notification productType is required")
	}
	if n.RefundType == nil {
		return fmt.Errorf("voided purchase notification refundType is required")
	}
	return nil
}

func (n TestNotification) Validate() error {
	if strings.TrimSpace(n.Version) == "" {
		return fmt.Errorf("test notification version is required")
	}
	return nil
}

func (m *PubSubMessage) UnmarshalJSON(data []byte) error {
	type pubSubMessageAlias PubSubMessage
	var raw struct {
		pubSubMessageAlias
		MessageIDSnake   string `json:"message_id,omitempty"`
		PublishTimeSnake string `json:"publish_time,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = PubSubMessage(raw.pubSubMessageAlias)
	if m.MessageID == "" {
		m.MessageID = raw.MessageIDSnake
	}
	if m.PublishTime == "" {
		m.PublishTime = raw.PublishTimeSnake
	}
	return nil
}

type Int64 struct {
	Value int64
}

func (n *Int64) UnmarshalJSON(data []byte) error {
	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		value, parseErr := number.Int64()
		if parseErr != nil {
			return parseErr
		}
		n.Value = value
		return nil
	}
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err != nil {
		return err
	}
	value, err := strconv.ParseInt(stringValue, 10, 64)
	if err != nil {
		return err
	}
	n.Value = value
	return nil
}

func (n Int64) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(n.Value, 10)), nil
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
