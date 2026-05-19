package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationsRTDNDecodeOutputsKindWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	notification := `{"version":"1.0","packageName":"com.example.app","eventTimeMillis":1700000000000,"testNotification":{"version":"1.0"}}`
	payload := fmt.Sprintf(`{"message":{"data":%q,"messageId":"136969346945"},"subscription":"projects/example/subscriptions/play-rtdn"}`, base64.StdEncoding.EncodeToString([]byte(notification)))
	path := writeRootTestPathContent(t, filepath.Join(t.TempDir(), "pubsub.json"), payload)

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"rtdn",
		"decode",
		"--file",
		path,
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"kind":"test"`,
		`"messageId":"136969346945"`,
		`"packageName":"com.example.app"`,
		`"testNotification":{"version":"1.0"}`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotificationsPubSubSetupDryRunDoesNotRequireAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"pubsub",
		"setup",
		"--project",
		"play-project",
		"--topic",
		"play-rtdn",
		"--subscription",
		"play-rtdn-sub",
		"--push-endpoint",
		"https://example.com/rtdn",
		"--dry-run",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"dryRun":true`,
		`"topicName":"projects/play-project/topics/play-rtdn"`,
		`"subscriptionName":"projects/play-project/subscriptions/play-rtdn-sub"`,
		`"publisherMember":"serviceAccount:google-play-developer-notifications@system.gserviceaccount.com"`,
		`"pushEndpoint":"https://example.com/rtdn"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestNotificationsPubSubPullRejectsAckWithoutConfirmBeforeAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"notifications",
		"pubsub",
		"pull",
		"--project",
		"play-project",
		"--subscription",
		"play-rtdn-sub",
		"--ack",
		"--output",
		"json",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected acknowledgement confirmation validation error")
	}
	if !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("error = %v, want confirmation validation", err)
	}
	if strings.Contains(err.Error(), "no active auth profile") {
		t.Fatalf("error = %v, did not expect auth error", err)
	}
}
