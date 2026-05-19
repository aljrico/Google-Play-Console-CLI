package notifications

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/googleclient"
	"google.golang.org/api/option"
	"google.golang.org/api/pubsub/v1"
)

const (
	GooglePlayPublisherMember = "serviceAccount:google-play-developer-notifications@system.gserviceaccount.com"
	pubSubPublisherRole       = "roles/pubsub.publisher"
)

var pubSubNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._~+%-]{2,254}$`)

type PubSubSetupOptions struct {
	ProjectID          string `json:"projectId"`
	TopicID            string `json:"topicId"`
	SubscriptionID     string `json:"subscriptionId"`
	PushEndpoint       string `json:"pushEndpoint,omitempty"`
	AckDeadlineSeconds int64  `json:"ackDeadlineSeconds,omitempty"`
	Confirm            bool   `json:"confirm"`
	DryRun             bool   `json:"dryRun"`
}

func (o PubSubSetupOptions) Validate() error {
	if err := validateProjectID(o.ProjectID); err != nil {
		return err
	}
	if err := validatePubSubID("topic ID", o.TopicID); err != nil {
		return err
	}
	if err := validatePubSubID("subscription ID", o.SubscriptionID); err != nil {
		return err
	}
	if o.PushEndpoint != "" {
		if err := validatePushEndpoint(o.PushEndpoint); err != nil {
			return err
		}
	}
	if o.AckDeadlineSeconds < 0 {
		return fmt.Errorf("ack deadline cannot be negative")
	}
	if o.AckDeadlineSeconds > 0 && (o.AckDeadlineSeconds < 10 || o.AckDeadlineSeconds > 600) {
		return fmt.Errorf("ack deadline must be between 10 and 600 seconds")
	}
	if o.Confirm && o.DryRun {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.Confirm && !o.DryRun {
		return fmt.Errorf("Pub/Sub setup requires --confirm or --dry-run")
	}
	return nil
}

func (o PubSubSetupOptions) TopicName() string {
	return fmt.Sprintf("projects/%s/topics/%s", o.ProjectID, o.TopicID)
}

func (o PubSubSetupOptions) SubscriptionName() string {
	return fmt.Sprintf("projects/%s/subscriptions/%s", o.ProjectID, o.SubscriptionID)
}

type PubSubSetupPlan struct {
	ProjectID          string   `json:"projectId"`
	TopicName          string   `json:"topicName"`
	SubscriptionName   string   `json:"subscriptionName"`
	PushEndpoint       string   `json:"pushEndpoint,omitempty"`
	AckDeadlineSeconds int64    `json:"ackDeadlineSeconds,omitempty"`
	PublisherMember    string   `json:"publisherMember"`
	Confirm            bool     `json:"confirm"`
	Steps              []string `json:"steps"`
	OperatorSteps      []string `json:"operatorSteps"`
}

type PubSubSetupResult struct {
	DryRun           bool            `json:"dryRun"`
	Applied          bool            `json:"applied"`
	TopicName        string          `json:"topicName"`
	SubscriptionName string          `json:"subscriptionName"`
	Plan             PubSubSetupPlan `json:"plan"`
}

type PubSubConfigurator interface {
	CreateTopic(ctx context.Context, name string) error
	CreateSubscription(ctx context.Context, name string, subscription PubSubSubscription) error
	GetTopicIAMPolicy(ctx context.Context, topicName string) (PubSubPolicy, error)
	SetTopicIAMPolicy(ctx context.Context, topicName string, policy PubSubPolicy) error
}

type PubSubSubscription struct {
	TopicName          string `json:"topicName"`
	PushEndpoint       string `json:"pushEndpoint,omitempty"`
	AckDeadlineSeconds int64  `json:"ackDeadlineSeconds,omitempty"`
}

type PubSubPolicy struct {
	Bindings []PubSubBinding `json:"bindings"`
	Etag     string          `json:"etag,omitempty"`
	Version  int64           `json:"version,omitempty"`
}

type PubSubBinding struct {
	Role      string           `json:"role"`
	Members   []string         `json:"members"`
	Condition *PubSubCondition `json:"condition,omitempty"`
}

type PubSubCondition struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Expression  string `json:"expression,omitempty"`
	Location    string `json:"location,omitempty"`
}

type PubSubClient struct {
	service *pubsub.Service
}

func NewPubSubClientFromActiveProfile(ctx context.Context) (*PubSubClient, error) {
	httpClient, err := googleclient.ActiveProfileHTTPClient(ctx, pubsub.PubsubScope)
	if err != nil {
		return nil, err
	}
	service, err := pubsub.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create Pub/Sub service: %w", err)
	}
	return &PubSubClient{service: service}, nil
}

func (c *PubSubClient) CreateTopic(ctx context.Context, name string) error {
	_, err := c.service.Projects.Topics.Create(name, &pubsub.Topic{Name: name}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create Pub/Sub topic %s: %w", name, err)
	}
	return nil
}

func (c *PubSubClient) CreateSubscription(ctx context.Context, name string, subscription PubSubSubscription) error {
	apiSubscription := &pubsub.Subscription{
		Name:               name,
		Topic:              subscription.TopicName,
		AckDeadlineSeconds: subscription.AckDeadlineSeconds,
	}
	if subscription.PushEndpoint != "" {
		apiSubscription.PushConfig = &pubsub.PushConfig{PushEndpoint: subscription.PushEndpoint}
	}
	_, err := c.service.Projects.Subscriptions.Create(name, apiSubscription).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create Pub/Sub subscription %s: %w", name, err)
	}
	return nil
}

func (c *PubSubClient) GetTopicIAMPolicy(ctx context.Context, topicName string) (PubSubPolicy, error) {
	apiPolicy, err := c.service.Projects.Topics.GetIamPolicy(topicName).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return PubSubPolicy{}, fmt.Errorf("get Pub/Sub topic IAM policy %s: %w", topicName, err)
	}
	return pubSubPolicyFromAPI(apiPolicy), nil
}

func (c *PubSubClient) SetTopicIAMPolicy(ctx context.Context, topicName string, policy PubSubPolicy) error {
	_, err := c.service.Projects.Topics.SetIamPolicy(topicName, &pubsub.SetIamPolicyRequest{
		Policy: pubSubPolicyToAPI(policy),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("set Pub/Sub topic IAM policy %s: %w", topicName, err)
	}
	return nil
}

func NewPubSubSetupPlan(options PubSubSetupOptions) (PubSubSetupPlan, error) {
	if err := options.Validate(); err != nil {
		return PubSubSetupPlan{}, err
	}
	topicName := options.TopicName()
	subscriptionName := options.SubscriptionName()
	steps := []string{
		"create Pub/Sub topic",
		"create Pub/Sub subscription",
		fmt.Sprintf("grant %s %s on topic", GooglePlayPublisherMember, pubSubPublisherRole),
	}
	operatorSteps := []string{
		fmt.Sprintf("In Play Console, set Real-time developer notifications for the app to %s.", topicName),
		"Send a test notification from Play Console and inspect it with gpc notifications rtdn decode.",
	}
	return PubSubSetupPlan{
		ProjectID:          options.ProjectID,
		TopicName:          topicName,
		SubscriptionName:   subscriptionName,
		PushEndpoint:       options.PushEndpoint,
		AckDeadlineSeconds: options.AckDeadlineSeconds,
		PublisherMember:    GooglePlayPublisherMember,
		Confirm:            options.Confirm,
		Steps:              steps,
		OperatorSteps:      operatorSteps,
	}, nil
}

func SetupPubSub(ctx context.Context, configurator PubSubConfigurator, options PubSubSetupOptions) (PubSubSetupResult, error) {
	plan, err := NewPubSubSetupPlan(options)
	if err != nil {
		return PubSubSetupResult{}, err
	}
	result := PubSubSetupResult{
		DryRun:           options.DryRun,
		Applied:          false,
		TopicName:        plan.TopicName,
		SubscriptionName: plan.SubscriptionName,
		Plan:             plan,
	}
	if options.DryRun {
		return result, nil
	}
	if configurator == nil {
		return PubSubSetupResult{}, fmt.Errorf("Pub/Sub configurator is required")
	}
	if err := configurator.CreateTopic(ctx, plan.TopicName); err != nil {
		return PubSubSetupResult{}, err
	}
	subscription := PubSubSubscription{
		TopicName:          plan.TopicName,
		PushEndpoint:       options.PushEndpoint,
		AckDeadlineSeconds: options.AckDeadlineSeconds,
	}
	if err := configurator.CreateSubscription(ctx, plan.SubscriptionName, subscription); err != nil {
		return PubSubSetupResult{}, err
	}
	policy, err := configurator.GetTopicIAMPolicy(ctx, plan.TopicName)
	if err != nil {
		return PubSubSetupResult{}, err
	}
	updatedPolicy, changed := policy.WithMember(pubSubPublisherRole, GooglePlayPublisherMember)
	if changed {
		if err := configurator.SetTopicIAMPolicy(ctx, plan.TopicName, updatedPolicy); err != nil {
			return PubSubSetupResult{}, err
		}
	}
	result.Applied = true
	return result, nil
}

func (p PubSubPolicy) WithMember(role string, member string) (PubSubPolicy, bool) {
	bindings := make([]PubSubBinding, len(p.Bindings))
	copy(bindings, p.Bindings)
	for index, binding := range bindings {
		if binding.Role != role {
			continue
		}
		if binding.Condition != nil {
			continue
		}
		for _, existingMember := range binding.Members {
			if existingMember == member {
				return PubSubPolicy{Bindings: bindings, Etag: p.Etag, Version: p.Version}, false
			}
		}
		binding.Members = append(binding.Members, member)
		sort.Strings(binding.Members)
		bindings[index] = binding
		return PubSubPolicy{Bindings: bindings, Etag: p.Etag, Version: p.Version}, true
	}
	bindings = append(bindings, PubSubBinding{Role: role, Members: []string{member}})
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Role < bindings[j].Role
	})
	return PubSubPolicy{Bindings: bindings, Etag: p.Etag, Version: p.Version}, true
}

func validateProjectID(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("project ID is required")
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("project ID cannot have leading or trailing whitespace")
	}
	return nil
}

func validatePubSubID(label string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", label)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s cannot have leading or trailing whitespace", label)
	}
	if strings.HasPrefix(value, "goog") {
		return fmt.Errorf("%s cannot start with goog", label)
	}
	if !pubSubNamePattern.MatchString(value) {
		return fmt.Errorf("%s must start with a letter and be 3-255 characters of letters, numbers, dashes, underscores, periods, tildes, plus signs, or percent signs", label)
	}
	return nil
}

func validatePushEndpoint(value string) error {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("parse push endpoint: %w", err)
	}
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("push endpoint must use https")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("push endpoint host is required")
	}
	if parsedURL.User != nil {
		return fmt.Errorf("push endpoint must not contain userinfo")
	}
	return nil
}

func pubSubPolicyFromAPI(apiPolicy *pubsub.Policy) PubSubPolicy {
	if apiPolicy == nil {
		return PubSubPolicy{}
	}
	bindings := make([]PubSubBinding, 0, len(apiPolicy.Bindings))
	for _, binding := range apiPolicy.Bindings {
		if binding == nil {
			continue
		}
		members := append([]string(nil), binding.Members...)
		sort.Strings(members)
		bindings = append(bindings, PubSubBinding{
			Role:      binding.Role,
			Members:   members,
			Condition: pubSubConditionFromAPI(binding.Condition),
		})
	}
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Role < bindings[j].Role
	})
	return PubSubPolicy{Bindings: bindings, Etag: apiPolicy.Etag, Version: apiPolicy.Version}
}

func pubSubPolicyToAPI(policy PubSubPolicy) *pubsub.Policy {
	bindings := make([]*pubsub.Binding, 0, len(policy.Bindings))
	for _, binding := range policy.Bindings {
		members := append([]string(nil), binding.Members...)
		sort.Strings(members)
		bindings = append(bindings, &pubsub.Binding{
			Role:      binding.Role,
			Members:   members,
			Condition: pubSubConditionToAPI(binding.Condition),
		})
	}
	return &pubsub.Policy{Bindings: bindings, Etag: policy.Etag, Version: policy.Version}
}

func pubSubConditionFromAPI(condition *pubsub.Expr) *PubSubCondition {
	if condition == nil {
		return nil
	}
	return &PubSubCondition{
		Title:       condition.Title,
		Description: condition.Description,
		Expression:  condition.Expression,
		Location:    condition.Location,
	}
}

func pubSubConditionToAPI(condition *PubSubCondition) *pubsub.Expr {
	if condition == nil {
		return nil
	}
	return &pubsub.Expr{
		Title:       condition.Title,
		Description: condition.Description,
		Expression:  condition.Expression,
		Location:    condition.Location,
	}
}
