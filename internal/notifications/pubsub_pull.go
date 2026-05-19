package notifications

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	"google.golang.org/api/pubsub/v1"
)

type PubSubPullOptions struct {
	ProjectID      string `json:"projectId"`
	SubscriptionID string `json:"subscriptionId"`
	MaxMessages    int64  `json:"maxMessages"`
	DecodeRTDN     bool   `json:"decodeRtdn"`
	Ack            bool   `json:"ack"`
	Confirm        bool   `json:"confirm"`
}

func (o PubSubPullOptions) Validate() error {
	if err := validateProjectID(o.ProjectID); err != nil {
		return err
	}
	if err := validatePubSubID("subscription ID", o.SubscriptionID); err != nil {
		return err
	}
	if o.MaxMessages <= 0 {
		return fmt.Errorf("max messages must be greater than 0")
	}
	if o.MaxMessages > 100 {
		return fmt.Errorf("max messages cannot exceed 100")
	}
	if o.Confirm && !o.Ack {
		return fmt.Errorf("--confirm can only be used with --ack")
	}
	if o.Ack && !o.Confirm {
		return fmt.Errorf("acknowledging Pub/Sub messages requires --confirm")
	}
	return nil
}

func (o PubSubPullOptions) SubscriptionName() string {
	return fmt.Sprintf("projects/%s/subscriptions/%s", o.ProjectID, o.SubscriptionID)
}

type PubSubPullResult struct {
	SubscriptionName string                `json:"subscriptionName"`
	MaxMessages      int64                 `json:"maxMessages"`
	DecodeRTDN       bool                  `json:"decodeRtdn"`
	Ack              bool                  `json:"ack"`
	Acknowledged     bool                  `json:"acknowledged"`
	Messages         []PulledPubSubMessage `json:"messages"`
	AckIDs           []string              `json:"ackIds,omitempty"`
}

type PulledPubSubMessage struct {
	AckID           string                  `json:"ackId,omitempty"`
	DeliveryAttempt int64                   `json:"deliveryAttempt,omitempty"`
	MessageID       string                  `json:"messageId,omitempty"`
	PublishTime     string                  `json:"publishTime,omitempty"`
	OrderingKey     string                  `json:"orderingKey,omitempty"`
	Attributes      map[string]string       `json:"attributes,omitempty"`
	Data            string                  `json:"data,omitempty"`
	RTDN            *PulledRTDNNotification `json:"rtdn,omitempty"`
}

type PulledRTDNNotification struct {
	Kind         string                `json:"kind"`
	Notification DeveloperNotification `json:"notification"`
}

type PubSubPuller interface {
	PullMessages(ctx context.Context, subscriptionName string, maxMessages int64) ([]PulledPubSubMessage, error)
}

type PubSubAcknowledger interface {
	AcknowledgeMessages(ctx context.Context, subscriptionName string, ackIDs []string) error
}

func PullPubSub(ctx context.Context, puller PubSubPuller, options PubSubPullOptions) (PubSubPullResult, error) {
	if err := options.Validate(); err != nil {
		return PubSubPullResult{}, err
	}
	if puller == nil {
		return PubSubPullResult{}, fmt.Errorf("Pub/Sub puller is required")
	}
	subscriptionName := options.SubscriptionName()
	messages, err := puller.PullMessages(ctx, subscriptionName, options.MaxMessages)
	if err != nil {
		return PubSubPullResult{}, err
	}
	if options.DecodeRTDN {
		for index, message := range messages {
			rtdn, err := DecodeRTDNMessageData(message.Data)
			if err != nil {
				return PubSubPullResult{}, fmt.Errorf("decode RTDN message %s: %w", message.MessageID, err)
			}
			messages[index].RTDN = &rtdn
		}
	}
	ackIDs := ackIDsFromMessages(messages)
	result := PubSubPullResult{
		SubscriptionName: subscriptionName,
		MaxMessages:      options.MaxMessages,
		DecodeRTDN:       options.DecodeRTDN,
		Ack:              options.Ack,
		Acknowledged:     false,
		Messages:         messages,
		AckIDs:           ackIDs,
	}
	return result, nil
}

func AcknowledgePulledPubSub(ctx context.Context, acknowledger PubSubAcknowledger, subscriptionName string, ackIDs []string) error {
	if len(ackIDs) == 0 {
		return nil
	}
	if acknowledger == nil {
		return fmt.Errorf("Pub/Sub acknowledger is required")
	}
	return acknowledger.AcknowledgeMessages(ctx, subscriptionName, ackIDs)
}

func DecodeRTDNMessageData(data string) (PulledRTDNNotification, error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return PulledRTDNNotification{}, fmt.Errorf("decode Pub/Sub message data: %w", err)
	}
	notification, err := decodeDeveloperNotification(decodedData)
	if err != nil {
		return PulledRTDNNotification{}, err
	}
	kind, err := notification.Kind()
	if err != nil {
		return PulledRTDNNotification{}, err
	}
	return PulledRTDNNotification{Kind: kind, Notification: notification}, nil
}

func (c *PubSubClient) PullMessages(ctx context.Context, subscriptionName string, maxMessages int64) ([]PulledPubSubMessage, error) {
	response, err := c.service.Projects.Subscriptions.Pull(subscriptionName, &pubsub.PullRequest{
		MaxMessages: maxMessages,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("pull Pub/Sub subscription %s: %w", subscriptionName, err)
	}
	messages := make([]PulledPubSubMessage, 0, len(response.ReceivedMessages))
	for _, receivedMessage := range response.ReceivedMessages {
		if receivedMessage == nil {
			continue
		}
		messages = append(messages, pulledPubSubMessageFromAPI(receivedMessage))
	}
	return messages, nil
}

func (c *PubSubClient) AcknowledgeMessages(ctx context.Context, subscriptionName string, ackIDs []string) error {
	if len(ackIDs) == 0 {
		return nil
	}
	_, err := c.service.Projects.Subscriptions.Acknowledge(subscriptionName, &pubsub.AcknowledgeRequest{
		AckIds: ackIDs,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("acknowledge Pub/Sub subscription %s: %w", subscriptionName, err)
	}
	return nil
}

func pulledPubSubMessageFromAPI(receivedMessage *pubsub.ReceivedMessage) PulledPubSubMessage {
	message := PulledPubSubMessage{
		AckID:           receivedMessage.AckId,
		DeliveryAttempt: receivedMessage.DeliveryAttempt,
	}
	if receivedMessage.Message != nil {
		message.MessageID = receivedMessage.Message.MessageId
		message.PublishTime = receivedMessage.Message.PublishTime
		message.OrderingKey = receivedMessage.Message.OrderingKey
		message.Attributes = sortedAttributes(receivedMessage.Message.Attributes)
		message.Data = receivedMessage.Message.Data
	}
	return message
}

func ackIDsFromMessages(messages []PulledPubSubMessage) []string {
	ackIDs := make([]string, 0, len(messages))
	for _, message := range messages {
		if message.AckID != "" {
			ackIDs = append(ackIDs, message.AckID)
		}
	}
	sort.Strings(ackIDs)
	return ackIDs
}
