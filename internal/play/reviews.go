package play

import (
	"context"
	"fmt"
)

const maxReviewReplyCharacters = 350

type ReviewID string

func NewReviewID(value string) (ReviewID, error) {
	if value == "" {
		return "", fmt.Errorf("review ID is required")
	}
	return ReviewID(value), nil
}

func (r ReviewID) String() string {
	return string(r)
}

type Timestamp struct {
	Seconds int64 `json:"seconds"`
	Nanos   int64 `json:"nanos,omitempty"`
}

type ReviewCommentKind string

const (
	ReviewCommentKindUser      ReviewCommentKind = "user"
	ReviewCommentKindDeveloper ReviewCommentKind = "developer"
)

type ReviewComment struct {
	Kind      ReviewCommentKind       `json:"kind"`
	Text      string                  `json:"text,omitempty"`
	UpdatedAt *Timestamp              `json:"updatedAt,omitempty"`
	User      *UserReviewComment      `json:"user,omitempty"`
	Developer *DeveloperReviewComment `json:"developer,omitempty"`
}

type UserReviewComment struct {
	ReviewerLanguage string          `json:"reviewerLanguage,omitempty"`
	StarRating       int64           `json:"starRating,omitempty"`
	AppVersionCode   int64           `json:"appVersionCode,omitempty"`
	AppVersionName   string          `json:"appVersionName,omitempty"`
	AndroidOSVersion int64           `json:"androidOsVersion,omitempty"`
	Device           string          `json:"device,omitempty"`
	OriginalText     string          `json:"originalText,omitempty"`
	ThumbsUpCount    int64           `json:"thumbsUpCount,omitempty"`
	ThumbsDownCount  int64           `json:"thumbsDownCount,omitempty"`
	DeviceMetadata   *DeviceMetadata `json:"deviceMetadata,omitempty"`
}

type DeviceMetadata struct {
	CPUMake            string `json:"cpuMake,omitempty"`
	CPUModel           string `json:"cpuModel,omitempty"`
	DeviceClass        string `json:"deviceClass,omitempty"`
	GLESVersion        int64  `json:"glEsVersion,omitempty"`
	Manufacturer       string `json:"manufacturer,omitempty"`
	NativePlatform     string `json:"nativePlatform,omitempty"`
	ProductName        string `json:"productName,omitempty"`
	RAMMegabytes       int64  `json:"ramMb,omitempty"`
	ScreenDensityDPI   int64  `json:"screenDensityDpi,omitempty"`
	ScreenHeightPixels int64  `json:"screenHeightPx,omitempty"`
	ScreenWidthPixels  int64  `json:"screenWidthPx,omitempty"`
}

type DeveloperReviewComment struct {
	LastEdited *Timestamp `json:"lastEdited,omitempty"`
}

type Review struct {
	ReviewID   ReviewID        `json:"reviewId"`
	AuthorName string          `json:"authorName,omitempty"`
	Comments   []ReviewComment `json:"comments"`
}

type ReviewPageInfo struct {
	ResultPerPage int64 `json:"resultPerPage,omitempty"`
	StartIndex    int64 `json:"startIndex,omitempty"`
	TotalResults  int64 `json:"totalResults,omitempty"`
}

type ReviewPagination struct {
	NextPageToken     string `json:"nextPageToken,omitempty"`
	PreviousPageToken string `json:"previousPageToken,omitempty"`
}

type ReviewListOptions struct {
	PackageName         PackageName `json:"packageName"`
	MaxResults          int64       `json:"maxResults,omitempty"`
	StartIndex          int64       `json:"startIndex,omitempty"`
	Token               string      `json:"token,omitempty"`
	TranslationLanguage string      `json:"translationLanguage,omitempty"`
}

func (o ReviewListOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if o.MaxResults < 0 {
		return fmt.Errorf("max results cannot be negative")
	}
	if o.MaxResults > 100 {
		return fmt.Errorf("max results cannot exceed 100")
	}
	if o.StartIndex < 0 {
		return fmt.Errorf("start index cannot be negative")
	}
	return nil
}

type ReviewListResult struct {
	PackageName PackageName       `json:"packageName"`
	Reviews     []Review          `json:"reviews"`
	PageInfo    *ReviewPageInfo   `json:"pageInfo,omitempty"`
	Pagination  *ReviewPagination `json:"pagination,omitempty"`
	Options     ReviewListOptions `json:"options"`
}

