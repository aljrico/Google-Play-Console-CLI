package play

import (
	"context"
	"fmt"

	"google.golang.org/api/androidpublisher/v3"
)

func (p *GooglePublisher) ListReviews(ctx context.Context, options ReviewListOptions) (ReviewListResult, error) {
	call := p.service.Reviews.List(options.PackageName.String()).Context(ctx)
	if options.MaxResults > 0 {
		call.MaxResults(options.MaxResults)
	}
	if options.StartIndex > 0 {
		call.StartIndex(options.StartIndex)
	}
	if options.Token != "" {
		call.Token(options.Token)
	}
	if options.TranslationLanguage != "" {
		call.TranslationLanguage(options.TranslationLanguage)
	}
	response, err := call.Do()
	if err != nil {
		return ReviewListResult{}, fmt.Errorf("list reviews for %s: %w", options.PackageName, err)
	}
	return reviewListResultFromAPI(options, response), nil
}

func (p *GooglePublisher) GetReview(ctx context.Context, packageName PackageName, reviewID ReviewID, translationLanguage string) (Review, error) {
	call := p.service.Reviews.Get(packageName.String(), reviewID.String()).Context(ctx)
	if translationLanguage != "" {
		call.TranslationLanguage(translationLanguage)
	}
	review, err := call.Do()
	if err != nil {
		return Review{}, fmt.Errorf("get review %s for %s: %w", reviewID, packageName, err)
	}
	return reviewFromAPI(review), nil
}

func (p *GooglePublisher) ReplyToReview(ctx context.Context, packageName PackageName, reviewID ReviewID, text string) (DeveloperReply, error) {
	request := &androidpublisher.ReviewsReplyRequest{
		ReplyText:       text,
		ForceSendFields: []string{"ReplyText"},
	}
	response, err := p.service.Reviews.Reply(packageName.String(), reviewID.String(), request).Context(ctx).Do()
	if err != nil {
		return DeveloperReply{}, fmt.Errorf("reply to review %s for %s: %w", reviewID, packageName, err)
	}
	if response == nil || response.Result == nil {
		return DeveloperReply{}, nil
	}
	return DeveloperReply{
		Text:       response.Result.ReplyText,
		LastEdited: timestampFromAPI(response.Result.LastEdited),
	}, nil
}

func reviewListResultFromAPI(options ReviewListOptions, response *androidpublisher.ReviewsListResponse) ReviewListResult {
	result := ReviewListResult{
		PackageName: options.PackageName,
		Reviews:     []Review{},
		Options:     options,
	}
	if response == nil {
		return result
	}
	for _, apiReview := range response.Reviews {
		result.Reviews = append(result.Reviews, reviewFromAPI(apiReview))
	}
	if response.PageInfo != nil {
		result.PageInfo = &ReviewPageInfo{
			ResultPerPage: response.PageInfo.ResultPerPage,
			StartIndex:    response.PageInfo.StartIndex,
			TotalResults:  response.PageInfo.TotalResults,
		}
	}
	if response.TokenPagination != nil {
		result.Pagination = &ReviewPagination{
			NextPageToken:     response.TokenPagination.NextPageToken,
			PreviousPageToken: response.TokenPagination.PreviousPageToken,
		}
	}
	return result
}

func reviewFromAPI(apiReview *androidpublisher.Review) Review {
	if apiReview == nil {
		return Review{Comments: []ReviewComment{}}
	}
	comments := make([]ReviewComment, 0, len(apiReview.Comments))
	for _, apiComment := range apiReview.Comments {
		comments = append(comments, reviewCommentFromAPI(apiComment))
	}
	return Review{
		ReviewID:   ReviewID(apiReview.ReviewId),
		AuthorName: apiReview.AuthorName,
		Comments:   comments,
	}
}

func reviewCommentFromAPI(apiComment *androidpublisher.Comment) ReviewComment {
	if apiComment == nil {
		return ReviewComment{}
	}
	if apiComment.UserComment != nil {
		userComment := apiComment.UserComment
		return ReviewComment{
			Kind:      ReviewCommentKindUser,
			Text:      userComment.Text,
			UpdatedAt: timestampFromAPI(userComment.LastModified),
			User: &UserReviewComment{
				ReviewerLanguage: userComment.ReviewerLanguage,
				StarRating:       userComment.StarRating,
				AppVersionCode:   userComment.AppVersionCode,
				AppVersionName:   userComment.AppVersionName,
				AndroidOSVersion: userComment.AndroidOsVersion,
				Device:           userComment.Device,
				OriginalText:     userComment.OriginalText,
				ThumbsUpCount:    userComment.ThumbsUpCount,
				ThumbsDownCount:  userComment.ThumbsDownCount,
				DeviceMetadata:   deviceMetadataFromAPI(userComment.DeviceMetadata),
			},
		}
	}
	if apiComment.DeveloperComment != nil {
		developerComment := apiComment.DeveloperComment
		return ReviewComment{
			Kind:      ReviewCommentKindDeveloper,
			Text:      developerComment.Text,
			UpdatedAt: timestampFromAPI(developerComment.LastModified),
			Developer: &DeveloperReviewComment{
				LastEdited: timestampFromAPI(developerComment.LastModified),
			},
		}
	}
	return ReviewComment{}
}

func deviceMetadataFromAPI(apiMetadata *androidpublisher.DeviceMetadata) *DeviceMetadata {
	if apiMetadata == nil {
		return nil
	}
	return &DeviceMetadata{
		CPUMake:            apiMetadata.CpuMake,
		CPUModel:           apiMetadata.CpuModel,
		DeviceClass:        apiMetadata.DeviceClass,
		GLESVersion:        apiMetadata.GlEsVersion,
		Manufacturer:       apiMetadata.Manufacturer,
		NativePlatform:     apiMetadata.NativePlatform,
		ProductName:        apiMetadata.ProductName,
		RAMMegabytes:       apiMetadata.RamMb,
		ScreenDensityDPI:   apiMetadata.ScreenDensityDpi,
		ScreenHeightPixels: apiMetadata.ScreenHeightPx,
		ScreenWidthPixels:  apiMetadata.ScreenWidthPx,
	}
}

func timestampFromAPI(apiTimestamp *androidpublisher.Timestamp) *Timestamp {
	if apiTimestamp == nil {
		return nil
	}
	return &Timestamp{Seconds: apiTimestamp.Seconds, Nanos: apiTimestamp.Nanos}
}
