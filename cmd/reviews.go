package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newReviewsCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var packageName string

	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "Read and reply to Google Play reviews",
	}
	cmd.PersistentFlags().StringVar(&packageName, "package", "", "Android package name, for example com.example.app")
	cmd.AddCommand(
		newReviewsListCommand(out, options, &packageName),
		newReviewsGetCommand(out, options, &packageName),
		newReviewsReplyCommand(out, options, &packageName),
	)
	return cmd
}

func newReviewsListCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		maxResults          int64
		startIndex          int64
		token               string
		translationLanguage string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Google Play reviews",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListReviews(cmd.Context(), publisher, play.ReviewListOptions{
				PackageName:         typedPackageName,
				MaxResults:          maxResults,
				StartIndex:          startIndex,
				Token:               token,
				TranslationLanguage: translationLanguage,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().Int64Var(&maxResults, "max-results", 0, "Maximum reviews to return")
	cmd.Flags().Int64Var(&startIndex, "start-index", 0, "Zero-based review offset")
	cmd.Flags().StringVar(&token, "token", "", "Pagination token from a previous response")
	cmd.Flags().StringVar(&translationLanguage, "translation-language", "", "Language localization code for translated reviews")
	return cmd
}

func newReviewsGetCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		reviewID            string
		translationLanguage string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get one Google Play review",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedReviewID, err := play.NewReviewID(reviewID)
			if err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			review, err := play.GetReview(cmd.Context(), publisher, play.ReviewGetOptions{
				PackageName:         typedPackageName,
				ReviewID:            typedReviewID,
				TranslationLanguage: translationLanguage,
			})
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, review)
		},
	}
	cmd.Flags().StringVar(&reviewID, "review-id", "", "Google Play review ID")
	cmd.Flags().StringVar(&translationLanguage, "translation-language", "", "Language localization code for translated review text")
	return cmd
}

func newReviewsReplyCommand(out io.Writer, options *globalOptions, packageName *string) *cobra.Command {
	var (
		reviewID string
		text     string
		confirm  bool
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to a Google Play review",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedPackageName, err := play.NewPackageName(*packageName)
			if err != nil {
				return err
			}
			typedReviewID, err := play.NewReviewID(reviewID)
			if err != nil {
				return err
			}
			replyOptions := play.ReviewReplyOptions{
				PackageName: typedPackageName,
				ReviewID:    typedReviewID,
				Text:        text,
				Confirm:     confirm,
				DryRun:      dryRun,
			}
			if dryRun {
				result, err := play.ReplyToReview(cmd.Context(), nil, replyOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			if _, err := play.NewReviewReplyPlan(replyOptions); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ReplyToReview(cmd.Context(), publisher, replyOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&reviewID, "review-id", "", "Google Play review ID")
	cmd.Flags().StringVar(&text, "text", "", "Public developer reply text")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the public review reply")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned review reply without calling Google Play")
	return cmd
}