type ReviewGetter interface {
	GetReview(ctx context.Context, packageName PackageName, reviewID ReviewID, translationLanguage string) (Review, error)
}

type ReviewLister interface {
	ListReviews(ctx context.Context, options ReviewListOptions) (ReviewListResult, error)
}

type ReviewReplier interface {
	ReplyToReview(ctx context.Context, packageName PackageName, reviewID ReviewID, text string) (DeveloperReply, error)
}

func ListReviews(ctx context.Context, lister ReviewLister, options ReviewListOptions) (ReviewListResult, error) {
	if err := options.Validate(); err != nil {
		return ReviewListResult{}, err
	}
	if lister == nil {
		return ReviewListResult{}, fmt.Errorf("review lister is required")
	}
	return lister.ListReviews(ctx, options)
}

type ReviewGetOptions struct {
	PackageName         PackageName `json:"packageName"`
	ReviewID            ReviewID    `json:"reviewId"`
	TranslationLanguage string      `json:"translationLanguage,omitempty"`
}

func (o ReviewGetOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewReviewID(o.ReviewID.String()); err != nil {
		return err
	}
	return nil
}

func GetReview(ctx context.Context, getter ReviewGetter, options ReviewGetOptions) (Review, error) {
	if err := options.Validate(); err != nil {
		return Review{}, err
	}
	if getter == nil {
		return Review{}, fmt.Errorf("review getter is required")
	}
	return getter.GetReview(ctx, options.PackageName, options.ReviewID, options.TranslationLanguage)
}

type ReviewReplyOptions struct {
	PackageName PackageName `json:"packageName"`
	ReviewID    ReviewID    `json:"reviewId"`
	Text        string      `json:"text"`
	Confirm     bool        `json:"confirm"`
	DryRun      bool        `json:"dryRun"`
}

func (o ReviewReplyOptions) Validate() error {
	if err := o.PackageName.Validate(); err != nil {
		return err
	}
	if _, err := NewReviewID(o.ReviewID.String()); err != nil {
		return err
	}
	if o.Text == "" {
		return fmt.Errorf("reply text is required")
	}
	if len([]rune(o.Text)) > maxReviewReplyCharacters {
		return fmt.Errorf("reply text cannot exceed %d characters", maxReviewReplyCharacters)
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("review reply requires --confirm or --dry-run")
	}
	return nil
}

type ReviewReplyPlan struct {
	PackageName PackageName `json:"packageName"`
	ReviewID    ReviewID    `json:"reviewId"`
	Confirm     bool        `json:"confirm"`
	Steps       []string    `json:"steps"`
}

type ReviewReplyResult struct {
	PackageName PackageName     `json:"packageName"`
	ReviewID    ReviewID        `json:"reviewId"`
	Text        string          `json:"text"`
	DryRun      bool            `json:"dryRun"`
	Applied     bool            `json:"applied"`
	Reply       *DeveloperReply `json:"reply,omitempty"`
	Plan        ReviewReplyPlan `json:"plan"`
}

type DeveloperReply struct {
	Text       string     `json:"text"`
	LastEdited *Timestamp `json:"lastEdited,omitempty"`
}

func NewReviewReplyPlan(options ReviewReplyOptions) (ReviewReplyPlan, error) {
	if err := options.Validate(); err != nil {
		return ReviewReplyPlan{}, err
	}
	steps := []string{"reply to review"}
	if options.DryRun {
		steps = []string{"plan review reply"}
	}
	return ReviewReplyPlan{
		PackageName: options.PackageName,
		ReviewID:    options.ReviewID,
		Confirm:     options.Confirm,
		Steps:       steps,
	}, nil
}

func ReplyToReview(ctx context.Context, replier ReviewReplier, options ReviewReplyOptions) (ReviewReplyResult, error) {
	plan, err := NewReviewReplyPlan(options)
	if err != nil {
		return ReviewReplyResult{}, err
	}
	result := ReviewReplyResult{
		PackageName: options.PackageName,
		ReviewID:    options.ReviewID,
		Text:        options.Text,
		DryRun:      options.DryRun,
		Applied:     false,
		Plan:        plan,
	}
	if options.DryRun {
		return result, nil
	}
	if replier == nil {
		return ReviewReplyResult{}, fmt.Errorf("review replier is required")
	}
	reply, err := replier.ReplyToReview(ctx, options.PackageName, options.ReviewID, options.Text)
	if err != nil {
		return ReviewReplyResult{}, err
	}
	result.Applied = true
	result.Reply = &reply
	return result, nil
}
