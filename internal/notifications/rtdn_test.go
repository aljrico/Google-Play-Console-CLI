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
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"subscriptionNotification":{"version":"1.0","notificationType":4,"purchaseToken":"token-123","subscriptionId":"pro_monthly"}}`
	payload := fmt.Sprintf(`{
		"message": {
			"attributes": {"source":"play"},
			"data": %q,
			"messageId": "136969346945",
			"publishTime": "2026-05-19T12:00:00Z"
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
	if result.Notification.SubscriptionNotification.SubscriptionID != "pro_monthly" {
		t.Fatalf("SubscriptionNotification = %#v", result.Notification.SubscriptionNotification)
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

func TestDecodeRTDNRejectsMultipleKinds(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"},"subscriptionNotification":{"version":"1.0","notificationType":1,"purchaseToken":"token","subscriptionId":"pro"}}`
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

func TestDecodeRTDNRejectsMissingInput(t *testing.T) {
	_, err := DecodeRTDN(RTDNDecodeOptions{})
	if err == nil {
		t.Fatal("DecodeRTDN() error = nil, want input validation")
	}
}
