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
	cmd.AddCommand(newNotifySendCommand(out, options), newNotifySlackCommand(out, options))
	return cmd
}

func newNotifySendCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify send")
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
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the webhook URL",
		WebhookURLFile: "File containing the webhook URL",
		Confirm:        "Send the notification webhook",
		DryRun:         "Print the notification payload without sending",
	})
	return cmd
}

func newNotifySlackCommand(out io.Writer, options *globalOptions) *cobra.Command {
	sendOptions := newNotifySendOptions("notify slack")
	cmd := &cobra.Command{
		Use:   "slack",
		Short: "Send a Slack incoming webhook notification",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := notify.SendSlack(cmd.Context(), nil, sendOptions)
			if result.Webhook == "" {
				return err
			}
			if writeErr := output.Write(out, options.output, options.pretty, result); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	addNotifyFlags(cmd, &sendOptions, notifyFlagHelp{
		WebhookURL:     "HTTPS Slack incoming webhook URL; http is allowed only for loopback hosts",
		WebhookURLEnv:  "Environment variable containing the Slack incoming webhook URL",
		WebhookURLFile: "File containing the Slack incoming webhook URL",
		Confirm:        "Send the Slack webhook",
		DryRun:         "Print the Slack payload without sending",
	})
	return cmd
}

type notifyFlagHelp struct {
	WebhookURL     string
	WebhookURLEnv  string
	WebhookURLFile string
	Confirm        string
	DryRun         string
}

func newNotifySendOptions(commandPath string) notify.SendOptions {
	return notify.SendOptions{
		CommandPath:   commandPath,
		WebhookURLEnv: notify.DefaultWebhookURLEnv,
	}
}

func addNotifyFlags(cmd *cobra.Command, sendOptions *notify.SendOptions, help notifyFlagHelp) {
	cmd.Flags().StringVar(&sendOptions.WebhookURL, "webhook-url", "", help.WebhookURL)
	cmd.Flags().StringVar(&sendOptions.WebhookURLEnv, "webhook-url-env", notify.DefaultWebhookURLEnv, help.WebhookURLEnv)
	cmd.Flags().StringVar(&sendOptions.WebhookURLFile, "webhook-url-file", "", help.WebhookURLFile)
	cmd.Flags().StringVar(&sendOptions.Title, "title", "", "Notification title")
	cmd.Flags().StringVar(&sendOptions.Message, "message", "", "Notification message")
	cmd.Flags().StringVar(&sendOptions.Severity, "severity", "", "Notification severity label")
	cmd.Flags().StringArrayVar(&sendOptions.Fields, "field", nil, "Notification field as name=value; repeatable")
	cmd.Flags().BoolVar(&sendOptions.Confirm, "confirm", false, help.Confirm)
	cmd.Flags().BoolVar(&sendOptions.DryRun, "dry-run", false, help.DryRun)
}
