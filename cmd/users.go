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
		Short: "Inspect and manage Google Play Console users",
	}
	cmd.AddCommand(
		newUsersListCommand(out, options),
		newUsersDeleteCommand(out, options),
	)
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

func newUsersDeleteCommand(out io.Writer, options *globalOptions) *cobra.Command {
	var (
		developer string
		userEmail string
		name      string
		confirm   bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove all developer-account access for a user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			userName, err := parseUserName(name, developer, userEmail)
			if err != nil {
				return err
			}
			deleteOptions := play.UserDeleteOptions{
				Name:    userName,
				Confirm: confirm,
				DryRun:  dryRun,
			}
			if err := deleteOptions.Validate(); err != nil {
				return err
			}
			if dryRun {
				result, err := play.DeleteUser(cmd.Context(), nil, deleteOptions)
				if err != nil {
					return err
				}
				return output.Write(out, options.output, options.pretty, result)
			}
			publisher, err := play.NewPublisherFromActiveProfile(cmd.Context())
			if err != nil {
				return err
			}
			result, err := play.DeleteUser(cmd.Context(), publisher, deleteOptions)
			if err != nil {
				return err
			}
			return output.Write(out, options.output, options.pretty, result)
		},
	}
	cmd.Flags().StringVar(&developer, "developer", "", "Developer account ID or resource, for example 1234567890 or developers/1234567890")
	cmd.Flags().StringVar(&userEmail, "user-email", "", "Play Console user email")
	cmd.Flags().StringVar(&name, "name", "", "User resource name, developers/{developer}/users/{email}")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Apply the user deletion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned user deletion without calling Google Play")
	return cmd
}

func parseUserName(name string, developer string, userEmail string) (play.UserName, error) {
	if name != "" {
		return play.NewUserName(name)
	}
	typedDeveloper, err := play.NewDeveloperAccount(developer)
	if err != nil {
		return "", err
	}
	typedUserEmail, err := play.NewGrantUserEmail(userEmail)
	if err != nil {
		return "", err
	}
	return play.NewUserNameFromParts(typedDeveloper, typedUserEmail), nil
}
