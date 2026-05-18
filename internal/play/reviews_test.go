package play

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestListReviewsPassesOptionsToLister(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	lister := &fakeReviewClient{
		listResult: ReviewListResult{
			PackageName: packageName,
			Reviews:     []Review{{ReviewID: "review-123"}},
		},
	}

	result, err := ListReviews(context.Background(), lister, ReviewListOptions{
		PackageName:         packageName,
		MaxResults:          25,
		Token:               "next",
		TranslationLanguage: "es",
	})
	if err != nil {
		t.Fatalf("ListReviews() error = %v", err)
	}
	if len(result.Reviews) != 1 {
		t.Fatalf("len(Reviews) = %d, want 1", len(result.Reviews))
	}
	if !reflect.DeepEqual(lister.listOptions, ReviewListOptions{
		PackageName:         packageName,
		MaxResults:          25,
		Token:               "next",
		TranslationLanguage: "es",
	}) {
		t.Fatalf("listOptions = %#v", lister.listOptions)
	}
}

func TestListReviewsRejectsNegativeMaxResults(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListReviews(context.Background(), nil, ReviewListOptions{
		PackageName: packageName,
		MaxResults:  -1,
	})
	if err == nil {
		t.Fatal("expected max results validation error")
	}
}

func TestListReviewsRejectsMaxResultsAboveGoogleLimit(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ListReviews(context.Background(), nil, ReviewListOptions{
		PackageName: packageName,
		MaxResults:  101,
	})
	if err == nil {
		t.Fatal("expected max results validation error")
	}
}

func TestGetReviewPassesReviewIDAndTranslationLanguage(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	getter := &fakeReviewClient{review: Review{ReviewID: "review-123"}}

	review, err := GetReview(context.Background(), getter, ReviewGetOptions{
		PackageName:         packageName,
		ReviewID:            "review-123",
		TranslationLanguage: "fr",
	})
	if err != nil {
		t.Fatalf("GetReview() error = %v", err)
	}
	if review.ReviewID != "review-123" {
		t.Fatalf("ReviewID = %q, want review-123", review.ReviewID)
	}
	if getter.getReviewID != "review-123" {
		t.Fatalf("getReviewID = %q, want review-123", getter.getReviewID)
	}
	if getter.translationLanguage != "fr" {
		t.Fatalf("translationLanguage = %q, want fr", getter.translationLanguage)
	}
}

func TestReplyToReviewDryRunDoesNotRequireReplier(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	result, err := ReplyToReview(context.Background(), nil, ReviewReplyOptions{
		PackageName: packageName,
		ReviewID:    "review-123",
		Text:        "Thanks for trying the app.",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("ReplyToReview() error = %v", err)
	}
	if !result.DryRun {
		t.Fatal("DryRun = false, want true")
	}
	if result.Applied {
		t.Fatal("Applied = true, want false")
	}
}

func TestReplyToReviewRequiresConfirmOutsideDryRun(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ReplyToReview(context.Background(), &fakeReviewClient{}, ReviewReplyOptions{
		PackageName: packageName,
		ReviewID:    "review-123",
		Text:        "Thanks.",
	})
	if err == nil {
		t.Fatal("expected confirmation validation error")
	}
}

func TestReplyToReviewAllowsGoogleMaximumLength(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ReplyToReview(context.Background(), nil, ReviewReplyOptions{
		PackageName: packageName,
		ReviewID:    "review-123",
		Text:        strings.Repeat("a", maxReviewReplyCharacters),
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("ReplyToReview() error = %v", err)
	}
}

func TestReplyToReviewRejectsTextOverGoogleMaximumLength(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}

	_, err = ReplyToReview(context.Background(), nil, ReviewReplyOptions{
		PackageName: packageName,
		ReviewID:    "review-123",
		Text:        strings.Repeat("a", maxReviewReplyCharacters+1),
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected reply length validation error")
	}
}

func TestReplyToReviewAppliesWithConfirm(t *testing.T) {
	packageName, err := NewPackageName("com.example.app")
	if err != nil {
		t.Fatalf("NewPackageName() error = %v", err)
	}
	replier := &fakeReviewClient{
		reply: DeveloperReply{Text: "Thanks."},
	}

	result, err := ReplyToReview(context.Background(), replier, ReviewReplyOptions{
		PackageName: packageName,
		ReviewID:    "review-123",
		Text:        "Thanks.",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("ReplyToReview() error = %v", err)
	}
	if !result.Applied {
		t.Fatal("Applied = false, want true")
	}
	if replier.replyText != "Thanks." {
		t.Fatalf("replyText = %q, want Thanks.", replier.replyText)
	}
}

type fakeReviewClient struct {
	listOptions         ReviewListOptions
	listResult          ReviewListResult
	review              Review
	getReviewID         ReviewID
	translationLanguage string
	replyText           string
	reply               DeveloperReply
}

func (c *fakeReviewClient) ListReviews(ctx context.Context, options ReviewListOptions) (ReviewListResult, error) {
	c.listOptions = options
	return c.listResult, nil
}

func (c *fakeReviewClient) GetReview(ctx context.Context, packageName PackageName, reviewID ReviewID, translationLanguage string) (Review, error) {
	c.getReviewID = reviewID
	c.translationLanguage = translationLanguage
	return c.review, nil
}

func (c *fakeReviewClient) ReplyToReview(ctx context.Context, packageName PackageName, reviewID ReviewID, text string) (DeveloperReply, error) {
	c.replyText = text
	return c.reply, nil
}
