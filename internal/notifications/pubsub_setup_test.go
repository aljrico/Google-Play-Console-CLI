package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
)

func TestSetupPubSubDryRunDoesNotRequireConfigurator(t *testing.T) {
	result, err := SetupPubSub(context.Background(), nil, PubSubSetupOptions{
		ProjectID:          "play-project",
		TopicID:            "play-rtdn",
		SubscriptionID:     "play-rtdn-sub",
		AckDeadlineSeconds: 30,
		DryRun:             true,
	})
	if err != nil {
		t.Fatalf("SetupPubSub() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run not applied", result)
	}
	if result.TopicName != "projects/play-project/topics/play-rtdn" {
		t.Fatalf("TopicName = %q", result.TopicName)
	}
	if !strings.Contains(result.Plan.OperatorSteps[0], result.TopicName) {
		t.Fatalf("operator steps = %#v, want Play Console topic step", result.Plan.OperatorSteps)
	}
}

func TestSetupPubSubCreatesResourcesAndGrantsPublisher(t *testing.T) {
	configurator := &fakePubSubConfigurator{
		policy: PubSubPolicy{
			Bindings: []PubSubBinding{
				{
					Role:      pubSubPublisherRole,
					Members:   []string{"serviceAccount:temporary@example.iam.gserviceaccount.com"},
					Condition: &PubSubCondition{Title: "temporary", Expression: `request.time < timestamp("2026-01-01T00:00:00Z")`},
				},
				{Role: "roles/viewer", Members: []string{"user:a@example.com"}},
			},
			Etag:    "etag-1",
			Version: 3,
		},
	}
	result, err := SetupPubSub(context.Background(), configurator, PubSubSetupOptions{
		ProjectID:          "play-project",
		TopicID:            "play-rtdn",
		SubscriptionID:     "play-rtdn-sub",
		PushEndpoint:       "https://example.com/rtdn",
		AckDeadlineSeconds: 20,
		Confirm:            true,
	})
	if err != nil {
		t.Fatalf("SetupPubSub() error = %v", err)
	}
	if !result.Applied {
		t.Fatalf("Applied = false, want true")
	}
	wantCalls := []string{"create-topic", "create-subscription", "get-policy", "set-policy"}
	if !reflect.DeepEqual(configurator.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", configurator.calls, wantCalls)
	}
	if configurator.subscription.PushEndpoint != "https://example.com/rtdn" {
		t.Fatalf("subscription = %#v, want push endpoint", configurator.subscription)
	}
	if !policyHasMember(configurator.updatedPolicy, pubSubPublisherRole, GooglePlayPublisherMember) {
		t.Fatalf("updatedPolicy = %#v, want Google Play publisher binding", configurator.updatedPolicy)
	}
	if configurator.updatedPolicy.Etag != "etag-1" {
		t.Fatalf("Etag = %q, want preserved etag", configurator.updatedPolicy.Etag)
	}
	if !policyHasCondition(configurator.updatedPolicy, "temporary") {
		t.Fatalf("updatedPolicy = %#v, want preserved conditional binding", configurator.updatedPolicy)
	}
	if !policyHasMember(configurator.updatedPolicy, pubSubPublisherRole, "serviceAccount:temporary@example.iam.gserviceaccount.com") {
		t.Fatalf("updatedPolicy = %#v, want preserved conditional member", configurator.updatedPolicy)
	}
}

func TestSetupPubSubSkipsIAMWriteWhenPublisherExists(t *testing.T) {
	configurator := &fakePubSubConfigurator{
		policy: PubSubPolicy{Bindings: []PubSubBinding{{Role: pubSubPublisherRole, Members: []string{GooglePlayPublisherMember}}}},
	}
	_, err := SetupPubSub(context.Background(), configurator, PubSubSetupOptions{
		ProjectID:          "play-project",
		TopicID:            "play-rtdn",
		SubscriptionID:     "play-rtdn-sub",
		AckDeadlineSeconds: 10,
		Confirm:            true,
	})
	if err != nil {
		t.Fatalf("SetupPubSub() error = %v", err)
	}
	wantCalls := []string{"create-topic", "create-subscription", "get-policy"}
	if !reflect.DeepEqual(configurator.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", configurator.calls, wantCalls)
	}
}

