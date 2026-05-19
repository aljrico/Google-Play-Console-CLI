package notifications

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
)

func TestPullPubSubDecodesRTDNMessages(t *testing.T) {
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`
	puller := &fakePubSubPuller{
		messages: []PulledPubSubMessage{
			{
				AckID:     "ack-1",
				MessageID: "message-1",
				Data:      base64.StdEncoding.EncodeToString([]byte(notification)),
			},
		},
	}

	result, err := PullPubSub(context.Background(), puller, PubSubPullOptions{
		ProjectID:      "play-project",
		SubscriptionID: "play-rtdn-sub",
		MaxMessages:    5,
		DecodeRTDN:     true,
	})
	if err != nil {
		t.Fatalf("PullPubSub() error = %v", err)
	}
	if result.SubscriptionName != "projects/play-project/subscriptions/play-rtdn-sub" {
		t.Fatalf("SubscriptionName = %q", result.SubscriptionName)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(result.Messages))
	}
	if result.Messages[0].RTDN == nil || result.Messages[0].RTDN.Kind != "test" {
		t.Fatalf("RTDN = %#v, want decoded test notification", result.Messages[0].RTDN)
	}
	if result.Acknowledged {
		t.Fatalf("Acknowledged = true, want false")
	}
}

func TestPullPubSubAcknowledgesOnlyWithConfirm(t *testing.T) {
	puller := &fakePubSubPuller{
		messages: []PulledPubSubMessage{{AckID: "ack-b", MessageID: "message-b"}, {AckID: "ack-a", MessageID: "message-a"}},
	}
	result, err := PullPubSub(context.Background(), puller, PubSubPullOptions{
		ProjectID:      "play-project",
		SubscriptionID: "play-rtdn-sub",
		MaxMessages:    10,
		Ack:            true,
		Confirm:        true,
	})
	if err != nil {
		t.Fatalf("PullPubSub() error = %v", err)
	}
	if result.Acknowledged {
		t.Fatalf("Acknowledged = true before explicit acknowledgement")
	}
	if err := AcknowledgePulledPubSub(context.Background(), puller, result.SubscriptionName, result.AckIDs); err != nil {
		t.Fatalf("AcknowledgePulledPubSub() error = %v", err)
	}
	wantAckIDs := []string{"ack-a", "ack-b"}
	if !reflect.DeepEqual(puller.ackIDs, wantAckIDs) {
		t.Fatalf("ackIDs = %#v, want %#v", puller.ackIDs, wantAckIDs)
	}
}

func TestPullPubSubOptionsRejectInvalidInputs(t *testing.T) {
	tests := []PubSubPullOptions{
		{},
		{ProjectID: " play-project", SubscriptionID: "play-rtdn-sub", MaxMessages: 1},
		{ProjectID: "play-project", SubscriptionID: "go", MaxMessages: 1},
		{ProjectID: "play-project", SubscriptionID: "play-rtdn-sub", MaxMessages: 0},
		{ProjectID: "play-project", SubscriptionID: "play-rtdn-sub", MaxMessages: 101},
		{ProjectID: "play-project", SubscriptionID: "play-rtdn-sub", MaxMessages: 1, Ack: true},
		{ProjectID: "play-project", SubscriptionID: "play-rtdn-sub", MaxMessages: 1, Confirm: true},
	}
	for _, options := range tests {
		if err := options.Validate(); err == nil {
			t.Fatalf("Validate(%#v) expected error", options)
		}
	}
}

func TestDecodeRTDNMessageDataRejectsInvalidData(t *testing.T) {
	_, err := DecodeRTDNMessageData("not-base64")
	if err == nil {
		t.Fatalf("DecodeRTDNMessageData() expected error")
	}
	if !strings.Contains(err.Error(), "decode Pub/Sub message data") {
		t.Fatalf("error = %v, want base64 decode error", err)
	}
}

func TestPubSubClientPullAndAcknowledge(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/play-project/subscriptions/play-rtdn-sub:pull":
			var request pubsub.PullRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode pull request error = %v", err)
			}
			if request.MaxMessages != 1 {
				t.Fatalf("MaxMessages = %d, want 1", request.MaxMessages)
			}
			_, _ = w.Write([]byte(`{"receivedMessages":[{"ackId":"ack-1","deliveryAttempt":2,"message":{"messageId":"message-1","data":"` + base64.StdEncoding.EncodeToString([]byte(`{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`)) + `","attributes":{"source":"play"},"publishTime":"2026-05-19T12:00:00Z"}}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/play-project/subscriptions/play-rtdn-sub:acknowledge":
			var request pubsub.AcknowledgeRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("Decode acknowledge request error = %v", err)
			}
			if !reflect.DeepEqual(request.AckIds, []string{"ack-1"}) {
				t.Fatalf("AckIds = %#v, want ack-1", request.AckIds)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	service, err := pubsub.NewService(
		context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	client := PubSubClient{service: service}
	messages, err := client.PullMessages(context.Background(), "projects/play-project/subscriptions/play-rtdn-sub", 1)
	if err != nil {
		t.Fatalf("PullMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].AckID != "ack-1" || messages[0].Attributes["source"] != "play" {
		t.Fatalf("messages = %#v, want pulled message", messages)
	}
	if err := client.AcknowledgeMessages(context.Background(), "projects/play-project/subscriptions/play-rtdn-sub", []string{"ack-1"}); err != nil {
		t.Fatalf("AcknowledgeMessages() error = %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("requests = %#v, want pull and acknowledge", requests)
	}
}

type fakePubSubPuller struct {
	messages []PulledPubSubMessage
	ackIDs   []string
}

func (p *fakePubSubPuller) PullMessages(ctx context.Context, subscriptionName string, maxMessages int64) ([]PulledPubSubMessage, error) {
	return p.messages, nil
}

func (p *fakePubSubPuller) AcknowledgeMessages(ctx context.Context, subscriptionName string, ackIDs []string) error {
	p.ackIDs = append([]string(nil), ackIDs...)
	return nil
}
