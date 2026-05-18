package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/aljrico/Google-Play-Console-CLI/internal/play"
	"github.com/spf13/cobra"
)

func newUsersCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Inspect Google Play Console users",
	}
	cmd.AddCommand(newUsersListCommand(out, options))
	return cmd
}

func newUsersListCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		developer string
		pageSize  int64
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users with access to a developer account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typedDeveloper, err := play.NewDeveloperAccount(developer)
			if err != nil {
				return err
			}
			listOptions := play.UserListOptions{
				Developer: typedDeveloper,
				PageSize:  pageSize,
				PageToken: pageToken,
			}
			if err := listOptions.Validate(); err != nil {
				return err
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.ListUsers(cmd.Context(), publisher, listOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&developer, "developer", "", "Developer account ID or resource, for example 1234567890 or developers/1234567890")
	cmd.Flags().Int64Var(&pageSize, "page-size", 0, "Maximum users to return; use -1 to disable pagination")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token from a previous response")
	return cmd
}