func TestPubSubSetupOptionsRejectInvalidInputs(t *testing.T) {
	tests := []PubSubSetupOptions{
		{},
		{ProjectID: " play-project", TopicID: "play-rtdn", SubscriptionID: "play-rtdn-sub", DryRun: true},
		{ProjectID: "play-project", TopicID: "go", SubscriptionID: "play-rtdn-sub", DryRun: true},
		{ProjectID: "play-project", TopicID: "goog-topic", SubscriptionID: "play-rtdn-sub", DryRun: true},
		{ProjectID: "play-project", TopicID: "play-rtdn", SubscriptionID: "bad space", DryRun: true},
		{ProjectID: "play-project", TopicID: "play-rtdn", SubscriptionID: "play-rtdn-sub", PushEndpoint: "http://example.com/rtdn", DryRun: true},
		{ProjectID: "play-project", TopicID: "play-rtdn", SubscriptionID: "play-rtdn-sub", AckDeadlineSeconds: 9, DryRun: true},
		{ProjectID: "play-project", TopicID: "play-rtdn", SubscriptionID: "play-rtdn-sub", Confirm: true, DryRun: true},
		{ProjectID: "play-project", TopicID: "play-rtdn", SubscriptionID: "play-rtdn-sub"},
	}
	for _, options := range tests {
		if err := options.Validate(); err == nil {
			t.Fatalf("Validate(%#v) expected error", options)
		}
	}
}

func TestPubSubClientMapsSetupCalls(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/v1/projects/play-project/topics/play-rtdn":
			_, _ = w.Write([]byte(`{"name":"projects/play-project/topics/play-rtdn"}`))
		case r.Method == http.MethodPut && r.URL.Path == "/v1/projects/play-project/subscriptions/play-rtdn-sub":
			_, _ = w.Write([]byte(`{"name":"projects/play-project/subscriptions/play-rtdn-sub"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/projects/play-project/topics/play-rtdn:getIamPolicy":
			if r.URL.Query().Get("options.requestedPolicyVersion") != "3" {
				t.Fatalf("query = %s, want requested policy version 3", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"bindings":[{"role":"roles/pubsub.publisher","members":["serviceAccount:temporary@example.iam.gserviceaccount.com"],"condition":{"title":"temporary","expression":"request.time < timestamp(\"2026-01-01T00:00:00Z\")"}}],"etag":"etag-1","version":3}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/play-project/topics/play-rtdn:setIamPolicy":
			_, _ = w.Write([]byte(`{"bindings":[{"role":"roles/pubsub.publisher","members":["serviceAccount:google-play-developer-notifications@system.gserviceaccount.com"]}],"etag":"etag-1"}`))
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
	if err := client.CreateTopic(context.Background(), "projects/play-project/topics/play-rtdn"); err != nil {
		t.Fatalf("CreateTopic() error = %v", err)
	}
	if err := client.CreateSubscription(context.Background(), "projects/play-project/subscriptions/play-rtdn-sub", PubSubSubscription{TopicName: "projects/play-project/topics/play-rtdn", AckDeadlineSeconds: 10}); err != nil {
		t.Fatalf("CreateSubscription() error = %v", err)
	}
	policy, err := client.GetTopicIAMPolicy(context.Background(), "projects/play-project/topics/play-rtdn")
	if err != nil {
		t.Fatalf("GetTopicIAMPolicy() error = %v", err)
	}
	updatedPolicy, _ := policy.WithMember(pubSubPublisherRole, GooglePlayPublisherMember)
	if err := client.SetTopicIAMPolicy(context.Background(), "projects/play-project/topics/play-rtdn", updatedPolicy); err != nil {
		t.Fatalf("SetTopicIAMPolicy() error = %v", err)
	}
	if len(requests) != 4 {
		t.Fatalf("requests = %#v, want 4 calls", requests)
	}
}

type fakePubSubConfigurator struct {
	calls         []string
	policy        PubSubPolicy
	updatedPolicy PubSubPolicy
	subscription  PubSubSubscription
}

func (c *fakePubSubConfigurator) CreateTopic(ctx context.Context, name string) error {
	c.calls = append(c.calls, "create-topic")
	return nil
}

func (c *fakePubSubConfigurator) CreateSubscription(ctx context.Context, name string, subscription PubSubSubscription) error {
	c.calls = append(c.calls, "create-subscription")
	c.subscription = subscription
	return nil
}

func (c *fakePubSubConfigurator) GetTopicIAMPolicy(ctx context.Context, topicName string) (PubSubPolicy, error) {
	c.calls = append(c.calls, "get-policy")
	return c.policy, nil
}

func (c *fakePubSubConfigurator) SetTopicIAMPolicy(ctx context.Context, topicName string, policy PubSubPolicy) error {
	c.calls = append(c.calls, "set-policy")
	c.updatedPolicy = policy
	return nil
}

func policyHasMember(policy PubSubPolicy, role string, member string) bool {
	for _, binding := range policy.Bindings {
		if binding.Role != role {
			continue
		}
		for _, existingMember := range binding.Members {
			if existingMember == member {
				return true
			}
		}
	}
	return false
}

func policyHasCondition(policy PubSubPolicy, title string) bool {
	for _, binding := range policy.Bindings {
		if binding.Condition != nil && binding.Condition.Title == title {
			return true
		}
	}
	return false
}
