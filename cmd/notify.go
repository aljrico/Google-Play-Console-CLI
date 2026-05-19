package cmd

import (
	"io"

	"github.com/aljrico/Google-Play-Console-CLI/internal/notify"
	"github.com/aljrico/Google-Play-Console-CLI/internal/output"
	"github.com/spf13/cobra"
)

func newNotifyCommand(out io.Writer, options *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Send release workflow notifications",
	}
	cmd.AddCommand(newNotifySendCommand(out, options))
	return cmd
}

func newNotifySendCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := notify.SendOptions{WebhookURLEnv: notify.DefaultWebhookURLEnv}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a JSON notification webhook",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.Send(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	cmd.Flags().StringVar(&sendOptions.WebhookURL, "webhook-url", "", "HTTPS webhook URL; http is allowed only for loopback hosts")
	cmd.Flags().StringVar(&sendOptions.WebhookURLEnv, "webhook-url-env", notify.DefaultWebhookURLEnv, "Environment variable containing the webhook URL")
	cmd.Flags().StringVar(&sendOptions.WebhookURLFile, "webhook-url-file", "", "File containing the webhook URL")
	cmd.Flags().StringVar(&sendOptions.Title, "title", "", "Notification title")
	cmd.Flags().StringVar(&sendOptions.Message, "message", "", "Notification message")
	cmd.Flags().StringVar(&sendOptions.Severity, "severity", "", "Notification severity label")
	cmd.Flags().StringArrayVar(&sendOptions.Fields, "field", nil, "Notification field as name=value; repeatable")
	cmd.Flags().BoolVar(&sendOptions.Confirm, "confirm", false, "Send the notification webhook")
	cmd.Flags().BoolVar(&sendOptions.DryRun, "dry-run", false, "Print the notification payload without sending")
	return cmd
}
