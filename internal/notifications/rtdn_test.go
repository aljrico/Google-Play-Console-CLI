package notifications

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeRTDNSubscriptionNotificationFromFile(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.some.thing","eventTimeMillis":"1503349566168","subscriptionNotification":{"version":"1.0","notificationType":4,"purchaseToken":"PURCHASE_TOKEN"}}`
	payload := fmt.Sprintf(`{
		"deliveryAttempt": 5,
		"message": {
			"attributes": {"source":"play"},
			"data": %q,
			"message_id": "136969346945",
			"orderingKey": "rtdn-key",
			"publish_time": "2026-05-19T12:00:00Z"
		},
		"subscription": "projects/example/subscriptions/play-rtdn"
	}`, base64.StdEncoding.EncodeToString([]byte(notification)))
	path := filepath.Join(t.TempDir(), "pubsub.json")
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := DecodeRTDN(RTDNDecodeOptions{File: path})
	if err != nil {
		t.Fatalf("DecodeRTDN() error = %v", err)
	}
	if result.Kind != "subscription" {
		t.Fatalf("Kind = %q, want subscription", result.Kind)
	}
	if result.MessageID != "136969346945" {
		t.Fatalf("MessageID = %q", result.MessageID)
	}
	if result.OrderingKey != "rtdn-key" {
		t.Fatalf("OrderingKey = %q", result.OrderingKey)
	}
	if result.DeliveryAttempt.Value != 5 {
		t.Fatalf("DeliveryAttempt = %#v", result.DeliveryAttempt)
	}
	if result.Notification.SubscriptionNotification.PurchaseToken != "PURCHASE_TOKEN" {
		t.Fatalf("SubscriptionNotification = %#v", result.Notification.SubscriptionNotification)
	}
	if result.Notification.EventTimeMillis.Value != 1503349566168 {
		t.Fatalf("EventTimeMillis = %#v", result.Notification.EventTimeMillis)
	}
}

func TestDecodeRTDNTestNotificationFromData(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	result, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err != nil {
		t.Fatalf("DecodeRTDN() error = %v", err)
	}
	if result.Kind != "test" {
		t.Fatalf("Kind = %q, want test", result.Kind)
	}
}

func TestDecodeRTDNUnwrappedTestNotificationFromData(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`

	result, err := DecodeRTDN(RTDNDecodeOptions{Data: notification, Unwrapped: true})
	if err != nil {
		t.Fatalf("DecodeRTDN() error = %v", err)
	}
	if result.Kind != "test" {
		t.Fatalf("Kind = %q, want test", result.Kind)
	}
	if result.MessageID != "" || result.Subscription != "" {
		t.Fatalf("result = %#v, did not expect Pub/Sub envelope metadata", result)
	}
}

func TestDecodeRTDNRejectsMultipleKinds(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"},"subscriptionNotification":{"version":"1.0","notificationType":1,"purchaseToken":"token"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want mutually exclusive kind validation")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error = %v, want mutually exclusive validation", err)
	}
}

func TestDecodeRTDNRejectsMissingRequiredNotificationFields(t *testing.T) {
	notification := `{"version":"1.0","testNotification":{"version":"1.0"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want required field validation")
	}
	if !strings.Contains(err.Error(), "packageName") {
		t.Fatalf("error = %v, want packageName validation", err)
	}
}

func TestDecodeRTDNRejectsMissingNestedSubscriptionFields(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"subscriptionNotification":{}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want nested validation")
	}
	if !strings.Contains(err.Error(), "subscription notification version") {
		t.Fatalf("error = %v, want subscription version validation", err)
	}
}

func TestDecodeRTDNRejectsMissingNestedOneTimeProductFields(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"oneTimeProductNotification":{"version":"1.0","notificationType":1,"purchaseToken":"token"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want nested validation")
	}
	if !strings.Contains(err.Error(), "sku") {
		t.Fatalf("error = %v, want sku validation", err)
	}
}

func TestDecodeRTDNRejectsMissingNestedVoidedPurchaseFields(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"voidedPurchaseNotification":{"purchaseToken":"token","orderId":"order","productType":1}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want nested validation")
	}
	if !strings.Contains(err.Error(), "refundType") {
		t.Fatalf("error = %v, want refundType validation", err)
	}
}

func TestDecodeRTDNRejectsTrailingEnvelopeJSON(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q}} {"extra":true}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want trailing JSON validation")
	}
}

func TestDecodeRTDNRejectsTrailingNotificationJSON(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}} {"extra":true}`
	payload := fmt.Sprintf(`{"message":{"data":%q}}`, base64.StdEncoding.EncodeToString([]byte(notification)))

	_, err := DecodeRTDN(RTDNDecodeOptions{Data: payload})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want trailing notification JSON validation")
	}
}

func TestDecodeRTDNRejectsMissingInput(t *testing.T) {
	_, err := DecodeRTDN(RTDNDecodeOptions{})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want input validation")
	}
}
